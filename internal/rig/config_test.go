package rig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_AllowsTaskDescription(t *testing.T) {
	dir := t.TempDir()
	config := `
[tasks]
alpha = { command = "echo hi", description = "  Says hi  " }

[tasks.dev]
command = "go run ."
watch = ["**/*.go"]
`
	if err := os.WriteFile(filepath.Join(dir, "rig.toml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write rig.toml: %v", err)
	}

	conf, _, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if got := conf.Tasks["alpha"].Description; got != "Says hi" {
		t.Fatalf("alpha.description=%q, want %q", got, "Says hi")
	}
	if got := strings.Join(conf.Tasks["dev"].Watch, ","); got != "**/*.go" {
		t.Fatalf("dev.watch=%q, want %q", got, "**/*.go")
	}
}

func TestLoadConfig_RejectsUnknownTaskField(t *testing.T) {
	dir := t.TempDir()
	config := `
[tasks]
bad = { command = "echo hi", no_such_field = "nope" }
`
	if err := os.WriteFile(filepath.Join(dir, "rig.toml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write rig.toml: %v", err)
	}

	_, _, err := LoadConfig(dir)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported field") {
		t.Fatalf("expected unsupported field error, got: %v", err)
	}
}
