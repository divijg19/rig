// internal/cli/build.go

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cfg "github.com/divijg19/rig/internal/config"
	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var (
	buildProfile string
	buildOutput  string
	buildTags    []string
	buildLdflags string
	buildGcflags string
	buildDir     string
	buildDryRun  bool
)

// buildCmd implements `rig build` with optional profiles.
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project using optional profiles from rig.toml",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, path, err := cfg.Load("")
		if err != nil {
			if errors.Is(err, cfg.ErrConfigNotFound) {
				return fmt.Errorf("no rig.toml found. run 'rig init' first")
			}
			return err
		}

		// Start composing go build command
		var parts []string
		parts = append(parts, "go", "build")

		// Apply profile if specified and exists
		var prof cfg.BuildProfile
		if buildProfile != "" {
			if conf.Profiles == nil {
				return fmt.Errorf("profile %q requested, but no [profile.*] defined in %s", buildProfile, path)
			}
			p, ok := conf.Profiles[buildProfile]
			if !ok {
				return fmt.Errorf("profile %q not found in %s", buildProfile, path)
			}
			prof = p
		}

		// Merge flags: CLI flags override profile
		ldflags := firstNonEmpty(buildLdflags, prof.Ldflags)
		gcflags := firstNonEmpty(buildGcflags, prof.Gcflags)
		tags := buildTags
		if len(tags) == 0 && len(prof.Tags) > 0 {
			tags = prof.Tags
		}
		out := firstNonEmpty(buildOutput, prof.Output)
		if out != "" {
			// Ensure output directory exists
			if dir := filepath.Dir(out); dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return fmt.Errorf("create output directory %s: %w", dir, err)
				}
			}
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

		// Package to build: default to current module
		parts = append(parts, ".")

		cmdline := strings.Join(parts, " ")

		// Prepare env from profile
		var env []string
		if prof.Env != nil {
			for k, v := range prof.Env {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
		}

		if buildDryRun {
			fmt.Printf("🧪 Dry run: would execute -> %s\n", cmdline)
			return nil
		}

		fmt.Printf("🔨 Building (profile=%q) using config %s\n", buildProfile, path)
		return core.ExecuteShell(cmdline, core.ExecOptions{Dir: buildDir, Env: env})
	},
}

func init() {
	buildCmd.Flags().StringVar(&buildProfile, "profile", "", "build profile from rig.toml [profile.<name>]")
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "output binary path")
	buildCmd.Flags().StringSliceVar(&buildTags, "tags", nil, "comma-separated build tags")
	buildCmd.Flags().StringVar(&buildLdflags, "ldflags", "", "custom -ldflags (overrides profile)")
	buildCmd.Flags().StringVar(&buildGcflags, "gcflags", "", "custom -gcflags (overrides profile)")
	buildCmd.Flags().StringVarP(&buildDir, "dir", "C", "", "working directory for build")
	buildCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "print the build command without executing")
	rootCmd.AddCommand(buildCmd)
}

// firstNonEmpty returns a if a != "", otherwise b.
func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return strings.TrimSpace(b)
}

// shellQuote provides minimal quoting for arguments that may contain spaces.
// Since we pass a command string to a shell, we quote values. This is not a
// full shell-escaping routine but is sufficient for common flags.
func shellQuote(s string) string {
	if s == "" {
		return s
	}
	// Wrap in double quotes and escape any existing quotes
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}
