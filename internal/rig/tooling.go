package rig

import (
	"path"
	"regexp"
	"strings"
)

// ToolIdentity splits a tool into:
// - Module: used for `go list -m` resolution (must be taggable)
// - InstallPath: used for `go install`
// - Bin: resulting binary name in .rig/bin
type ToolIdentity struct {
	Module      string
	InstallPath string
	Bin         string
}

// ToolShortNameMap maps commonly used tool short names to their identities.
var ToolShortNameMap = map[string]ToolIdentity{
	"golangci-lint": {Module: "github.com/golangci/golangci-lint", InstallPath: "github.com/golangci/golangci-lint/cmd/golangci-lint", Bin: "golangci-lint"},
	"mockery":       {Module: "github.com/vektra/mockery/v2", InstallPath: "github.com/vektra/mockery/v2", Bin: "mockery"},
	"staticcheck":   {Module: "honnef.co/go/tools", InstallPath: "honnef.co/go/tools/cmd/staticcheck", Bin: "staticcheck"},
	"revive":        {Module: "github.com/mgechev/revive", InstallPath: "github.com/mgechev/revive", Bin: "revive"},
	"air":           {Module: "github.com/cosmtrek/air", InstallPath: "github.com/cosmtrek/air", Bin: "air"},
	"reflex":        {Module: "github.com/cespare/reflex", InstallPath: "github.com/cespare/reflex", Bin: "reflex"},
	// Common extras
	"dlv":       {Module: "github.com/go-delve/delve", InstallPath: "github.com/go-delve/delve/cmd/dlv", Bin: "dlv"},
	"gotestsum": {Module: "gotest.tools/gotestsum", InstallPath: "gotest.tools/gotestsum", Bin: "gotestsum"},
	"gci":       {Module: "github.com/daixiang0/gci", InstallPath: "github.com/daixiang0/gci", Bin: "gci"},
	"gofumpt":   {Module: "mvdan.cc/gofumpt", InstallPath: "mvdan.cc/gofumpt", Bin: "gofumpt"},
}

// ResolveToolIdentity resolves a tool identifier (short name or module path) into a ToolIdentity.
//
// If the tool was specified by a known short name, the identity is explicit.
// Otherwise, Module and InstallPath are assumed to be the provided value.
func ResolveToolIdentity(name string) ToolIdentity {
	name = strings.TrimSpace(name)
	if id, ok := ToolShortNameMap[name]; ok {
		return id
	}
	bin := inferBinFromInstallPath(name)
	return ToolIdentity{Module: name, InstallPath: name, Bin: bin}
}

func inferBinFromInstallPath(installPath string) string {
	installPath = strings.TrimSpace(installPath)
	if installPath == "" {
		return ""
	}
	base := path.Base(installPath)
	// If the last segment looks like a Go major-version suffix (v2, v3, ...), use the previous segment.
	if majorSuffixRE.MatchString(base) {
		dir := path.Dir(installPath)
		prev := path.Base(dir)
		if prev != "." && prev != "/" && prev != "" {
			return prev
		}
	}
	return base
}

var majorSuffixRE = regexp.MustCompile(`^v[0-9]+$`)

// ResolveModuleAndBin is a compatibility wrapper for older callers.
// Prefer ResolveToolIdentity.
func ResolveModuleAndBin(name string) (module string, bin string) {
	id := ResolveToolIdentity(name)
	return id.InstallPath, id.Bin
}

// NormalizeSemver removes a leading 'v' from a version string for comparison.
func NormalizeSemver(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// EnsureSemverPrefixV adds a leading 'v' to a semver if missing (except 'latest').
func EnsureSemverPrefixV(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "latest" || strings.HasPrefix(v, "v") {
		return v
	}
	if len(v) > 0 && v[0] >= '0' && v[0] <= '9' {
		return "v" + v
	}
	return v
}

// ParseVersionFromOutput extracts a semver (with or without v) from arbitrary --version output.
func ParseVersionFromOutput(s string) string {
	re := regexp.MustCompile(`v?\d+\.\d+\.\d+`)
	m := re.FindString(s)
	return NormalizeSemver(m)
}
