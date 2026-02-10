package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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

func installEntrypoints(t *testing.T, outDir string) map[string]string {
	t.Helper()
	root := projectRoot(t)
	cmd := exec.Command("go", "install", "./cmd/...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOBIN="+outDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go install ./cmd/... failed: %v\n%s", err, out)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	paths := map[string]string{
		"rig": filepath.Join(outDir, "rig"+ext),
		"rir": filepath.Join(outDir, "rir"+ext),
		"ric": filepath.Join(outDir, "ric"+ext),
		"rid": filepath.Join(outDir, "rid"+ext),
		"ris": filepath.Join(outDir, "ris"+ext),
	}
	for name, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %s entrypoint to exist at %s: %v", name, p, err)
		}
	}
	return paths
}

func runRigCmdInDir(t *testing.T, dir string, args ...string) (string, error) {
	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
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
	writeFile(t, filepath.Join(dir, "rig.lock"), `schema = 0

[[tools]]
kind = "go-binary"
requested = "mockery@2.0.0"
resolved = "github.com/vektra/mockery/v2@v2.0.0"
module = "github.com/vektra/mockery/v2"
bin = "mockery"
`, 0o644)

	bin := filepath.Join(dir, ".rig", "bin", "mockery")
	writeFile(t, bin, "#!/bin/sh\necho mockery v2.0.0\n", 0o755)

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
	writeFile(t, filepath.Join(dir, "rig.lock"), `schema = 0

[[tools]]
kind = "go-binary"
requested = "mockery@2.0.0"
resolved = "github.com/vektra/mockery/v2@v2.0.0"
module = "github.com/vektra/mockery/v2"
bin = "mockery"
`, 0o644)
	writeFile(t, filepath.Join(dir, ".rig", "bin", "mockery"), "#!/bin/sh\necho mockery v2.0.0\n", 0o755)

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

	binDir := t.TempDir()
	paths := installEntrypoints(t, binDir)
	rigBin := paths["rig"]
	rirBin := paths["rir"]

	cmd1 := exec.Command(rigBin, "run", "--list")
	cmd1.Dir = work
	out1, err := cmd1.CombinedOutput()
	if err != nil {
		t.Fatalf("rig run --list failed: %v\n%s", err, out1)
	}

	cmd2 := exec.Command(rirBin, "--list")
	cmd2.Dir = work
	out2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("rir --list failed: %v\n%s", err, out2)
	}
	if strings.TrimSpace(string(out1)) != strings.TrimSpace(string(out2)) {
		t.Fatalf("outputs differ\nrig: %q\nrir: %q", strings.TrimSpace(string(out1)), strings.TrimSpace(string(out2)))
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

	binDir := t.TempDir()
	paths := installEntrypoints(t, binDir)
	rigBin := paths["rig"]
	ricBin := paths["ric"]

	cmd1 := exec.Command(rigBin, "check")
	cmd1.Dir = work
	out1, err := cmd1.CombinedOutput()
	if err == nil {
		t.Fatalf("expected rig check to fail without lock, got none. output=%s", out1)
	}
	if !strings.Contains(string(out1), "\"ok\":false") {
		t.Fatalf("expected JSON ok=false, got: %s", out1)
	}

	cmd2 := exec.Command(ricBin)
	cmd2.Dir = work
	out2, err := cmd2.CombinedOutput()
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

	binDir := t.TempDir()
	paths := installEntrypoints(t, binDir)
	rigBin := paths["rig"]
	ridBin := paths["rid"]

	cmd1 := exec.Command(rigBin, "dev")
	cmd1.Dir = work
	out1, err := cmd1.CombinedOutput()
	if err == nil {
		t.Fatalf("expected rig dev to fail without lock, got none. output=%s", out1)
	}

	cmd2 := exec.Command(ridBin)
	cmd2.Dir = work
	out2, err := cmd2.CombinedOutput()
	if err == nil {
		t.Fatalf("expected rid invocation to fail without lock, got none. output=%s", out2)
	}
	if strings.TrimSpace(string(out1)) != strings.TrimSpace(string(out2)) {
		t.Fatalf("outputs differ\nrig dev: %q\nrid: %q", strings.TrimSpace(string(out1)), strings.TrimSpace(string(out2)))
	}
}

func TestEntrypointRisMatchesRigStartStub(t *testing.T) {
	work := t.TempDir()
	binDir := t.TempDir()
	paths := installEntrypoints(t, binDir)
	rigBin := paths["rig"]
	risBin := paths["ris"]

	cmd1 := exec.Command(rigBin, "start")
	cmd1.Dir = work
	out1, err := cmd1.CombinedOutput()
	if err == nil {
		t.Fatalf("expected rig start to fail (stub), got none. output=%s", out1)
	}

	cmd2 := exec.Command(risBin)
	cmd2.Dir = work
	out2, err := cmd2.CombinedOutput()
	if err == nil {
		t.Fatalf("expected ris invocation to fail (stub), got none. output=%s", out2)
	}
	if strings.TrimSpace(string(out1)) != strings.TrimSpace(string(out2)) {
		t.Fatalf("outputs differ\nrig start: %q\nris: %q", strings.TrimSpace(string(out1)), strings.TrimSpace(string(out2)))
	}
}

func TestAliasCommandOutputIsStable(t *testing.T) {
	work := t.TempDir()
	binDir := t.TempDir()
	paths := installEntrypoints(t, binDir)
	rigBin := paths["rig"]

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
