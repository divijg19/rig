package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const aliasInfoText = `Command entrypoints:

	rig  → main CLI
	rir  → rig run
	ric  → rig check
	rid  → rig dev
	ris  → rig start
`

var aliasCmd = &cobra.Command{
	Use:     "alias",
	Short:   "Explain rig command aliases",
	Long:    "Aliases are logical: rig is the only binary; behavior depends on invocation name (argv[0]).",
	Args:    cobra.NoArgs,
	Example: "  rig alias\n",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(aliasInfoText)
	},
}

var aliasesDeprecatedCmd = &cobra.Command{
	Use:    "aliases",
	Short:  "Explain rig command aliases (deprecated)",
	Hidden: true,
	Args:   cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Backwards compatibility with earlier releases that used `rig aliases`.
		// Kept hidden to avoid reinforcing the plural form.
		fmt.Fprintln(cmd.ErrOrStderr(), "warning: `rig aliases` is deprecated; use `rig alias`")
		fmt.Print(aliasInfoText)
	},
}

func init() {
	rootCmd.AddCommand(aliasCmd)
	rootCmd.AddCommand(aliasesDeprecatedCmd)
}
