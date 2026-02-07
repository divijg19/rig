package rig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/shlex"
)

func parseCommand(command string) ([]string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil, errors.New("command must be non-empty")
	}
	argv, err := shlex.Split(command)
	if err != nil {
		return nil, fmt.Errorf("parse command: %w", err)
	}
	if len(argv) == 0 {
		return nil, errors.New("command must not be empty")
	}
	return argv, nil
}

func resolveCwd(configPath string, taskCwd string) (string, error) {
	baseDir := filepath.Dir(configPath)
	if strings.TrimSpace(taskCwd) == "" {
		return filepath.Abs(baseDir)
	}
	cwd := taskCwd
	if !filepath.IsAbs(cwd) {
		cwd = filepath.Join(baseDir, cwd)
	}
	return filepath.Abs(cwd)
}

func resolveExecutable(cmd string, cwd string, env []string) (string, error) {
	if cmd == "" {
		return "", errors.New("empty executable")
	}
	if filepath.IsAbs(cmd) {
		return cmd, nil
	}
	if strings.ContainsRune(cmd, os.PathSeparator) {
		p := cmd
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, p)
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		if err := ensureExecutable(abs); err != nil {
			return "", err
		}
		return abs, nil
	}

	pathVal := ""
	for _, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			pathVal = strings.TrimPrefix(kv, "PATH=")
			break
		}
	}
	if pathVal == "" {
		pathVal = os.Getenv("PATH")
	}

	dirs := strings.Split(pathVal, string(os.PathListSeparator))
	candidates := []string{cmd}
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(cmd), ".exe") {
		candidates = append([]string{cmd + ".exe"}, candidates...)
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		for _, c := range candidates {
			p := filepath.Join(dir, c)
			abs, err := filepath.Abs(p)
			if err != nil {
				continue
			}
			if ensureExecutable(abs) == nil {
				return abs, nil
			}
		}
	}

	return "", fmt.Errorf("executable %q not found on PATH", cmd)
}

func ensureExecutable(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	if st.Mode()&0o111 == 0 {
		return fmt.Errorf("%s is not executable", path)
	}
	return nil
}
