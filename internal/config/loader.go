// internal/config/loader.go

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

var ErrConfigNotFound = errors.New("rig.toml not found; run 'rig init' to create one")

// LocateConfig searches from the provided start directory upward for a rig.toml file.
// Returns the absolute path to the first match or ErrConfigNotFound.
func LocateConfig(start string) (string, error) {
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
		start = wd
	}
	start, _ = filepath.Abs(start)

	dir := start
	for {
		candidate := filepath.Join(dir, "rig.toml")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached root
			break
		}
		dir = parent
	}
	return "", ErrConfigNotFound
}

// Load reads rig.toml (starting from startDir upwards) into a Config struct.
// Returns the config and the path that was loaded.
func Load(startDir string) (*Config, string, error) {
	path, err := LocateConfig(startDir)
	if err != nil {
		return nil, "", err
	}

	// Read and decode the base config
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read config %s: %w", path, err)
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return nil, "", fmt.Errorf("unmarshal base config: %w", err)
	}

	// Resolve include paths relative to the base file (and support .rig/ fallbacks)
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
				continue // skip missing include
			}
		}
		incData, err := os.ReadFile(incPath)
		if err != nil {
			return nil, "", fmt.Errorf("read include %s: %w", incPath, err)
		}
		var inc Config
		if err := toml.Unmarshal(incData, &inc); err != nil {
			return nil, "", fmt.Errorf("unmarshal include %s: %w", incPath, err)
		}
		if inc.Tasks != nil {
			if c.Tasks == nil {
				c.Tasks = map[string]string{}
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
				c.Profiles = map[string]BuildProfile{}
			}
			for k, v := range inc.Profiles {
				c.Profiles[k] = v
			}
		}
	}
	if c.Tasks == nil {
		c.Tasks = map[string]string{}
	}
	return &c, path, nil
}

// parseIncludeList extracts a top-level include array as []string from TOML bytes.
func parseIncludeList(b []byte) []string {
	// Simple, lenient single-line parser: include = ["a.toml", "b.toml"]
	// For multi-line arrays, extend as needed.
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
