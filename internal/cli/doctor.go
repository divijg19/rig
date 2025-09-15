// internal/cli/doctor.go

package cli

import (
	"errors"
	"fmt"

	cfg "github.com/divijg19/rig/internal/config"
	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check your development environment and tooling",
	Long:  "Verifies the Go toolchain and (if present) tools pinned in rig.toml using the local .rig/bin PATH.",
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
			env := envWithLocalBin(path, nil, false)
			for name, ver := range conf.Tools {
				_, bin := core.ResolveModuleAndBin(name)
				want := core.NormalizeSemver(core.EnsureSemverPrefixV(ver))
				out, err := execCommandEnv(bin, []string{"--version"}, env)
				if err != nil {
					fmt.Printf("  ❌ %s not found (want %s)\n", bin, ver)
					continue
				}
				have := core.ParseVersionFromOutput(out)
				if have == "" || want == "" {
					fmt.Printf("  ✅ %s present\n", bin)
					continue
				}
				if have != want {
					fmt.Printf("  ❌ %s version mismatch (have %s, want %s)\n", bin, have, want)
				} else {
					fmt.Printf("  ✅ %s %s\n", bin, ver)
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
