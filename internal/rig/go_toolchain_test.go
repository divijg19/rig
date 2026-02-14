package rig

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNormalizeGoToolchainRequested(t *testing.T) {
	tc := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"1.22.3", "1.22.3", false},
		{"go1.22.3", "1.22.3", false},
		{"v1.22.3", "1.22.3", false},
		{"  1.22.3 ", "1.22.3", false},
		{"1.22", "", true},
		{"", "", true},
	}
	for _, c := range tc {
		got, err := NormalizeGoToolchainRequested(c.in)
		if (err != nil) != c.wantErr {
			t.Fatalf("NormalizeGoToolchainRequested(%q) err=%v wantErr=%v", c.in, err, c.wantErr)
		}
		if err == nil && got != c.want {
			t.Fatalf("NormalizeGoToolchainRequested(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestParseGoToolchainDetectedFromGoVersionOutput(t *testing.T) {
	out := "go version go1.23.4 linux/amd64"
	got, err := ParseGoToolchainDetectedFromGoVersionOutput(out)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got != "1.23.4" {
		t.Fatalf("got %q", got)
	}
}

func TestCheckAndRunValidateGoToolchain(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a shebang script as a fake go binary")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rig.toml"), []byte(strings.Join([]string{
		"[project]",
		"name='tmp'",
		"version='0.0.0'",
		"",
		"[tasks]",
		"hello='go version'",
		"",
		"[tools]",
		"go='1.23.4'",
	}, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write rig.toml: %v", err)
	}

	lock := strings.Join([]string{
		"schema = 0",
		"",
		"[toolchain.go]",
		"kind = \"go-toolchain\"",
		"requested = \"1.23.4\"",
		"detected = \"1.23.4\"",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, "rig.lock"), []byte(lock), 0o644); err != nil {
		t.Fatalf("write rig.lock: %v", err)
	}

	binDir := filepath.Join(dir, "fakebin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir fakebin: %v", err)
	}
	goPath := filepath.Join(binDir, "go")
	go1234 := "#!/bin/sh\necho 'go version go1.23.4 linux/amd64'\n"
	if err := os.WriteFile(goPath, []byte(go1234), 0o755); err != nil {
		t.Fatalf("write fake go: %v", err)
	}
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)

	rep, err := Check(dir)
	if err != nil {
		t.Fatalf("Check err: %v", err)
	}
	if !rep.OK {
		t.Fatalf("expected OK; got: %#v", rep)
	}
	if rep.Go == nil || rep.Go.Status != "ok" {
		t.Fatalf("expected go ok; got: %#v", rep.Go)
	}

	if err := Run(dir, "hello", nil); err != nil {
		t.Fatalf("Run err: %v", err)
	}

	go1235 := "#!/bin/sh\necho 'go version go1.23.5 linux/amd64'\n"
	if err := os.WriteFile(goPath, []byte(go1235), 0o755); err != nil {
		t.Fatalf("update fake go: %v", err)
	}

	rep2, err := Check(dir)
	if err != nil {
		t.Fatalf("Check(2) err: %v", err)
	}
	if rep2.OK {
		t.Fatalf("expected NOT ok")
	}
	if rep2.Go == nil || rep2.Go.Status != "mismatch" {
		t.Fatalf("expected go mismatch; got: %#v", rep2.Go)
	}

	if err := Run(dir, "hello", nil); err == nil {
		t.Fatalf("expected Run failure due to go mismatch")
	}
}
