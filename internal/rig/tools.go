package rig

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type ToolState string

const (
	ToolOK       ToolState = "ok"
	ToolMissing  ToolState = "missing"
	ToolMismatch ToolState = "mismatch"
)

// ToolStatusRow is a stable, machine-friendly representation of tool state.
// It is used by v0.2 `rig check` (and by `rig run` preflight validation).
//
// Status values: ok | missing | mismatch
type ToolStatusRow struct {
	Name   string `json:"name"`
	Bin    string `json:"bin"`
	Want   string `json:"want"`
	Have   string `json:"have"`
	Status string `json:"status"`
}

func rigLockPathForConfig(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "rig.lock")
}

func localBinDirForConfig(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), ".rig", "bin")
}

// Tool execution authority (v0.3 invariant):
//
//   - All rig-managed tools are installed, checked, and executed exclusively from .rig/bin.
//   - Rig does not fall back to PATH, GOBIN, or GOPATH/bin for tool discovery or execution.
//
// Explicit exception:
//   - The Go toolchain (`go`) is resolved from PATH and validated via `go version`.
//     Rig never installs Go.

func ToolBinPath(configPath string, bin string) string {
	name := bin
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		name += ".exe"
	}
	return filepath.Join(localBinDirForConfig(configPath), name)
}

func normalizeExeNameForMatch(name string) string {
	name = strings.TrimSpace(name)
	if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(name), ".exe") {
		return strings.TrimSuffix(name, ".exe")
	}
	return name
}

// ResolveManagedToolExecutable returns an absolute path in .rig/bin if argv0 refers to a
// tool declared in rig.lock. If argv0 is not a managed tool binary, ok=false.
//
// This is the only supported way for rig to run managed tools in v0.3.
func ResolveManagedToolExecutable(configPath string, lock Lockfile, argv0 string) (path string, ok bool, err error) {
	argv0 = strings.TrimSpace(argv0)
	if argv0 == "" {
		return "", false, nil
	}
	// Only bare command names are eligible. Paths are treated as explicit user input.
	if filepath.IsAbs(argv0) || strings.ContainsRune(argv0, os.PathSeparator) {
		return "", false, nil
	}

	want := normalizeExeNameForMatch(argv0)
	for _, lt := range lock.Tools {
		toolName, _, perr := ParseRequested(lt.Requested)
		if perr != nil {
			return "", false, perr
		}
		bin := strings.TrimSpace(lt.Bin)
		if bin == "" {
			bin = ResolveToolIdentity(toolName).Bin
		}
		if normalizeExeNameForMatch(bin) != want {
			continue
		}
		binPath := ToolBinPath(configPath, bin)
		if err := ensureExecutable(binPath); err != nil {
			return "", true, fmt.Errorf("%s not installed in .rig/bin (run 'rig sync'): %w", bin, err)
		}
		return binPath, true, nil
	}
	return "", false, nil
}

