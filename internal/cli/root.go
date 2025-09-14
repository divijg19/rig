// internal/cli/root.go

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	cfg "github.com/divijg19/rig/internal/config"
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

// Common error messages for consistency
const msgNoConfig = "no rig.toml found. run 'rig init' first"

// loadConfigOrFail loads the config and returns a standardized error if not found.
// This eliminates duplicate error handling across CLI commands.
func loadConfigOrFail() (*cfg.Config, string, error) {
	conf, path, err := cfg.Load("")
	if err != nil {
		if errors.Is(err, cfg.ErrConfigNotFound) {
			return nil, "", errors.New(msgNoConfig)
		}
		return nil, "", err
	}
	return conf, path, nil
}

// loadConfigOptional loads the config but allows ErrConfigNotFound to be handled by caller.
// Used by commands like doctor that can work without a config file.
func loadConfigOptional() (*cfg.Config, string, error) {
	return cfg.Load("")
}

// firstNonEmpty returns a if a != "", otherwise b.
// This utility function eliminates duplicate implementations across CLI commands.
func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return strings.TrimSpace(b)
}

// execCommand captures command output and errors for diagnostic purposes.
// Used by commands that need to check tool availability or versions.
func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// execCommandSilent runs a command without capturing output, for installation tasks.
// Returns only the error status for success/failure determination.
func execCommandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
