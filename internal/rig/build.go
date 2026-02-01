// internal/rig/build.go

package rig

import (
	"path/filepath"
	"strings"

	cfg "github.com/divijg19/rig/internal/config"
)

// BuildOverrides represents CLI-provided overrides for build flags.
type BuildOverrides struct {
	Output  string
	Tags    []string
	Ldflags string
	Gcflags string
}

// ComposeBuildCommand returns the go build command line and env based on the
// provided profile and CLI overrides. The working directory handling is done by callers.
func ComposeBuildCommand(prof cfg.BuildProfile, o BuildOverrides) (cmdline string, env []string) {
	var parts []string
	parts = append(parts, "go", "build")

	// Merge flags (CLI overrides profile)
	ldflags := firstNonEmpty(o.Ldflags, prof.Ldflags)
	gcflags := firstNonEmpty(o.Gcflags, prof.Gcflags)
	tags := o.Tags
	if len(tags) == 0 && len(prof.Tags) > 0 {
		tags = prof.Tags
	}
	out := firstNonEmpty(o.Output, prof.Output)
	if out != "" {
		parts = append(parts, "-o", shellQuote(filepath.Clean(out)))
	}
	if ldflags != "" {
		parts = append(parts, "-ldflags", shellQuote(ldflags))
	}
	if gcflags != "" {
		parts = append(parts, "-gcflags", shellQuote(gcflags))
	}
	if len(tags) > 0 {
		parts = append(parts, "-tags", shellQuote(strings.Join(tags, ",")))
	}
	if len(prof.Flags) > 0 {
		parts = append(parts, prof.Flags...)
	}

	// default package
	parts = append(parts, ".")

	// Env
	if prof.Env != nil {
		for k, v := range prof.Env {
			env = append(env, k+"="+v)
		}
	}

	return strings.Join(parts, " "), env
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return strings.TrimSpace(b)
}

// Minimal quoting helpers duplicated here to avoid external dependencies.
func shellQuote(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}
