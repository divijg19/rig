package rig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cfg "github.com/divijg19/rig/internal/config"
	toml "github.com/pelletier/go-toml/v2"
)

// LoadConfig loads rig.toml like config.Load, but enforces the strict task schema:
//
// - [tasks].<name> is either a string, or a table
// - task tables may only contain: command, description, env, cwd, depends_on
// - [tasks.dev] additionally supports: watch
// - no other task fields are permitted
func LoadConfig(startDir string) (*cfg.Config, string, error) {
	path, err := cfg.LocateConfig(startDir)
	if err != nil {
		return nil, "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read config %s: %w", path, err)
	}

	base, err := parseConfigBytes(data)
	if err != nil {
		return nil, "", fmt.Errorf("unmarshal base config: %w", err)
	}

	c := base
	baseDir := filepath.Dir(path)

	includes := c.Includes
	if len(includes) == 0 {
		includes = append(includes, parseIncludeList(data)...)
	}

	for _, rel := range includes {
		incPath := rel
		if !filepath.IsAbs(incPath) {
			incPath = filepath.Join(baseDir, rel)
		}
		if _, err := os.Stat(incPath); err != nil {
			alt := filepath.Join(baseDir, ".rig", rel)
			if _, err2 := os.Stat(alt); err2 == nil {
				incPath = alt
			} else {
				continue
			}
		}
		incData, err := os.ReadFile(incPath)
		if err != nil {
			return nil, "", fmt.Errorf("read include %s: %w", incPath, err)
		}
		inc, err := parseConfigBytes(incData)
		if err != nil {
			return nil, "", fmt.Errorf("unmarshal include %s: %w", incPath, err)
		}

		if inc.Tasks != nil {
			if c.Tasks == nil {
				c.Tasks = cfg.TasksMap{}
			}
			for k, v := range inc.Tasks {
				c.Tasks[k] = v
			}
		}
		if inc.Tools != nil {
			if c.Tools == nil {
				c.Tools = map[string]string{}
			}
			for k, v := range inc.Tools {
				c.Tools[k] = v
			}
		}
		if inc.Profiles != nil {
			if c.Profiles == nil {
				c.Profiles = map[string]cfg.BuildProfile{}
			}
			for k, v := range inc.Profiles {
				c.Profiles[k] = v
			}
		}
	}

	if c.Tasks == nil {
		c.Tasks = cfg.TasksMap{}
	}
	return &c, path, nil
}

type rawConfig struct {
	Project  cfg.Project                 `toml:"project"`
	Tasks    map[string]any              `toml:"tasks"`
	Tools    map[string]string           `toml:"tools"`
	Includes []string                    `toml:"include"`
	Profiles map[string]cfg.BuildProfile `toml:"profile"`
}

func parseConfigBytes(b []byte) (cfg.Config, error) {
	var raw rawConfig
	if err := toml.Unmarshal(b, &raw); err != nil {
		return cfg.Config{}, err
	}
	c := cfg.Config{
		Project:  raw.Project,
		Tools:    raw.Tools,
		Includes: raw.Includes,
		Profiles: raw.Profiles,
	}
	if len(raw.Tasks) > 0 {
		tasks, err := parseTasks(raw.Tasks)
		if err != nil {
			return cfg.Config{}, err
		}
		c.Tasks = tasks
	}
	if c.Tasks == nil {
		c.Tasks = cfg.TasksMap{}
	}
	return c, nil
}

func parseTasks(raw map[string]any) (cfg.TasksMap, error) {
	out := make(cfg.TasksMap, len(raw))
	for name, v := range raw {
		t, err := parseTask(name, v)
		if err != nil {
			return nil, fmt.Errorf("task %q: %w", name, err)
		}
		out[name] = t
	}
	return out, nil
}

