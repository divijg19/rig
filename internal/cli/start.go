package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "(stub) Start the application (future)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("rig start is not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
