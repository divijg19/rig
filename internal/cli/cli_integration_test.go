package cli

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"

	core "github.com/divijg19/rig/internal/rig"
)

func projectRoot(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	// Find the project root by looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (go.mod)")
		}
		dir = parent
	}
}

func buildRigBinary(t *testing.T, outDir string) string {
	binName := "rig"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(outDir, binName)
	root := projectRoot(t)
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/rig")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build rig failed: %v\n%s", err, out)
	}
	return binPath
}

func runWithInvocationName(t *testing.T, bin string, dir string, invocation string, args ...string) (string, error) {
	t.Helper()
	// We want to execute the same rig binary, but set argv[0] to a different
	// invocation name (rir/ric/rid/ris). This simulates manual symlinks/renames
	// without requiring filesystem mutation.
	cmd := exec.Command(bin)
	cmd.Args = append([]string{invocation}, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestSingleBinaryCmdTree(t *testing.T) {
	root := projectRoot(t)
	entries, err := os.ReadDir(filepath.Join(root, "cmd"))
	if err != nil {
		t.Fatalf("readdir cmd/: %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) != 1 || dirs[0] != "rig" {
		t.Fatalf("cmd/ must contain exactly one entrypoint directory (rig). got=%v", dirs)
	}
}

func TestRootHelpOutputUnified(t *testing.T) {
	dir := t.TempDir()
	outDefault, err := runRigCmdInDir(t, dir)
	if err != nil {
		t.Fatalf("rig failed: %v\n%s", err, outDefault)
	}
	outShort, err := runRigCmdInDir(t, dir, "-h")
	if err != nil {
		t.Fatalf("rig -h failed: %v\n%s", err, outShort)
	}
	outLong, err := runRigCmdInDir(t, dir, "--help")
	if err != nil {
		t.Fatalf("rig --help failed: %v\n%s", err, outLong)
	}
	outHelpCmd, err := runRigCmdInDir(t, dir, "help")
	if err != nil {
		t.Fatalf("rig help failed: %v\n%s", err, outHelpCmd)
	}

	if outDefault != outShort || outDefault != outLong || outDefault != outHelpCmd {
		t.Fatalf("expected identical help output across invocations\nrig:\n%s\nrig -h:\n%s\nrig --help:\n%s\nrig help:\n%s", outDefault, outShort, outLong, outHelpCmd)
	}
	if !strings.HasPrefix(outDefault, "rig ") {
		t.Fatalf("expected version header prefix, got: %q", outDefault)
	}
	if !strings.Contains(outDefault, "\n\n") {
		t.Fatalf("expected blank line between version and help body, got: %q", outDefault)
	}
	if !strings.Contains(outDefault, "Run \"rig version\" for build information.") {
		t.Fatalf("expected help footer note, got: %q", outDefault)
	}
}

func TestVersionOutputUnified(t *testing.T) {
	dir := t.TempDir()
	outCmd, err := runRigCmdInDir(t, dir, "version")
	if err != nil {
		t.Fatalf("rig version failed: %v\n%s", err, outCmd)
	}
	outShort, err := runRigCmdInDir(t, dir, "-v")
	if err != nil {
		t.Fatalf("rig -v failed: %v\n%s", err, outShort)
	}
	outLong, err := runRigCmdInDir(t, dir, "--version")
	if err != nil {
		t.Fatalf("rig --version failed: %v\n%s", err, outLong)
	}

	if outCmd != outShort || outCmd != outLong {
		t.Fatalf("expected identical version output across invocations\nrig version:\n%s\nrig -v:\n%s\nrig --version:\n%s", outCmd, outShort, outLong)
	}
	if strings.Contains(outCmd, "Usage:") || strings.Contains(outCmd, "Available Commands") {
		t.Fatalf("version output must not include help text, got: %s", outCmd)
	}
	if strings.Contains(outCmd, "\x1b[") {
		t.Fatalf("version output must not include ANSI color, got: %q", outCmd)
	}
	if !strings.Contains(outCmd, "\ncommit: ") || !strings.Contains(outCmd, "\nbuilt: ") {
		t.Fatalf("expected commit/built lines in version output, got: %q", outCmd)
	}
	if !strings.Contains(outCmd, "\ngo: "+runtime.Version()+"\n") {
		t.Fatalf("expected go runtime line in version output, got: %q", outCmd)
	}
}

func TestRootAndVersionExitCodes(t *testing.T) {
	dir := t.TempDir()
	if out, err := runRigCmdInDir(t, dir); err != nil {
		t.Fatalf("expected rig to exit 0, got err=%v output=%s", err, out)
	}
	if out, err := runRigCmdInDir(t, dir, "-h"); err != nil {
		t.Fatalf("expected rig -h to exit 0, got err=%v output=%s", err, out)
	}
	if out, err := runRigCmdInDir(t, dir, "version"); err != nil {
		t.Fatalf("expected rig version to exit 0, got err=%v output=%s", err, out)
	}

	out, err := runRigCmdInDir(t, dir, "definitely-unknown-command")
	if err == nil {
		t.Fatalf("expected unknown command to exit non-zero, got output=%s", out)
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected ExitError for unknown command, got %T", err)
	}
	if ee.ExitCode() != 1 {
		t.Fatalf("expected exit code 1 for unknown command, got %d (output=%s)", ee.ExitCode(), out)
	}
}

func runRigCmdInDir(t *testing.T, dir string, args ...string) (string, error) {
	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runRigCmdInDirWithEnv(t *testing.T, dir string, env []string, args ...string) (string, error) {
	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func envWithPrependedPath(prepend string) []string {
	base := os.Environ()
	old := ""
	for _, kv := range base {
		if strings.HasPrefix(kv, "PATH=") {
			old = strings.TrimPrefix(kv, "PATH=")
			break
		}
	}
	newPath := prepend
	if old != "" {
		newPath = prepend + string(os.PathListSeparator) + old
	}
	out := make([]string, 0, len(base)+1)
	for _, kv := range base {
		if strings.HasPrefix(kv, "PATH=") {
			continue
		}
		out = append(out, kv)
	}
	out = append(out, "PATH="+newPath)
	return out
}

func readUntilContains(t *testing.T, f *os.File, substr string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var buf strings.Builder
	chunk := make([]byte, 256)
	for time.Now().Before(deadline) {
		_ = f.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		n, err := f.Read(chunk)
		if n > 0 {
			buf.Write(chunk[:n])
			if strings.Contains(buf.String(), substr) {
				return buf.String()
			}
		}
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
		}
	}
	t.Fatalf("timeout waiting for %q; output so far: %s", substr, buf.String())
	return ""
}

func writeFile(t *testing.T, path string, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCheckFailsWithoutLock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
mockery = "2.0.0"

[tasks]
noop = "true"
`, 0o644)

	out, err := runRigCmdInDir(t, dir, "check")
	if err == nil {
		t.Fatalf("expected error, got none. output=%s", out)
	}
	if !strings.Contains(out, "\"ok\":false") {
		t.Fatalf("expected JSON ok=false, got: %s", out)
	}
}

func TestCheckOKWithLockAndInstalledTools(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
mockery = "2.0.0"

[tasks]
ver = "mockery --version"
`, 0o644)
	bin := filepath.Join(dir, ".rig", "bin", "mockery")
	writeFile(t, bin, "#!/bin/sh\necho mockery v2.0.0\n", 0o755)
	sha, err := core.ComputeFileSHA256(bin)
	if err != nil {
		t.Fatalf("sha256 mockery: %v", err)
	}
	writeFile(t, filepath.Join(dir, "rig.lock"), fmt.Sprintf(`schema = 0

[[tools]]
kind = "go-binary"
requested = "mockery@2.0.0"
resolved = "github.com/vektra/mockery/v2@v2.0.0"
module = "github.com/vektra/mockery/v2"
bin = "mockery"
sha256 = %q
`, sha), 0o644)

	out, err := runRigCmdInDir(t, dir, "check")
	if err != nil {
		t.Fatalf("expected success, got error: %v\n%s", err, out)
	}
	if !strings.Contains(out, "\"ok\":true") {
		t.Fatalf("expected JSON ok=true, got: %s", out)
	}
}

func TestRunRequiresLock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
mockery = "2.0.0"

[tasks]
ver = "mockery --version"
`, 0o644)
	writeFile(t, filepath.Join(dir, ".rig", "bin", "mockery"), "#!/bin/sh\necho mockery v2.0.0\n", 0o755)

	out, err := runRigCmdInDir(t, dir, "run", "ver")
	if err == nil {
		t.Fatalf("expected error, got none. output=%s", out)
	}
}

func TestRunDeterministicDepsAndPassthrough(t *testing.T) {
	dir := t.TempDir()
	append := filepath.Join(dir, "append")
	writeFile(t, append, "#!/bin/sh\nprintf '%s\\n' \"$*\" >> out.txt\n", 0o755)

	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
mockery = "2.0.0"

[tasks]
dep1 = "./append dep1"
dep2 = { command = "./append dep2", depends_on = ["dep1"] }
main = { command = "./append main", depends_on = ["dep2"] }
`, 0o644)
	mockBin := filepath.Join(dir, ".rig", "bin", "mockery")
	writeFile(t, mockBin, "#!/bin/sh\necho mockery v2.0.0\n", 0o755)
	sha, err := core.ComputeFileSHA256(mockBin)
	if err != nil {
		t.Fatalf("sha256 mockery: %v", err)
	}
	writeFile(t, filepath.Join(dir, "rig.lock"), fmt.Sprintf(`schema = 0

[[tools]]
kind = "go-binary"
requested = "mockery@2.0.0"
resolved = "github.com/vektra/mockery/v2@v2.0.0"
module = "github.com/vektra/mockery/v2"
bin = "mockery"
sha256 = %q
`, sha), 0o644)

	out, err := runRigCmdInDir(t, dir, "run", "main", "--", "extra")
	if err != nil {
		t.Fatalf("expected success, got error: %v\n%s", err, out)
	}
	b, rerr := os.ReadFile(filepath.Join(dir, "out.txt"))
	if rerr != nil {
		t.Fatalf("read out.txt: %v", rerr)
	}
	got := strings.TrimSpace(string(b))
	if got != "dep1\ndep2\nmain extra" {
		t.Fatalf("unexpected task order/passthrough; got:\n%s", got)
	}
}

func TestEntrypointRirMatchesRigRunList(t *testing.T) {
	work := t.TempDir()
	writeFile(t, filepath.Join(work, "rig.toml"), `
[tasks]
a = "true"
b = "true"
`, 0o644)

	rigBin := buildRigBinary(t, t.TempDir())

	cmd1 := exec.Command(rigBin, "run", "--list")
	cmd1.Dir = work
	out1, err := cmd1.CombinedOutput()
	if err != nil {
		t.Fatalf("rig run --list failed: %v\n%s", err, out1)
	}
	out2, err := runWithInvocationName(t, rigBin, work, "rir", "--list")
	if err != nil {
		t.Fatalf("rir --list failed: %v\n%s", err, out2)
	}
	if strings.TrimSpace(string(out1)) != strings.TrimSpace(string(out2)) {
		t.Fatalf("outputs differ\nrig: %q\nrir: %q", strings.TrimSpace(string(out1)), strings.TrimSpace(string(out2)))
	}
}

func TestRigRunListShowsDescriptionsAligned(t *testing.T) {
	work := t.TempDir()
	writeFile(t, filepath.Join(work, "rig.toml"), `
[tasks]
a = { command = "true", description = "Alpha" }
b = "true"
longtask = { command = "true", description = "Long description" }
`, 0o644)

	rigBin := buildRigBinary(t, t.TempDir())

	cmd := exec.Command(rigBin, "run", "--list")
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rig run --list failed: %v\n%s", err, out)
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), string(out))
	}
	if !strings.HasPrefix(lines[0], "a") {
		t.Fatalf("expected first line to start with task name 'a', got: %q", lines[0])
	}
	if lines[1] != "b" {
		t.Fatalf("expected second line to be 'b', got: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "longtask") {
		t.Fatalf("expected third line to start with task name 'longtask', got: %q", lines[2])
	}

	idxA := strings.Index(lines[0], "Alpha")
	idxL := strings.Index(lines[2], "Long description")
	if idxA < 0 || idxL < 0 {
		t.Fatalf("expected descriptions in output, got: %q", string(out))
	}
	if strings.TrimSpace(lines[0][:idxA]) != "a" {
		t.Fatalf("unexpected prefix for 'a' line: %q", lines[0])
	}
	if strings.TrimSpace(lines[2][:idxL]) != "longtask" {
		t.Fatalf("unexpected prefix for 'longtask' line: %q", lines[2])
	}
	if idxA != idxL {
		t.Fatalf("expected aligned descriptions, got positions %d and %d\n%s", idxA, idxL, string(out))
	}
}

