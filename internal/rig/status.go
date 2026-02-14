package rig

import "os"

type StatusReport struct {
	ConfigPath string `json:"configPath"`
	LockPath   string `json:"lockPath"`

	HasLock bool `json:"hasLock"`

	LockMatchesConfig bool         `json:"lockMatchesConfig"`
	ToolsOK           bool         `json:"toolsOk"`
	Missing           int          `json:"missing"`
	Mismatched        int          `json:"mismatched"`
	Extras            int          `json:"extras"`
	Go                *GoStatusRow `json:"go,omitempty"`
}

func Status(startDir string) (StatusReport, error) {
	conf, confPath, err := LoadConfig(startDir)
	if err != nil {
		return StatusReport{}, err
	}
	lockPath := rigLockPathForConfig(confPath)
	lock, err := ReadLockfile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return StatusReport{ConfigPath: confPath, LockPath: lockPath, HasLock: false}, nil
		}
		return StatusReport{}, err
	}

	rows, missing, mismatched, extras, err := CheckInstalledTools(conf.Tools, lock, confPath)
	if err != nil {
		return StatusReport{ConfigPath: confPath, LockPath: lockPath, HasLock: true, LockMatchesConfig: false}, nil
	}

	ok := missing == 0 && mismatched == 0
	_ = rows
	goRow, goOK := checkGoAgainstLockIfRequired(conf.Tools, lock, confPath)

	return StatusReport{
		ConfigPath:        confPath,
		LockPath:          lockPath,
		HasLock:           true,
		LockMatchesConfig: true,
		ToolsOK:           ok && goOK,
		Missing:           missing,
		Mismatched:        mismatched,
		Extras:            len(extras),
		Go:                goRow,
	}, nil
}
