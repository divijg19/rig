package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ExecuteEntrypoint is the single bootstrap for the rig binary.
// It resolves intent from argv[0] (optional user-created symlink/rename) and then runs the Cobra CLI.
func ExecuteEntrypoint() {
	rewriteArgsForInvocation()
	Execute()
}

func rewriteArgsForInvocation() {
	base := strings.ToLower(filepath.Base(os.Args[0]))
	if runtime.GOOS == "windows" {
		base = strings.TrimSuffix(base, ".exe")
	}

	switch base {
	case "rir":
		os.Args = append([]string{"rig", "run"}, os.Args[1:]...)
	case "ric":
		os.Args = append([]string{"rig", "check"}, os.Args[1:]...)
	case "rid":
		os.Args = append([]string{"rig", "dev"}, os.Args[1:]...)
	case "ris":
		os.Args = append([]string{"rig", "start"}, os.Args[1:]...)
	default:
		// no rewrite
	}
}
