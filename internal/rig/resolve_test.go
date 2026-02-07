package rig

import (
	"reflect"
	"testing"
)

func TestResolveLockedToolsUsesModuleRootForResolution(t *testing.T) {
	old := goListModuleVersion
	t.Cleanup(func() { goListModuleVersion = old })

	var gotModule string
	var gotVersion string
	goListModuleVersion = func(module, version, workDir string, env []string) (string, string, error) {
		gotModule = module
		gotVersion = version
		return "v1.59.1", "h1:sum", nil
	}

	locked, err := ResolveLockedTools(map[string]string{"golangci-lint": "1.59.1"}, "", nil)
	if err != nil {
		t.Fatalf("ResolveLockedTools: %v", err)
	}
	if gotModule != "github.com/golangci/golangci-lint" {
		t.Fatalf("module=%q", gotModule)
	}
	if gotVersion != "v1.59.1" {
		t.Fatalf("version=%q", gotVersion)
	}
	if len(locked) != 1 {
		t.Fatalf("len=%d", len(locked))
	}
	want := LockedTool{
		Kind:      "go-binary",
		Requested: "golangci-lint@1.59.1",
		Resolved:  "github.com/golangci/golangci-lint@v1.59.1",
		Module:    "github.com/golangci/golangci-lint",
		Bin:       "golangci-lint",
		Checksum:  "h1:sum",
	}
	if !reflect.DeepEqual(locked[0], want) {
		t.Fatalf("locked[0]=%#v\nwant=%#v", locked[0], want)
	}
}

func TestResolveToolIdentityMajorSuffixBin(t *testing.T) {
	id := ResolveToolIdentity("github.com/vektra/mockery/v2")
	if id.Bin != "mockery" {
		t.Fatalf("bin=%q", id.Bin)
	}
}
