// internal/cli/tools.go

package cli

import (
	stdjson "encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var (
	toolsCheck     bool
	outdatedJSON   bool
	toolsCheckJSON bool
)

// toolsCmd represents the tools command
var toolsCmd = &cobra.Command{
	Use:     "tools",
	Short:   "Manage project tools",
	Long:    "Manage project tools defined in [tools] section of rig.toml. Also available via shortcuts: 'rig sync', 'rig check', 'rig outdated'.",
	Aliases: []string{"t"},
}

// toolsCheckCmd: convenience alias for `rig tools sync --check`
var toolsCheckCmd = &cobra.Command{
	Use:     "check",
	Aliases: []string{"status"},
	Short:   "Verify tools are in sync without installing",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, path, err := loadConfigOrFail()
		if err != nil {
			return err
		}
		tools := conf.Tools
		if len(args) > 0 {
			// accept optional tools.txt style files for parity
			extra, err := parseToolsFiles(args)
			if err != nil {
				return err
			}
			tools = mergeTools(tools, extra)
		}
		if len(tools) == 0 {
			if toolsCheckJSON {
				// Emit an empty diff JSON for CI (mirrors sync --check --json)
				payload := struct {
					Status  []ToolStatusRow `json:"status"`
					Summary struct {
						Missing    int      `json:"missing"`
						Mismatched int      `json:"mismatched"`
						Extra      int      `json:"extra"`
						Extras     []string `json:"extras"`
					} `json:"summary"`
				}{Status: []ToolStatusRow{}}
				b, err := stdjson.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}
			fmt.Printf("ℹ️  No [tools] specified in %s or provided via .txt\n", path)
			return nil
		}
		lockFile := manifestLockPath(path)
		currentHash := computeToolsHash(tools)
		return checkToolsSync(tools, lockFile, currentHash, path)
	},
}

// toolsSyncCmd represents the tools sync command
var toolsSyncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Sync tools from rig.toml to .rig/bin",
	Long:    "Install/update tools defined in [tools] section to .rig/bin and create manifest lock. Shortcut: 'rig sync'.",
	Aliases: []string{"s"},
	Example: `
	rig tools sync
	rig tools sync --check
	rig tools sync --check --json | jq .
	rig tools sync tools.txt
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate flag combinations early for better UX
		if toolsCheckJSON && !toolsCheck {
			return fmt.Errorf("--json is only valid with --check")
		}
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
			if toolsCheck && toolsCheckJSON {
				// Emit an empty diff JSON for CI
				payload := struct {
					Status  []ToolStatusRow `json:"status"`
					Summary struct {
						Missing    int      `json:"missing"`
						Mismatched int      `json:"mismatched"`
						Extra      int      `json:"extra"`
						Extras     []string `json:"extras"`
					} `json:"summary"`
				}{Status: []ToolStatusRow{}}
				b, err := stdjson.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}
			fmt.Printf("ℹ️  No [tools] specified in %s or provided via .txt\n", path)
			return nil
		}

		// Check if tools are already in sync
		lockFile := manifestLockPath(path)
		currentHash := computeToolsHash(tools)

		if toolsCheck {
			return checkToolsSync(tools, lockFile, currentHash, path)
		}

		fmt.Printf("🔧 Syncing tools from %s\n", path)

		// Ensure local bin dir exists and prepare env with GOBIN and PATH
		binDir := localBinDirFor(path)
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return fmt.Errorf("create local bin dir: %w", err)
		}
		env := envWithLocalBin(path, nil, true)

		// Concurrent installs with deterministic reporting
		names := make([]string, 0, len(tools))
		for n := range tools {
			names = append(names, n)
		}
		sort.Strings(names)

		type result struct {
			name, bin, ver string
			err            error
		}
		results := make([]result, len(names))
		// Concurrency: up to NumCPU, but no more than the number of tools
		conc := max(1, min(len(names), runtime.NumCPU()))
		sem := make(chan struct{}, conc)
		var wg sync.WaitGroup
		for i, name := range names {
			i, name := i, name
			ver := tools[name]
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				normalized := core.EnsureSemverPrefixV(ver)
				module, bin := core.ResolveModuleAndBin(name)
				moduleWithVer := module + "@" + normalized
				err := execCommandSilentEnv("go", []string{"install", moduleWithVer}, env)
				results[i] = result{name: name, bin: bin, ver: normalized, err: err}
			}()
		}
		wg.Wait()
		for _, r := range results {
			if r.err != nil {
				return fmt.Errorf("install %s: %w", r.name, r.err)
			}
			fmt.Printf("✅ %s %s installed\n", r.bin, r.ver)
		}

		// Write lock file with hash of current tools
		if err := os.WriteFile(lockFile, []byte(currentHash), 0o644); err != nil {
			return fmt.Errorf("write manifest lock: %w", err)
		}

		fmt.Printf("🔒 Tools synced and locked in %s\n", lockFile)
		return nil
	},
}

// toolsOutdatedCmd reports tools that are missing or have a version mismatch without making changes.
var toolsOutdatedCmd = &cobra.Command{
	Use:     "outdated",
	Short:   "Show missing or mismatched tools",
	Long:    "Checks installed tools in .rig/bin against rig.toml versions and lists any that are missing or mismatched. Shortcut: 'rig outdated'.",
	Aliases: []string{"o"},
	Example: `
	rig tools outdated
	rig tools outdated --json | jq .
	rig tools outdated tools.txt
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, path, err := loadConfigOrFail()
		if err != nil {
			return err
		}

		extraTools, err := parseToolsFiles(args)
		if err != nil {
			return err
		}
		tools := mergeTools(conf.Tools, extraTools)
		if len(tools) == 0 {
			if outdatedJSON {
				fmt.Println("[]")
				return nil
			}
			fmt.Printf("ℹ️  No [tools] specified in %s or provided via .txt\n", path)
			return nil
		}

		if outdatedJSON {
			rows, missing, mismatched := collectToolStatus(tools, path)
			issues := missing + mismatched
			b, err := stdjson.MarshalIndent(rows, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			if issues > 0 {
				return fmt.Errorf("%d tool(s) need update. Run 'rig tools sync'", issues)
			}
			return nil
		}

		// Human output branch
		fmt.Printf("🔍 Checking tools status in %s:\n", path)
		rows, missing, mismatched := collectToolStatus(tools, path)
		issues := missing + mismatched
		for _, r := range rows {
			switch r.Status {
			case "missing":
				fmt.Printf("  ❌ %s not found (want %s)\n", r.Bin, r.Want)
			case "mismatch":
				fmt.Printf("  ❌ %s version mismatch (have %s, want %s)\n", r.Bin, r.Have, r.Want)
			default:
				fmt.Printf("  ✅ %s %s\n", r.Bin, r.Want)
			}
		}
		if issues > 0 {
			return fmt.Errorf("%d tool(s) need update. Run 'rig tools sync'", issues)
		}
		fmt.Println("✅ All tools up to date")
		return nil
	},
}