func execCapture(name string, args []string, dir string, env []string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func ParseRequested(requested string) (name string, version string, err error) {
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

func SplitResolved(resolved string) (module string, version string) {
	resolved = strings.TrimSpace(resolved)
	i := strings.LastIndex(resolved, "@")
	if i <= 0 || i == len(resolved)-1 {
		return resolved, ""
	}
	return strings.TrimSpace(resolved[:i]), strings.TrimSpace(resolved[i+1:])
}

func NormalizeToolVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if v == "latest" {
		return "latest"
	}
	return NormalizeSemver(EnsureSemverPrefixV(v))
}

func lockMatchesGoToolchain(lock Lockfile, tools map[string]string) error {
	goReqRaw, _ := tools["go"]
	if strings.TrimSpace(goReqRaw) == "" {
		// If config does not declare a Go toolchain, lock must not either.
		if lock.Toolchain != nil && lock.Toolchain.Go != nil {
			return fmt.Errorf("rig.lock has toolchain.go but rig.toml does not declare tools.go")
		}
		return nil
	}
	goReq, err := NormalizeGoToolchainRequested(goReqRaw)
	if err != nil {
		return fmt.Errorf("tools.go: %w", err)
	}
	if lock.Toolchain == nil || lock.Toolchain.Go == nil {
		return fmt.Errorf("rig.lock missing [toolchain.go] (required by tools.go)")
	}
	if strings.TrimSpace(lock.Toolchain.Go.Kind) != "go-toolchain" {
		return fmt.Errorf("rig.lock toolchain.go.kind must be %q", "go-toolchain")
	}
	if NormalizeToolVersion(lock.Toolchain.Go.Requested) != NormalizeToolVersion(goReq) {
		return fmt.Errorf("go toolchain requested mismatch: rig.toml wants %q, rig.lock has %q", goReqRaw, lock.Toolchain.Go.Requested)
	}
	return nil
}

// LockMatchesTools verifies that rig.lock is consistent with the [tools] map.
// It is intentionally strict.
func LockMatchesTools(lock Lockfile, tools map[string]string) error {
	if err := ValidateLockfile(lock); err != nil {
		return err
	}
	if err := lockMatchesGoToolchain(lock, tools); err != nil {
		return err
	}
	_, tools = splitToolsAndGoRequirement(tools)

	byName := make(map[string]LockedTool, len(lock.Tools))
	for _, lt := range lock.Tools {
		name, _, err := ParseRequested(lt.Requested)
		if err != nil {
			return err
		}
		if _, ok := byName[name]; ok {
			return fmt.Errorf("rig.lock has duplicate tool %q", name)
		}
		kind := strings.TrimSpace(lt.Kind)
		if (kind == "go" || kind == "go-binary") && strings.TrimSpace(lt.Module) == "" {
			return fmt.Errorf("rig.lock tool %q: missing module field", name)
		}
		byName[name] = lt
	}

	for name, wantVer := range tools {
		lt, ok := byName[name]
		if !ok {
			return fmt.Errorf("rig.lock missing tool %q", name)
		}
		_, lockReqVer, err := ParseRequested(lt.Requested)
		if err != nil {
			return err
		}
		if NormalizeToolVersion(lockReqVer) != NormalizeToolVersion(wantVer) {
			return fmt.Errorf("tool %q version mismatch: rig.toml wants %q, rig.lock has %q", name, wantVer, lockReqVer)
		}
		id := ResolveToolIdentity(name)
		if strings.TrimSpace(lt.Module) != strings.TrimSpace(id.Module) {
			return fmt.Errorf("tool %q module mismatch: expected %q, rig.lock has %q", name, id.Module, lt.Module)
		}
		resMod, _ := SplitResolved(lt.Resolved)
		if strings.TrimSpace(resMod) != strings.TrimSpace(id.Module) {
			return fmt.Errorf("tool %q resolved mismatch: expected %q@..., rig.lock has %q", name, id.Module, lt.Resolved)
		}
		if strings.TrimSpace(lt.Bin) != "" && strings.TrimSpace(lt.Bin) != strings.TrimSpace(id.Bin) {
			return fmt.Errorf("tool %q bin mismatch: expected %q, rig.lock has %q", name, id.Bin, lt.Bin)
		}
	}

	for name := range byName {
		if _, ok := tools[name]; !ok {
			return fmt.Errorf("rig.lock has extra tool %q not present in rig.toml", name)
		}
	}
	return nil
}

// ReadRigLockForConfig reads rig.lock next to rig.toml.
func ReadRigLockForConfig(configPath string) (Lockfile, error) {
	return ReadLockfile(rigLockPathForConfig(configPath))
}

// CheckInstalledTools compares .rig/bin tool versions against rig.lock.
// It returns deterministic rows ordered by tool name and also reports "extras".
func CheckInstalledTools(tools map[string]string, lock Lockfile, configPath string) (rows []ToolStatusRow, missing int, mismatched int, extras []string, err error) {
	if err := LockMatchesTools(lock, tools); err != nil {
		return nil, 0, 0, nil, err
	}
	_, tools = splitToolsAndGoRequirement(tools)

	byName := make(map[string]LockedTool, len(lock.Tools))
	for _, lt := range lock.Tools {
		name, _, err := ParseRequested(lt.Requested)
		if err != nil {
			return nil, 0, 0, nil, err
		}
		byName[name] = lt
	}

	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)

	rows = make([]ToolStatusRow, 0, len(names))
	declaredBins := map[string]struct{}{}
	for _, name := range names {
		lt := byName[name]
		_, resolvedVer := SplitResolved(lt.Resolved)
		want := NormalizeToolVersion(resolvedVer)
		bin := strings.TrimSpace(lt.Bin)
		if bin == "" {
			bin = ResolveToolIdentity(name).Bin
		}
		declaredBins[bin] = struct{}{}

		binPath := ToolBinPath(configPath, bin)
		status := ToolOK
		have := ""
		// Missing is strictly about presence in .rig/bin (no PATH fallback).
		if err := ensureExecutable(binPath); err != nil {
			status = ToolMissing
			missing++
		} else {
			expected := strings.TrimSpace(lt.SHA256)
			got, herr := ComputeFileSHA256(binPath)
			if herr != nil {
				status = ToolMismatch
				mismatched++
			} else if expected == "" || got != expected {
				status = ToolMismatch
				mismatched++
			}
		}
		rows = append(rows, ToolStatusRow{Name: name, Bin: bin, Want: want, Have: have, Status: string(status)})
	}

	binDir := localBinDirForConfig(configPath)
	if entries, derr := os.ReadDir(binDir); derr == nil {
		for _, ent := range entries {
			if ent.IsDir() {
				continue
			}
			name := ent.Name()
			if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(name), ".exe") {
				name = strings.TrimSuffix(name, ".exe")
			}
			if _, ok := declaredBins[name]; !ok {
				extras = append(extras, name)
			}
		}
		sort.Strings(extras)
	}

	return rows, missing, mismatched, extras, nil
}
