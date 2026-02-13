package rig

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var goVersionTokenRE = regexp.MustCompile(`\bgo\d+\.\d+\.\d+\b`)
var semverNoVRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func NormalizeGoToolchainRequested(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", fmt.Errorf("empty go toolchain version")
	}
	if v == "latest" {
		return v, nil
	}
	if strings.HasPrefix(v, "go") {
		v = strings.TrimSpace(strings.TrimPrefix(v, "go"))
	}
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	if !semverNoVRE.MatchString(v) {
		return "", fmt.Errorf("invalid go toolchain version %q (expected x.y.z)", v)
	}
	return v, nil
}

func ParseGoToolchainDetectedFromGoVersionOutput(out string) (string, error) {
	out = strings.TrimSpace(out)
	m := goVersionTokenRE.FindString(out)
	if m == "" {
		return "", fmt.Errorf("unable to parse go version output: %q", out)
	}
	return strings.TrimPrefix(m, "go"), nil
}

func DetectGoToolchainVersion(workDir string, env []string) (string, error) {
	out, err := execCapture("go", []string{"version"}, workDir, env)
	if err != nil {
		return "", fmt.Errorf("go version failed: %w: %s", err, out)
	}
	return ParseGoToolchainDetectedFromGoVersionOutput(out)
}

func splitToolsAndGoRequirement(tools map[string]string) (goReq string, rest map[string]string) {
	rest = make(map[string]string, len(tools))
	for k, v := range tools {
		if k == "go" {
			goReq = v
			continue
		}
		rest[k] = v
	}
	return goReq, rest
}

// GoStatusRow is a stable, machine-friendly representation of Go toolchain state.
// Status values: ok | missing | mismatch
type GoStatusRow struct {
	Requested string `json:"requested"`
	Locked    string `json:"locked"`
	Have      string `json:"have"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

func checkGoAgainstLockIfRequired(tools map[string]string, lock Lockfile, configPath string) (*GoStatusRow, bool) {
	goReqRaw, _ := tools["go"]
	if strings.TrimSpace(goReqRaw) == "" {
		return nil, true
	}
	goReq, err := NormalizeGoToolchainRequested(goReqRaw)
	if err != nil {
		return &GoStatusRow{Requested: goReqRaw, Locked: "", Have: "", Status: "mismatch", Error: err.Error()}, false
	}
	if lock.Toolchain == nil || lock.Toolchain.Go == nil {
		return &GoStatusRow{Requested: goReq, Locked: "", Have: "", Status: "missing", Error: "rig.lock missing [toolchain.go]"}, false
	}
	lockedDetected := strings.TrimSpace(lock.Toolchain.Go.Detected)
	have, derr := DetectGoToolchainVersion(filepath.Dir(configPath), nil)
	if derr != nil {
		return &GoStatusRow{Requested: goReq, Locked: lockedDetected, Have: "", Status: "missing", Error: derr.Error()}, false
	}
	if strings.TrimSpace(lockedDetected) != strings.TrimSpace(have) {
		return &GoStatusRow{Requested: goReq, Locked: lockedDetected, Have: have, Status: "mismatch"}, false
	}
	return &GoStatusRow{Requested: goReq, Locked: lockedDetected, Have: have, Status: "ok"}, true
}

// CheckGoToolchainAgainstLock validates the Go toolchain requirement when tools.go is declared.
func CheckGoToolchainAgainstLock(tools map[string]string, lock Lockfile, configPath string) (*GoStatusRow, bool) {
	return checkGoAgainstLockIfRequired(tools, lock, configPath)
}