func init() {
	toolsSyncCmd.Flags().BoolVar(&toolsCheck, "check", false, "verify tools are in sync without installing")
	toolsSyncCmd.Flags().BoolVar(&toolsCheckJSON, "json", false, "use with --check to print machine-readable JSON summary")
	toolsCheckCmd.Flags().BoolVar(&toolsCheckJSON, "json", false, "print machine-readable JSON summary")
	toolsOutdatedCmd.Flags().BoolVar(&outdatedJSON, "json", false, "print machine-readable JSON status")

	toolsCmd.AddCommand(toolsSyncCmd)
	toolsCmd.AddCommand(toolsCheckCmd)
	toolsCmd.AddCommand(toolsOutdatedCmd)
	rootCmd.AddCommand(toolsCmd)
}

// checkToolsSync verifies if tools are in sync with the manifest
func checkToolsSync(tools map[string]string, lockFile, currentHash, configPath string) error {
	// Check if lock file exists and matches
	if lockData, err := os.ReadFile(lockFile); err == nil {
		if strings.TrimSpace(string(lockData)) == currentHash {
			if toolsCheckJSON {
				// Emit an empty diff summary in JSON for CI-friendly checks
				payload := struct {
					Status  []ToolStatusRow `json:"status"`
					Summary struct {
						Missing    int `json:"missing"`
						Mismatched int `json:"mismatched"`
						Extra      int `json:"extra"`
					} `json:"summary"`
				}{Status: []ToolStatusRow{}}
				b, err := stdjson.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
			} else {
				fmt.Printf("✅ Tools are in sync with %s\n", configPath)
			}
			return nil
		}
	}

	if !toolsCheckJSON {
		fmt.Printf("❌ Tools are out of sync. Run 'rig tools sync' to update.\n")
	}

	// Show detailed status

	if !toolsCheckJSON {
		fmt.Printf("🔍 Checking individual tools from %s:\n", configPath)
	}
	rows, missing, mismatched := collectToolStatus(tools, configPath)

	// Detect extra binaries present in .rig/bin that aren't declared in tools
	var extras []string
	binDir := localBinDirFor(configPath)
	if entries, err := os.ReadDir(binDir); err == nil {
		declaredBins := map[string]struct{}{}
		for k := range tools {
			_, b := core.ResolveModuleAndBin(k)
			declaredBins[b] = struct{}{}
		}
		for _, ent := range entries {
			if ent.IsDir() {
				continue
			}
			name := ent.Name()
			// Normalize for Windows
			if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(name), ".exe") {
				name = strings.TrimSuffix(name, ".exe")
			}
			if _, ok := declaredBins[name]; !ok {
				if !toolsCheckJSON {
					fmt.Printf("  ⚠️  extra binary not in manifest: %s\n", name)
				}
				extras = append(extras, name)
			}
		}
		sort.Strings(extras)
	}
	extra := len(extras)

	if toolsCheckJSON {
		payload := struct {
			Status  []ToolStatusRow `json:"status"`
			Summary struct {
				Missing    int      `json:"missing"`
				Mismatched int      `json:"mismatched"`
				Extra      int      `json:"extra"`
				Extras     []string `json:"extras"`
			} `json:"summary"`
		}{Status: rows}
		payload.Summary.Missing = missing
		payload.Summary.Mismatched = mismatched
		payload.Summary.Extra = extra
		// Use precomputed, normalized extras
		payload.Summary.Extras = extras
		b, err := stdjson.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	} else {
		fmt.Printf("\nSummary: %d missing, %d mismatched, %d extra\n", missing, mismatched, extra)
	}
	return fmt.Errorf("tools out of sync")
}
