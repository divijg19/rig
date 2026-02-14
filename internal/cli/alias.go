package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const aliasInfoText = "Command entrypoints:\n\n" +
	"    rig  → main CLI\n" +
	"    rir  → rig run\n" +
	"    ric  → rig check\n" +
	"    ril  → rig tools ls\n" +
	"    rip  → rig tools path\n" +
	"    riw  → rig tools why\n" +
	"    rid  → rig dev\n" +
	"    ris  → rig start\n\n" +
	"Aliases are created automatically when installed via the official installer.\n" +
	"If installed via go install, create symlinks manually.\n\n" +
	"Unix examples:\n" +
	"    ln -sf /usr/local/bin/rig /usr/local/bin/rir\n" +
	"    ln -sf /usr/local/bin/rig /usr/local/bin/ric\n" +
	"    ln -sf /usr/local/bin/rig /usr/local/bin/ril\n" +
	"    ln -sf /usr/local/bin/rig /usr/local/bin/rip\n" +
	"    ln -sf /usr/local/bin/rig /usr/local/bin/riw\n" +
	"    ln -sf /usr/local/bin/rig /usr/local/bin/rid\n" +
	"    ln -sf /usr/local/bin/rig /usr/local/bin/ris\n"

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

func init() {
	rootCmd.AddCommand(aliasCmd)
}
