package cli

import (
	"fmt"
	"os"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade rig to latest release",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		res, err := core.UpgradeSelf(core.UpgradeOptions{
			CurrentVersion: version,
			ExecutablePath: exePath,
		})
		if err != nil {
			return err
		}
		if res.UpToDate {
			fmt.Printf("rig is up to date (%s)\n", res.Current)
			return nil
		}
		fmt.Printf("upgraded rig: %s -> %s\n", res.Current, res.Latest)
		fmt.Printf("asset: %s\n", res.AssetName)
		fmt.Printf("checksum: %s\n", res.ChecksumName)
		fmt.Printf("path: %s\n", res.ExecutableOut)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
