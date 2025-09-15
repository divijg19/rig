// internal/cli/run.go

package cli

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var (
	runWorkingDir string
	runList       bool
	runDryRun     bool
	runEnv        []string
)

// runCmd represents the `rig run <task>` command.
var runCmd = &cobra.Command{
	Use:   "run <task> [-- extra args]",
	Short: "Run a named task from rig.toml",
	Long:  "Run a task from the [tasks] section of rig.toml. Use --list to discover tasks.",
	Example: `
	rig run --list
	rig run test
	rig run build -- -v
	rig run build -C ./cmd/rig
	rig run lint --dry-run
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if runList {
			return nil // allow zero args when listing
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

		if runList {
			if len(conf.Tasks) == 0 {
				fmt.Printf("ℹ️  No tasks defined in %s\n", path)
				return nil
			}
			fmt.Printf("📝 Tasks in %s:\n", path)
			names := make([]string, 0, len(conf.Tasks))
			for name := range conf.Tasks {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Printf("  • %s: %s\n", name, conf.Tasks[name])
			}
			return nil
		}

		taskName := args[0]
		extraArgs := args[1:]

		task, ok := conf.Tasks[taskName]
		if !ok {
			return fmt.Errorf("task %q not found in %s", taskName, path)
		}

		// If extra args are provided, append them to the command string.
		if len(extraArgs) > 0 {
			var parts []string
			parts = append(parts, task)
			parts = append(parts, extraArgs...)
			task = strings.Join(parts, " ")
		}

		// Ensure local tool bin is preferred on PATH
		env := envWithLocalBin(path, runEnv, false)

		if runDryRun {
			fmt.Printf("🧪 Dry run: would execute -> %s\n", task)
			return nil
		}

		fmt.Printf("🚀 Running task %q (from %s)\n", taskName, path)
		if err := core.ExecuteShell(task, core.ExecOptions{Dir: runWorkingDir, Env: env}); err != nil {
			return err
		}
		fmt.Println("✅ Done")
		return nil
	},
}

func init() {
	runCmd.Flags().StringVarP(&runWorkingDir, "dir", "C", "", "working directory to run the task in")
	runCmd.Flags().BoolVar(&runList, "list", false, "list tasks defined in rig.toml")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "print the command without executing")
	runCmd.Flags().StringArrayVar(&runEnv, "env", nil, "environment variables (KEY=VALUE), can be repeated")
	rootCmd.AddCommand(runCmd)
}
