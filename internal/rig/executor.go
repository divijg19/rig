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
