package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitDefaults(t *testing.T) {
	dir := t.TempDir()

	out, err := runRigCmdInDir(t, dir, "init", "--yes")
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, "rig.toml"))
	if err != nil {
		t.Fatalf("read rig.toml: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "[project]") {
		t.Fatalf("expected [project], got:\n%s", content)
	}
	if !strings.Contains(content, "[tasks]") {
		t.Fatalf("expected [tasks], got:\n%s", content)
	}
	if !strings.Contains(content, "[tools]") {
		t.Fatalf("expected [tools], got:\n%s", content)
	}
	if !strings.Contains(content, "go = \"") {
		t.Fatalf("expected go toolchain pin, got:\n%s", content)
	}
	if strings.Contains(content, "[tasks.dev]") {
		t.Fatalf("did not expect [tasks.dev] in defaults, got:\n%s", content)
	}
	if strings.Contains(content, "reflex = \"latest\"") {
		t.Fatalf("did not expect reflex in defaults, got:\n%s", content)
	}
}

func TestInitDev(t *testing.T) {
	dir := t.TempDir()
	out, err := runRigCmdInDir(t, dir, "init", "--yes", "--dev")
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, "rig.toml"))
	if err != nil {
		t.Fatalf("read rig.toml: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "[tasks.dev]") {
		t.Fatalf("expected [tasks.dev] in rig.toml, got:\n%s", content)
	}
	if !strings.Contains(content, "command = \"go run .\"") {
		t.Fatalf("expected dev command in rig.toml, got:\n%s", content)
	}
	if !strings.Contains(content, "watch = [\"**/*.go\"]") {
		t.Fatalf("expected dev watch in rig.toml, got:\n%s", content)
	}
	if !strings.Contains(content, "go = \"") {
		t.Fatalf("expected go toolchain pin, got:\n%s", content)
	}
	if !strings.Contains(content, "reflex = \"latest\"") {
		t.Fatalf("expected reflex in [tools], got:\n%s", content)
	}
}

func TestInitMinimal(t *testing.T) {
	dir := t.TempDir()
	out, err := runRigCmdInDir(t, dir, "init", "--yes", "--minimal")
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, "rig.toml"))
	if err != nil {
		t.Fatalf("read rig.toml: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "[project]") {
		t.Fatalf("expected [project], got:\n%s", content)
	}
	if !strings.Contains(content, "[tools]") {
		t.Fatalf("expected [tools], got:\n%s", content)
	}
	if !strings.Contains(content, "go = \"") {
		t.Fatalf("expected go toolchain pin, got:\n%s", content)
	}
	if strings.Contains(content, "[tasks]") || strings.Contains(content, "[tasks.") {
		t.Fatalf("did not expect tasks in --minimal, got:\n%s", content)
	}
	if strings.Contains(content, "[profile.") {
		t.Fatalf("did not expect profiles in --minimal, got:\n%s", content)
	}
}

func TestInitCI(t *testing.T) {
	dir := t.TempDir()
	out, err := runRigCmdInDir(t, dir, "init", "--yes", "--ci")
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, "rig.toml"))
	if err != nil {
		t.Fatalf("read rig.toml: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "[tasks.ci]") {
		t.Fatalf("expected [tasks.ci], got:\n%s", content)
	}
	if !strings.Contains(content, "command = \"rig check && rig run test\"") {
		t.Fatalf("expected CI command, got:\n%s", content)
	}
}

func TestInitMonorepo(t *testing.T) {
	dir := t.TempDir()
	out, err := runRigCmdInDir(t, dir, "init", "--yes", "--monorepo")
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, "rig.toml"))
	if err != nil {
		t.Fatalf("read rig.toml: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "include = [\"rig.tasks.toml\", \"rig.tools.toml\"]") {
		t.Fatalf("expected include entries in rig.toml, got:\n%s", content)
	}
	tasksBytes, err := os.ReadFile(filepath.Join(dir, ".rig", "rig.tasks.toml"))
	if err != nil {
		t.Fatalf("read .rig/rig.tasks.toml: %v", err)
	}
	if !strings.Contains(string(tasksBytes), "[tasks]") {
		t.Fatalf("expected [tasks] in monorepo tasks file, got:\n%s", string(tasksBytes))
	}
	toolsBytes, err := os.ReadFile(filepath.Join(dir, ".rig", "rig.tools.toml"))
	if err != nil {
		t.Fatalf("read .rig/rig.tools.toml: %v", err)
	}
	if !strings.Contains(string(toolsBytes), "[tools]") || !strings.Contains(string(toolsBytes), "go = \"") {
		t.Fatalf("expected [tools] with go pin in monorepo tools file, got:\n%s", string(toolsBytes))
	}
}

func TestInitDevAndCI(t *testing.T) {
	dir := t.TempDir()
	out, err := runRigCmdInDir(t, dir, "init", "--yes", "--dev", "--ci")
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, "rig.toml"))
	if err != nil {
		t.Fatalf("read rig.toml: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "[tasks.dev]") || !strings.Contains(content, "[tasks.ci]") {
		t.Fatalf("expected both dev and ci tasks, got:\n%s", content)
	}
	if !strings.Contains(content, "reflex = \"latest\"") {
		t.Fatalf("expected reflex for --dev, got:\n%s", content)
	}
}

func TestInitHelpFlagSurface(t *testing.T) {
	dir := t.TempDir()
	out, err := runRigCmdInDir(t, dir, "init", "--help")
	if err != nil {
		t.Fatalf("init --help failed: %v\n%s", err, out)
	}
	for _, flag := range []string{
		"--help",
		"--dir string",
		"--dev",
		"--minimal",
		"--ci",
		"--monorepo",
		"--force",
		"--name string",
		"--license string",
		"--version string",
		"--yes",
	} {
		if !strings.Contains(out, flag) {
			t.Fatalf("expected help to contain %q, got:\n%s", flag, out)
		}
	}
	for _, legacy := range []string{"--developer", "--dev-watcher", "--no-tools", "--no-tasks", "--profiles", "--dx"} {
		if strings.Contains(out, legacy) {
			t.Fatalf("expected help to exclude legacy flag %q, got:\n%s", legacy, out)
		}
	}
}

func TestInitCreatesGitignoreWithRigEntry(t *testing.T) {
	dir := t.TempDir()
	out, err := runRigCmdInDir(t, dir, "init", "--yes")
	if err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if string(b) != ".rig/\n" {
		t.Fatalf("expected .gitignore to be exactly .rig/, got:\n%s", string(b))
	}
}

func TestInitAppendsGitignoreRigEntryWithoutDuplicates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".gitignore"), "bin/\n", 0o644)

	out, err := runRigCmdInDir(t, dir, "init", "--yes")
	if err != nil {
		t.Fatalf("first init failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore after first init: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, "bin/\n") || !strings.Contains(got, ".rig/\n") {
		t.Fatalf("expected existing lines preserved and .rig/ appended, got:\n%s", got)
	}
	if strings.Count(got, ".rig/\n") != 1 {
		t.Fatalf("expected .rig/ exactly once after first init, got:\n%s", got)
	}

	out, err = runRigCmdInDir(t, dir, "init", "--yes", "--force")
	if err != nil {
		t.Fatalf("second init with --force failed: %v\n%s", err, out)
	}
	b, err = os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore after second init: %v", err)
	}
	got = string(b)
	if strings.Count(got, ".rig/\n") != 1 {
		t.Fatalf("expected no duplicate .rig/ entry, got:\n%s", got)
	}
}
