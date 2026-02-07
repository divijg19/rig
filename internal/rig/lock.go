package rig

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const LockSchema0 = 0

// LockedTool is a single tool entry in rig.lock.
//
// Contract (schema = 0):
// - schema = 0
// - tools are sorted lexicographically by Requested
// - fields are written in a fixed order
// - module and url are mutually exclusive
// - checksum is optional
//
// Notes:
//   - requested and resolved are intentionally opaque strings; they are meant to be
//     human-inspectable and stable for diffs.
//   - kind is currently "go-binary" for installable tools.
//
// TOML layout:
//
//	schema = 0
//
//	[[tools]]
//	kind = "go-binary"
//	requested = "golangci-lint@1.62.0"
//	resolved = "github.com/golangci/golangci-lint@v1.62.0"
//	module = "github.com/golangci/golangci-lint"
//	bin = "golangci-lint"
//	checksum = "h1:..." # optional
//
// (No comments are generated in the lock file.)
type LockedTool struct {
	Kind      string `toml:"kind"`
	Requested string `toml:"requested"`
	Resolved  string `toml:"resolved"`

	Module   string `toml:"module,omitempty"`
	Bin      string `toml:"bin,omitempty"`
	URL      string `toml:"url,omitempty"`
	Checksum string `toml:"checksum,omitempty"`
}

// GoToolchainLock captures the Go toolchain requirement for this repo.
//
// TOML layout (schema = 0):
//
//	[toolchain.go]
//	kind = "go-toolchain"
//	requested = "1.22.0"
//	detected = "1.22.0"
type GoToolchainLock struct {
	Kind      string `toml:"kind"`
	Requested string `toml:"requested"`
	Detected  string `toml:"detected"`
}

type ToolchainLock struct {
	Go *GoToolchainLock `toml:"go,omitempty"`
}

type Lockfile struct {
	Schema    int            `toml:"schema"`
	Toolchain *ToolchainLock `toml:"toolchain,omitempty"`
	Tools     []LockedTool   `toml:"tools"`
}

func (t LockedTool) validate() error {
	if strings.TrimSpace(t.Kind) == "" {
		return errors.New("tool.kind is required")
	}
	if strings.TrimSpace(t.Requested) == "" {
		return errors.New("tool.requested is required")
	}
	if strings.TrimSpace(t.Resolved) == "" {
		return errors.New("tool.resolved is required")
	}
	if t.Module != "" && t.URL != "" {
		return errors.New("tool.module and tool.url are mutually exclusive")
	}
	return nil
}

func ValidateLockfile(l Lockfile) error {
	if l.Schema != LockSchema0 {
		return fmt.Errorf("unsupported rig.lock schema %d", l.Schema)
	}
	if l.Toolchain != nil && l.Toolchain.Go != nil {
		gt := l.Toolchain.Go
		if strings.TrimSpace(gt.Kind) == "" {
			return errors.New("toolchain.go.kind is required")
		}
		if strings.TrimSpace(gt.Kind) != "go-toolchain" {
			return fmt.Errorf("toolchain.go.kind must be %q", "go-toolchain")
		}
		if strings.TrimSpace(gt.Requested) == "" {
			return errors.New("toolchain.go.requested is required")
		}
		if strings.TrimSpace(gt.Detected) == "" {
			return errors.New("toolchain.go.detected is required")
		}
	}
	for i, t := range l.Tools {
		if err := t.validate(); err != nil {
			return fmt.Errorf("tools[%d]: %w", i, err)
		}
	}
	return nil
}

func ReadLockfile(path string) (Lockfile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Lockfile{}, err
	}
	var l Lockfile
	if err := toml.Unmarshal(b, &l); err != nil {
		return Lockfile{}, fmt.Errorf("parse rig.lock: %w", err)
	}
	if err := ValidateLockfile(l); err != nil {
		return Lockfile{}, err
	}
	return l, nil
}

// MarshalLockfile renders a lockfile deterministically.
// It does not rely on the TOML encoder to avoid nondeterministic map ordering.
func MarshalLockfile(l Lockfile) ([]byte, error) {
	if err := ValidateLockfile(l); err != nil {
		return nil, err
	}

	tools := make([]LockedTool, 0, len(l.Tools))
	tools = append(tools, l.Tools...)
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Requested < tools[j].Requested
	})

	var buf bytes.Buffer
	buf.WriteString("schema = 0\n")

	if l.Toolchain != nil && l.Toolchain.Go != nil {
		buf.WriteString("\n")
		buf.WriteString("[toolchain.go]\n")
		writeTOMLKV(&buf, "kind", l.Toolchain.Go.Kind)
		writeTOMLKV(&buf, "requested", l.Toolchain.Go.Requested)
		writeTOMLKV(&buf, "detected", l.Toolchain.Go.Detected)
	}

	if len(tools) > 0 {
		buf.WriteString("\n")
	}

	for i, t := range tools {
		buf.WriteString("[[tools]]\n")
		writeTOMLKV(&buf, "kind", t.Kind)
		writeTOMLKV(&buf, "requested", t.Requested)
		writeTOMLKV(&buf, "resolved", t.Resolved)
		if t.Module != "" {
			writeTOMLKV(&buf, "module", t.Module)
		} else if t.URL != "" {
			writeTOMLKV(&buf, "url", t.URL)
		}
		if t.Bin != "" {
			writeTOMLKV(&buf, "bin", t.Bin)
		}
		if t.Checksum != "" {
			writeTOMLKV(&buf, "checksum", t.Checksum)
		}
		if i != len(tools)-1 {
			buf.WriteString("\n")
		}
	}

	// Always end with a trailing newline.
	if buf.Len() == 0 || !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteString("\n")
	}
	return buf.Bytes(), nil
}

func writeTOMLKV(buf *bytes.Buffer, key, value string) {
	buf.WriteString(key)
	buf.WriteString(" = ")
	buf.WriteString(tomlQuote(value))
	buf.WriteString("\n")
}

func tomlQuote(s string) string {
	repl := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	)
	return "\"" + repl.Replace(s) + "\""
}

// WriteLockfile writes rig.lock to path by fully overwriting it.
// The write is atomic (write to a temp file in the same directory then rename).
func WriteLockfile(path string, l Lockfile) error {
	b, err := MarshalLockfile(l)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "rig.lock.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return nil
}