func TestEntrypointRicMatchesRigCheck(t *testing.T) {
	work := t.TempDir()
	writeFile(t, filepath.Join(work, "rig.toml"), `
[tools]
mockery = "2.0.0"

[tasks]
noop = "true"
`, 0o644)

	rigBin := buildRigBinary(t, t.TempDir())

	cmd1 := exec.Command(rigBin, "check")
	cmd1.Dir = work
	out1, err := cmd1.CombinedOutput()
	if err == nil {
		t.Fatalf("expected rig check to fail without lock, got none. output=%s", out1)
	}
	if !strings.Contains(string(out1), "\"ok\":false") {
		t.Fatalf("expected JSON ok=false, got: %s", out1)
	}

	out2, err := runWithInvocationName(t, rigBin, work, "ric")
	if err == nil {
		t.Fatalf("expected ric invocation to fail without lock, got none. output=%s", out2)
	}
	if strings.TrimSpace(string(out1)) != strings.TrimSpace(string(out2)) {
		t.Fatalf("outputs differ\nrig check: %q\nric: %q", strings.TrimSpace(string(out1)), strings.TrimSpace(string(out2)))
	}
}

func TestEntrypointRidMatchesRigDev(t *testing.T) {
	work := t.TempDir()
	writeFile(t, filepath.Join(work, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "go run ."
watch = ["**/*.go"]
`, 0o644)

	rigBin := buildRigBinary(t, t.TempDir())

	cmd1 := exec.Command(rigBin, "dev")
	cmd1.Dir = work
	out1, err := cmd1.CombinedOutput()
	if err == nil {
		t.Fatalf("expected rig dev to fail without lock, got none. output=%s", out1)
	}

	out2, err := runWithInvocationName(t, rigBin, work, "rid")
	if err == nil {
		t.Fatalf("expected rid invocation to fail without lock, got none. output=%s", out2)
	}
	if strings.TrimSpace(string(out1)) != strings.TrimSpace(string(out2)) {
		t.Fatalf("outputs differ\nrig dev: %q\nrid: %q", strings.TrimSpace(string(out1)), strings.TrimSpace(string(out2)))
	}
}

func TestEntrypointRisMatchesRigStartStub(t *testing.T) {
	work := t.TempDir()
	rigBin := buildRigBinary(t, t.TempDir())

	cmd1 := exec.Command(rigBin, "start")
	cmd1.Dir = work
	out1, err := cmd1.CombinedOutput()
	if err == nil {
		t.Fatalf("expected rig start to fail (stub), got none. output=%s", out1)
	}

	out2, err := runWithInvocationName(t, rigBin, work, "ris")
	if err == nil {
		t.Fatalf("expected ris invocation to fail (stub), got none. output=%s", out2)
	}
	if strings.TrimSpace(string(out1)) != strings.TrimSpace(string(out2)) {
		t.Fatalf("outputs differ\nrig start: %q\nris: %q", strings.TrimSpace(string(out1)), strings.TrimSpace(string(out2)))
	}
}

func TestAliasCommandOutputIsStable(t *testing.T) {
	work := t.TempDir()
	rigBin := buildRigBinary(t, t.TempDir())

	cmd := exec.Command(rigBin, "alias")
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rig alias failed: %v\n%s", err, out)
	}
	got := string(out)
	want := aliasInfoText
	if got != want {
		t.Fatalf("unexpected rig alias output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestRunRejectsUnsupportedTaskFields(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
mockery = "2.0.0"

[tasks]
bad = { command = "echo hi", shell = "bash" }
`, 0o644)
	out, err := runRigCmdInDir(t, dir, "run", "bad")
	if err == nil {
		t.Fatalf("expected error, got none. output=%s", out)
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if !strings.Contains(out, "unsupported field") {
		t.Fatalf("expected unsupported field error, got: %s", out)
	}
}

func TestToolAuthority_PositiveCheckAndDev(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nexit 0\n", 0o755)
	reflexBin := filepath.Join(dir, ".rig", "bin", "reflex")
	writeFile(t, reflexBin, "#!/bin/sh\n# minimal reflex stub for tests\n# must NOT be invoked with --version by rig\nif [ \"$1\" = \"--version\" ]; then\n  echo unexpected --version >&2\n  exit 2\nfi\n# run args after -- once, then exit\nwhile [ $# -gt 0 ]; do\n  if [ \"$1\" = \"--\" ]; then\n    shift\n    exec \"$@\"\n  fi\n  shift\ndone\nexit 0\n", 0o755)
	sha, err := core.ComputeFileSHA256(reflexBin)
	if err != nil {
		t.Fatalf("sha256 reflex: %v", err)
	}
	writeFile(t, filepath.Join(dir, "rig.lock"), fmt.Sprintf(`schema = 0

[[tools]]
kind = "go-binary"
requested = "reflex@latest"
resolved = "github.com/cespare/reflex@v9.9.9"
module = "github.com/cespare/reflex"
bin = "reflex"
sha256 = %q
`, sha), 0o644)

	out, err := runRigCmdInDir(t, dir, "check")
	if err != nil {
		t.Fatalf("expected check success, got error: %v\n%s", err, out)
	}
	if !strings.Contains(out, "\"ok\":true") {
		t.Fatalf("expected JSON ok=true, got: %s", out)
	}

	out, err = runRigCmdInDir(t, dir, "dev")
	if err != nil {
		t.Fatalf("expected dev success, got error: %v\n%s", err, out)
	}
}

func TestToolAuthority_GlobalToolDoesNotCount(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
`, 0o644)
	writeFile(t, filepath.Join(dir, "rig.lock"), `schema = 0

[[tools]]
kind = "go-binary"
requested = "reflex@latest"
resolved = "github.com/cespare/reflex@v9.9.9"
module = "github.com/cespare/reflex"
bin = "reflex"
sha256 = "00"
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nexit 0\n", 0o755)

	globalBin := filepath.Join(dir, "global-bin")
	writeFile(t, filepath.Join(globalBin, "reflex"), "#!/bin/sh\n# global reflex should be ignored\nif [ \"$1\" = \"--version\" ]; then\n  echo unexpected --version >&2\n  exit 2\nfi\nexit 0\n", 0o755)
	env := envWithPrependedPath(globalBin)

	out, err := runRigCmdInDirWithEnv(t, dir, env, "check")
	if err == nil {
		t.Fatalf("expected check to fail (no .rig/bin/reflex), got none. output=%s", out)
	}
	if !strings.Contains(out, "\"ok\":false") {
		t.Fatalf("expected JSON ok=false, got: %s", out)
	}

	out, err = runRigCmdInDirWithEnv(t, dir, env, "dev")
	if err == nil {
		t.Fatalf("expected dev to fail (no .rig/bin/reflex), got none. output=%s", out)
	}
	// The exact error string is UX, but it must fail even if reflex exists globally.
	if !strings.Contains(out, "dev watcher 'reflex' missing in .rig/bin") {
		t.Fatalf("expected missing-reflex failure, got: %s", out)
	}
}

func TestDevRestartOnChildExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nexit 0\n", 0o755)
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\nstate=\"$RIG_TEST_STATE\"\nif [ ! -f \"$state\" ]; then\n  echo first > \"$state\"\n  exit 1\nfi\nexit 0\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, "dev")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "RIG_TEST_STATE="+state)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected dev success, got error: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "▶ restarting…") {
		t.Fatalf("expected restart log, got: %s", out)
	}
}

