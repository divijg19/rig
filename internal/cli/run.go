// internal/cli/run.go

package cli

import (
	"fmt"
	"sort"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

func newRunLikeCommand(use string, short string) *cobra.Command {
	var list bool
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args: func(cmd *cobra.Command, args []string) error {
			if list {
				if cmd.ArgsLenAtDash() >= 0 {
					return fmt.Errorf("usage: %s --list", cmd.CommandPath())
				}
				if len(args) != 0 {
					return fmt.Errorf("usage: %s --list", cmd.CommandPath())
				}
				return nil
			}
			dash := cmd.ArgsLenAtDash()
			if dash >= 0 {
				if dash != 1 {
					return fmt.Errorf("usage: %s <task> [-- args...]", cmd.CommandPath())
				}
				args = args[:dash]
			}
			if len(args) != 1 {
				return fmt.Errorf("usage: %s <task> [-- args...]", cmd.CommandPath())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				conf, _, err := core.LoadConfig("")
				if err != nil {
					return err
				}
				names := make([]string, 0, len(conf.Tasks))
				for name := range conf.Tasks {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					fmt.Println(name)
				}
				return nil
			}
			dash := cmd.ArgsLenAtDash()
			passthrough := []string(nil)
			if dash >= 0 {
				passthrough = append([]string(nil), args[dash:]...)
				args = args[:dash]
			}
			if len(args) != 1 {
				return fmt.Errorf("usage: %s <task> [-- args...]", cmd.CommandPath())
			}
			return core.Run("", args[0], passthrough)
		},
	}
	cmd.Flags().BoolVar(&list, "list", false, "list available tasks and exit")
	return cmd
}

// runCmd represents the v0.2 `rig run <task>` command.
var runCmd = newRunLikeCommand("run", "Run a named task from rig.toml")


func init() {
	rootCmd.AddCommand(runCmd)
}
