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
	Short: "Run a managed tool from .rig/bin",
	Long:  "Run a rig-managed tool from .rig/bin as pinned by rig.lock. No PATH fallback and no auto-install; use 'rig tools sync' first.",
	Example: `
  rig x golangci-lint -- run
  rig x mockery -- --help
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("usage: rig x <tool[@version]|module[@version]> [-- args]")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Try to load config for project root; allow missing config
		_, configPath, err := loadConfigOptional()
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

		// Parse target
		target := args[0]
		if strings.Contains(target, "@") {
			return fmt.Errorf("rig x does not install; omit @version and run 'rig tools sync' instead")
		}

		lockPath := filepath.Join(filepath.Dir(configPath), "rig.lock")
		lock, err := core.ReadLockfile(lockPath)
		if err != nil {
			return fmt.Errorf("rig.lock required (%s): %w", lockPath, err)
		}

		binPath, ok, rerr := core.ResolveManagedToolExecutable(configPath, lock, target)
		if rerr != nil {
			return rerr
		}
		if !ok {
			return fmt.Errorf("%s is not a managed tool (declare it in [tools] and run 'rig tools sync')", target)
		}

		// Verify binary integrity against rig.lock before executing.
		want := ""
		for _, lt := range lock.Tools {
			name, _, perr := core.ParseRequested(lt.Requested)
			if perr != nil {
				return perr
			}
			bin := strings.TrimSpace(lt.Bin)
			if bin == "" {
				bin = core.ResolveToolIdentity(name).Bin
			}
			if core.ToolBinPath(configPath, bin) != binPath {
				continue
			}
			want = strings.TrimSpace(lt.SHA256)
			break
		}
		if want == "" {
			return fmt.Errorf("unable to locate %s in rig.lock", target)
		}
		have, herr := core.ComputeFileSHA256(binPath)
		if herr != nil {
			return fmt.Errorf("hash %s: %w", target, herr)
		}
		if have != want {
			return fmt.Errorf("%s integrity mismatch (run 'rig tools sync')", target)
		}

		// Prepare env; tools are executed via absolute .rig/bin paths.
		execDir := strings.TrimSpace(xDir)
		envRun := envWithLocalBin(configPath, xEnv, false)

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
			pretty := target
			if len(toolArgs) > 0 {
				pretty = pretty + " " + strings.Join(toolArgs, " ")
			}
			fmt.Printf("🧪 Dry run: would execute -> %s\n", pretty)
			return nil
		}

		return core.Execute(binPath, toolArgs, core.ExecOptions{Dir: execDir, Env: envRun})
	},
}

func init() {
	xCmd.Flags().BoolVar(&xNoInstall, "no-install", false, "fail if tool is missing or version mismatched")
	xCmd.Flags().BoolVar(&xDryRun, "dry-run", false, "print the command without executing")
	xCmd.Flags().StringVarP(&xDir, "dir", "C", "", "working directory to run the tool in")
	xCmd.Flags().StringArrayVar(&xEnv, "env", nil, "environment variables (KEY=VALUE), can be repeated")
	rootCmd.AddCommand(xCmd)
}
