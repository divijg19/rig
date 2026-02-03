// internal/cli/tools.go

package cli

import (
	stdjson "encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	toolsOffline   bool
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
		return checkToolsSync(tools, path)
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

		if toolsCheck {
			return checkToolsSync(tools, path)
		}

		fmt.Printf("🔧 Syncing tools from %s\n", path)

		// Ensure local bin dir exists and prepare env with GOBIN and PATH
		binDir := localBinDirFor(path)
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return fmt.Errorf("create local bin dir: %w", err)
		}
		env := envWithLocalBin(path, toolsOfflineEnv(toolsOffline), true)

		// Resolve tools into a deterministic rig.lock representation.
		// This enables offline installs/checks and ensures sync is reproducible.
		lockedTools, err := core.ResolveLockedTools(tools, filepath.Dir(path), env)
		if err != nil {
			return err
		}

		// Concurrent installs with deterministic reporting
		sort.Slice(lockedTools, func(i, j int) bool {
			return lockedTools[i].Requested < lockedTools[j].Requested
		})

		type result struct {
			name, bin, ver string
			err            error
		}
		results := make([]result, len(lockedTools))
		// Concurrency: up to NumCPU, but no more than the number of tools
		conc := max(1, min(len(lockedTools), runtime.NumCPU()))
		sem := make(chan struct{}, conc)
		var wg sync.WaitGroup
		for i, lt := range lockedTools {
			i, lt := i, lt
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				toolName, _, _ := parseRequested(lt.Requested)
				module, resolvedVer := splitResolved(lt.Resolved)
				_, bin := core.ResolveModuleAndBin(toolName)
				moduleWithVer := module + "@" + resolvedVer
				err := execCommandSilentEnv("go", []string{"install", moduleWithVer}, env)
				results[i] = result{name: lt.Requested, bin: bin, ver: resolvedVer, err: err}
			}()
		}
		wg.Wait()
		for _, r := range results {
			if r.err != nil {
				return fmt.Errorf("install %s: %w", r.name, r.err)
			}
			fmt.Printf("✅ %s %s installed\n", r.bin, r.ver)
		}

		// Only write lock files after successful installs.
		// This prevents partial or misleading lockfile updates.
		rigLock := core.Lockfile{Schema: core.LockSchema0, Tools: lockedTools}
		rigLockPath := rigLockPathFor(path)
		if err := core.WriteLockfile(rigLockPath, rigLock); err != nil {
			return fmt.Errorf("write rig.lock: %w", err)
		}

		// Write a fast manifest hash lock as a cache (derived from the declared tools map).
		manifestPath := manifestLockPath(path)
		currentHash := computeToolsHash(tools)
		if err := os.WriteFile(manifestPath, []byte(currentHash), 0o644); err != nil {
			return fmt.Errorf("write manifest lock: %w", err)
		}

		fmt.Printf("🔒 Tools synced (rig.lock: %s, manifest: %s)\n", rigLockPath, manifestPath)
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
	toolsSyncCmd.Flags().BoolVar(&toolsOffline, "offline", false, "do not download modules (sets GOPROXY=off, GOSUMDB=off)")
	toolsCheckCmd.Flags().BoolVar(&toolsCheckJSON, "json", false, "print machine-readable JSON summary")
	toolsOutdatedCmd.Flags().BoolVar(&outdatedJSON, "json", false, "print machine-readable JSON status")

	toolsCmd.AddCommand(toolsSyncCmd)
	toolsCmd.AddCommand(toolsCheckCmd)
	toolsCmd.AddCommand(toolsOutdatedCmd)
	rootCmd.AddCommand(toolsCmd)
}

// checkToolsSync verifies rig.lock is consistent with rig.toml, then checks installed binaries.
func checkToolsSync(tools map[string]string, configPath string) error {
	lockPath := rigLockPathFor(configPath)
	lock, err := core.ReadLockfile(lockPath)
	if err != nil {
		if toolsCheckJSON {
			payload := struct {
				Status  []ToolStatusRow `json:"status"`
				Summary struct {
					Missing    int      `json:"missing"`
					Mismatched int      `json:"mismatched"`
					Extra      int      `json:"extra"`
					Extras     []string `json:"extras"`
					Error      string   `json:"error"`
				} `json:"summary"`
			}{Status: []ToolStatusRow{}}
			payload.Summary.Error = fmt.Sprintf("rig.lock missing or unreadable: %v", err)
			b, jerr := stdjson.MarshalIndent(payload, "", "  ")
			if jerr != nil {
				return jerr
			}
			fmt.Println(string(b))
		}
		return fmt.Errorf("rig.lock missing or unreadable (%s); run 'rig tools sync' to generate it", lockPath)
	}
	if err := lockMatchesTools(lock, tools); err != nil {
		if toolsCheckJSON {
			payload := struct {
				Status  []ToolStatusRow `json:"status"`
				Summary struct {
					Missing    int      `json:"missing"`
					Mismatched int      `json:"mismatched"`
					Extra      int      `json:"extra"`
					Extras     []string `json:"extras"`
					Error      string   `json:"error"`
				} `json:"summary"`
			}{Status: []ToolStatusRow{}}
			payload.Summary.Error = err.Error()
			b, jerr := stdjson.MarshalIndent(payload, "", "  ")
			if jerr != nil {
				return jerr
			}
			fmt.Println(string(b))
		}
		return fmt.Errorf("rig.lock out of date; run 'rig tools sync' (%w)", err)
	}

	if !toolsCheckJSON {
		fmt.Printf("🔍 Checking tools status in %s:\n", configPath)
	}
	rows, missing, mismatched := collectToolStatusWithLock(tools, lock, configPath)

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

	issues := missing + mismatched + extra
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
		payload.Summary.Extras = extras
		b, jerr := stdjson.MarshalIndent(payload, "", "  ")
		if jerr != nil {
			return jerr
		}
		fmt.Println(string(b))
	} else {
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
		if issues == 0 {
			fmt.Println("✅ All tools up to date")
			return nil
		}
		fmt.Printf("\nSummary: %d missing, %d mismatched, %d extra\n", missing, mismatched, extra)
	}
	if issues == 0 {
		return nil
	}
	return fmt.Errorf("tools out of sync")
}