func TestDevRejectsUnsupportedDevTaskFields(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
description = "nope"
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nexit 0\n", 0o755)

	// Even with a valid lock + installed reflex, strict parsing must reject extra fields.
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\nexit 0\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	out, err := runRigCmdInDir(t, dir, "dev")
	if err == nil {
		t.Fatalf("expected error, got none. output=%s", out)
	}
	if !strings.Contains(out, "unsupported field") {
		t.Fatalf("expected unsupported field error, got: %s", out)
	}
}

func TestDevRestartsOnGoFileChange(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	ready := filepath.Join(dir, "ready")
	watchFile := filepath.Join(dir, "watchme.go")

	writeFile(t, watchFile, "package main\n", 0o644)
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nexit 0\n", 0o755)

	reflexStub := "#!/bin/sh\n" +
		"seen_r=0\n" +
		"regex=\"\"\n" +
		"while [ $# -gt 0 ]; do\n" +
		"  if [ \"$1\" = \"-g\" ]; then\n" +
		"    echo 'glob mode disallowed' >&2\n" +
		"    exit 3\n" +
		"  fi\n" +
		"  if [ \"$1\" = \"-r\" ]; then\n" +
		"    seen_r=1\n" +
		"    shift\n" +
		"    if [ $# -eq 0 ]; then\n" +
		"      echo 'missing regex argument' >&2\n" +
		"      exit 4\n" +
		"    fi\n" +
		"    regex=\"$1\"\n" +
		"  fi\n" +
		"  shift\n" +
		"done\n" +
		"if [ \"$seen_r\" -ne 1 ]; then\n" +
		"  echo 'missing -r mode' >&2\n" +
		"  exit 5\n" +
		"fi\n" +
		"if [ \"$regex\" != '\\.go$' ]; then\n" +
		"  echo \"unexpected regex: $regex\" >&2\n" +
		"  exit 6\n" +
		"fi\n" +
		"state=\"$RIG_TEST_STATE\"\n" +
		"ready=\"$RIG_TEST_READY\"\n" +
		"watch=\"$RIG_TEST_WATCH\"\n" +
		"mtime() { stat -c %Y \"$1\" 2>/dev/null || stat -f %m \"$1\"; }\n" +
		"if [ ! -f \"$state\" ]; then\n" +
		"  echo first > \"$state\"\n" +
		"  if [ -n \"$ready\" ]; then echo ready > \"$ready\"; fi\n" +
		"  base=\"$(mtime \"$watch\")\"\n" +
		"  while true; do\n" +
		"    now=\"$(mtime \"$watch\")\"\n" +
		"    if [ \"$now\" != \"$base\" ]; then\n" +
		"      exit 1\n" +
		"    fi\n" +
		"    sleep 0.05\n" +
		"  done\n" +
		"fi\n" +
		"exit 0\n"

	reflexPath, reflexSHA := writeTool(t, dir, "reflex", reflexStub)
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, "dev", "--color=never")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "RIG_TEST_STATE="+state, "RIG_TEST_READY="+ready, "RIG_TEST_WATCH="+watchFile)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		t.Fatalf("start dev: %v", err)
	}
	defer func() { _ = cmd.Process.Signal(syscall.SIGTERM) }()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if _, err := os.Stat(ready); err != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		_ = cmd.Wait()
		t.Fatalf("timeout waiting for reflex ready; output so far: %s", buf.String())
	}

	// Ensure we cross a second boundary before rewriting (stub uses seconds mtime).
	time.Sleep(1200 * time.Millisecond)
	writeFile(t, watchFile, "package main\n// changed\n", 0o644)

	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	select {
	case err := <-waitCh:
		if err != nil {
			t.Fatalf("expected clean exit, got: %v\noutput=%s", err, buf.String())
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Signal(syscall.SIGTERM)
		<-waitCh
		t.Fatalf("timeout waiting for dev to exit; output so far: %s", buf.String())
	}

	out := buf.String()
	if !strings.Contains(out, "🔁 change detected") {
		t.Fatalf("expected change detected log, got: %s", out)
	}
	if !strings.Contains(out, "▶ restarting…") {
		t.Fatalf("expected restarting log, got: %s", out)
	}
}

