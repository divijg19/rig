// internal/cli/doctor.go

package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	cfg "github.com/divijg19/rig/internal/config"
	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check your development environment and tooling",
	Long:  "Verifies the Go toolchain and (if present) tools pinned in rig.toml. Rig-managed tools are checked exclusively from .rig/bin.",
	Example: `
	rig doctor
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🩺 Rig Doctor")

		// Go version
		goVer, err := execCommand("go", "version")
		if err != nil {
			fmt.Println("❌ go toolchain not found in PATH")
		} else {
			fmt.Printf("✅ %s\n", goVer)
		}

		conf, path, err := loadConfigOptional()
		if err != nil && !errors.Is(err, cfg.ErrConfigNotFound) {
			return err
		}
		if conf != nil && len(conf.Tools) > 0 {
			fmt.Println("🔍 Checking tools:")
			lockPath := filepath.Join(filepath.Dir(path), "rig.lock")
			lock, lerr := core.ReadLockfile(lockPath)
			if lerr != nil {
				fmt.Printf("  ❌ rig.lock missing or unreadable (%s): %v\n", lockPath, lerr)
			} else {
				rows, missing, mismatched, _, cerr := core.CheckInstalledTools(conf.Tools, lock, path)
				if cerr != nil {
					fmt.Printf("  ❌ tooling check failed: %v\n", cerr)
				} else {
					for _, r := range rows {
						switch r.Status {
						case "missing":
							fmt.Printf("  ❌ %s missing (run 'rig tools sync')\n", r.Bin)
						case "mismatch":
							fmt.Printf("  ❌ %s integrity mismatch (run 'rig tools sync')\n", r.Bin)
						default:
							fmt.Printf("  ✅ %s\n", r.Bin)
						}
					}
					_ = missing
					_ = mismatched
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
