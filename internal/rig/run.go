package rig

import (
	"errors"
	"fmt"

	cfg "github.com/divijg19/rig/internal/config"
)

func Run(startDir string, taskName string, passthrough []string) error {
	conf, confPath, err := LoadConfig(startDir)
	if err != nil {
		return err
	}

	lock, err := ReadRigLockForConfig(confPath)
	if err != nil {
		return fmt.Errorf("rig.lock required: %w", err)
	}

	rows, missing, mismatched, extras, err := CheckInstalledTools(conf.Tools, lock, confPath)
	if err != nil {
		return err
	}
	if missing > 0 || mismatched > 0 {
		return fmt.Errorf("tools are out of sync with rig.lock (missing=%d mismatched=%d extras=%d)", missing, mismatched, len(extras))
	}
	_ = rows // reserved for future diagnostics

	if goRow, ok := checkGoAgainstLockIfRequired(conf.Tools, lock, confPath); !ok {
		if goRow != nil {
			if goRow.Error != "" {
				return fmt.Errorf("go toolchain check failed (%s): %s", goRow.Status, goRow.Error)
			}
			return fmt.Errorf("go toolchain check failed (%s): have %q, want %q", goRow.Status, goRow.Have, goRow.Locked)
		}
		return fmt.Errorf("go toolchain check failed")
	}

	task, ok := conf.Tasks[taskName]
	if !ok {
		return fmt.Errorf("task %q not found", taskName)
	}
	if task.Command == "" {
		return fmt.Errorf("task %q missing command", taskName)
	}

	order, err := resolveTaskOrder(conf.Tasks, taskName)
	if err != nil {
		return err
	}

	for i, name := range order {
		t := conf.Tasks[name]
		argv, err := parseCommand(t.Command)
		if err != nil {
			return fmt.Errorf("task %q: %w", name, err)
		}

		// Passthrough applies only to the root task (last in order).
		if i == len(order)-1 && len(passthrough) > 0 {
			argv = append(argv, passthrough...)
		}

		cwd, err := resolveCwd(confPath, t.Cwd)
		if err != nil {
			return fmt.Errorf("task %q: resolve cwd: %w", name, err)
		}

		env := buildEnv(confPath, t.Env)

		exe := ""
		// Managed tools are executed exclusively from .rig/bin (no PATH fallback).
		// Explicit exception: `go` is resolved from PATH (toolchain), and is never installed by rig.
		if argv[0] != "go" {
			if p, ok, rerr := ResolveManagedToolExecutable(confPath, lock, argv[0]); rerr != nil {
				return fmt.Errorf("task %q: %w", name, rerr)
			} else if ok {
				exe = p
			}
		}
		if exe == "" {
			exe, err = resolveExecutable(argv[0], cwd, env)
			if err != nil {
				return fmt.Errorf("task %q: %w", name, err)
			}
		}

		if err := Execute(exe, argv[1:], ExecOptions{Dir: cwd, Env: env, EnvExact: true}); err != nil {
			return fmt.Errorf("task %q failed: %w", name, err)
		}
	}

	return nil
}

func resolveTaskOrder(tasks cfg.TasksMap, root string) ([]string, error) {
	adj := make(map[string][]string, len(tasks))
	for name, t := range tasks {
		if len(t.DependsOn) > 0 {
			deps := make([]string, len(t.DependsOn))
			copy(deps, t.DependsOn)
			adj[name] = deps
		} else {
			adj[name] = nil
		}
	}
	if _, ok := adj[root]; !ok {
		return nil, fmt.Errorf("task %q not found", root)
	}

	state := make(map[string]int, len(adj))
	var order []string
	var stack []string

	var dfs func(string) error
	dfs = func(u string) error {
		st := state[u]
		if st == 1 {
			// cycle: report a minimal path
			idx := 0
			for i, s := range stack {
				if s == u {
					idx = i
					break
				}
			}
			cycle := append(stack[idx:], u)
			return fmt.Errorf("cycle detected: %v", cycle)
		}
		if st == 2 {
			return nil
		}
		state[u] = 1
		stack = append(stack, u)
		for _, v := range adj[u] {
			if _, ok := adj[v]; !ok {
				return fmt.Errorf("task %q depends_on unknown task %q", u, v)
			}
			if err := dfs(v); err != nil {
				return err
			}
		}
		stack = stack[:len(stack)-1]
		state[u] = 2
		order = append(order, u)
		return nil
	}

	if err := dfs(root); err != nil {
		return nil, err
	}
	if len(order) == 0 {
		return nil, errors.New("empty task plan")
	}
	return order, nil
}
