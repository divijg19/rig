// internal/cli/doctor.go

package cli

import (
	"errors"
	"fmt"

	cfg "github.com/divijg19/rig/internal/config"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check your development environment and tooling",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🩺 Rig Doctor")

		// Go version
		goVer, err := execCommand("go", "version")
		if err != nil {
			fmt.Println("❌ go toolchain not found in PATH")
		} else {
			fmt.Printf("✅ %s\n", goVer)
		}

		conf, _, err := loadConfigOptional()
		if err != nil && !errors.Is(err, cfg.ErrConfigNotFound) {
			return err
		}
		if conf != nil && len(conf.Tools) > 0 {
			fmt.Println("🔍 Checking tools:")
			for name := range conf.Tools {
				if _, err := execCommand(name, "--version"); err != nil {
					fmt.Printf("  ❌ %s not found\n", name)
				} else {
					fmt.Printf("  ✅ %s present\n", name)
				}
			}
		}

		fmt.Println("✔️  Doctor check complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
