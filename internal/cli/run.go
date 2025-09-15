// internal/cli/run.go

package cli

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/divijg19/rig/internal/config"
	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var (
	runWorkingDir string
	runList       bool
	runListJSON   bool
	runDryRun     bool
	runEnv        []string
)

// runCmd represents the `rig run <task>` command.
var runCmd = &cobra.Command{
	Use:     "run <task> [-- extra args]",
	Short:   "Run a named task from rig.toml",
	Long:    "Run a task from the [tasks] section of rig.toml. Supports both simple string commands and structured tasks with environment variables and dependencies. Use --list to discover tasks. Shortcuts: 'rig r', 'rig ls' (for --list).",
	Aliases: []string{"r"},
	Example: `
	rig run --list
	rig run test
	rig run build -- -v
	rig run build -C ./cmd/rig
	rig run lint --dry-run
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if runList {
			// Listing mode: no task required; --json only valid with --list
			if !runListJSON {
				return nil
			}
			return nil
		}
		if runListJSON {
			return errors.New("--json is only valid with --list")
		}
		if len(args) < 1 {
			return errors.New("missing task name; usage: rig run <task> [-- extra args] or rig run --list")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, path, err := loadConfigOrFail()
		if err != nil {
			return err
		}

		// Fast tool sync check if tools are defined; skip in list mode for speed
		if !runList {
			if len(conf.Tools) > 0 && !runDryRun {
				if err := checkToolsSyncFast(conf.Tools, path); err != nil {
					return err
				}
			}
		}

		if runList {
			if len(conf.Tasks) == 0 {
				fmt.Printf("ℹ️  No tasks defined in %s\n", path)
				return nil
			}
			names := make([]string, 0, len(conf.Tasks))
			for name := range conf.Tasks {
				names = append(names, name)
			}
			sort.Strings(names)
			if runListJSON {
				type tinfo struct {
					Name        string   `json:"name"`
					Description string   `json:"description,omitempty"`
					Command     string   `json:"command,omitempty"`
					Argv        []string `json:"argv,omitempty"`
					DependsOn   []string `json:"dependsOn,omitempty"`
					EnvKeys     []string `json:"envKeys,omitempty"`
					Shell       string   `json:"shell,omitempty"`
				}
				infos := make([]tinfo, 0, len(names))
				for _, name := range names {
					tk := conf.Tasks[name]
					// collect env keys deterministically
					var envKeys []string
					if len(tk.Env) > 0 {
						for k := range tk.Env {
							envKeys = append(envKeys, k)
						}
						sort.Strings(envKeys)
					}
					infos = append(infos, tinfo{
						Name:        name,
						Description: tk.Description,
						Command:     tk.Command,
						Argv:        tk.Argv,
						DependsOn:   tk.DependsOn,
						EnvKeys:     envKeys,
						Shell:       tk.Shell,
					})
				}
				js, err := jsonMarshal(infos)
				if err != nil {
					return err
				}
				fmt.Println(string(js))
			} else {
				fmt.Printf("📝 Tasks in %s:\n", path)
				for _, name := range names {
					task := conf.Tasks[name]
					display := task.Command
					if len(task.Argv) > 0 {
						display = strings.Join(task.Argv, " ")
					}
					if task.Description != "" {
						fmt.Printf("  • %s: %s (%s)\n", name, display, task.Description)
					} else {
						fmt.Printf("  • %s: %s\n", name, display)
					}
				}
			}
			return nil
		}

		taskName := args[0]
		extraArgs := args[1:]

		taskConfig, ok := conf.Tasks[taskName]
		if !ok {
			return fmt.Errorf("task %q not found in %s", taskName, path)
		}

		// Extract task information
		command := taskConfig.Command
		if command == "" && len(taskConfig.Argv) == 0 {
			return fmt.Errorf("task %q has no command defined", taskName)
		}
		taskEnv := taskConfig.Env

		// Resolve dependencies with cycle detection and topological ordering
		order, err := resolveTaskOrder(conf.Tasks, taskName)
		if err != nil {
			return err
		}
		// order includes the root task as last; execute all but the last as dependencies
		if len(order) > 1 {
			fmt.Printf("🔗 Dependency plan for %q: %v\n", taskName, order[:len(order)-1])
		}
		for _, depName := range order[:len(order)-1] {
			depConfig := conf.Tasks[depName]
			depCommand := depConfig.Command
			depEnv := depConfig.Env

			// Merge environment for dependency
			env := mergeEnvWithLocalBin(path, runEnv, depEnv, false)

			fmt.Printf("  🔸 Running dependency: %s\n", depName)
			if runDryRun {
				if len(depConfig.Argv) > 0 {
					fmt.Printf("    ↪ would run: %s\n", strings.Join(depConfig.Argv, " "))
				} else {
					fmt.Printf("    ↪ would run: %s\n", depCommand)
				}
				continue
			}
			var err error
			if len(depConfig.Argv) > 0 {
				err = core.Execute(depConfig.Argv[0], depConfig.Argv[1:], core.ExecOptions{Dir: runWorkingDir, Env: env})
			} else if depConfig.Shell != "" {
				err = core.ExecuteShellWith(depConfig.Shell, depCommand, core.ExecOptions{Dir: runWorkingDir, Env: env})
			} else {
				err = core.ExecuteShell(depCommand, core.ExecOptions{Dir: runWorkingDir, Env: env})
			}
			if err != nil {
				return fmt.Errorf("dependency %q failed: %w", depName, err)
			}
		}

		// If extra args are provided, append them to the command string.
		if len(extraArgs) > 0 {
			if len(taskConfig.Argv) > 0 {
				taskConfig.Argv = append(taskConfig.Argv, extraArgs...)
			} else {
				var parts []string
				parts = append(parts, command)
				parts = append(parts, extraArgs...)
				command = strings.Join(parts, " ")
			}
		}

		// Ensure local tool bin is preferred on PATH and merge task environment
		env := mergeEnvWithLocalBin(path, runEnv, taskEnv, false)

		if runDryRun {
			if len(taskConfig.Argv) > 0 {
				fmt.Printf("🧪 Dry run: would execute -> %s\n", strings.Join(taskConfig.Argv, " "))
			} else {
				fmt.Printf("🧪 Dry run: would execute -> %s\n", command)
			}
			if len(taskEnv) > 0 {
				fmt.Printf("🧪 Environment variables: %v\n", taskEnv)
			}
			return nil
		}

		fmt.Printf("🚀 Running task %q (from %s)\n", taskName, path)
		var execErr error
		if len(taskConfig.Argv) > 0 {
			execErr = core.Execute(taskConfig.Argv[0], taskConfig.Argv[1:], core.ExecOptions{Dir: runWorkingDir, Env: env})
		} else if taskConfig.Shell != "" {
			execErr = core.ExecuteShellWith(taskConfig.Shell, command, core.ExecOptions{Dir: runWorkingDir, Env: env})
		} else {
			execErr = core.ExecuteShell(command, core.ExecOptions{Dir: runWorkingDir, Env: env})
		}
		if execErr != nil {
			return execErr
		}
		fmt.Println("✅ Done")
		return nil
	},
}

// resolveTaskOrder returns a deterministic topological order of dependencies ending with root.
// On cycle, returns an error describing a minimal cycle path.
func resolveTaskOrder(tasks config.TasksMap, root string) ([]string, error) {
	// Build adjacency list
	adj := make(map[string][]string, len(tasks))
	for name, t := range tasks {
		// copy slice to avoid aliasing
		if len(t.DependsOn) > 0 {
			deps := make([]string, len(t.DependsOn))
			copy(deps, t.DependsOn)
			// stable order for determinism
			sort.Strings(deps)
			adj[name] = deps
		} else {
			adj[name] = nil
		}
	}
	if _, ok := adj[root]; !ok {
		return nil, fmt.Errorf("task %q not found", root)
	}

	// states: 0=unseen,1=visiting,2=done
	state := make(map[string]int, len(adj))
	var order []string
	var stack []string

	var dfs func(string) error
	dfs = func(u string) error {
		st := state[u]
		if st == 1 {
			// found a cycle; extract minimal cycle from stack
			idx := -1
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i] == u {
					idx = i
					break
				}
			}
			if idx >= 0 {
				cycle := append(append([]string{}, stack[idx:]...), u)
				return fmt.Errorf("♻️  dependency cycle detected: %s", strings.Join(cycle, " -> "))
			}
			return fmt.Errorf("♻️  dependency cycle detected involving %q", u)
		}
		if st == 2 {
			return nil
		}
		state[u] = 1
		stack = append(stack, u)
		for _, v := range adj[u] {
			if _, exists := adj[v]; !exists {
				return fmt.Errorf("dependency %q referenced by %q not found", v, u)
			}
			if err := dfs(v); err != nil {
				return err
			}
		}
		// pop
		stack = stack[:len(stack)-1]
		state[u] = 2
		order = append(order, u)
		return nil
	}

	if err := dfs(root); err != nil {
		return nil, err
	}
	// remove duplicates while preserving order (shouldn't be any due to state guards)
	return order, nil
}

func init() {
	runCmd.Flags().StringVarP(&runWorkingDir, "dir", "C", "", "working directory to run the task in")
	runCmd.Flags().BoolVarP(&runList, "list", "l", false, "list tasks defined in rig.toml")
	runCmd.Flags().BoolVarP(&runListJSON, "json", "j", false, "use with --list to print machine-readable JSON")
	runCmd.Flags().BoolVarP(&runDryRun, "dry-run", "n", false, "print the command without executing")
	runCmd.Flags().StringArrayVarP(&runEnv, "env", "E", nil, "environment variables (KEY=VALUE), can be repeated")
	rootCmd.AddCommand(runCmd)
}

// mergeEnvWithLocalBin wraps envWithLocalBin and appends task env map entries deterministically.
func mergeEnvWithLocalBin(configPath string, base []string, kv map[string]string, includeGOBIN bool) []string {
	env := envWithLocalBin(configPath, base, includeGOBIN)
	if len(kv) == 0 {
		return env
	}
	// sort keys for determinism (useful for tests and stability)
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		env = append(env, fmt.Sprintf("%s=%s", k, kv[k]))
	}
	return env
}

// jsonMarshal is a tiny wrapper to keep imports localized
func jsonMarshal(v any) ([]byte, error) {
	return stdjson.MarshalIndent(v, "", "  ")
}