func TestDevCtrlCExitsWithoutRestart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nsleep 2\n", 0o755)
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\ntrap 'exit 0' INT TERM\nsleep 5\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, "dev", "--color=never")
	cmd.Dir = dir
	ptmx, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("start pty: %v", err)
	}
	defer func() { _ = ptmx.Close() }()

	_ = readUntilContains(t, ptmx, "🚀 dev started", 2*time.Second)
	_, _ = ptmx.Write([]byte{0x03})
	out := readUntilContains(t, ptmx, "🛑 dev stopped", 2*time.Second)

	if err := cmd.Wait(); err != nil {
		t.Fatalf("expected clean exit, got: %v", err)
	}
	if strings.Contains(out, "▶ restarting…") || strings.Contains(out, "🔁 change detected") {
		t.Fatalf("unexpected restart after Ctrl+C: %s", out)
	}
}

func TestDevCtrlRTriggersRestart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	lockFile := filepath.Join(dir, "child.lock")
	parallel := filepath.Join(dir, "parallel")
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nexit 0\n", 0o755)
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\nlock=\"$RIG_TEST_LOCK\"\nparallel=\"$RIG_TEST_PARALLEL\"\nif [ -f \"$lock\" ]; then\n  echo parallel > \"$parallel\"\n  exit 42\nfi\necho $$ > \"$lock\"\ntrap 'rm -f \"$lock\"; exit 0' INT TERM\nsleep 10\nrm -f \"$lock\"\nexit 0\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, "dev", "--color=never")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "RIG_TEST_LOCK="+lockFile, "RIG_TEST_PARALLEL="+parallel)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("start pty: %v", err)
	}
	defer func() { _ = ptmx.Close() }()

	_ = readUntilContains(t, ptmx, "🚀 dev started", 2*time.Second)
	_, _ = ptmx.Write([]byte{0x12})
	out := readUntilContains(t, ptmx, "🔄 manual reload", 2*time.Second)
	if !strings.Contains(out, "▶ restarting…") {
		_ = readUntilContains(t, ptmx, "▶ restarting…", 2*time.Second)
	}

	_ = cmd.Process.Signal(syscall.SIGTERM)
	_ = cmd.Wait()

	if _, err := os.Stat(parallel); err == nil {
		t.Fatalf("detected parallel child process")
	}
}

