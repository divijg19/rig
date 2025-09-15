// internal/cli/build.go

package cli

import (
	"fmt"
	"os"
	"path/filepath"

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
	Long:  "Compose and run 'go build' using flags from rig.toml profiles and CLI overrides.",
	Example: `
	rig build --dry-run
	rig build --profile release
	rig build --tags netgo --ldflags "-s -w" -o bin/app
	rig build -C ./cmd/rig
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, path, err := loadConfigOrFail()
		if err != nil {
			return err
		}

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

		// Determine effective output and ensure output directory exists
		out := firstNonEmpty(buildOutput, prof.Output)
		if out != "" {
			if dir := filepath.Dir(out); dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return fmt.Errorf("create output directory %s: %w", dir, err)
				}
			}
		}

		// Compose command via core package
		cmdline, env := core.ComposeBuildCommand(prof, core.BuildOverrides{
			Output:  out,
			Tags:    buildTags,
			Ldflags: buildLdflags,
			Gcflags: buildGcflags,
		})
		// Ensure local .rig/bin is preferred on PATH
		env = envWithLocalBin(path, env, false)

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
