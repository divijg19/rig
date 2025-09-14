// internal/cli/setup.go

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install or verify pinned tools from [tools] in rig.toml",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, path, err := loadConfigOrFail()
		if err != nil {
			return err
		}
		if len(conf.Tools) == 0 {
			fmt.Printf("ℹ️  No [tools] specified in %s\n", path)
			return nil
		}
		fmt.Printf("🔧 Setting up tools from %s\n", path)
		for name, ver := range conf.Tools {
			switch name {
			case "golangci-lint":
				// Install via official go install pattern
				// Note: On Windows, this installs to GOPATH/bin or GOBIN
				module := "github.com/golangci/golangci-lint/cmd/golangci-lint@v" + strings.TrimPrefix(ver, "v")
				if err := execCommandSilent("go", "install", module); err != nil {
					return fmt.Errorf("install %s: %w", name, err)
				}
				fmt.Printf("✅ %s %s installed\n", name, ver)
			default:
				// Generic: try go install <name>@<ver> if it looks like a module path
				if strings.Contains(name, "/") {
					module := name + "@" + ver
					if err := execCommandSilent("go", "install", module); err != nil {
						return fmt.Errorf("install %s: %w", name, err)
					}
					fmt.Printf("✅ %s %s installed\n", name, ver)
				} else {
					fmt.Printf("⚠️  Unknown tool '%s' with version '%s' - skipping (add installer mapping)\n", name, ver)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
