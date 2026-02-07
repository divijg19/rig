package rig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type goModuleInfo struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
	Sum     string `json:"Sum"`
}

// ResolveLockedTools resolves a [tools] map from rig.toml into LockedTool facts.
//
// It does not write any files.
// It performs deterministic resolution using `go list -m -json <module>@<requested>`.
func ResolveLockedTools(tools map[string]string, workDir string, env []string) ([]LockedTool, error) {
	if len(tools) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(tools))
	for k := range tools {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	locked := make([]LockedTool, 0, len(keys))
	for _, name := range keys {
		reqVer := strings.TrimSpace(tools[name])
		if reqVer == "" {
			return nil, fmt.Errorf("tool %q: empty version is not allowed (use an explicit version or \"latest\")", name)
		}
		normalized := EnsureSemverPrefixV(reqVer)
		id := ResolveToolIdentity(name)
		resolvedVer, sum, err := goListModuleVersion(id.Module, normalized, workDir, env)
		if err != nil {
			return nil, fmt.Errorf("resolve %s@%s: %w", id.Module, normalized, err)
		}

		lt := LockedTool{
			Kind:      "go-binary",
			Requested: fmt.Sprintf("%s@%s", name, reqVer),
			Resolved:  fmt.Sprintf("%s@%s", id.Module, resolvedVer),
			Module:    id.Module,
			Bin:       id.Bin,
			Checksum:  strings.TrimSpace(sum),
		}
		locked = append(locked, lt)
	}
	return locked, nil
}

var goListModuleVersion = resolveGoModuleVersion

func resolveGoModuleVersion(module, version, workDir string, env []string) (resolvedVersion string, sum string, err error) {
	target := module + "@" + version
	cmd := exec.Command("go", "list", "-m", "-json", target)
	if workDir != "" {
		cmd.Dir = filepath.Clean(workDir)
	}
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("go list -m failed: %w: %s", err, strings.TrimSpace(out.String()))
	}
	var info goModuleInfo
	if err := json.Unmarshal(out.Bytes(), &info); err != nil {
		return "", "", fmt.Errorf("parse go list output: %w", err)
	}
	if strings.TrimSpace(info.Version) == "" {
		return "", "", fmt.Errorf("go list returned empty version")
	}
	return info.Version, info.Sum, nil
}
