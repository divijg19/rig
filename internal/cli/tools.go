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

var toolsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List managed tools",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := core.ToolsLS("")
		if err != nil {
			return err
		}
		for _, it := range items {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", it.Name, it.Requested, it.Resolved, it.Path, string(it.Status))
		}
		return nil
	},
}

var toolsPathCmd = &cobra.Command{
	Use:   "path <name>",
	Short: "Print absolute path of a managed tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := core.ToolPath("", args[0])
		if err != nil {
			return err
		}
		fmt.Println(p)
		return nil
	},
}

var toolsWhyCmd = &cobra.Command{
	Use:   "why <name>",
	Short: "Explain tool provenance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := core.ToolWhy("", args[0])
		if err != nil {
			return err
		}
		fmt.Printf("name: %s\n", info.Name)
		fmt.Printf("requested: %s\n", info.Requested)
		fmt.Printf("resolved: %s\n", info.Resolved)
		fmt.Printf("sha256: %s\n", info.SHA256)
		fmt.Printf("path: %s\n", info.Path)
		return nil
	},
}

var toolsDoctorCmd = &cobra.Command{
	Use:   "doctor [name]",
	Short: "Diagnose managed tools",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		reports, err := core.ToolsDoctor("", name)
		if err != nil {
			return err
		}
		for _, r := range reports {
			fmt.Printf("name: %s\n", r.Name)
			fmt.Printf("path: %s\n", r.Path)
			fmt.Printf("exists: %t\n", r.Exists)
			fmt.Printf("executable: %t\n", r.Executable)
			fmt.Printf("sha_expected: %s\n", r.SHAExpected)
			fmt.Printf("sha_actual: %s\n", r.SHAActual)
			fmt.Printf("sha_match: %t\n", r.SHAMatch)
			fmt.Printf("resolved_path: %s\n", r.ResolvedPath)
			fmt.Printf("resolved_ok: %t\n", r.ResolvedOK)
			fmt.Printf("status: %s\n", r.Status)
			if strings.TrimSpace(r.Error) != "" {
				fmt.Printf("error: %s\n", r.Error)
			}
		}
		return nil
	},
}

// toolsCmd represents the tools command
var toolsCmd = &cobra.Command{
	Use:     "tools",
	Short:   "Manage project tools",
	Long:    "Manage project tools defined in [tools] section of rig.toml. Also available via shortcuts: 'rig sync', 'rig outdated'.",
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
					Status  []core.ToolStatusRow `json:"status"`
					Summary struct {
						Missing    int      `json:"missing"`
						Mismatched int      `json:"mismatched"`
						Extra      int      `json:"extra"`
						Extras     []string `json:"extras"`
					} `json:"summary"`
				}{Status: []core.ToolStatusRow{}}
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
		goReqRaw := tools["go"]
		toolsNoGo := stripGoToolchain(tools)

		if len(toolsNoGo) == 0 && strings.TrimSpace(goReqRaw) == "" {
			if toolsCheck && toolsCheckJSON {
				// Emit an empty diff JSON for CI
				payload := struct {
					Status  []core.ToolStatusRow `json:"status"`
					Summary struct {
						Missing    int      `json:"missing"`
						Mismatched int      `json:"mismatched"`
						Extra      int      `json:"extra"`
						Extras     []string `json:"extras"`
					} `json:"summary"`
				}{Status: []core.ToolStatusRow{}}
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

		// Validate Go toolchain requirement (tools.go) if present.
		var toolchain *core.ToolchainLock
		if strings.TrimSpace(goReqRaw) != "" {
			normReq, err := core.NormalizeGoToolchainRequested(goReqRaw)
			if err != nil {
				return err
			}
			detected, err := core.DetectGoToolchainVersion(filepath.Dir(path), nil)
			if err != nil {
				return err
			}
			if normReq != "latest" && strings.TrimSpace(detected) != strings.TrimSpace(normReq) {
				return fmt.Errorf("go toolchain mismatch: have %q, want %q", detected, normReq)
			}
			toolchain = &core.ToolchainLock{Go: &core.GoToolchainLock{Kind: "go-toolchain", Requested: normReq, Detected: detected}}
		}

		// Ensure local bin dir exists and prepare env with GOBIN and PATH
		binDir := localBinDirFor(path)
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return fmt.Errorf("create local bin dir: %w", err)
		}
		env := envWithLocalBin(path, toolsOfflineEnv(toolsOffline), true)

		// Resolve tools into a deterministic rig.lock representation.
		// This enables offline installs/checks and ensures sync is reproducible.
		lockedTools, err := core.ResolveLockedTools(toolsNoGo, filepath.Dir(path), env)
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
				toolName, _, perr := core.ParseRequested(lt.Requested)
				if perr != nil {
					results[i] = result{name: lt.Requested, err: perr}
					return
				}
				_, resolvedVer := core.SplitResolved(lt.Resolved)
				id := core.ResolveToolIdentity(toolName)
				installWithVer := id.InstallPath + "@" + resolvedVer
				err := execCommandSilentEnv("go", []string{"install", installWithVer}, env)
				bin := lt.Bin
				if strings.TrimSpace(bin) == "" {
					bin = id.Bin
				}
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

		// Compute and record binary integrity after successful installs.
		for i := range lockedTools {
			lt := lockedTools[i]
			toolName, _, perr := core.ParseRequested(lt.Requested)
			if perr != nil {
				return perr
			}
			bin := strings.TrimSpace(lt.Bin)
			if bin == "" {
				bin = core.ResolveToolIdentity(toolName).Bin
			}
			binPath := core.ToolBinPath(path, bin)
			sum, herr := core.ComputeFileSHA256(binPath)
			if herr != nil {
				return fmt.Errorf("compute sha256 for %s: %w", bin, herr)
			}
			lockedTools[i].SHA256 = sum
		}

		// Only write lock files after successful installs.
		// This prevents partial or misleading lockfile updates.
		rigLock := core.Lockfile{Schema: core.LockSchema0, Toolchain: toolchain, Tools: lockedTools}
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
	toolsSetupCmd.Flags().BoolVar(&setupCheck, "check", false, "verify installed tool versions against rig.toml (no install)")

	toolsCmd.AddCommand(toolsSyncCmd)
	toolsCmd.AddCommand(toolsCheckCmd)
	toolsCmd.AddCommand(toolsOutdatedCmd)
	toolsCmd.AddCommand(toolsSetupCmd)
	toolsCmd.AddCommand(toolsLsCmd)
	toolsCmd.AddCommand(toolsPathCmd)
	toolsCmd.AddCommand(toolsWhyCmd)
	toolsCmd.AddCommand(toolsDoctorCmd)
	rootCmd.AddCommand(toolsCmd)
}

