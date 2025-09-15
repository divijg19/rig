package cli

import (
	"testing"

	core "github.com/divijg19/rig/internal/rig"
)

func TestNormalizeSemver(t *testing.T) {
	cases := map[string]string{
		"v1.2.3":   "1.2.3",
		"1.2.3":    "1.2.3",
		" v1.2.3 ": "1.2.3",
	}
	for in, want := range cases {
		if got := core.NormalizeSemver(in); got != want {
			t.Fatalf("normalizeSemver(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEnsureSemverPrefixV(t *testing.T) {
	cases := map[string]string{
		"1.2.3":  "v1.2.3",
		"v1.2.3": "v1.2.3",
		"latest": "latest",
		"":       "",
	}
	for in, want := range cases {
		if got := core.EnsureSemverPrefixV(in); got != want {
			t.Fatalf("ensureSemverPrefixV(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseVersionFromOutput(t *testing.T) {
	cases := map[string]string{
		"golangci-lint has version v1.62.0 build": "1.62.0",
		"mockery version 2.46.0":                  "2.46.0",
		"version: v0.0.0-dev":                     "0.0.0",
		"no version here":                         "",
	}
	for in, want := range cases {
		if got := core.ParseVersionFromOutput(in); got != want {
			t.Fatalf("parseVersionFromOutput(%q) = %q, want %q", in, got, want)
		}
	}
}
