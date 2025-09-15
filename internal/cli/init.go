// internal/cli/init.go

package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/divijg19/rig/internal/config"
	"github.com/spf13/cobra"
)

const configFileName = "rig.toml"

// Command-line flags for the init command
var (
	initDirectory     string
	initYes           bool
	initForce         bool
	initDeveloperMode bool
	initMinimal       bool
	initName          string
	initVersion       string
	initLicense       string
	initMonorepo      bool
	initNoTools       bool
	initNoTasks       bool
	initProfiles      []string
	initDevWatcher    string // none|reflex|air
	initCI            bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new rig.toml file",
	Long: `Creates a rig.toml manifest with sensible defaults or interactive prompts.

This file is the single source of truth for your project: define tasks, pin tools, and configure build profiles.
Supports monorepos via .rig/ includes.`,
	Example: `
  rig init            # defaults
  rig init -y         # non-interactive
  rig init -C ./app   # write manifest in subfolder
  rig init --developer --monorepo --dev-watcher reflex
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve target directory
		targetDirectory := initDirectory
		if targetDirectory == "" {
			targetDirectory = "."
		}
		if err := os.MkdirAll(targetDirectory, 0o755); err != nil {
			return fmt.Errorf("create target dir: %w", err)
		}
		configPath := filepath.Join(targetDirectory, configFileName)

		if _, err := os.Stat(configPath); err == nil && !initForce {
			return fmt.Errorf("%s already exists. Use --force to overwrite", configPath)
		}

		// Validate mutually exclusive options
		if initDeveloperMode && initMinimal {
			return errors.New("--developer and --minimal are mutually exclusive")
		}

		// Interactive mode with smart defaults
		if !initYes {
			fmt.Println("🚀 This will create a rig.toml file in the current directory.")
			fmt.Println()
		}

		// Determine project metadata from flags, prompts, or smart defaults
		projectName := initName
		if projectName == "" {
			// Smart default: directory base name
			base := filepath.Base(targetDirectory)
			if base == "." || base == string(os.PathSeparator) || base == "" {
				projectName = config.GetDefaultProjectName()
			} else {
				projectName = strings.ToLower(base)
			}
			if !initYes {
				projectName = askString(fmt.Sprintf("? project name: (%s)", projectName), projectName)
			}
		}

		version := firstNonEmpty(initVersion, "0.1.0")
		if !initYes && initVersion == "" {
			version = askString(fmt.Sprintf("? version: (%s)", version), version)
		}

		license := firstNonEmpty(initLicense, "MIT")
		if !initYes && initLicense == "" {
			license = askString(fmt.Sprintf("? license: (%s)", license), license)
		}

		// Smart default: author from git config
		var authors []string
		if gitAuthor := getGitAuthor(); gitAuthor != "" && !initYes {
			authorInput := askString(fmt.Sprintf("? author: (%s)", gitAuthor), gitAuthor)
			if authorInput != "" {
				authors = []string{authorInput}
			}
		}

		// Smart default: Go version detection
		goVersion := getGoVersion()
		if !initYes && goVersion != "" {
			add := askConfirm(fmt.Sprintf("? Go version detected: %s (add to tools?) (Y/n)", goVersion), true)
			if !add {
				// Respect user's choice to not include; set empty to skip later
				goVersion = ""
			}
		}

		if !initYes {
			fmt.Println()
		}

		developerMode := initDeveloperMode
		if !initYes && !initDeveloperMode && !initMinimal {
			// Ask which mode
			developerMode = askConfirm("Enable developer mode with advanced features? (y/N)", false)
		}
		// Profiles default
		profiles := initProfiles
		if len(profiles) == 0 {
			if developerMode {
				profiles = []string{"dev", "release"}
			} else {
				profiles = []string{"release"}
			}
		}

		monorepo := initMonorepo
		if !initYes && !initMinimal && !initDeveloperMode {
			monorepo = askConfirm("Use monorepo layout with .rig/ includes? (y/N)", false)
		}

		devWatcher := initDevWatcher
		if devWatcher == "" {
			devWatcher = "none"
			if developerMode {
				devWatcher = "reflex"
			}
		}
		if !isValidOption(devWatcher, "none", "reflex", "air") {
			return fmt.Errorf("invalid --dev-watcher: %s", devWatcher)
		}

		useTools := !initNoTools
		useTasks := !initNoTasks
		if !initYes && developerMode {
			// Quick prompts
			useTools = askConfirm("Pin common tools (golangci-lint, watcher)? (Y/n)", true)
			useTasks = askConfirm("Generate default tasks? (Y/n)", true)
		}

		// Generate configuration files
		mainToml := buildMainConfig(projectName, version, license, authors, profiles, monorepo)
		var includes []string
		var tasksToml, toolsToml string
		if monorepo {
			if useTasks {
				tasksToml = buildTasksConfig(developerMode, devWatcher, initCI)
				includes = append(includes, "rig.tasks.toml")
			}
			if useTools {
				toolsToml = buildToolsConfig(developerMode, devWatcher, goVersion)
				includes = append(includes, "rig.tools.toml")
			}
			if len(includes) > 0 {
				mainToml = injectInclude(mainToml, includes)
			}
		} else {
			// Single-file: append tasks/tools into main
			if useTasks {
				mainToml += "\n" + buildTasksConfig(developerMode, devWatcher, initCI)
			}
			if useTools {
				mainToml += "\n" + buildToolsConfig(developerMode, devWatcher, goVersion)
			}
		}

		// Write files
		if err := os.WriteFile(configPath, []byte(mainToml), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", configPath, err)
		}
		wrote := []string{getRelativePath(configPath)}
		if monorepo {
			rigDirectory := filepath.Join(targetDirectory, ".rig")
			if err := os.MkdirAll(rigDirectory, 0o755); err != nil {
				return fmt.Errorf("create .rig dir: %w", err)
			}
			if tasksToml != "" {
				p := filepath.Join(rigDirectory, "rig.tasks.toml")
				if err := os.WriteFile(p, []byte(tasksToml), 0o644); err != nil {
					return fmt.Errorf("write %s: %w", p, err)
				}
				wrote = append(wrote, getRelativePath(p))
			}
			if toolsToml != "" {
				p := filepath.Join(rigDirectory, "rig.tools.toml")
				if err := os.WriteFile(p, []byte(toolsToml), 0o644); err != nil {
					return fmt.Errorf("write %s: %w", p, err)
				}
				wrote = append(wrote, getRelativePath(p))
			}
		}

		if !initYes {
			fmt.Println()
		}
		fmt.Printf("✅ rig.toml created successfully!\n")
		fmt.Println("📋 Created:")
		for _, p := range wrote {
			fmt.Printf("  • %s\n", p)
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initDirectory, "dir", "C", "", "target directory (default current)")
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "accept defaults without prompts")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing rig.toml if present")
	initCmd.Flags().BoolVar(&initDeveloperMode, "developer", false, "developer-focused template (watchers, lint, dev profile)")
	initCmd.Flags().BoolVar(&initMinimal, "minimal", false, "minimal template (release profile only)")
	initCmd.Flags().StringVar(&initName, "name", "", "project name (defaults to directory name)")
	initCmd.Flags().StringVar(&initVersion, "version", "0.1.0", "project version")
	initCmd.Flags().StringVar(&initLicense, "license", "MIT", "project license")
	initCmd.Flags().BoolVar(&initMonorepo, "monorepo", false, "use .rig/ with includes for monorepos")
	initCmd.Flags().BoolVar(&initNoTools, "no-tools", false, "do not generate [tools] section")
	initCmd.Flags().BoolVar(&initNoTasks, "no-tasks", false, "do not generate [tasks] section")
	initCmd.Flags().StringSliceVar(&initProfiles, "profiles", nil, "profiles to create (comma-separated)")
	initCmd.Flags().StringVar(&initDevWatcher, "dev-watcher", "", "dev watcher: none|reflex|air")
	initCmd.Flags().BoolVar(&initCI, "ci", false, "add a simple CI task")

	// Add backward compatibility alias for DX flag
	initCmd.Flags().BoolVar(&initDeveloperMode, "dx", false, "alias for --developer (deprecated)")
	initCmd.Flags().MarkHidden("dx")

	rootCmd.AddCommand(initCmd)
}

// Helper functions
func askString(prompt, defaultValue string) string {
	if initYes {
		return defaultValue
	}
	fmt.Printf("%s ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue
	}
	return line
}

func askConfirm(prompt string, defaultValue bool) bool {
	if initYes {
		return defaultValue
	}
	fmt.Printf("%s ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" {
		return defaultValue
	}
	return strings.HasPrefix(line, "y")
}

func getGitAuthor() string {
	// Try to get git user name and email
	name, _ := execCommand("git", "config", "user.name")
	email, _ := execCommand("git", "config", "user.email")
	if name != "" && email != "" {
		return fmt.Sprintf("%s <%s>", name, email)
	}
	return ""
}

func getGoVersion() string {
	// Try to detect current Go version
	output, err := execCommand("go", "version")
	if err != nil {
		return ""
	}
	// Parse "go version go1.21.5 ..." to extract "1.21.5"
	parts := strings.Fields(output)
	if len(parts) >= 3 && strings.HasPrefix(parts[2], "go") {
		return strings.TrimPrefix(parts[2], "go")
	}
	return ""
}

func isValidOption(value string, validOptions ...string) bool {
	for _, option := range validOptions {
		if value == option {
			return true
		}
	}
	return false
}

func getRelativePath(absolutePath string) string {
	cwd, _ := os.Getwd()
	if relativePath, err := filepath.Rel(cwd, absolutePath); err == nil {
		return relativePath
	}
	return absolutePath
}

func buildMainConfig(name, version, license string, authors []string, profiles []string, monorepo bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[project]\nname = \"%s\"\nversion = \"%s\"\n", name, version)

	// Authors
	if len(authors) > 0 {
		b.WriteString("authors = [")
		for i, author := range authors {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "\"%s\"", author)
		}
		b.WriteString("]\n")
	} else {
		b.WriteString("authors = []\n")
	}

	fmt.Fprintf(&b, "license = \"%s\"\n\n", license)

	// Profiles
	for _, p := range profiles {
		switch p {
		case "release":
			b.WriteString("[profile.release]\nldflags = \"-s -w\"\ntags = []\ngcflags = \"\"\noutput = \"bin/app\"\n\n")
		case "dev":
			b.WriteString("[profile.dev]\nflags = [\"-race\"]\ntags = []\ngcflags = \"\"\n\n")
		default:
			// create empty block
			fmt.Fprintf(&b, "[profile.%s]\n\n", p)
		}
	}
	// In monorepo mode, tasks/tools are in includes
	if monorepo {
		b.WriteString("# include = [\"rig.tasks.toml\", \"rig.tools.toml\"]\n")
	}
	return b.String()
}

func buildTasksConfig(developerMode bool, watcher string, includeCI bool) string {
	var builder strings.Builder
	builder.WriteString("[tasks]\n")
	builder.WriteString("list = \"rig run --list\"\n")
	builder.WriteString("help = \"rig --help\"\n")
	builder.WriteString("build = \"go build ./...\"\n")
	builder.WriteString("test = \"go test ./...\"\n")
	builder.WriteString("vet = \"go vet ./...\"\n")
	builder.WriteString("fmt = \"gofmt -s -w .\"\n")
	builder.WriteString("run = \"go run .\"\n")
	if developerMode {
		builder.WriteString("lint = \"golangci-lint run ./...\"\n")
		switch watcher {
		case "reflex":
			// Prefer argv-style for robust cross-platform execution
			builder.WriteString("dev.argv = [\"reflex\", \"-r\", \\\"\\\\.go$\\\" , \"--\", \"go\", \"run\", \".\"]\n")
		case "air":
			builder.WriteString("dev = \"air\"\n")
		}
	}
	if includeCI {
		// Simple CI task: vet and test
		builder.WriteString("ci = \"go vet ./... && go test ./...\"\n")
	}
	return builder.String()
}

func buildToolsConfig(developerMode bool, watcher string, goVersion string) string {
	var builder strings.Builder
	builder.WriteString("[tools]\n")

	// Add Go version if detected
	if goVersion != "" {
		fmt.Fprintf(&builder, "go = \"%s\"\n", goVersion)
	}

	if developerMode {
		builder.WriteString("golangci-lint = \"1.62.0\"\n")
		switch watcher {
		case "reflex":
			// module path for generic go install
			builder.WriteString("github.com/cespare/reflex = \"latest\"\n")
		case "air":
			builder.WriteString("github.com/cosmtrek/air = \"latest\"\n")
		}
	}
	return builder.String()
}

func injectInclude(mainToml string, files []string) string {
	// Place include after [project] block
	var b strings.Builder
	b.WriteString(mainToml)
	b.WriteString("\ninclude = [")
	for i, f := range files {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("\"")
		b.WriteString(f)
		b.WriteString("\"")
	}
	b.WriteString("]\n")
	return b.String()
}
