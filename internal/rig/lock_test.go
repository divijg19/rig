package rig

import (
	"os"
	"strings"
	"testing"
)

func TestMarshalLockfileDeterministicOrderingAndFields(t *testing.T) {
	l := Lockfile{
		Schema: LockSchema0,
		Tools: []LockedTool{
			{
				Kind:      "go-binary",
				Requested: "zeta@latest",
				Resolved:  "example.com/zeta@v1.0.0",
				Module:    "example.com/zeta",
				Bin:       "zeta",
				SHA256:    "00",
			},
			{
				Kind:      "go-binary",
				Requested: "aardvark@v1.2.3",
				Resolved:  "example.com/aardvark@v1.2.3",
				Module:    "example.com/aardvark",
				Bin:       "aardvark",
				Checksum:  "h1:abc",
				SHA256:    "11",
			},
		},
	}

	b1, err := MarshalLockfile(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	b2, err := MarshalLockfile(l)
	if err != nil {
		t.Fatalf("marshal(2): %v", err)
	}
	if string(b1) != string(b2) {
		t.Fatalf("marshal not deterministic:\n%s\n---\n%s", b1, b2)
	}

	out := string(b1)
	if !strings.HasPrefix(out, "schema = 0\n") {
		t.Fatalf("expected schema header, got: %q", out)
	}
	// Tools must be sorted by requested (aardvark before zeta).
	idxAardvark := strings.Index(out, "requested = \"aardvark@v1.2.3\"")
	idxZeta := strings.Index(out, "requested = \"zeta@latest\"")
	if idxAardvark < 0 || idxZeta < 0 || idxAardvark > idxZeta {
		t.Fatalf("expected tools sorted by requested; got:\n%s", out)
	}

	// Fixed field order inside a tool.
	aardvarkBlockStart := strings.Index(out, "[[tools]]\n")
	if aardvarkBlockStart < 0 {
		t.Fatalf("expected tools blocks")
	}
	aardvarkBlock := out[aardvarkBlockStart:]
	wantOrder := []string{
		"kind = \"go-binary\"\n",
		"requested = \"aardvark@v1.2.3\"\n",
		"resolved = \"example.com/aardvark@v1.2.3\"\n",
		"module = \"example.com/aardvark\"\n",
		"bin = \"aardvark\"\n",
		"checksum = \"h1:abc\"\n",
		"sha256 = \"11\"\n",
	}
	pos := 0
	for _, needle := range wantOrder {
		i := strings.Index(aardvarkBlock[pos:], needle)
		if i < 0 {
			t.Fatalf("missing field %q in output:\n%s", needle, out)
		}
		pos += i + len(needle)
	}
}

func TestReadLockfileRoundTrip(t *testing.T) {
	l := Lockfile{Schema: LockSchema0, Tools: []LockedTool{{
		Kind:      "go-binary",
		Requested: "mockery@latest",
		Resolved:  "github.com/vektra/mockery/v2@v2.46.0",
		Module:    "github.com/vektra/mockery/v2",
		Bin:       "mockery",
		SHA256:    "aa",
	}}}

	b, err := MarshalLockfile(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed Lockfile
	if err := ValidateLockfile(l); err != nil {
		t.Fatalf("validate: %v", err)
	}
	// Parse via toml to ensure the file is valid TOML.
	parsed, err = func() (Lockfile, error) {
		// Use the internal reader path by writing to disk in a temp dir.
		dir := t.TempDir()
		p := dir + "/rig.lock"
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return Lockfile{}, err
		}
		return ReadLockfile(p)
	}()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if parsed.Schema != LockSchema0 {
		t.Fatalf("schema=%d", parsed.Schema)
	}
	if len(parsed.Tools) != 1 || parsed.Tools[0].Requested != "mockery@latest" {
		t.Fatalf("unexpected parsed lock: %#v", parsed)
	}
}
