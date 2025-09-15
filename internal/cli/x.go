// internal/cli/x.go

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
	xNoInstall bool
	xDryRun    bool
	xDir       string
	xEnv       []string
)

// xCmd provides an ephemeral runner similar to npx/bunx/uvx.
// Usage: rig x <tool[@version] | module[@version]> [-- args]
var xCmd = &cobra.Command{
	Use:   "x <tool[@version]|module[@version]> [-- args]",
	Short: "Run a tool ephemerally (installs if needed)",
	Long:  "Run a tool ephemerally by name or module path, optionally with @version. Installs into .rig/bin if missing or mismatched, then executes with provided args.",
	Example: `
  rig x golangci-lint@v1.59.0 run
  rig x mockery -- --help
  rig x github.com/vektra/mockery/v2@v2.42.0 -- --all
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("usage: rig x <tool[@version]|module[@version]> [-- args]")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Try to load config for project root; allow missing config
		conf, configPath, err := loadConfigOptional()
		if err != nil && !errors.Is(err, cfg.ErrConfigNotFound) {
			return err
		}
		if errors.Is(err, cfg.ErrConfigNotFound) || configPath == "" {
			// Fallback to CWD to compute .rig/bin path
			cwd, cwderr := os.Getwd()
			if cwderr != nil {
				return cwderr
			}
			configPath = filepath.Join(cwd, "rig.toml")
		}

		// Parse target and version
		target := args[0]
		var wantVer string
		if i := strings.LastIndex(target, "@"); i > 0 {
			wantVer = target[i+1:]
			target = target[:i]
		}

		module, bin := core.ResolveModuleAndBin(target)

		// If version not given, try to use version from [tools] in config
		if wantVer == "" && conf != nil && conf.Tools != nil {
			if v, ok := conf.Tools[target]; ok {
				wantVer = v
			} else if v, ok := conf.Tools[module]; ok {
				wantVer = v
			} else if v, ok := conf.Tools[bin]; ok {
				wantVer = v
			}
		}
		if wantVer == "" {
			wantVer = "latest"
		}
		wantVer = core.EnsureSemverPrefixV(wantVer)

		// Prepare env with local bin first; set working dir if provided
		execDir := firstNonEmpty(xDir, xDir)
		env := envWithLocalBin(configPath, xEnv, true)

		// Check installed version by invoking bin --version
		out, _ := execCommandEnv(bin, []string{"--version"}, env)
		have := core.ParseVersionFromOutput(out)
		needInstall := have == "" || (core.NormalizeSemver(have) != core.NormalizeSemver(wantVer))

		if needInstall {
			if xNoInstall {
				return fmt.Errorf("%s not installed (want %s) and --no-install set", bin, wantVer)
			}
			if xDryRun {
				fmt.Printf("🧪 Dry run: would install %s@%s and run it\n", module, wantVer)
				return nil
			}
			// Ensure .rig/bin exists
			if err := os.MkdirAll(localBinDirFor(configPath), 0o755); err != nil {
				return fmt.Errorf("create .rig/bin: %w", err)
			}
			fmt.Printf("🔧 Installing %s %s\n", bin, wantVer)
			if err := execCommandSilentEnv("go", []string{"install", module + "@" + wantVer}, env); err != nil {
				return fmt.Errorf("install %s@%s: %w", module, wantVer, err)
			}
		}

		// Execute with remaining args after the first; support `--` pass-through
		toolArgs := []string{}
		if len(args) > 1 {
			toolArgs = args[1:]
			// Remove leading "--" if present (Cobra already splits, but support manual style)
			if len(toolArgs) > 0 && toolArgs[0] == "--" {
				toolArgs = toolArgs[1:]
			}
		}

		if xDryRun {
			pretty := bin
			if len(toolArgs) > 0 {
				pretty = pretty + " " + strings.Join(toolArgs, " ")
			}
			fmt.Printf("🧪 Dry run: would execute -> %s\n", pretty)
			return nil
		}

		return core.Execute(bin, toolArgs, core.ExecOptions{Dir: execDir, Env: env})
	},
}

func init() {
	xCmd.Flags().BoolVar(&xNoInstall, "no-install", false, "fail if tool is missing or version mismatched")
	xCmd.Flags().BoolVar(&xDryRun, "dry-run", false, "print the command without executing")
	xCmd.Flags().StringVarP(&xDir, "dir", "C", "", "working directory to run the tool in")
	xCmd.Flags().StringArrayVar(&xEnv, "env", nil, "environment variables (KEY=VALUE), can be repeated")
	rootCmd.AddCommand(xCmd)
}
