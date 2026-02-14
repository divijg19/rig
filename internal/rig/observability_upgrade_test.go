package rig

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeTestFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func setupToolsFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	rigToml := strings.Join([]string{
		"[project]",
		"name='tmp'",
		"version='0.0.0'",
		"",
		"[tools]",
		"mockery='v2.46.0'",
		"golangci-lint='1.62.0'",
		"",
		"[tasks]",
		"noop='echo ok'",
	}, "\n") + "\n"
	writeTestFile(t, filepath.Join(dir, "rig.toml"), rigToml, 0o644)

	binDir := filepath.Join(dir, ".rig", "bin")
	writeTestFile(t, filepath.Join(binDir, "mockery"), "#!/bin/sh\necho mockery\n", 0o755)
	writeTestFile(t, filepath.Join(binDir, "golangci-lint"), "#!/bin/sh\necho golangci-lint\n", 0o755)

	mockerySHA, err := ComputeFileSHA256(filepath.Join(binDir, "mockery"))
	if err != nil {
		t.Fatalf("sha mockery: %v", err)
	}
	golangciLintSHA, err := ComputeFileSHA256(filepath.Join(binDir, "golangci-lint"))
	if err != nil {
		t.Fatalf("sha golangci-lint: %v", err)
	}

	lock := fmt.Sprintf(`schema = 0

[[tools]]
kind = "go-binary"
requested = "golangci-lint@1.62.0"
resolved = "github.com/golangci/golangci-lint@v1.62.0"
module = "github.com/golangci/golangci-lint"
bin = "golangci-lint"
sha256 = %q

[[tools]]
kind = "go-binary"
requested = "mockery@v2.46.0"
resolved = "github.com/vektra/mockery/v2@v2.46.0"
module = "github.com/vektra/mockery/v2"
bin = "mockery"
sha256 = %q
`, golangciLintSHA, mockerySHA)
	writeTestFile(t, filepath.Join(dir, "rig.lock"), lock, 0o644)
	return dir
}

func TestToolsLSDeterministicOrder(t *testing.T) {
	dir := setupToolsFixture(t)
	items, err := ToolsLS(dir)
	if err != nil {
		t.Fatalf("ToolsLS: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "golangci-lint" || items[1].Name != "mockery" {
		t.Fatalf("unexpected order: %#v", items)
	}
}

func TestToolsPathMissing(t *testing.T) {
	dir := setupToolsFixture(t)
	if err := os.Remove(filepath.Join(dir, ".rig", "bin", "mockery")); err != nil {
		t.Fatalf("remove binary: %v", err)
	}
	_, err := ToolPath(dir, "mockery")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "not found in .rig/bin") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToolsWhyOutputShape(t *testing.T) {
	dir := setupToolsFixture(t)
	info, err := ToolWhy(dir, "mockery")
	if err != nil {
		t.Fatalf("ToolWhy: %v", err)
	}
	if info.Name == "" || info.Requested == "" || info.Resolved == "" || info.SHA256 == "" || info.Path == "" {
		t.Fatalf("incomplete why info: %#v", info)
	}
}

func TestToolsDoctorMissing(t *testing.T) {
	dir := setupToolsFixture(t)
	if err := os.Remove(filepath.Join(dir, ".rig", "bin", "mockery")); err != nil {
		t.Fatalf("remove binary: %v", err)
	}
	reports, err := ToolsDoctor(dir, "mockery")
	if err != nil {
		t.Fatalf("ToolsDoctor: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report")
	}
	if reports[0].Status != ToolMissing {
		t.Fatalf("expected missing status, got %s", reports[0].Status)
	}
}

func TestToolsDoctorShaMismatch(t *testing.T) {
	dir := setupToolsFixture(t)
	writeTestFile(t, filepath.Join(dir, ".rig", "bin", "mockery"), "#!/bin/sh\necho changed\n", 0o755)
	reports, err := ToolsDoctor(dir, "mockery")
	if err != nil {
		t.Fatalf("ToolsDoctor: %v", err)
	}
	if reports[0].Status != ToolMismatch {
		t.Fatalf("expected mismatch status, got %s", reports[0].Status)
	}
}

func TestDoctorNoLock(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "rig.toml"), "[project]\nname='x'\nversion='0.0.0'\n", 0o644)
	rep, err := Doctor(dir, "v0.4.0", filepath.Join(dir, "rig"))
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if rep.HasLock {
		t.Fatalf("expected no lock")
	}
}

func TestDoctorGoMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script go stub")
	}
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "rig.toml"), strings.Join([]string{
		"[project]",
		"name='x'",
		"version='0.0.0'",
		"",
		"[tools]",
		"go='1.23.4'",
	}, "\n")+"\n", 0o644)
	writeTestFile(t, filepath.Join(dir, "rig.lock"), strings.Join([]string{
		"schema = 0",
		"",
		"[toolchain.go]",
		"kind = \"go-toolchain\"",
		"requested = \"1.23.4\"",
		"detected = \"1.23.4\"",
	}, "\n")+"\n", 0o644)

	binDir := filepath.Join(dir, "fakebin")
	writeTestFile(t, filepath.Join(binDir, "go"), "#!/bin/sh\necho 'go version go1.23.5 linux/amd64'\n", 0o755)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	rep, err := Doctor(dir, "v0.4.0", filepath.Join(dir, "rig"))
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if rep.GoMatchesLock {
		t.Fatalf("expected go mismatch")
	}
}

