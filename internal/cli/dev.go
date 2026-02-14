package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	cfg "github.com/divijg19/rig/internal/config"
	core "github.com/divijg19/rig/internal/rig"
	"github.com/spf13/cobra"
)

var devColorMode string

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run the dev loop (watch + restart)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rt, err := loadDevRuntime(devColorMode, os.Stdout, os.Stderr)
		if err != nil {
			return err
		}
		return rt.Run()
	},
}

func init() {
	devCmd.Flags().StringVar(&devColorMode, "color", "auto", "color output: auto|always|never")
	rootCmd.AddCommand(devCmd)
}

// DevRuntime is the v0.3 dev runtime entrypoint.
// It validates invariants, constructs the watcher, and supervises restarts.
type DevRuntime struct {
	Task      cfg.Task
	Lock      core.Lockfile
	Toolchain core.GoToolchainLock

	configPath  string
	tools       map[string]string
	watchGlobs  []string
	command     string
	cwd         string
	env         []string
	watcherPath string
	watcherArgs []string
	colorMode   string
	colorOn     bool
	out         io.Writer
	errOut      io.Writer
}

// Supervisor manages a single child process at a time.
type Supervisor struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func loadDevRuntime(colorMode string, out io.Writer, errOut io.Writer) (*DevRuntime, error) {
	conf, confPath, err := core.LoadConfig("")
	if err != nil {
		if errors.Is(err, cfg.ErrConfigNotFound) {
			return nil, errors.New(msgNoConfig)
		}
		return nil, err
	}
	lockPath := filepath.Join(filepath.Dir(confPath), "rig.lock")
	lock, err := core.ReadLockfile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("error: rig.lock required; run 'rig sync'")
		}
		return nil, err
	}

	colorOn, err := resolveColorEnabled(colorMode, os.Stdout)
	if err != nil {
		return nil, err
	}

	devTask, ok := conf.Tasks["dev"]
	if !ok {
		return nil, errors.New("error: [tasks.dev] is required")
	}
	if strings.TrimSpace(devTask.Command) == "" {
		return nil, errors.New("error: [tasks.dev] must define 'command'")
	}
	if len(devTask.Watch) == 0 {
		return nil, errors.New("error: [tasks.dev] must define 'watch'")
	}

	rt := &DevRuntime{
		Task:       devTask,
		Lock:       lock,
		configPath: confPath,
		tools:      conf.Tools,
		watchGlobs: devTask.Watch,
		colorMode:  colorMode,
		colorOn:    colorOn,
		out:        out,
		errOut:     errOut,
	}
	if lock.Toolchain != nil && lock.Toolchain.Go != nil {
		rt.Toolchain = *lock.Toolchain.Go
	}
	if err := rt.Validate(); err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *DevRuntime) Validate() error {
	if strings.TrimSpace(r.Task.Command) == "" {
		return errors.New("error: [tasks.dev] must define 'command'")
	}
	if len(r.Task.Watch) == 0 {
		return errors.New("error: [tasks.dev] must define 'watch'")
	}
	for _, g := range r.Task.Watch {
		if strings.TrimSpace(g) == "" {
			return errors.New("error: [tasks.dev] must define 'watch'")
		}
	}

	if !hasTool(r.tools, "reflex") {
		return errors.New("error: dev watcher 'reflex' must be declared in [tools]")
	}

	if goRow, ok := core.CheckGoToolchainAgainstLock(r.tools, r.Lock, r.configPath); !ok {
		if goRow != nil {
			if goRow.Error != "" {
				return fmt.Errorf("error: go toolchain check failed (%s): %s", goRow.Status, goRow.Error)
			}
			return fmt.Errorf("error: go toolchain check failed (%s): have %q, want %q", goRow.Status, goRow.Have, goRow.Locked)
		}
		return errors.New("error: go toolchain check failed")
	}

	rows, missing, mismatched, extras, err := core.CheckInstalledTools(r.tools, r.Lock, r.configPath)
	if err != nil {
		return fmt.Errorf("error: %s", err)
	}
	_ = rows
	if missing > 0 || mismatched > 0 {
		if err := r.ensureWatcherInstalled(); err != nil {
			return err
		}
		return fmt.Errorf("error: tools are out of sync with rig.lock (missing=%d mismatched=%d extras=%d)", missing, mismatched, len(extras))
	}

	if err := r.ensureWatcherInstalled(); err != nil {
		return err
	}
	if err := ensureShellAvailable(); err != nil {
		return err
	}
	cmdCwd, err := resolveDevCwd(r.configPath, r.Task.Cwd)
	if err != nil {
		return err
	}

	r.command = strings.TrimSpace(r.Task.Command)
	r.cwd = cmdCwd
	r.env = buildDevEnv(r.configPath, r.Task.Env)
	r.watcherPath = core.ToolBinPath(r.configPath, "reflex")
	r.watcherArgs = buildWatcherArgs(r.Task.Watch, r.command)

	return nil
}

