package cli

import (
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

func runRigCmd(t *testing.T, args ...string) string {
	root := projectRoot(t)
	bin := buildRigBinary(t, t.TempDir())
	cmd := exec.Command(bin, args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rig %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func TestRigBuildDryRun(t *testing.T) {
	out := runRigCmd(t, "build", "--dry-run")
	if !strings.Contains(out, "🧪 Dry run: would execute -> go build") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestRigRunList(t *testing.T) {
	out := runRigCmd(t, "run", "--list")
	if !strings.Contains(out, "📝 Tasks in") {
		t.Errorf("expected task list output, got: %s", out)
	}
}

func TestRigSetupHelp(t *testing.T) {
	out := runRigCmd(t, "setup", "--help")
	if !strings.Contains(out, "Reads [tools] from rig.toml") {
		t.Errorf("expected setup help output, got: %s", out)
	}
}

func TestRigDoctorHelp(t *testing.T) {
	out := runRigCmd(t, "doctor", "--help")
	if !strings.Contains(out, "Verifies the Go toolchain") {
		t.Errorf("expected doctor help output, got: %s", out)
	}
}

func TestRigXHelp(t *testing.T) {
	out := runRigCmd(t, "x", "--help")
	if !strings.Contains(out, "Run a tool ephemerally") {
		t.Errorf("expected x help output, got: %s", out)
	}
}
