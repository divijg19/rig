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

// rigLockPathFor returns the path to rig.lock next to rig.toml.
func rigLockPathFor(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "rig.lock")
}

func toolsOfflineEnv(offline bool) []string {
	if !offline {
		return nil
	}
	// Prevent downloading modules; commands will fail if deps are not in module cache.
	return []string{"GOPROXY=off", "GOSUMDB=off"}
}

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

func stripGoToolchain(tools map[string]string) map[string]string {
	if len(tools) == 0 {
		return tools
	}
	if _, ok := tools["go"]; !ok {
		return tools
	}
	out := make(map[string]string, len(tools)-1)
	for k, v := range tools {
		if k == "go" {
			continue
		}
		out[k] = v
	}
	return out
}

// collectToolStatus concurrently checks installed tool versions against the manifest.
// It returns deterministic rows ordered by tool name, along with counts of missing and mismatched tools.
func collectToolStatus(tools map[string]string, configPath string) ([]core.ToolStatusRow, int, int) {
	tools = stripGoToolchain(tools)
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
		row core.ToolStatusRow
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
			want := core.NormalizeToolVersion(ver)
			_, bin := core.ResolveModuleAndBin(name)
			binPath := core.ToolBinPath(configPath, bin)
			out, _ := execCommandEnv(binPath, []string{"--version"}, env)
			have := core.ParseVersionFromOutput(out)
			status := "ok"
			if have == "" {
				status = "missing"
			} else if want != "latest" && have != want {
				status = "mismatch"
			}
			outCh <- rowRes{idx: i, row: core.ToolStatusRow{Name: name, Bin: bin, Want: want, Have: have, Status: status}}
		}()
	}
	wg.Wait()
	close(outCh)
	rows := make([]core.ToolStatusRow, len(names))
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
