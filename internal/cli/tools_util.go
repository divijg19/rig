// internal/cli/tools_util.go

package cli

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	core "github.com/divijg19/rig/internal/rig"
)

// manifestLockPath returns the path to the tools lockfile relative to rig.toml
func manifestLockPath(configPath string) string {
	binDir := localBinDirFor(configPath)
	return filepath.Join(filepath.Dir(binDir), "manifest.lock")
}

// computeToolsHash creates a hash of the tools configuration for lock file
func computeToolsHash(tools map[string]string) string {
	h := sha256.New()

	// Sort keys for consistent hashing
	var keys []string
	for k := range tools {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		h.Write([]byte(k + "=" + tools[k] + "\n"))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// checkToolsSyncFast performs a quick check if tools are in sync
func checkToolsSyncFast(tools map[string]string, configPath string) error {
	if len(tools) == 0 {
		return nil
	}

	lockFile := manifestLockPath(configPath)
	currentHash := computeToolsHash(tools)

	// Check if lock file exists and matches
	if lockData, err := os.ReadFile(lockFile); err == nil {
		if strings.TrimSpace(string(lockData)) == currentHash {
			return nil // Tools are in sync
		}
	}

	return fmt.Errorf("❌ Tools are out of sync with rig.toml. Run 'rig tools sync' (lock: %s)", lockFile)
}

// parseToolsFiles reads one or more .txt files and returns a map of tool -> version
// Accepts lines in forms:
//
//	name = version
//	module@version
//	name            (assumes latest)
func parseToolsFiles(paths []string) (map[string]string, error) {
	out := map[string]string{}
	for _, arg := range paths {
		if !strings.HasSuffix(arg, ".txt") {
			continue
		}
		f, err := os.Open(arg)
		if err != nil {
			return nil, fmt.Errorf("read tools file %s: %w", arg, err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if i := strings.Index(line, "="); i > 0 {
				name := strings.TrimSpace(line[:i])
				ver := strings.TrimSpace(line[i+1:])
				out[name] = ver
			} else if i := strings.Index(line, "@"); i > 0 {
				name := strings.TrimSpace(line[:i])
				ver := strings.TrimSpace(line[i+1:])
				out[name] = ver
			} else {
				out[line] = "latest"
			}
		}
		if err := scanner.Err(); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("scan tools file %s: %w", arg, err)
		}
		_ = f.Close()
	}
	return out, nil
}

// mergeTools overlays b onto a (b wins on conflicts) and returns a new map
func mergeTools(a, b map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// ToolStatusRow represents the status of a single tool for reporting.
type ToolStatusRow struct {
	Name   string `json:"name"`
	Bin    string `json:"bin"`
	Want   string `json:"want"`
	Have   string `json:"have"`
	Status string `json:"status"` // ok|missing|mismatch
}

// collectToolStatus concurrently checks installed tool versions against the manifest.
// It returns deterministic rows ordered by tool name, along with counts of missing and mismatched tools.
func collectToolStatus(tools map[string]string, configPath string) ([]ToolStatusRow, int, int) {
	if len(tools) == 0 {
		return nil, 0, 0
	}
	env := envWithLocalBin(configPath, nil, false)
	// Stable order of names
	names := make([]string, 0, len(tools))
	for n := range tools {
		names = append(names, n)
	}
	sort.Strings(names)

	type rowRes struct {
		idx int
		row ToolStatusRow
	}
	outCh := make(chan rowRes, len(names))
	sem := make(chan struct{}, max(1, runtime.NumCPU()))
	var wg sync.WaitGroup

	for i, name := range names {
		i, name := i, name
		ver := tools[name]
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			normalized := core.EnsureSemverPrefixV(ver)
			_, bin := core.ResolveModuleAndBin(name)
			out, _ := execCommandEnv(bin, []string{"--version"}, env)
			have := core.ParseVersionFromOutput(out)
			want := core.NormalizeSemver(normalized)
			status := "ok"
			if have == "" {
				status = "missing"
			} else if want != "latest" && have != want {
				status = "mismatch"
			}
			outCh <- rowRes{idx: i, row: ToolStatusRow{Name: name, Bin: bin, Want: want, Have: have, Status: status}}
		}()
	}
	wg.Wait()
	close(outCh)
	rows := make([]ToolStatusRow, len(names))
	var missing, mismatched int
	for rr := range outCh {
		rows[rr.idx] = rr.row
	}
	for _, r := range rows {
		switch r.Status {
		case "missing":
			missing++
		case "mismatch":
			mismatched++
		}
	}
	return rows, missing, mismatched
}
