// internal/cli/init.go

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/divijg19/rig/internal/config"
	"github.com/spf13/cobra"
)

const configFileName = "rig.toml"

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new rig.toml file in the current directory",
	Long: `Creates a rig.toml file with default values.

This file serves as the central manifest for your project, allowing you to
define tasks, manage tools, and configure your build process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := filepath.Join(".", configFileName)
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("✅ A %s file already exists in this directory.\n", configFileName)
			return nil
		}

		projectName := config.GetDefaultProjectName()
		configFileContent := fmt.Sprintf(config.DefaultConfigTemplate, projectName)
		if err := os.WriteFile(configPath, []byte(configFileContent), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", configFileName, err)
		}
		fmt.Printf("🚀 Created a new %s file.\n", configFileName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
