// internal/cli/run.go

package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	cfg "github.com/divijg19/rig/internal/config"
	"github.com/spf13/cobra"
)

var (
	runWorkingDir string
	runList       bool
	runDryRun     bool
)

// runCmd represents the `rig run <task>` command.
var runCmd = &cobra.Command{
	Use:   "run <task> [-- extra args]",
	Short: "Run a named task from rig.toml",
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
		conf, path, err := cfg.Load("")
		if err != nil {
			if errors.Is(err, cfg.ErrConfigNotFound) {
				return fmt.Errorf("no rig.toml found. run 'rig init' first")
			}
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
			task = task + " " + strings.Join(extraArgs, " ")
		}

		if runDryRun {
			fmt.Printf("� Dry run: would execute -> %s\n", task)
			return nil
		}

		fmt.Printf("�🚀 Running task %q (from %s)\n", taskName, path)

		// On Windows, use `cmd /c`, on Unix use `sh -c`.
		var execCmd *exec.Cmd
		if runtime.GOOS == "windows" {
			execCmd = exec.Command("cmd", "/c", task)
		} else {
			execCmd = exec.Command("sh", "-c", task)
		}

		if runWorkingDir != "" {
			execCmd.Dir = runWorkingDir
		}
		// Stream output directly.
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin

		if err := execCmd.Run(); err != nil {
			// Preserve exit code by returning error to Cobra.
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
	rootCmd.AddCommand(runCmd)
}
