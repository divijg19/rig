package rig

import (
	"strings"
	"testing"

	cfg "github.com/divijg19/rig/internal/config"
)

func TestComposeBuildCommand_ProfileAndOverrides(t *testing.T) {
	prof := cfg.BuildProfile{
		Ldflags: "-s -w",
		Gcflags: "all=-N -l",
		Tags:    []string{"prod"},
		Flags:   []string{"-trimpath"},
		Env:     map[string]string{"FOO": "bar"},
		Output:  "bin/app",
	}
	cmd, env := ComposeBuildCommand(prof, BuildOverrides{
		Output:  "out/rig.exe",
		Tags:    []string{"prod", "win"},
		Ldflags: "-X main.version=1.0.0",
		Gcflags: "",
	})

	if !strings.Contains(cmd, "go build") {
		t.Fatalf("expected go build in cmd, got %s", cmd)
	}
	// On Windows, filepath.Clean uses backslashes
	if !strings.Contains(cmd, "-o \"out\\rig.exe\"") && !strings.Contains(cmd, "-o \"out/rig.exe\"") {
		t.Errorf("expected output override, got %s", cmd)
	}
	if !strings.Contains(cmd, "-ldflags \"-X main.version=1.0.0\"") {
		t.Errorf("expected ldflags override, got %s", cmd)
	}
	// Empty override should preserve profile gcflags
	if !strings.Contains(cmd, "-gcflags \"all=-N -l\"") {
		t.Errorf("expected gcflags from profile, got %s", cmd)
	}
	if !strings.Contains(cmd, "-tags \"prod,win\"") {
		t.Errorf("expected merged tags, got %s", cmd)
	}
	if !strings.Contains(cmd, "-trimpath") {
		t.Errorf("expected profile flags appended, got %s", cmd)
	}
	if len(env) != 1 || env[0] != "FOO=bar" {
		t.Errorf("expected env from profile, got %v", env)
	}
}
