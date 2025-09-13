// cmd/rig/cmd/root.go

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rig",
	Short: "rig is an all-in-one project manager and task runner for Go.",
	Long: `rig enhances the Go toolchain with a single, declarative manifest,
solving common pain points like toolchain management and script cross-compatibility.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	// Placeholder for persistent and local flags on root when needed.
	// Example:
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rig.toml)")
}
