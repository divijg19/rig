// internal/cli/setup.go

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var setupCheck bool

var setupCmd = &cobra.Command{
	Use:    "setup",
	Short:  "Install pinned tools into .rig/bin (from [tools])",
	Long:   "Reads [tools] from rig.toml and installs version-locked tools locally into .rig/bin using 'go install'.",
	Hidden: true,
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
		tools = stripGoToolchain(tools) // Go is a toolchain, not a rig-managed installable tool.
		if len(tools) == 0 {
			fmt.Printf("ℹ️  No [tools] specified in %s or provided via .txt\n", path)
			return nil
		}
		if setupCheck {
			fmt.Printf("🔍 Checking pinned tools from %s\n", path)
		} else {
			fmt.Printf("🔧 Setting up tools from %s\n", path)
		}

		if setupCheck {
			return checkToolsSync(mergeTools(conf.Tools, extraTools), path)
		}

		// Ensure local bin dir exists and prepare env with GOBIN and PATH
		binDir := localBinDirFor(path)
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return fmt.Errorf("create local bin dir: %w", err)
		}
		env := envWithLocalBin(path, nil, true)

		// Resolve and install deterministically.
		lockedTools, err := core.ResolveLockedTools(tools, filepath.Dir(path), env)
		if err != nil {
			return err
		}
		sort.Slice(lockedTools, func(i, j int) bool { return lockedTools[i].Requested < lockedTools[j].Requested })
		for i, lt := range lockedTools {
			toolName, _, perr := core.ParseRequested(lt.Requested)
			if perr != nil {
				return perr
			}
			_, resolvedVer := core.SplitResolved(lt.Resolved)
			id := core.ResolveToolIdentity(toolName)
			installWithVer := id.InstallPath + "@" + resolvedVer
			if err := execCommandSilentEnv("go", []string{"install", installWithVer}, env); err != nil {
				return fmt.Errorf("install %s: %w", lt.Requested, err)
			}
			bin := strings.TrimSpace(lt.Bin)
			if bin == "" {
				bin = id.Bin
			}
			sum, herr := core.ComputeFileSHA256(core.ToolBinPath(path, bin))
			if herr != nil {
				return fmt.Errorf("compute sha256 for %s: %w", bin, herr)
			}
			lockedTools[i].SHA256 = sum
			fmt.Printf("✅ %s %s installed\n", bin, resolvedVer)
		}

		rigLock := core.Lockfile{Schema: core.LockSchema0, Toolchain: nil, Tools: lockedTools}
		rigLockPath := rigLockPathFor(path)
		if err := core.WriteLockfile(rigLockPath, rigLock); err != nil {
			return fmt.Errorf("write rig.lock: %w", err)
		}
		manifestPath := manifestLockPath(path)
		currentHash := computeToolsHash(mergeTools(conf.Tools, extraTools))
		if err := os.WriteFile(manifestPath, []byte(currentHash), 0o644); err != nil {
			return fmt.Errorf("write manifest lock: %w", err)
		}
		return nil
	},
}

func init() {
	setupCmd.Flags().BoolVar(&setupCheck, "check", false, "verify installed tool versions against rig.toml (no install)")
	rootCmd.AddCommand(setupCmd)
}

var toolsSetupCmd = &cobra.Command{
	Use:     "setup",
	Short:   "Install pinned tools into .rig/bin (from [tools])",
	Long:    "Reads [tools] from rig.toml and installs version-locked tools locally into .rig/bin using 'go install'.",
	Example: setupCmd.Example,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setupCmd.RunE(setupCmd, args)
	},
}
