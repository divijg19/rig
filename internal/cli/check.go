package cli

import (
	"errors"
	"fmt"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Verify rig.lock and installed tools",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rep, err := core.Check("")
		if b, mErr := rep.MarshalJSONStable(); mErr == nil {
			fmt.Println(string(b))
		}
		if err != nil {
			return err
		}
		if !rep.OK {
			return errors.New("check failed")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
