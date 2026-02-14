package rig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DoctorReport struct {
	VersionPresent bool
	GoAvailable    bool
	GoVersion      string
	GoMatchesLock  bool

	ConfigPath   string
	LockPath     string
	HasConfig    bool
	HasLock      bool
	LockValid    bool
	BinDir       string
	BinDirExists bool
	BinWritable  bool

	ExecutablePath     string
	ExecutableWritable bool

	Errors []string
}

func Doctor(startDir string, currentVersion string, executablePath string) (DoctorReport, error) {
	rep := DoctorReport{VersionPresent: strings.TrimSpace(currentVersion) != "" && strings.TrimSpace(currentVersion) != "dev"}

	goVer, gerr := execCapture("go", []string{"version"}, "", nil)
	if gerr == nil {
		rep.GoAvailable = true
		rep.GoVersion = goVer
	} else {
		rep.Errors = append(rep.Errors, "go toolchain not found in PATH")
	}

	conf, confPath, err := LoadConfig(startDir)
	if err != nil {
		return rep, err
	}
	rep.HasConfig = true
	rep.ConfigPath = confPath
	rep.BinDir = localBinDirForConfig(confPath)
	rep.BinDirExists = dirExists(rep.BinDir)
	rep.BinWritable = isDirWritable(rep.BinDir)

	lockPath := rigLockPathForConfig(confPath)
	rep.LockPath = lockPath
	lock, lerr := ReadLockfile(lockPath)
	if lerr != nil {
		rep.HasLock = false
		rep.LockValid = false
		rep.Errors = append(rep.Errors, fmt.Sprintf("rig.lock missing or invalid: %v", lerr))
	} else {
		rep.HasLock = true
		rep.LockValid = true
		if strings.TrimSpace(conf.Tools["go"]) != "" {
			row, ok := checkGoAgainstLockIfRequired(conf.Tools, lock, confPath)
			rep.GoMatchesLock = ok
			if row != nil && row.Status != "ok" {
				rep.Errors = append(rep.Errors, fmt.Sprintf("go mismatch: have=%q want=%q", row.Have, row.Locked))
			}
		} else {
			rep.GoMatchesLock = true
		}
	}

	rep.ExecutablePath = executablePath
	rep.ExecutableWritable = isFileReplaceWritable(executablePath)
	if !rep.ExecutableWritable {
		rep.Errors = append(rep.Errors, fmt.Sprintf("binary path not writable: %s", executablePath))
	}

	return rep, nil
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.IsDir()
}

func isDirWritable(path string) bool {
	if !dirExists(path) {
		return false
	}
	f, err := os.CreateTemp(path, ".rig-writecheck-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

func isFileReplaceWritable(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	dir := filepath.Dir(path)
	if !dirExists(dir) {
		return false
	}
	f, err := os.CreateTemp(dir, ".rig-upgrade-check-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}
