package config

import (
	"os"
	"path/filepath"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

func write(t *testing.T, path, content string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestLoad_WithIncludes(t *testing.T) {
	dir := t.TempDir()
	main := write(t, filepath.Join(dir, "rig.toml"), `
[project]
name = "test"

include = ["extra.toml"]

[tasks]
build = "go build ."
`)
	_ = main
	// Sanity: ensure include is detectable from TOML
	b, err := os.ReadFile(main)
	if err != nil {
		t.Fatal(err)
	}
	incList := parseIncludeList(b)
	if len(incList) != 1 || incList[0] != "extra.toml" {
		t.Fatalf("include not parsed from base: %v", incList)
	}
	write(t, filepath.Join(dir, "extra.toml"), `
[tasks]
test = "go test ./..."

[tools]
golangci-lint = "1.62.0"

[profile.release]
ldflags = "-s -w"
`)

	c, path, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if path != filepath.Join(dir, "rig.toml") {
		t.Fatalf("unexpected path: %s", path)
	}
	if c.Tasks["build"].Command == "" || c.Tasks["test"].Command == "" {
		t.Fatalf("expected tasks from both files, got %+v", c.Tasks)
	}
	if c.Tools["golangci-lint"] == "" {
		t.Fatalf("expected tool from include")
	}
	if c.Profiles["release"].Ldflags != "-s -w" {
		t.Fatalf("expected profile from include, got %+v", c.Profiles["release"])
	}
}

func TestDecodeIncludeStandalone(t *testing.T) {
	dir := t.TempDir()
	path := write(t, filepath.Join(dir, "inc.toml"), `
[tasks]
test = "go test ./..."

[profile.release]
ldflags = "-s -w"
`)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var incRaw rawConfig
	if err := toml.Unmarshal(data, &incRaw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	inc, err := toTyped(incRaw)
	if err != nil {
		t.Fatalf("toTyped: %v", err)
	}
	if inc.Tasks["test"].Command == "" {
		t.Fatalf("expected task from include, got %+v", inc.Tasks)
	}
	if inc.Profiles["release"].Ldflags != "-s -w" {
		t.Fatalf("expected profile from include, got %+v", inc.Profiles)
	}
}

func TestDecodeBaseIncludes(t *testing.T) {
	dir := t.TempDir()
	p := write(t, filepath.Join(dir, "rig.toml"), `
[project]
name = "x"

include = ["a.toml", "b.toml"]
`)
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	inc := parseIncludeList(data)
	if len(inc) != 2 || inc[0] != "a.toml" || inc[1] != "b.toml" {
		t.Fatalf("expected includes parsed, got %#v", inc)
	}
}
