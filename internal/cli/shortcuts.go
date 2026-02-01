// internal/cli/shortcuts.go

package cli

import "github.com/spf13/cobra"

// syncCmd (top-level): shortcut for `rig tools sync`
var syncCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"tool-sync"},
	Short:   "Sync tools from rig.toml to .rig/bin",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delegate to toolsSyncCmd implementation
		return toolsSyncCmd.RunE(toolsSyncCmd, args)
	},
}

// checkCmd (top-level): shortcut for `rig tools check` (aka `tools sync --check`)
var checkCmd = &cobra.Command{
	Use:     "check",
	Aliases: []string{"status"},
	Short:   "Verify tools are in sync without installing",
	RunE: func(cmd *cobra.Command, args []string) error {
		return toolsCheckCmd.RunE(toolsCheckCmd, args)
	},
}

// outdatedCmd (top-level): shortcut for `rig tools outdated`
var outdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "Show missing or mismatched tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		return toolsOutdatedCmd.RunE(toolsOutdatedCmd, args)
	},
}

func init() {
	// Mirror relevant flags so they affect the same underlying variables
	syncCmd.Flags().BoolVar(&toolsCheck, "check", false, "verify tools are in sync without installing")
	syncCmd.Flags().BoolVar(&toolsCheckJSON, "json", false, "use with --check to print machine-readable JSON summary")

	checkCmd.Flags().BoolVar(&toolsCheckJSON, "json", false, "print machine-readable JSON summary")

	outdatedCmd.Flags().BoolVar(&outdatedJSON, "json", false, "print machine-readable JSON status")

	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(outdatedCmd)
}
