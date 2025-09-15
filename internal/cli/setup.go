// internal/cli/setup.go

package cli

import (
	"fmt"
	"os"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var setupCheck bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install pinned tools into .rig/bin (from [tools])",
	Long:  "Reads [tools] from rig.toml and installs version-locked tools locally into .rig/bin using 'go install'.",
	Example: `
	rig setup
	# pin tools in rig.toml first, e.g.:
	# [tools]\n# golangci-lint = "1.62.0"\n# github.com/vektra/mockery/v2 = "v2.46.0"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, path, err := loadConfigOrFail()
		if err != nil {
			return err
		}
		// Optionally read extra tools from a file (like pip requirements.txt)
		extraTools, err := parseToolsFiles(args)
		if err != nil {
			return err
		}
		// Merge conf.Tools and extraTools
		tools := mergeTools(conf.Tools, extraTools)
		if len(tools) == 0 {
			fmt.Printf("ℹ️  No [tools] specified in %s or provided via .txt\n", path)
			return nil
		}
		if setupCheck {
			fmt.Printf("🔍 Checking pinned tools from %s\n", path)
		} else {
			fmt.Printf("🔧 Setting up tools from %s\n", path)
		}
		// Ensure local bin dir exists and prepare env with GOBIN and PATH
		binDir := localBinDirFor(path)
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return fmt.Errorf("create local bin dir: %w", err)
		}
		env := envWithLocalBin(path, nil, true)
		for name, ver := range tools {
			normalized := core.EnsureSemverPrefixV(ver)
			module, bin := core.ResolveModuleAndBin(name)
			moduleWithVer := module + "@" + normalized
			if setupCheck {
				out, _ := execCommandEnv(bin, []string{"--version"}, env)
				have := core.ParseVersionFromOutput(out)
				want := core.NormalizeSemver(normalized)
				if have == "" {
					fmt.Printf("  ❌ %s not found (want %s)\n", bin, normalized)
				} else if have != want {
					fmt.Printf("  ❌ %s version mismatch (have %s, want %s)\n", bin, have, want)
				} else {
					fmt.Printf("  ✅ %s %s\n", bin, normalized)
				}
				continue
			}
			if err := execCommandSilentEnv("go", []string{"install", moduleWithVer}, env); err != nil {
				return fmt.Errorf("install %s: %w", name, err)
			}
			fmt.Printf("✅ %s %s installed\n", bin, normalized)
		}
		return nil
	},
}

func init() {
	setupCmd.Flags().BoolVar(&setupCheck, "check", false, "verify installed tool versions against rig.toml (no install)")
	rootCmd.AddCommand(setupCmd)
}
