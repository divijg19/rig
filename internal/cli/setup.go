// internal/cli/setup.go

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
		extraTools := map[string]string{}
		for _, arg := range args {
			if strings.HasSuffix(arg, ".txt") {
				f, err := os.Open(arg)
				if err != nil {
					return fmt.Errorf("read tools file %s: %w", arg, err)
				}
				defer f.Close()
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					// Format: name = version or module@version
					if i := strings.Index(line, "="); i > 0 {
						name := strings.TrimSpace(line[:i])
						ver := strings.TrimSpace(line[i+1:])
						extraTools[name] = ver
					} else if i := strings.Index(line, "@"); i > 0 {
						name := strings.TrimSpace(line[:i])
						ver := strings.TrimSpace(line[i+1:])
						extraTools[name] = ver
					} else {
						// If just a name, default to latest
						extraTools[line] = "latest"
					}
				}
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("scan tools file %s: %w", arg, err)
				}
			}
		}
		// Merge conf.Tools and extraTools
		tools := map[string]string{}
		for k, v := range conf.Tools {
			tools[k] = v
		}
		for k, v := range extraTools {
			tools[k] = v
		}
		if len(tools) == 0 {
			fmt.Printf("ℹ️  No [tools] specified in %s or provided via .txt\n", path)
			return nil
		}
		if setupCheck {
			fmt.Printf("🔍 Checking pinned tools from %s\n", path)
		} else {
			fmt.Printf("� Setting up tools from %s\n", path)
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