func (r *DevRuntime) Run() error {
	reloadCh, exitCh, cleanup := r.startKeyListener()
	defer cleanup()

	r.logStart()
	err := r.supervise(reloadCh, exitCh)
	r.logStop()
	return err
}

func (r *DevRuntime) supervise(reloadCh <-chan struct{}, exitCh <-chan struct{}) error {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	manualExit := false

	for {
		s := &Supervisor{}
		ctx, cancel := context.WithCancel(context.Background())
		cmd, err := r.spawn(ctx)
		if err != nil {
			cancel()
			return err
		}
		s.cmd = cmd
		s.cancel = cancel
		if err := cmd.Start(); err != nil {
			cancel()
			return err
		}

		waitCh := make(chan error, 1)
		go func() { waitCh <- cmd.Wait() }()

		select {
		case <-exitCh:
			manualExit = true
			s.stop(syscall.SIGTERM)
			waitForExit(waitCh, cancel)
			return nil
		case <-reloadCh:
			r.logManualReload()
			r.logRestarting()
			s.stop(syscall.SIGTERM)
			waitForExit(waitCh, cancel)
			continue
		case sig := <-sigCh:
			switch sig {
			case os.Interrupt:
				manualExit = true
				s.stop(syscall.SIGTERM)
				waitForExit(waitCh, cancel)
				return nil
			default:
				s.stop(syscall.SIGTERM)
				waitForExit(waitCh, cancel)
				return nil
			}
		case err := <-waitCh:
			if err == nil {
				return nil
			}
			if errors.Is(err, context.Canceled) || manualExit {
				return nil
			}
			r.logChangeDetected()
			r.logRestarting()
			continue
		}
	}
}

func (r *DevRuntime) spawn(ctx context.Context) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, r.watcherPath, r.watcherArgs...)
	cmd.Dir = r.cwd
	cmd.Env = r.env
	cmd.Stdout = r.out
	if r.colorOn {
		cmd.Stderr = &ansiWriter{w: r.errOut, prefix: ansiRed, suffix: ansiReset}
	} else {
		cmd.Stderr = r.errOut
	}
	cmd.Stdin = os.Stdin
	return cmd, nil
}

func (r *DevRuntime) logStart() {
	start := "🚀 dev started"
	watch := fmt.Sprintf("👀 watching: %s", strings.Join(r.Task.Watch, ", "))
	cmd := fmt.Sprintf("▶ %s", r.command)
	if r.colorOn {
		start = ansiBoldCyan + start + ansiReset
		watch = ansiBoldCyan + watch + ansiReset
		cmd = ansiBoldCyan + cmd + ansiReset
	}
	fmt.Fprintln(r.out, start)
	fmt.Fprintln(r.out, watch)
	fmt.Fprintln(r.out, cmd)
}

func (r *DevRuntime) logChangeDetected() {
	msg := "🔁 change detected"
	if r.colorOn {
		msg = ansiYellow + msg + ansiReset
	}
	fmt.Fprintln(r.out, msg)
}

func (r *DevRuntime) logManualReload() {
	msg := "🔄 manual reload"
	if r.colorOn {
		msg = ansiYellow + msg + ansiReset
	}
	fmt.Fprintln(r.out, msg)
}

func (r *DevRuntime) logRestarting() {
	msg := "▶ restarting…"
	if r.colorOn {
		msg = ansiYellow + msg + ansiReset
	}
	fmt.Fprintln(r.out, msg)
}

func (r *DevRuntime) logStop() {
	msg := "🛑 dev stopped"
	if r.colorOn {
		msg = ansiRed + msg + ansiReset
	}
	fmt.Fprintln(r.out, msg)
}

