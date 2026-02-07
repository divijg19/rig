package cli

import (
	"fmt"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show rig status (read-only)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rep, err := core.Status("")
		if err != nil {
			return err
		}
		fmt.Printf("config: %s\n", rep.ConfigPath)
		if !rep.HasLock {
			fmt.Printf("lock: %s (missing)\n", rep.LockPath)
			return nil
		}
		fmt.Printf("lock: %s\n", rep.LockPath)
		fmt.Printf("lockMatchesConfig: %t\n", rep.LockMatchesConfig)
		fmt.Printf("toolsOk: %t\n", rep.ToolsOK)
		fmt.Printf("missing: %d\n", rep.Missing)
		fmt.Printf("mismatched: %d\n", rep.Mismatched)
		fmt.Printf("extras: %d\n", rep.Extras)
		if rep.Go != nil {
			fmt.Printf("goRequested: %s\n", rep.Go.Requested)
			fmt.Printf("goLocked: %s\n", rep.Go.Locked)
			fmt.Printf("goHave: %s\n", rep.Go.Have)
			fmt.Printf("goStatus: %s\n", rep.Go.Status)
			if rep.Go.Error != "" {
				fmt.Printf("goError: %s\n", rep.Go.Error)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
