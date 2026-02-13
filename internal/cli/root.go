// internal/cli/root.go

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	cfg "github.com/divijg19/rig/internal/config"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = ""
	date    = ""

	rootShowVersion bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rig",
	Short: "All-in-one project manager and task runner for Go",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if rootShowVersion {
			printVersion(cmd.OutOrStdout())
			return nil
		}
		return cmd.Help()
	},
	Long: `rig enhances the Go toolchain with a single, declarative manifest (rig.toml).

Command aliases are logical and resolved by invocation name (argv[0]):
	• rir → rig run
	• ric → rig check
	• rid → rig dev
	• ris → rig start (stub)

To use them as standalone commands, symlink or rename the rig binary.
See: rig alias

Features:
	• Interactive Setup: 'rig init' with smart defaults (git config, Go version detection)
	• Unified Manifest: [project], [tasks], [tools], [profile.*], and includes
	• Strict Tasks: task values are either strings or tables with {command, description, env, cwd, depends_on}
	• Reproducible Tooling: 'rig sync' writes rig.lock and installs tools into .rig/bin
	• Daily Dev Loop: 'rig dev' runs a watcher-backed dev task (requires rig.lock)
	• Lock-backed Execution: 'rig run' and 'rig check' require rig.lock and enforce tool parity

Tip: run 'rig init' to create an interactive rig.toml, then 'rig tools sync' to install tools.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion(cmd.OutOrStdout())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// ExecuteWithArgs runs the CLI with an explicit argv (excluding argv[0]).
// This is used by wrapper binaries that forward to a specific subcommand.
func ExecuteWithArgs(args []string) {
	rootCmd.SetArgs(args)
	Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&rootShowVersion, "version", "v", false, "print version information")

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		printVersion(out)
		fmt.Fprintln(out)
		defaultHelp(cmd, args)
		fmt.Fprintln(out, "Run \"rig version\" for build information.")
	})

	rootCmd.AddCommand(versionCmd)
}

func printVersion(w io.Writer) {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}
	c := strings.TrimSpace(commit)
	if c == "" {
		c = "unknown"
	}
	d := strings.TrimSpace(date)
	if d == "" {
		d = "unknown"
	}
	goVer := strings.TrimSpace(runtime.Version())
	if goVer == "" {
		goVer = "unknown"
	}
	fmt.Fprintf(w, "rig %s\n", v)
	fmt.Fprintf(w, "commit: %s\n", c)
	fmt.Fprintf(w, "built: %s\n", d)
	fmt.Fprintf(w, "go: %s\n", goVer)
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
