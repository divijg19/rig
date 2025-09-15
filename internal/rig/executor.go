// internal/rig/executor.go

package rig

import (
	"os"
	"os/exec"
	"runtime"
)

// ExecOptions describes how a task should be executed.
type ExecOptions struct {
	// Dir sets the working directory. Empty means current.
	Dir string
	// Env allows adding/overriding environment variables (KEY=VALUE form).
	Env []string
}

// ExecuteShell runs a shell command string via the platform shell, streaming stdio.
// On Windows: cmd /c <command>; on others: sh -c <command>.
func ExecuteShell(command string, opts ExecOptions) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// ExecuteShellWith selects a specific shell by name: "sh", "bash", "pwsh", "cmd".
// Falls back to platform default if unknown.
func ExecuteShellWith(shell, command string, opts ExecOptions) error {
	var cmd *exec.Cmd
	switch shell {
	case "bash":
		cmd = exec.Command("bash", "-lc", command)
	case "sh":
		cmd = exec.Command("sh", "-c", command)
	case "pwsh":
		cmd = exec.Command("pwsh", "-NoLogo", "-NoProfile", "-Command", command)
	case "cmd":
		cmd = exec.Command("cmd", "/c", command)
	default:
		return ExecuteShell(command, opts)
	}
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Execute runs a binary with argv directly (no shell), streaming stdio.
func Execute(name string, args []string, opts ExecOptions) error {
	cmd := exec.Command(name, args...)
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