func TestDoctorBinaryWritable(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "rig.toml"), "[project]\nname='x'\nversion='0.0.0'\n", 0o644)
	lock := "schema = 0\n"
	writeTestFile(t, filepath.Join(dir, "rig.lock"), lock, 0o644)
	exe := filepath.Join(dir, "rig")
	writeTestFile(t, exe, "bin", 0o755)
	rep, err := Doctor(dir, "v0.4.0", exe)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if !rep.ExecutableWritable {
		t.Fatalf("expected executable writable")
	}
}

func makeTarGzWithSingle(name string, content []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))})
	_, _ = tw.Write(content)
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func checksumLine(name string, data []byte) string {
	s := sha256.Sum256(data)
	return fmt.Sprintf("%s  %s\n", hex.EncodeToString(s[:]), name)
}

func TestUpgradeUpToDate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v0.4.0","assets":[]}`))
	}))
	defer ts.Close()

	exeDir := t.TempDir()
	exe := filepath.Join(exeDir, "rig")
	writeTestFile(t, exe, "old", 0o755)

	res, err := UpgradeSelf(UpgradeOptions{CurrentVersion: "v0.4.0", ExecutablePath: exe, LatestURL: ts.URL, GOOS: "linux", GOARCH: "amd64"})
	if err != nil {
		t.Fatalf("UpgradeSelf: %v", err)
	}
	if !res.UpToDate {
		t.Fatalf("expected up to date")
	}
}

func TestUpgradeChecksumMismatch(t *testing.T) {
	assetName := "rig_linux_amd64.tar.gz"
	asset := makeTarGzWithSingle("rig", []byte("newbin"))
	badChecksum := "deadbeef  " + assetName + "\n"

	var baseURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			_, _ = w.Write([]byte(`{"tag_name":"v0.5.0","assets":[{"name":"` + assetName + `","browser_download_url":"` + baseURL + `/asset"},{"name":"` + assetName + `.sha256","browser_download_url":"` + baseURL + `/sum"}]}`))
		case "/asset":
			_, _ = w.Write(asset)
		case "/sum":
			_, _ = w.Write([]byte(badChecksum))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	baseURL = ts.URL
	defer ts.Close()

	exeDir := t.TempDir()
	exe := filepath.Join(exeDir, "rig")
	writeTestFile(t, exe, "old", 0o755)

	_, err := UpgradeSelf(UpgradeOptions{CurrentVersion: "v0.4.0", ExecutablePath: exe, LatestURL: ts.URL + "/latest", GOOS: "linux", GOARCH: "amd64"})
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got: %v", err)
	}
}

func TestUpgradeMissingAsset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v0.5.0","assets":[]}`))
	}))
	defer ts.Close()

	exeDir := t.TempDir()
	exe := filepath.Join(exeDir, "rig")
	writeTestFile(t, exe, "old", 0o755)

	_, err := UpgradeSelf(UpgradeOptions{CurrentVersion: "v0.4.0", ExecutablePath: exe, LatestURL: ts.URL, GOOS: "linux", GOARCH: "amd64"})
	if err == nil || !strings.Contains(err.Error(), "asset not found") {
		t.Fatalf("expected asset missing error, got: %v", err)
	}
}

func TestUpgradeHappyPath(t *testing.T) {
	assetName := "rig_linux_amd64.tar.gz"
	asset := makeTarGzWithSingle("rig", []byte("newbin"))
	sum := checksumLine(assetName, asset)

	var baseURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			_, _ = w.Write([]byte(`{"tag_name":"v0.5.0","assets":[{"name":"` + assetName + `","browser_download_url":"` + baseURL + `/asset"},{"name":"` + assetName + `.sha256","browser_download_url":"` + baseURL + `/sum"}]}`))
		case "/asset":
			_, _ = w.Write(asset)
		case "/sum":
			_, _ = w.Write([]byte(sum))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	baseURL = ts.URL
	defer ts.Close()

	exeDir := t.TempDir()
	exe := filepath.Join(exeDir, "rig")
	writeTestFile(t, exe, "old", 0o755)

	res, err := UpgradeSelf(UpgradeOptions{CurrentVersion: "v0.4.0", ExecutablePath: exe, LatestURL: ts.URL + "/latest", GOOS: "linux", GOARCH: "amd64"})
	if err != nil {
		t.Fatalf("UpgradeSelf: %v", err)
	}
	if res.UpToDate {
		t.Fatalf("expected upgrade performed")
	}
	b, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("read exe: %v", err)
	}
	if string(b) != "newbin" {
		t.Fatalf("unexpected upgraded content: %q", string(b))
	}
}
