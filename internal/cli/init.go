// internal/cli/init.go

package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/divijg19/rig/internal/config"
	"github.com/spf13/cobra"
)

const configFileName = "rig.toml"

// Command-line flags for the init command
var (
	initDirectory string
	initYes       bool
	initForce     bool
	initDev       bool
	initMinimal   bool
	initCI        bool
	initMonorepo  bool
	initName      string
	initVersion   string
	initLicense   string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new rig.toml file",
	Long: `Create a rig.toml manifest.

Default output is a minimal app template with project metadata, starter tasks, and a pinned Go toolchain.
Use --dev to add a watcher-backed dev task and reflex tool support.`,
	Example: `
  rig init
  rig init --yes
  rig init --dev --ci
  rig init --minimal
  rig init --monorepo -C ./workspace
`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if initDev && initMinimal {
			return fmt.Errorf("--dev and --minimal are mutually exclusive")
		}
		if initMinimal && initCI {
			return fmt.Errorf("--minimal and --ci are mutually exclusive")
		}

		if !initYes {
			fmt.Printf("Create rig.toml in %s\n\n", targetDirectory)
		}

		projectName := initName
		if projectName == "" {
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
		if !initYes {
			version = askString(fmt.Sprintf("? version: (%s)", version), version)
		}

		license := firstNonEmpty(initLicense, "MIT")
		if !initYes {
			license = askString(fmt.Sprintf("? license: (%s)", license), license)
		}

		goVersion := getGoVersion()
		if goVersion == "" {
			goVersion = strings.TrimPrefix(runtime.Version(), "go")
		}

		mainToml := buildMainConfig(projectName, version, license)
		var includes []string
		var tasksToml, toolsToml string
		includeTasks := !initMinimal
		if initMonorepo {
			if includeTasks {
				tasksToml = buildTasksConfig(initDev, initCI)
				includes = append(includes, "rig.tasks.toml")
			}
			toolsToml = buildToolsConfig(goVersion, initDev)
			includes = append(includes, "rig.tools.toml")
			if len(includes) > 0 {
				mainToml = injectInclude(mainToml, includes)
			}
		} else {
			if includeTasks {
				mainToml += "\n" + buildTasksConfig(initDev, initCI)
			}
			mainToml += "\n" + buildToolsConfig(goVersion, initDev)
		}

		// Write files
		if err := os.WriteFile(configPath, []byte(mainToml), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", configPath, err)
		}
		wrote := []string{getRelativePath(configPath)}
		if initMonorepo {
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

		if err := ensureRigIgnored(targetDirectory); err != nil {
			return err
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
	initCmd.Flags().StringVarP(&initDirectory, "dir", "C", "", "Write manifest in <dir> (default cwd)")
	initCmd.Flags().BoolVar(&initDev, "dev", false, "Include dev watcher + reflex tool")
	initCmd.Flags().BoolVar(&initMinimal, "minimal", false, "Minimal manifest (no tasks, no profiles)")
	initCmd.Flags().BoolVar(&initCI, "ci", false, "Add a simple CI task")
	initCmd.Flags().BoolVar(&initMonorepo, "monorepo", false, "Monorepo layout (--includes)")
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing rig.toml")
	initCmd.Flags().StringVar(&initName, "name", "", "Project name (default: directory)")
	initCmd.Flags().StringVar(&initLicense, "license", "MIT", "Project license")
	initCmd.Flags().StringVar(&initVersion, "version", "0.1.0", "Project version")
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Accept defaults (non-interactive)")

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

func getRelativePath(absolutePath string) string {
	cwd, _ := os.Getwd()
	if relativePath, err := filepath.Rel(cwd, absolutePath); err == nil {
		return relativePath
	}
	return absolutePath
}

func buildMainConfig(name, version, license string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[project]\nname = \"%s\"\nversion = \"%s\"\nlicense = \"%s\"\n", name, version, license)
	return b.String()
}

func buildTasksConfig(includeDev bool, includeCI bool) string {
	var builder strings.Builder
	builder.WriteString("[tasks]\n")
	builder.WriteString("build = \"go build ./...\"\n")
	builder.WriteString("test = \"go test ./...\"\n")
	builder.WriteString("run = \"go run .\"\n")
	if includeCI {
		builder.WriteString("\n[tasks.ci]\n")
		builder.WriteString("command = \"rig check && rig run test\"\n")
	}
	if includeDev {
		builder.WriteString("\n[tasks.dev]\n")
		builder.WriteString("command = \"go run .\"\n")
		builder.WriteString("watch = [\"**/*.go\"]\n")
	}
	return builder.String()
}

func buildToolsConfig(goVersion string, includeDev bool) string {
	var builder strings.Builder
	builder.WriteString("[tools]\n")
	fmt.Fprintf(&builder, "go = \"%s\"\n", goVersion)
	if includeDev {
		builder.WriteString("reflex = \"latest\"\n")
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

func ensureRigIgnored(targetDirectory string) error {
	ignorePath := filepath.Join(targetDirectory, ".gitignore")
	b, err := os.ReadFile(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(ignorePath, []byte(".rig/\n"), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", ignorePath, err)
			}
			return nil
		}
		return fmt.Errorf("read %s: %w", ignorePath, err)
	}

	content := string(b)
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == ".rig/" {
			return nil
		}
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += ".rig/\n"
	if err := os.WriteFile(ignorePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", ignorePath, err)
	}
	return nil
}
