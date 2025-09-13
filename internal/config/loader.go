// internal/config/loader.go

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")
	// Environment variable overrides (optional, namespaced by PROJECT_)
	v.SetEnvPrefix("RIG")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, "", fmt.Errorf("read config %s: %w", path, err)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, "", fmt.Errorf("unmarshal config: %w", err)
	}
	if c.Tasks == nil {
		c.Tasks = map[string]string{}
	}
	return &c, path, nil
}
