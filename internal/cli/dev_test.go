package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	core "github.com/divijg19/rig/internal/rig"
)

func TestDevRuntimeMissingWatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "go run ."
`, 0o644)
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\nexit 0\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	t.Chdir(dir)
	_, err := loadDevRuntime("never", io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "must define 'watch'") {
		t.Fatalf("expected watch error, got: %v", err)
	}
}

func TestDevRuntimeMissingLock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "go run ."
watch = ["**/*.go"]
`, 0o644)

	t.Chdir(dir)
	_, err := loadDevRuntime("never", io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "rig.lock required") {
		t.Fatalf("expected rig.lock error, got: %v", err)
	}
}

func TestDevRuntimeMissingTool(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tasks.dev]
command = "go run ."
watch = ["**/*.go"]
`, 0o644)
	writeRigLock(t, dir, []core.LockedTool{})

	t.Chdir(dir)
	_, err := loadDevRuntime("never", io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "reflex") {
		t.Fatalf("expected reflex tool error, got: %v", err)
	}
}

func TestDevRuntimeMissingDevTask(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tasks]
test = "go test ./..."
`, 0o644)
	writeRigLock(t, dir, []core.LockedTool{})

	t.Chdir(dir)
	_, err := loadDevRuntime("never", io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "[tasks.dev] is required") {
		t.Fatalf("expected missing dev task error, got: %v", err)
	}
}

func TestDevRuntimeHashMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "go run ."
watch = ["**/*.go"]
`, 0o644)
	_, _ = writeTool(t, dir, "reflex", "#!/bin/sh\nexit 0\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntryWithSHA("reflex", "deadbeef")})

	t.Chdir(dir)
	_, err := loadDevRuntime("never", io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "mismatched=1") {
		t.Fatalf("expected hash mismatch error, got: %v", err)
	}
}

func TestDevWatcherConstruction(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"

[tasks.dev]
command = "go run ."
watch = ["*.go", "cmd/**/*.go"]
`, 0o644)
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\nexit 0\n")
	writeRigLock(t, dir, []core.LockedTool{lockToolEntry("reflex", reflexPath, reflexSHA)})

	t.Chdir(dir)
	rt, err := loadDevRuntime("never", io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"-s", "-r", `\.go$`, "--", "sh", "-c", "go run ."}
	if !equalStrings(rt.watcherArgs, want) {
		t.Fatalf("unexpected watcher args: %#v", rt.watcherArgs)
	}
	for _, a := range rt.watcherArgs {
		if a == "-g" {
			t.Fatalf("unexpected glob flag in watcher args: %#v", rt.watcherArgs)
		}
	}
}

func TestComputeWatchRegexOnlyDot(t *testing.T) {
	got := computeWatchRegex([]string{"."})
	if got != "." {
		t.Fatalf("expected dot regex, got %q", got)
	}
}

func TestComputeWatchRegexAlternation(t *testing.T) {
	got := computeWatchRegex([]string{"foo", "bar"})
	if got != "(foo|bar)" {
		t.Fatalf("expected alternation regex, got %q", got)
	}
}

func TestDevCommandIsPassedVerbatim(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script based test")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "rig.toml"), `
[tools]
reflex = "latest"
mockery = "latest"

[tasks.dev]
command = "mockery --help"
watch = ["**/*.go"]
`, 0o644)
	reflexPath, reflexSHA := writeTool(t, dir, "reflex", "#!/bin/sh\nexit 0\n")
	mockeryPath, mockerySHA := writeTool(t, dir, "mockery", "#!/bin/sh\nexit 0\n")
	writeRigLock(t, dir, []core.LockedTool{
		lockToolEntry("reflex", reflexPath, reflexSHA),
		lockToolEntry("mockery", mockeryPath, mockerySHA),
	})

	t.Chdir(dir)
	rt, err := loadDevRuntime("never", io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantSuffix := []string{"--", "sh", "-c", "mockery --help"}
	if len(rt.watcherArgs) < len(wantSuffix) {
		t.Fatalf("unexpected watcher args: %#v", rt.watcherArgs)
	}
	gotSuffix := rt.watcherArgs[len(rt.watcherArgs)-len(wantSuffix):]
	if !equalStrings(gotSuffix, wantSuffix) {
		t.Fatalf("expected watcher args suffix %#v, got %#v", wantSuffix, gotSuffix)
	}
}

func writeTool(t *testing.T, dir string, bin string, content string) (string, string) {
	t.Helper()
	configPath := filepath.Join(dir, "rig.toml")
	path := core.ToolBinPath(configPath, bin)
	writeFile(t, path, content, 0o755)
	sha, err := core.ComputeFileSHA256(path)
	if err != nil {
		t.Fatalf("sha256 %s: %v", bin, err)
	}
	return path, sha
}

func writeRigLock(t *testing.T, dir string, tools []core.LockedTool) {
	t.Helper()
	lock := core.Lockfile{Schema: core.LockSchema0, Tools: tools}
	b, err := core.MarshalLockfile(lock)
	if err != nil {
		t.Fatalf("marshal lock: %v", err)
	}
	writeFile(t, filepath.Join(dir, "rig.lock"), string(b), 0o644)
}

func lockToolEntry(name string, binPath string, sha string) core.LockedTool {
	id := core.ResolveToolIdentity(name)
	return core.LockedTool{
		Kind:      "go-binary",
		Requested: fmt.Sprintf("%s@latest", name),
		Resolved:  fmt.Sprintf("%s@v0.0.0", id.Module),
		Module:    id.Module,
		Bin:       filepath.Base(binPath),
		SHA256:    sha,
	}
}

func lockToolEntryWithSHA(name string, sha string) core.LockedTool {
	id := core.ResolveToolIdentity(name)
	return core.LockedTool{
		Kind:      "go-binary",
		Requested: fmt.Sprintf("%s@latest", name),
		Resolved:  fmt.Sprintf("%s@v0.0.0", id.Module),
		Module:    id.Module,
		Bin:       id.Bin,
		SHA256:    sha,
	}
}

func equalStrings(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
