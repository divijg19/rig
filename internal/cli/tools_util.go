// internal/cli/tools_util.go

package cli

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

// collectToolStatus checks installed tool integrity against rig.lock.
// It returns deterministic rows ordered by tool name, along with counts of missing and mismatched tools.
func collectToolStatus(tools map[string]string, configPath string) ([]core.ToolStatusRow, int, int) {
	lockPath := rigLockPathFor(configPath)
	lock, err := core.ReadLockfile(lockPath)
	if err != nil {
		// If rig.lock is missing/unreadable, treat all tools as mismatched for the purpose of outdated.
		// The caller prints the error context.
		tools = stripGoToolchain(tools)
		if len(tools) == 0 {
			return nil, 0, 0
		}
		names := make([]string, 0, len(tools))
		for n := range tools {
			names = append(names, n)
		}
		sort.Strings(names)
		rows := make([]core.ToolStatusRow, 0, len(names))
		for _, name := range names {
			_, bin := core.ResolveModuleAndBin(name)
			rows = append(rows, core.ToolStatusRow{Name: name, Bin: bin, Want: "", Have: "", Status: "mismatch"})
		}
		return rows, 0, len(rows)
	}

	rows, missing, mismatched, _, cerr := core.CheckInstalledTools(tools, lock, configPath)
	if cerr != nil {
		// Treat schema/consistency errors as a full mismatch.
		tools = stripGoToolchain(tools)
		if len(tools) == 0 {
			return nil, 0, 0
		}
		names := make([]string, 0, len(tools))
		for n := range tools {
			names = append(names, n)
		}
		sort.Strings(names)
		fallback := make([]core.ToolStatusRow, 0, len(names))
		for _, name := range names {
			_, bin := core.ResolveModuleAndBin(name)
			fallback = append(fallback, core.ToolStatusRow{Name: name, Bin: bin, Want: "", Have: "", Status: "mismatch"})
		}
		return fallback, 0, len(fallback)
	}
	return rows, missing, mismatched
}
