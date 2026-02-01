// internal/cli/root.go

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	cfg "github.com/divijg19/rig/internal/config"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rig",
	Short: "All-in-one project manager and task runner for Go",
	Long: `rig enhances the Go toolchain with a single, declarative manifest (rig.toml).

Features:
	• Interactive Setup: 'rig init' with smart defaults (git config, Go version detection)
	• Unified Manifest: [project], [tasks], [tools], [profile.*], and includes
	• Structured Tasks: Support simple strings or advanced tables with env vars and dependencies
	• Explicit Tool Management: 'rig tools sync' with manifest.lock for fast verification
	• Reproducible Tooling: pins and installs tools into .rig/bin per project
	• Friendly DX: emoji output, clear errors, and cross-platform commands

Tip: run 'rig init' to create an interactive rig.toml, then 'rig tools sync' to install tools.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// lsCmd offers a short alias for listing tasks (equivalent to `rig run --list`).
var (
	lsJSON bool
)

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List tasks",
	Long:    "List tasks defined in rig.toml. Shortcuts: 'rig ls', 'rig run --list'.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delegate to runCmd behavior but force list mode
		runList = true
		runListJSON = lsJSON
		return runCmd.RunE(runCmd, nil)
	},
}

func init() {
	lsCmd.Flags().BoolVarP(&lsJSON, "json", "j", false, "print machine-readable JSON")
	rootCmd.AddCommand(lsCmd)
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

// localBinDirFor returns the project-local tool bin directory based on rig.toml path.
func localBinDirFor(configPath string) string {
	base := filepath.Dir(configPath)
	return filepath.Join(base, ".rig", "bin")
}

// envWithLocalBin returns env entries that ensure the local .rig/bin is preferred on PATH.
// If includeGOBIN is true, also sets GOBIN to .rig/bin so `go install` writes there.
// Any extra entries provided are preserved and PATH/GOBIN are appended last to win on duplicates.
func envWithLocalBin(configPath string, extra []string, includeGOBIN bool) []string {
	localBin := localBinDirFor(configPath)
	// Determine base PATH from extra env if present; otherwise use process PATH
	basePath := os.Getenv("PATH")
	if len(extra) > 0 {
		for _, kv := range extra {
			if strings.HasPrefix(kv, "PATH=") {
				basePath = strings.TrimPrefix(kv, "PATH=")
				break
			}
		}
	}
	// Build PATH with localBin first, removing duplicates
	// Split current PATH and filter duplicates (case-insensitive on Windows)
	parts := []string{}
	if basePath != "" {
		parts = strings.Split(basePath, string(os.PathListSeparator))
	}
	dedup := make([]string, 0, len(parts)+1)
	seen := make(map[string]struct{}, len(parts)+1)
	// Always put localBin first
	keyLocal := localBin
	if runtime.GOOS == "windows" {
		keyLocal = strings.ToLower(filepath.Clean(localBin))
	} else {
		keyLocal = filepath.Clean(localBin)
	}
	seen[keyLocal] = struct{}{}
	dedup = append(dedup, localBin)
	for _, p := range parts {
		if p == "" {
			continue
		}
		// Normalize key
		key := p
		if runtime.GOOS == "windows" {
			key = strings.ToLower(filepath.Clean(p))
		} else {
			key = filepath.Clean(p)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		// Avoid equality check via equals; we've normalized to key
		seen[key] = struct{}{}
		dedup = append(dedup, p)
	}
	newPath := strings.Join(dedup, string(os.PathListSeparator))
	env := make([]string, 0, len(extra)+2)
	if len(extra) > 0 {
		env = append(env, extra...)
	}
	// Append GOBIN first or PATH first? Order doesn't matter between them, but both should be last overall.
	if includeGOBIN {
		env = append(env, "GOBIN="+localBin)
	}
	env = append(env, "PATH="+newPath)
	return env
}

// max returns the larger of a and b.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// execCommandEnv runs a command with additional env entries.
func execCommandEnv(name string, args []string, env []string) (string, error) {
	cmd := exec.Command(name, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// execCommandSilentEnv runs a command with env and no output capture.
func execCommandSilentEnv(name string, args []string, env []string) error {
	cmd := exec.Command(name, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
