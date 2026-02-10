package rig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func Dev(startDir string) error {
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

	devTask, ok := conf.Tasks["dev"]
	if !ok {
		return errors.New("task \"dev\" not found")
	}
	if strings.TrimSpace(devTask.Command) == "" {
		return errors.New("task \"dev\" missing command")
	}
	if len(devTask.Watch) == 0 {
		return errors.New("task \"dev\" missing watch patterns")
	}

	// v0.3 watcher: tool-backed (reflex). We intentionally do not implement a native watcher.
	if !hasTool(conf.Tools, "reflex") {
		return errors.New("rig dev requires a watcher tool in [tools]; add reflex = \"<version>\" and run 'rig sync'")
	}

	// Build reflex args: reflex -r <regex> -- <cmd...>
	watchRE, err := watchGlobsToRegex(devTask.Watch)
	if err != nil {
		return fmt.Errorf("task \"dev\": invalid watch: %w", err)
	}
	argv, err := parseCommand(devTask.Command)
	if err != nil {
		return fmt.Errorf("task \"dev\": %w", err)
	}

	cwd, err := resolveCwd(confPath, devTask.Cwd)
	if err != nil {
		return fmt.Errorf("task \"dev\": resolve cwd: %w", err)
	}
	// Ensure the underlying executable resolves before we start the long-running loop.
	env := buildEnv(confPath, devTask.Env)
	if _, err := resolveExecutable(argv[0], cwd, env); err != nil {
		return fmt.Errorf("task \"dev\": %w", err)
	}

	reflexPath := ToolBinPath(confPath, "reflex")
	if err := ensureExecutable(reflexPath); err != nil {
		return fmt.Errorf("reflex not installed in .rig/bin (run 'rig sync'): %w", err)
	}

	reflexArgs := []string{"-r", watchRE, "--"}
	reflexArgs = append(reflexArgs, argv...)

	return superviseWithSignals(func(ctx context.Context) (*exec.Cmd, error) {
		cmd := exec.CommandContext(ctx, reflexPath, reflexArgs...)
		cmd.Dir = cwd
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd, nil
	})
}

func hasTool(tools map[string]string, want string) bool {
	if tools == nil {
		return false
	}
	if _, ok := tools[want]; ok {
		return true
	}
	// Also accept fully-qualified module keys.
	id := ResolveToolIdentity(want)
	if _, ok := tools[id.Module]; ok {
		return true
	}
	if _, ok := tools[id.InstallPath]; ok {
		return true
	}
	return false
}

func watchGlobsToRegex(globs []string) (string, error) {
	if len(globs) == 0 {
		return "", errors.New("watch is empty")
	}

	parts := make([]string, 0, len(globs))
	for _, g := range globs {
		g = strings.TrimSpace(g)
		if g == "" {
			return "", errors.New("watch patterns must be non-empty")
		}
		r, err := globToPathRegex(g)
		if err != nil {
			return "", err
		}
		parts = append(parts, "("+r+")")
	}
	return strings.Join(parts, "|"), nil
}

func globToPathRegex(glob string) (string, error) {
	// Very small glob->regex adapter. Supports: **, *, ? and path separators.
	// We intentionally avoid complex glob semantics; the watcher tool does the heavy lifting.
	var b strings.Builder
	b.WriteString("^")
	runes := []rune(glob)
	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case '*':
			if i+1 < len(runes) && runes[i+1] == '*' {
				b.WriteString(".*")
				i++
				continue
			}
			b.WriteString("[^\\\\/]*")
		case '?':
			b.WriteString(".")
		case '/':
			b.WriteString("[\\\\/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(runes[i])))
		}
	}
	b.WriteString("$")

	// Quick sanity check that we produced a valid regexp.
	if _, err := regexp.Compile(b.String()); err != nil {
		return "", fmt.Errorf("watch pattern %q produced invalid regex: %w", glob, err)
	}
	return b.String(), nil
}

func superviseWithSignals(spawn func(ctx context.Context) (*exec.Cmd, error)) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 2)
	// SIGINT restarts; SIGTERM exits.
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	var lastStart time.Time
	restart := func() bool {
		// Basic backoff to avoid tight restart loops.
		if !lastStart.IsZero() && time.Since(lastStart) < 200*time.Millisecond {
			time.Sleep(200 * time.Millisecond)
		}
		lastStart = time.Now()
		return true
	}

	for {
		if !restart() {
			return nil
		}
		cmd, err := spawn(ctx)
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}

		waitCh := make(chan error, 1)
		go func() { waitCh <- cmd.Wait() }()

		select {
		case sig := <-sigCh:
			switch sig {
			case os.Interrupt:
				// Restart: forward SIGINT if possible, then loop.
				_ = interruptProcess(cmd)
				<-waitCh
				continue
			default:
				// SIGTERM (or anything else): forward termination and exit.
				_ = terminateProcess(cmd)
				<-waitCh
				return nil
			}
		case err := <-waitCh:
			// Child exited on its own.
			if err == nil {
				return nil
			}
			// If the context was canceled, treat it as clean shutdown.
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
	}
}

func interruptProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		return cmd.Process.Kill()
	}
	return cmd.Process.Signal(syscall.SIGINT)
}

func terminateProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		return cmd.Process.Kill()
	}
	// Prefer SIGTERM, fallback to SIGKILL after a short grace.
	_ = cmd.Process.Signal(syscall.SIGTERM)
	grace := 2 * time.Second
	deadline := time.Now().Add(grace)
	for time.Now().Before(deadline) {
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}
	return cmd.Process.Kill()
}
