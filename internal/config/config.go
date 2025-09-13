// internal/config/config.go

package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Define the structs that will hold our configuration.
// We use `mapstructure` tags for Viper to know how to unmarshal the TOML data.

type Project struct {
	Name    string   `mapstructure:"name"`
	Version string   `mapstructure:"version"`
	Authors []string `mapstructure:"authors"`
}

type Config struct {
	Project Project           `mapstructure:"project"`
	Tasks   map[string]string `mapstructure:"tasks"`
}

// DefaultConfigTemplate is the content that will be written to a new rig.toml file.
// Using a multiline string literal makes it clean and easy to edit.
const DefaultConfigTemplate = `
# rig.toml: The single source of truth for your Go project.
# For more information, see: https://github.com/your-org/rig

[project]
name = "%s"
version = "0.1.0"
authors = []

[tasks]
# Define your cross-platform tasks here.
# Example: rig run test
test = "go test -v -race ./..."
lint = "golangci-lint run"
run = "go run ."
`

// GetDefaultProjectName infers a project name from the current directory.
// This makes the `rig init` command feel smarter.
func GetDefaultProjectName() string {
	// Get the current working directory.
	wd, err := os.Getwd()
	if err != nil {
		// Fallback if we can't get the directory.
		return "my-go-project"
	}
	// Return the last part of the path (the directory name).
	return strings.ToLower(filepath.Base(wd))
}
