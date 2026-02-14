package rig

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const defaultLatestReleaseURL = "https://api.github.com/repos/divijg19/rig/releases/latest"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type UpgradeOptions struct {
	CurrentVersion string
	ExecutablePath string
	GOOS           string
	GOARCH         string
	LatestURL      string
	Client         HTTPClient
}

type UpgradeResult struct {
	UpToDate      bool
	Current       string
	Latest        string
	AssetName     string
	ChecksumName  string
	ExecutableOut string
}

type githubLatestRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func UpgradeSelf(opts UpgradeOptions) (UpgradeResult, error) {
	if strings.TrimSpace(opts.ExecutablePath) == "" {
		return UpgradeResult{}, errors.New("executable path is required")
	}
	if !isFileReplaceWritable(opts.ExecutablePath) {
		return UpgradeResult{}, fmt.Errorf("binary path not writable: %s", opts.ExecutablePath)
	}
	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}
	if strings.TrimSpace(opts.LatestURL) == "" {
		opts.LatestURL = defaultLatestReleaseURL
	}
	if strings.TrimSpace(opts.GOOS) == "" {
		opts.GOOS = runtime.GOOS
	}
	if strings.TrimSpace(opts.GOARCH) == "" {
		opts.GOARCH = runtime.GOARCH
	}

	rel, err := fetchLatestRelease(opts.Client, opts.LatestURL)
	if err != nil {
		return UpgradeResult{}, err
	}
	res := UpgradeResult{Current: strings.TrimSpace(opts.CurrentVersion), Latest: strings.TrimSpace(rel.TagName)}
	if res.Current != "" && res.Current == res.Latest {
		res.UpToDate = true
		return res, nil
	}

	assetName, checksumName, err := expectedAssetNames(opts.GOOS, opts.GOARCH)
	if err != nil {
		return UpgradeResult{}, err
	}
	res.AssetName = assetName
	res.ChecksumName = checksumName

	assetURL, ok := findAssetURL(rel, assetName)
	if !ok {
		return UpgradeResult{}, fmt.Errorf("release asset not found: %s", assetName)
	}
	checksumURL, ok := findAssetURL(rel, checksumName)
	if !ok {
		return UpgradeResult{}, fmt.Errorf("release checksum not found: %s", checksumName)
	}

	assetData, err := fetchBytes(opts.Client, assetURL)
	if err != nil {
		return UpgradeResult{}, err
	}
	checksumData, err := fetchBytes(opts.Client, checksumURL)
	if err != nil {
		return UpgradeResult{}, err
	}
	if err := verifyChecksum(assetName, assetData, checksumData); err != nil {
		return UpgradeResult{}, err
	}

	binaryName := "rig"
	if opts.GOOS == "windows" {
		binaryName = "rig.exe"
	}
	binaryData, err := extractSingleBinary(assetName, assetData, binaryName)
	if err != nil {
		return UpgradeResult{}, err
	}

	if err := replaceExecutableAtomically(opts.ExecutablePath, binaryData); err != nil {
		if opts.GOOS == "windows" {
			return UpgradeResult{}, fmt.Errorf("upgrade failed to replace running binary; close all rig processes and retry: %w", err)
		}
		return UpgradeResult{}, err
	}

	res.ExecutableOut = opts.ExecutablePath
	return res, nil
}

func fetchLatestRelease(client HTTPClient, url string) (githubLatestRelease, error) {
	body, err := fetchBytes(client, url)
	if err != nil {
		return githubLatestRelease{}, err
	}
	var rel githubLatestRelease
	if err := json.Unmarshal(body, &rel); err != nil {
		return githubLatestRelease{}, fmt.Errorf("parse latest release: %w", err)
	}
	if strings.TrimSpace(rel.TagName) == "" {
		return githubLatestRelease{}, errors.New("latest release missing tag_name")
	}
	return rel, nil
}

func fetchBytes(client HTTPClient, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("request failed (%d) for %s", resp.StatusCode, url)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func expectedAssetNames(goos, goarch string) (asset string, checksum string, err error) {
	switch goos {
	case "linux", "darwin":
		asset = fmt.Sprintf("rig_%s_%s.tar.gz", goos, goarch)
	case "windows":
		asset = fmt.Sprintf("rig_windows_%s.zip", goarch)
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", goos)
	}
	return asset, asset + ".sha256", nil
}

func findAssetURL(rel githubLatestRelease, name string) (string, bool) {
	for _, a := range rel.Assets {
		if strings.TrimSpace(a.Name) == name {
			return strings.TrimSpace(a.BrowserDownloadURL), true
		}
	}
	return "", false
}

func verifyChecksum(assetName string, data []byte, checksumFile []byte) error {
	line := strings.TrimSpace(string(checksumFile))
	if line == "" {
		return errors.New("empty checksum file")
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return fmt.Errorf("invalid checksum format: %q", line)
	}
	expected := strings.TrimSpace(fields[0])
	file := strings.TrimSpace(fields[len(fields)-1])
	if file != assetName {
		return fmt.Errorf("checksum filename mismatch: got %q want %q", file, assetName)
	}
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func extractSingleBinary(assetName string, data []byte, wantName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".tar.gz") {
		g, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer g.Close()
		tr := tar.NewReader(g)
		count := 0
		var out []byte
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if h.FileInfo().IsDir() {
				continue
			}
			count++
			if h.Name != wantName {
				return nil, fmt.Errorf("archive entry %q invalid (want exactly %q)", h.Name, wantName)
			}
			b, rerr := io.ReadAll(tr)
			if rerr != nil {
				return nil, rerr
			}
			out = b
		}
		if count != 1 {
			return nil, fmt.Errorf("archive must contain exactly one file (%q), got %d", wantName, count)
		}
		return out, nil
	}
	if strings.HasSuffix(assetName, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		files := 0
		var out []byte
		for _, f := range zr.File {
			if f.FileInfo().IsDir() {
				continue
			}
			files++
			if f.Name != wantName {
				return nil, fmt.Errorf("archive entry %q invalid (want exactly %q)", f.Name, wantName)
			}
			rc, rerr := f.Open()
			if rerr != nil {
				return nil, rerr
			}
			b, rerr := io.ReadAll(rc)
			_ = rc.Close()
			if rerr != nil {
				return nil, rerr
			}
			out = b
		}
		if files != 1 {
			return nil, fmt.Errorf("archive must contain exactly one file (%q), got %d", wantName, files)
		}
		return out, nil
	}
	return nil, fmt.Errorf("unsupported asset format: %s", assetName)
}

func replaceExecutableAtomically(path string, data []byte) error {
	if len(data) == 0 {
		return errors.New("empty binary data")
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "rig-upgrade-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		// On Windows, os.Rename does not reliably replace an existing destination.
		// Best-effort two-step replacement: move old aside, move new into place.
		backup := path + ".old"
		_ = os.Remove(backup)
		if err := os.Rename(path, backup); err != nil {
			return err
		}
		if err := os.Rename(tmpName, path); err != nil {
			_ = os.Rename(backup, path)
			return err
		}
		_ = os.Remove(backup)
		return nil
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return nil
}
