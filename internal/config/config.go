// internal/config/config.go

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Define the structs that will hold our configuration.
// We use `mapstructure` tags for Viper to know how to unmarshal the TOML data.

type Project struct {
	Name    string   `mapstructure:"name" toml:"name"`
	Version string   `mapstructure:"version" toml:"version"`
	Authors []string `mapstructure:"authors" toml:"authors"`
	License string   `mapstructure:"license" toml:"license"`
}

// Task represents either a simple command string or a structured task configuration
type Task struct {
	Command     string            `mapstructure:"command" toml:"command,omitempty"`
	Argv        []string          `mapstructure:"argv" toml:"argv,omitempty"`
	Shell       string            `mapstructure:"shell" toml:"shell,omitempty"`
	Description string            `mapstructure:"description" toml:"description,omitempty"`
	Watch       []string          `mapstructure:"watch" toml:"watch,omitempty"`
	Env         map[string]string `mapstructure:"env" toml:"env,omitempty"`
	Cwd         string            `mapstructure:"cwd" toml:"cwd,omitempty"`
	DependsOn   []string          `mapstructure:"depends_on" toml:"depends_on,omitempty"`
}

// UnmarshalTOML allows Task to be decoded from either a string (command) or a table.
// Compatible with github.com/pelletier/go-toml/v2 where value is one of: string | map[string]any
func (t *Task) UnmarshalTOML(v any) error {
	return t.fromAny(v)
}

// fromAny decodes a Task from arbitrary TOML value.
func (t *Task) fromAny(v any) error {
	switch val := v.(type) {
	case string:
		t.Command = val
		return nil
	case map[string]any:
		// argv (takes precedence if present)
		if arr, ok := val["argv"].([]any); ok {
			argv, err := toStringSlice(arr)
			if err != nil {
				return fmt.Errorf("argv: %w", err)
			}
			t.Argv = argv
		}
		// command can be string or array (ignored if argv already set)
		if len(t.Argv) == 0 {
			if cmd, ok := val["command"].(string); ok {
				t.Command = cmd
			} else if arr, ok := val["command"].([]any); ok {
				argv, err := toStringSlice(arr)
				if err != nil {
					return fmt.Errorf("command array: %w", err)
				}
				t.Argv = argv
			}
		}
		// description
		if desc, ok := val["description"].(string); ok {
			t.Description = desc
		}
		// shell
		if sh, ok := val["shell"].(string); ok {
			t.Shell = sh
		}
		// env
		if envRaw, ok := val["env"].(map[string]any); ok {
			if t.Env == nil {
				t.Env = make(map[string]string, len(envRaw))
			}
			for k, v := range envRaw {
				if s, ok := v.(string); ok {
					t.Env[k] = s
				} else {
					return fmt.Errorf("env %q must be a string, got %T", k, v)
				}
			}
		}
		// watch
		if watchRaw, ok := val["watch"].([]any); ok {
			watch, err := toStringSlice(watchRaw)
			if err != nil {
				return fmt.Errorf("watch: %w", err)
			}
			t.Watch = watch
		}
		// cwd
		if cwd, ok := val["cwd"].(string); ok {
			t.Cwd = cwd
		}
		// args (optional extra args)
		if argsRaw, ok := val["args"].([]any); ok {
			args, err := toStringSlice(argsRaw)
			if err != nil {
				return fmt.Errorf("args: %w", err)
			}
			if len(t.Argv) > 0 {
				t.Argv = append(t.Argv, args...)
			} else if t.Command != "" {
				t.Argv = append([]string{t.Command}, args...)
				t.Command = ""
			} else {
				// args provided but no command/argv base; treat as error
				return fmt.Errorf("args provided without a base command")
			}
		}
		// depends_on
		if depsRaw, ok := val["depends_on"].([]any); ok {
			for _, d := range depsRaw {
				if s, ok := d.(string); ok {
					t.DependsOn = append(t.DependsOn, s)
				} else {
					return fmt.Errorf("depends_on items must be strings, got %T", d)
				}
			}
		}
		return nil
	case nil:
		// treat as empty
		return nil
	default:
		return fmt.Errorf("task must be string or table, got %T", v)
	}
}

// TasksMap is a custom type to allow decoding [tasks] where values can be strings or tables.
type TasksMap map[string]Task

// UnmarshalTOML implements custom decoding for tasks table.
func (m *TasksMap) UnmarshalTOML(v any) error {
	tbl, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("tasks must be a table, got %T", v)
	}
	out := make(TasksMap, len(tbl))
	for name, raw := range tbl {
		var t Task
		if err := t.fromAny(raw); err != nil {
			return fmt.Errorf("task %q: %w", name, err)
		}
		out[name] = t
	}
	*m = out
	return nil
}

// toStringSlice converts a []any to []string with validation.
func toStringSlice(v []any) ([]string, error) {
	out := make([]string, 0, len(v))
	for _, it := range v {
		s, ok := it.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", it)
		}
		out = append(out, s)
	}
	return out, nil
}

type Config struct {
	Project Project           `mapstructure:"project" toml:"project"`
	Tasks   TasksMap          `mapstructure:"tasks" toml:"tasks"`
	Tools   map[string]string `mapstructure:"tools" toml:"tools"`
	// Include allows splitting configuration across files.
	// Paths are resolved relative to the main rig.toml directory. For monorepos,
	// paths under .rig/ are also attempted if not found alongside the main file.
	Includes []string `mapstructure:"include" toml:"include"`
	// Profile-specific build settings (e.g., [profile.release])
	Profiles map[string]BuildProfile `mapstructure:"profile" toml:"profile"`
}

// BuildProfile captures optional build-time configuration that can be
// selected via `rig build --profile <name>`.
type BuildProfile struct {
	// Go build flags
	Ldflags string   `mapstructure:"ldflags" toml:"ldflags"`
	Gcflags string   `mapstructure:"gcflags" toml:"gcflags"`
	Tags    []string `mapstructure:"tags" toml:"tags"`
	Flags   []string `mapstructure:"flags" toml:"flags"`

	// Optional environment to apply during build (KEY=VALUE)
	Env map[string]string `mapstructure:"env" toml:"env"`

	// Optional default output path/name (overridden by --output)
	Output string `mapstructure:"output" toml:"output"`
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
license = "MIT"

[tasks]
# Define your cross-platform tasks here.
# Example: rig run test
test = "go test -v -race ./..."
lint = "golangci-lint run"
run = "go run ."

[tools]
# Pin tool versions for reproducible CI/dev
# go = "1.25.1"
# golangci-lint = "1.62.0"

# include = ["rig.tasks.toml", "rig.tools.toml"]

# Optional build profiles for \"rig build --profile <name>\"
[profile.release]
# Strip debug, smaller binary
ldflags = "-s -w"
tags = []
gcflags = ""
output = "bin/app"
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
