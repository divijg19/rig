package cli

import (
	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run the dev loop (watch + restart)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return core.Dev("")
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}
