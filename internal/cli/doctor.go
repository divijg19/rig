// internal/cli/doctor.go

package cli

import (
	"fmt"
	"os"
	"strings"

	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor [name]",
	Short: "Check your development environment and tooling",
	Long:  "Diagnose the toolchain and rig-managed tools.",
	Args:  cobra.MaximumNArgs(1),
	Example: `
	rig doctor
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return toolsDoctorCmd.RunE(toolsDoctorCmd, args)
		}

		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		rep, err := core.Doctor("", version, exePath)
		if err != nil {
			return err
		}
		fmt.Printf("version_present: %t\n", rep.VersionPresent)
		fmt.Printf("go_available: %t\n", rep.GoAvailable)
		fmt.Printf("go_version: %s\n", rep.GoVersion)
		fmt.Printf("go_matches_lock: %t\n", rep.GoMatchesLock)
		fmt.Printf("config_path: %s\n", rep.ConfigPath)
		fmt.Printf("lock_path: %s\n", rep.LockPath)
		fmt.Printf("has_config: %t\n", rep.HasConfig)
		fmt.Printf("has_lock: %t\n", rep.HasLock)
		fmt.Printf("lock_valid: %t\n", rep.LockValid)
		fmt.Printf("bin_dir: %s\n", rep.BinDir)
		fmt.Printf("bin_dir_exists: %t\n", rep.BinDirExists)
		fmt.Printf("bin_dir_writable: %t\n", rep.BinWritable)
		fmt.Printf("executable_path: %s\n", rep.ExecutablePath)
		fmt.Printf("executable_writable: %t\n", rep.ExecutableWritable)
		for _, e := range rep.Errors {
			if strings.TrimSpace(e) != "" {
				fmt.Printf("error: %s\n", e)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
