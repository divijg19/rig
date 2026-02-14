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

// outdatedCmd (top-level): shortcut for `rig tools outdated`
var outdatedCmd = &cobra.Command{
	Use:    "outdated",
	Short:  "Show missing or mismatched tools",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return toolsOutdatedCmd.RunE(toolsOutdatedCmd, args)
	},
}

var lsToolsCmd = &cobra.Command{
	Use:    "ls",
	Short:  "List managed tools",
	Args:   cobra.NoArgs,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return toolsLsCmd.RunE(toolsLsCmd, args)
	},
}

var pathCmd = &cobra.Command{
	Use:    "path <name>",
	Short:  "Print absolute path of a managed tool",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return toolsPathCmd.RunE(toolsPathCmd, args)
	},
}

var whyCmd = &cobra.Command{
	Use:    "why <name>",
	Short:  "Explain tool provenance",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return toolsWhyCmd.RunE(toolsWhyCmd, args)
	},
}

func init() {
	// Mirror relevant flags so they affect the same underlying variables
	syncCmd.Flags().BoolVar(&toolsCheck, "check", false, "verify tools are in sync without installing")
	syncCmd.Flags().BoolVar(&toolsCheckJSON, "json", false, "use with --check to print machine-readable JSON summary")
	syncCmd.Flags().BoolVar(&toolsOffline, "offline", false, "do not download modules (sets GOPROXY=off, GOSUMDB=off)")

	outdatedCmd.Flags().BoolVar(&outdatedJSON, "json", false, "print machine-readable JSON status")

	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(outdatedCmd)
	rootCmd.AddCommand(lsToolsCmd)
	rootCmd.AddCommand(pathCmd)
	rootCmd.AddCommand(whyCmd)
}
