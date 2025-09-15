package rig

import (
	"regexp"
	"strings"
)

// ToolShortNameMap maps commonly used tool short names to their module paths.
var ToolShortNameMap = map[string]string{
	"golangci-lint": "github.com/golangci/golangci-lint/cmd/golangci-lint",
	"mockery":       "github.com/vektra/mockery/v2",
	"staticcheck":   "honnef.co/go/tools/cmd/staticcheck",
	"revive":        "github.com/mgechev/revive",
}

// ResolveModuleAndBin takes a tool identifier (short name or module path)
// and returns the canonical module path and the expected binary name.
// Examples:
//
//	"mockery" => ("github.com/vektra/mockery/v2", "mockery")
//	"github.com/vektra/mockery/v2" => ("github.com/vektra/mockery/v2", "mockery")
//	"golangci-lint" => ("github.com/golangci/golangci-lint/cmd/golangci-lint", "golangci-lint")
func ResolveModuleAndBin(name string) (module string, bin string) {
	module = name
	if mapped, ok := ToolShortNameMap[name]; ok {
		module = mapped
	}
	bin = name
	if i := strings.LastIndex(module, "/"); i >= 0 {
		bin = module[i+1:]
	}
	return module, bin
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