func (r *DevRuntime) startKeyListener() (<-chan struct{}, <-chan struct{}, func()) {
	if !isTTY(os.Stdin) || runtime.GOOS == "windows" {
		return nil, nil, func() {}
	}
	fd := int(os.Stdin.Fd())
	restoreMode, err := setDevInputMode(fd)
	if err != nil {
		return nil, nil, func() {}
	}
	reloadCh := make(chan struct{}, 1)
	exitCh := make(chan struct{}, 1)
	done := make(chan struct{})

	go func() {
		buf := []byte{0}
		for {
			select {
			case <-done:
				return
			default:
			}
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				return
			}
			switch buf[0] {
			case 0x12: // Ctrl+R
				select {
				case reloadCh <- struct{}{}:
				default:
				}
			case 0x03: // Ctrl+C
				select {
				case exitCh <- struct{}{}:
				default:
				}
			}
		}
	}()

	cleanup := func() {
		close(done)
		restoreMode()
	}
	return reloadCh, exitCh, cleanup
}

func (s *Supervisor) stop(sig os.Signal) {
	if s.cmd == nil || s.cmd.Process == nil {
		return
	}
	if runtime.GOOS == "windows" {
		_ = s.cmd.Process.Kill()
		return
	}
	if sig == nil {
		_ = s.cmd.Process.Kill()
		return
	}
	_ = s.cmd.Process.Signal(sig)
}

func waitForExit(waitCh <-chan error, cancel context.CancelFunc) {
	select {
	case <-waitCh:
		return
	case <-time.After(200 * time.Millisecond):
		cancel()
		<-waitCh
	}
}

func ensureShellAvailable() error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if _, err := exec.LookPath("sh"); err != nil {
		return errors.New("error: sh not found on PATH")
	}
	return nil
}

func (r *DevRuntime) ensureWatcherInstalled() error {
	path := core.ToolBinPath(r.configPath, "reflex")
	if err := ensureExecutable(path); err != nil {
		return errors.New("error: dev watcher 'reflex' missing in .rig/bin")
	}
	return nil
}

func buildWatcherArgs(globs []string, command string) []string {
	regex := computeWatchRegex(globs)
	args := []string{"-s", "-r", regex, "--", "sh", "-c", command}
	return args
}

func computeWatchRegex(globs []string) string {
	trimmed := make([]string, 0, len(globs))
	hasGo := false
	onlyDot := len(globs) > 0
	for _, g := range globs {
		v := strings.TrimSpace(g)
		if v == "" {
			continue
		}
		trimmed = append(trimmed, v)
		if strings.Contains(v, ".go") {
			hasGo = true
		}
		if v != "." {
			onlyDot = false
		}
	}
	if hasGo {
		return `\.go$`
	}
	if onlyDot {
		return "."
	}
	if len(trimmed) == 0 {
		return "."
	}
	if len(trimmed) == 1 {
		return trimmed[0]
	}
	return "(" + strings.Join(trimmed, "|") + ")"
}

func resolveDevCwd(configPath string, taskCwd string) (string, error) {
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

func buildDevEnv(configPath string, taskEnv map[string]string) []string {
	base := map[string]string{}
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		base[k] = v
	}

	localBin := filepath.Join(filepath.Dir(configPath), ".rig", "bin")

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
	sortStrings(keys)

	env := make([]string, 0, len(keys))
	for _, k := range keys {
		env = append(env, k+"="+base[k])
	}
	return env
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

func hasTool(tools map[string]string, want string) bool {
	if tools == nil {
		return false
	}
	if _, ok := tools[want]; ok {
		return true
	}
	id := core.ResolveToolIdentity(want)
	if _, ok := tools[id.Module]; ok {
		return true
	}
	if _, ok := tools[id.InstallPath]; ok {
		return true
	}
	return false
}

type ansiWriter struct {
	w      io.Writer
	prefix string
	suffix string
}

func (w *ansiWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if _, err := io.WriteString(w.w, w.prefix); err != nil {
		return 0, err
	}
	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}
	if _, err := io.WriteString(w.w, w.suffix); err != nil {
		return n, err
	}
	return len(p), nil
}

func sortStrings(s []string) {
	if len(s) < 2 {
		return
	}
	for i := 1; i < len(s); i++ {
		j := i
		for j > 0 && s[j] < s[j-1] {
			s[j], s[j-1] = s[j-1], s[j]
			j--
		}
	}
}
