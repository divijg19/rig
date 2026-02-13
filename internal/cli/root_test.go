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

func TestMinHelper(t *testing.T) {
	cases := []struct{ a, b, want int }{
		{1, 2, 1},
		{2, 1, 1},
		{3, 3, 3},
		{0, 5, 0},
	}
	for _, c := range cases {
		if got := min(c.a, c.b); got != c.want {
			t.Fatalf("min(%d,%d)=%d, want %d", c.a, c.b, got, c.want)
		}
	}
}
