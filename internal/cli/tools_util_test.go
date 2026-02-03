package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestComputeToolsHashDeterministic(t *testing.T) {
	a := map[string]string{"golangci-lint": "v1.59.0", "mockery": "v2.42.0"}
	b := map[string]string{"mockery": "v2.42.0", "golangci-lint": "v1.59.0"}
	ha := computeToolsHash(a)
	hb := computeToolsHash(b)
	if ha != hb {
		t.Fatalf("hash should be deterministic regardless of map order: %s vs %s", ha, hb)
	}
}

func TestParseToolsFilesBasic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tools.txt")
	content := strings.Join([]string{
		"# comment",
		"golangci-lint = v1.59.0",
		"github.com/vektra/mockery/v2@v2.42.0",
		"air",
		"",
	}, "\n")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp tools.txt: %v", err)
	}
	m, err := parseToolsFiles([]string{p})
	if err != nil {
		t.Fatalf("parse tools: %v", err)
	}
	if len(m) != 3 {
		t.Fatalf("expected 3 tools, got %d: %#v", len(m), m)
	}
	if m["golangci-lint"] != "v1.59.0" {
		t.Fatalf("unexpected version for golangci-lint: %s", m["golangci-lint"])
	}
	if m["github.com/vektra/mockery/v2"] != "v2.42.0" {
		t.Fatalf("unexpected version for mockery: %s", m["github.com/vektra/mockery/v2"])
	}
	if m["air"] != "latest" {
		t.Fatalf("unexpected default version for air: %s", m["air"])
	}
}

func TestMergeToolsOverlay(t *testing.T) {
	base := map[string]string{"a": "1", "b": "1"}
	extra := map[string]string{"b": "2", "c": "3"}
	got := mergeTools(base, extra)
	if got["a"] != "1" || got["b"] != "2" || got["c"] != "3" {
		t.Fatalf("unexpected merge result: %#v", got)
	}
	if base["b"] != "1" {
		t.Fatalf("base map mutated: %#v", base)
	}
	keys := make([]string, 0, len(got))
	for k := range got {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	expected := []string{"a", "b", "c"}
	if !slices.Equal(keys, expected) {
		t.Fatalf("unexpected keys: %#v", keys)
	}
}

func TestOutdatedJSONNoToolsPrintsEmptyArray(t *testing.T) {
	dir := t.TempDir()
	rigToml := "[project]\nname='tmp'\nversion='0.0.0'\n"
	if err := os.WriteFile(filepath.Join(dir, "rig.toml"), []byte(rigToml), 0o644); err != nil {
		t.Fatalf("write rig.toml: %v", err)
	}
	out, err := runRig(dir, "tools", "outdated", "--json")
	if err != nil {
		t.Fatalf("rig tools outdated --json failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Fatalf("expected empty JSON array, got: %q", strings.TrimSpace(out))
	}
}

func TestSyncCheckJSONWhenInSyncPrintsEmptySummary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a shebang script as a fake tool binary")
	}
	dir, err := os.MkdirTemp(projectRootForTest(), "rig-test-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	rigToml := "[project]\nname='tmp'\nversion='0.0.0'\n[tools]\nmockery='v2.46.0'\n"
	if err := os.WriteFile(filepath.Join(dir, "rig.toml"), []byte(rigToml), 0o644); err != nil {
		t.Fatalf("write rig.toml: %v", err)
	}
	// Fake installed binary in .rig/bin that reports the expected version.
	binDir := filepath.Join(dir, ".rig", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir .rig/bin: %v", err)
	}
	fake := "#!/bin/sh\necho \"mockery v2.46.0\"\n"
	fakePath := filepath.Join(binDir, "mockery")
	if err := os.WriteFile(fakePath, []byte(fake), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
	// Deterministic rig.lock matching rig.toml.
	rigLock := "schema = 0\n\n[[tools]]\nkind = \"go\"\nrequested = \"mockery@v2.46.0\"\nresolved = \"github.com/vektra/mockery/v2@v2.46.0\"\nmodule = \"github.com/vektra/mockery/v2\"\n"
	if err := os.WriteFile(filepath.Join(dir, "rig.lock"), []byte(rigLock), 0o644); err != nil {
		t.Fatalf("write rig.lock: %v", err)
	}
	out, err := runRig(dir, "tools", "sync", "--check", "--json")
	if err != nil {
		t.Fatalf("rig tools sync --check --json failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "\"missing\": 0") || !strings.Contains(out, "\"mismatched\": 0") || !strings.Contains(out, "\"extra\": 0") {
		t.Fatalf("expected zero counts in summary, got: %s", out)
	}
}

func TestSyncOfflineDoesNotWriteLockOnResolveFailure(t *testing.T) {
	// This test is hermetic: GOPROXY=off ensures no network access and go list fails fast.
	dir, err := os.MkdirTemp(projectRootForTest(), "rig-test-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	rigToml := "[project]\nname='tmp'\nversion='0.0.0'\n[tools]\nexample.com/does/not/exist='v0.0.0'\n"
	if err := os.WriteFile(filepath.Join(dir, "rig.toml"), []byte(rigToml), 0o644); err != nil {
		t.Fatalf("write rig.toml: %v", err)
	}

	out, runErr := runRig(dir, "tools", "sync", "--offline")
	if runErr == nil {
		t.Fatalf("expected failure, got success\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "rig.lock")); err == nil {
		t.Fatalf("rig.lock should not be written on failure")
	}
	if _, err := os.Stat(manifestLockPath(filepath.Join(dir, "rig.toml"))); err == nil {
		t.Fatalf("manifest.lock should not be written on failure")
	}
}

func runRig(dir string, args ...string) (string, error) {
	binDir := dir
	binName := "rig"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)
	build := exec.Command("go", "build", "-o", binPath, "./cmd/rig")
	build.Dir = projectRootForTest()
	if out, err := build.CombinedOutput(); err != nil {
		return string(out), err
	}
	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func projectRootForTest() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}