func TestDevSigtermStopsCleanly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nexit 0\n", 0o755)
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\ntrap 'exit 0' INT TERM\nsleep 2\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, "dev", "--color=never")
	cmd.Dir = dir
	var buf strings.Builder
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		t.Fatalf("start dev: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	_ = cmd.Process.Signal(syscall.SIGTERM)
	_ = cmd.Wait()
	if !strings.Contains(buf.String(), "🛑 dev stopped") {
		t.Fatalf("expected stop log, got: %s", buf.String())
	}
}

func TestDevNoColorWhenNotTTY(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "./ok"
watch = ["**/*.go"]
`, 0o644)
	writeFile(t, filepath.Join(dir, "ok"), "#!/bin/sh\nexit 0\n", 0o755)
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\nexit 0\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, "dev", "--color=always")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected dev success, got error: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "\x1b[") {
		t.Fatalf("expected no ANSI color in non-TTY output, got: %q", string(out))
	}
}

func TestEmojiAbsentOutsideDevAndJSONUnaffected(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), "[project]\nname = \"t\"\nversion = \"0.0.0\"\n", 0o644)
	writeFile(t, filepath.Join(dir, "rig.lock"), "schema = 0\n", 0o644)

	out, err := runRigCmdInDir(t, dir, "check")
	if err != nil {
		t.Fatalf("expected check success, got error: %v\n%s", err, out)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("expected JSON output, got: %s", out)
	}
	if strings.Contains(out, "🚀") || strings.Contains(out, "🔁") || strings.Contains(out, "🛑") {
		t.Fatalf("unexpected emoji outside dev: %s", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("unexpected ANSI in JSON output: %q", out)
	}
}

func TestRunNoGoModFailsCleanly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tasks]
test = "go test ./..."
`, 0o644)
	writeFile(t, filepath.Join(dir, "rig.lock"), "schema = 0\n", 0o644)

	out, err := runRigCmdInDir(t, dir, "run", "test")
	if err == nil {
		t.Fatalf("expected error, got none. output=%s", out)
	}
	// Ensure we failed due to Go/module state, not due to rig tool sync state.
	if strings.Contains(out, "tools are out of sync") {
		t.Fatalf("unexpected tooling-state error; got: %s", out)
	}
	outLower := strings.ToLower(out)
	if !strings.Contains(outLower, "go.mod") && !strings.Contains(outLower, "main module") {
		t.Fatalf("expected output to mention go.mod/module state, got: %s", out)
	}
}