func parseTask(name string, v any) (cfg.Task, error) {
	switch val := v.(type) {
	case string:
		cmd := strings.TrimSpace(val)
		if cmd == "" {
			return cfg.Task{}, errors.New("command must be non-empty")
		}
		return cfg.Task{Command: cmd}, nil
	case map[string]any:
		// v0.3: [tasks.dev] is a strict schema: only { command, watch }.
		// We intentionally defer "non-empty" validation to the dev runtime so
		// that dev UX error strings remain stable.
		if name == "dev" {
			allowed := map[string]struct{}{
				"command": {},
				"watch":   {},
			}
			for k := range val {
				if _, ok := allowed[k]; !ok {
					return cfg.Task{}, fmt.Errorf("unsupported field %q (allowed: command, watch)", k)
				}
			}

			cmd := ""
			if cmdRaw, ok := val["command"]; ok {
				s, ok := cmdRaw.(string)
				if !ok {
					return cfg.Task{}, fmt.Errorf("command must be a string, got %T", cmdRaw)
				}
				cmd = strings.TrimSpace(s)
			}

			var watch []string
			if watchRaw, ok := val["watch"]; ok {
				arr, ok := watchRaw.([]any)
				if !ok {
					return cfg.Task{}, fmt.Errorf("watch must be an array of strings, got %T", watchRaw)
				}
				watch = make([]string, 0, len(arr))
				for _, it := range arr {
					s, ok := it.(string)
					if !ok {
						return cfg.Task{}, fmt.Errorf("watch items must be strings, got %T", it)
					}
					watch = append(watch, strings.TrimSpace(s))
				}
			}

			return cfg.Task{Command: cmd, Watch: watch}, nil
		}

		allowed := map[string]struct{}{
			"command":     {},
			"description": {},
			"env":         {},
			"cwd":         {},
			"depends_on":  {},
		}
		for k := range val {
			if _, ok := allowed[k]; !ok {
				return cfg.Task{}, fmt.Errorf("unsupported field %q (allowed: command, description, env, cwd, depends_on)", k)
			}
		}

		cmdRaw, ok := val["command"]
		if !ok {
			return cfg.Task{}, errors.New("missing required field \"command\"")
		}
		cmd, ok := cmdRaw.(string)
		if !ok {
			return cfg.Task{}, fmt.Errorf("command must be a string, got %T", cmdRaw)
		}
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			return cfg.Task{}, errors.New("command must be non-empty")
		}

		desc := ""
		if descRaw, ok := val["description"]; ok {
			s, ok := descRaw.(string)
			if !ok {
				return cfg.Task{}, fmt.Errorf("description must be a string, got %T", descRaw)
			}
			desc = strings.TrimSpace(s)
		}

		var env map[string]string
		if envRaw, ok := val["env"]; ok {
			tbl, ok := envRaw.(map[string]any)
			if !ok {
				return cfg.Task{}, fmt.Errorf("env must be a table, got %T", envRaw)
			}
			env = make(map[string]string, len(tbl))
			for k, v := range tbl {
				s, ok := v.(string)
				if !ok {
					return cfg.Task{}, fmt.Errorf("env %q must be a string, got %T", k, v)
				}
				env[k] = s
			}
		}

		cwd := ""
		if cwdRaw, ok := val["cwd"]; ok {
			s, ok := cwdRaw.(string)
			if !ok {
				return cfg.Task{}, fmt.Errorf("cwd must be a string, got %T", cwdRaw)
			}
			cwd = strings.TrimSpace(s)
		}

		depsRaw, hasDeps := val["depends_on"], false
		if _, ok := val["depends_on"]; ok {
			hasDeps = true
		}
		var deps []string
		if hasDeps {
			arr, ok := depsRaw.([]any)
			if !ok {
				return cfg.Task{}, fmt.Errorf("depends_on must be an array of strings, got %T", depsRaw)
			}
			for _, it := range arr {
				s, ok := it.(string)
				if !ok {
					return cfg.Task{}, fmt.Errorf("depends_on items must be strings, got %T", it)
				}
				deps = append(deps, s)
			}
		}

		return cfg.Task{Command: cmd, Description: desc, Env: env, Cwd: cwd, DependsOn: deps}, nil
	default:
		return cfg.Task{}, fmt.Errorf("task must be string or table, got %T", v)
	}
}

func parseIncludeList(b []byte) []string {
	s := string(b)
	re := regexp.MustCompile(`(?m)^\s*include\s*=\s*\[([^\]]*)\]`)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return nil
	}
	inner := m[1]
	parts := strings.Split(inner, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
