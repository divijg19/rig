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

func splitResolved(resolved string) (module string, version string) {
	resolved = strings.TrimSpace(resolved)
	i := strings.LastIndex(resolved, "@")
	if i <= 0 || i == len(resolved)-1 {
		return resolved, ""
	}
	return strings.TrimSpace(resolved[:i]), strings.TrimSpace(resolved[i+1:])
}

func parseRequested(requested string) (name string, version string, err error) {
	requested = strings.TrimSpace(requested)
	i := strings.LastIndex(requested, "@")
	if i <= 0 || i == len(requested)-1 {
		return "", "", fmt.Errorf("invalid requested %q (expected name@version)", requested)
	}
	name = strings.TrimSpace(requested[:i])
	version = strings.TrimSpace(requested[i+1:])
	if name == "" || version == "" {
		return "", "", fmt.Errorf("invalid requested %q (empty name or version)", requested)
	}
	return name, version, nil
}

func normalizeToolVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if v == "latest" {
		return "latest"
	}
	return core.NormalizeSemver(core.EnsureSemverPrefixV(v))
}

func lockMatchesTools(lock core.Lockfile, tools map[string]string) error {
	if len(tools) == 0 {
		if len(lock.Tools) == 0 {
			return nil
		}
		return fmt.Errorf("rig.lock has %d tool(s) but rig.toml has none", len(lock.Tools))
	}

	byName := make(map[string]core.LockedTool, len(lock.Tools))
	for _, lt := range lock.Tools {
		name, reqVer, err := parseRequested(lt.Requested)
		if err != nil {
			return err
		}
		if _, ok := byName[name]; ok {
			return fmt.Errorf("rig.lock has duplicate tool %q", name)
		}
		// For Go tools we expect module to be present.
		if strings.TrimSpace(lt.Kind) == "go" && strings.TrimSpace(lt.Module) == "" {
			return fmt.Errorf("rig.lock tool %q: missing module field", name)
		}
		// Normalize requested version for stable comparison.
		_ = reqVer
		byName[name] = lt
	}

	for name, wantVer := range tools {
		lt, ok := byName[name]
		if !ok {
			return fmt.Errorf("rig.lock missing tool %q", name)
		}
		_, lockReqVer, err := parseRequested(lt.Requested)
		if err != nil {
			return err
		}
		if normalizeToolVersion(lockReqVer) != normalizeToolVersion(wantVer) {
			return fmt.Errorf("tool %q version mismatch: rig.toml wants %q, rig.lock has %q", name, wantVer, lockReqVer)
		}
		expModule, _ := core.ResolveModuleAndBin(name)
		if strings.TrimSpace(lt.Module) != strings.TrimSpace(expModule) {
			return fmt.Errorf("tool %q module mismatch: expected %q, rig.lock has %q", name, expModule, lt.Module)
		}
		resMod, _ := splitResolved(lt.Resolved)
		if strings.TrimSpace(resMod) != strings.TrimSpace(expModule) {
			return fmt.Errorf("tool %q resolved mismatch: expected %q@..., rig.lock has %q", name, expModule, lt.Resolved)
		}
	}

	for name := range byName {
		if _, ok := tools[name]; !ok {
			return fmt.Errorf("rig.lock has extra tool %q not present in rig.toml", name)
		}
	}
	return nil
}

// manifestLockPath returns the path to the tools lockfile relative to rig.toml
func manifestLockPath(configPath string) string {
	binDir := localBinDirFor(configPath)
	return filepath.Join(filepath.Dir(binDir), "manifest.lock")
}

func toolBinPath(configPath string, bin string) string {
	name := bin
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		name += ".exe"
	}
	return filepath.Join(localBinDirFor(configPath), name)
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
			binPath := toolBinPath(configPath, bin)
			out, _ := execCommandEnv(binPath, []string{"--version"}, env)
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

func collectToolStatusWithLock(tools map[string]string, lock core.Lockfile, configPath string) ([]ToolStatusRow, int, int) {
	if len(tools) == 0 {
		return nil, 0, 0
	}
	if err := lockMatchesTools(lock, tools); err != nil {
		// Callers should have checked already; avoid silently returning misleading status.
		return nil, 0, 0
	}

	byName := make(map[string]core.LockedTool, len(lock.Tools))
	for _, lt := range lock.Tools {
		name, _, err := parseRequested(lt.Requested)
		if err != nil {
			continue
		}
		byName[name] = lt
	}

	// Reuse the status collector shape but use resolved versions from rig.lock.
	desired := make(map[string]string, len(tools))
	for name, v := range tools {
		if lt, ok := byName[name]; ok {
			_, resolvedVer := splitResolved(lt.Resolved)
			if strings.TrimSpace(resolvedVer) == "" {
				desired[name] = v
				continue
			}
			desired[name] = resolvedVer
			continue
		}
		desired[name] = v
	}

	// We need a version-aware collector; implement a local variant.
	env := envWithLocalBin(configPath, nil, false)
	names := make([]string, 0, len(desired))
	for n := range desired {
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
		ver := desired[name]
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			want := normalizeToolVersion(ver)
			_, bin := core.ResolveModuleAndBin(name)
			binPath := toolBinPath(configPath, bin)
			out, _ := execCommandEnv(binPath, []string{"--version"}, env)
			have := core.ParseVersionFromOutput(out)
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
