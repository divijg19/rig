package rig

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func buildEnv(configPath string, taskEnv map[string]string) []string {
	base := map[string]string{}
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		base[k] = v
	}

	localBin := localBinDirForConfig(configPath)

	basePath := base["PATH"]
	parts := []string{}
	if basePath != "" {
		parts = strings.Split(basePath, string(os.PathListSeparator))
	}

	seen := map[string]struct{}{}
	dedup := make([]string, 0, len(parts)+1)
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
		key := p
		if runtime.GOOS == "windows" {
			key = strings.ToLower(filepath.Clean(p))
		} else {
			key = filepath.Clean(p)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		dedup = append(dedup, p)
	}
	base["PATH"] = strings.Join(dedup, string(os.PathListSeparator))

	for k, v := range taskEnv {
		base[k] = v
	}

	keys := make([]string, 0, len(base))
	for k := range base {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	env := make([]string, 0, len(keys))
	for _, k := range keys {
		env = append(env, k+"="+base[k])
	}
	return env
}