// checkToolsSync verifies rig.lock is consistent with rig.toml, then checks installed binaries.
func checkToolsSync(tools map[string]string, configPath string) error {
	lockPath := rigLockPathFor(configPath)
	lock, err := core.ReadLockfile(lockPath)
	if err != nil {
		if toolsCheckJSON {
			payload := struct {
				Status  []core.ToolStatusRow `json:"status"`
				Summary struct {
					Missing    int      `json:"missing"`
					Mismatched int      `json:"mismatched"`
					Extra      int      `json:"extra"`
					Extras     []string `json:"extras"`
					Error      string   `json:"error"`
				} `json:"summary"`
			}{Status: []core.ToolStatusRow{}}
			payload.Summary.Error = fmt.Sprintf("rig.lock missing or unreadable: %v", err)
			b, jerr := stdjson.MarshalIndent(payload, "", "  ")
			if jerr != nil {
				return jerr
			}
			fmt.Println(string(b))
		}
		return fmt.Errorf("rig.lock missing or unreadable (%s); run 'rig tools sync' to generate it", lockPath)
	}
	if err := core.LockMatchesTools(lock, tools); err != nil {
		if toolsCheckJSON {
			payload := struct {
				Status  []core.ToolStatusRow `json:"status"`
				Summary struct {
					Missing    int      `json:"missing"`
					Mismatched int      `json:"mismatched"`
					Extra      int      `json:"extra"`
					Extras     []string `json:"extras"`
					Error      string   `json:"error"`
				} `json:"summary"`
			}{Status: []core.ToolStatusRow{}}
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
	rows, missing, mismatched, extras, err := core.CheckInstalledTools(tools, lock, configPath)
	if err != nil {
		return err
	}
	if goReqRaw := strings.TrimSpace(tools["go"]); goReqRaw != "" {
		want, nerr := core.NormalizeGoToolchainRequested(goReqRaw)
		have, herr := core.DetectGoToolchainVersion(filepath.Dir(configPath), nil)
		status := "ok"
		if nerr != nil {
			status = "mismatch"
			mismatched++
			have = ""
		} else if herr != nil {
			status = "missing"
			missing++
			have = ""
		} else if lock.Toolchain == nil || lock.Toolchain.Go == nil {
			status = "missing"
			missing++
		} else if strings.TrimSpace(lock.Toolchain.Go.Detected) != strings.TrimSpace(have) {
			status = "mismatch"
			mismatched++
		}
		rows = append(rows, core.ToolStatusRow{Name: "go", Bin: "go", Want: want, Have: have, Status: status})
	}
	for _, name := range extras {
		if !toolsCheckJSON {
			fmt.Printf("  ⚠️  extra binary not in manifest: %s\n", name)
		}
	}
	extra := len(extras)

	issues := missing + mismatched + extra
	if toolsCheckJSON {
		payload := struct {
			Status  []core.ToolStatusRow `json:"status"`
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
