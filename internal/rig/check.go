package rig

import (
	"encoding/json"
	"os"
)

type CheckReport struct {
	ConfigPath string          `json:"configPath"`
	LockPath   string          `json:"lockPath"`
	OK         bool            `json:"ok"`
	Error      string          `json:"error,omitempty"`
	Missing    int             `json:"missing"`
	Mismatched int             `json:"mismatched"`
	Extras     []string        `json:"extras,omitempty"`
	Tools      []ToolStatusRow `json:"tools"`
	Go         *GoStatusRow    `json:"go,omitempty"`
}

func Check(startDir string) (CheckReport, error) {
	conf, confPath, err := LoadConfig(startDir)
	if err != nil {
		return CheckReport{}, err
	}

	lockPath := rigLockPathForConfig(confPath)
	lock, err := ReadLockfile(lockPath)
	if err != nil {
		rep := CheckReport{ConfigPath: confPath, LockPath: lockPath, OK: false, Tools: []ToolStatusRow{}}
		if os.IsNotExist(err) {
			rep.Error = "rig.lock not found: run 'rig sync' first"
			return rep, nil
		}
		rep.Error = err.Error()
		return rep, nil
	}

	rows, missing, mismatched, extras, err := CheckInstalledTools(conf.Tools, lock, confPath)
	if err != nil {
		rep := CheckReport{ConfigPath: confPath, LockPath: lockPath, OK: false, Tools: []ToolStatusRow{}}
		rep.Error = err.Error()
		return rep, nil
	}

	goRow, goOK := checkGoAgainstLockIfRequired(conf.Tools, lock, confPath)

	ok := missing == 0 && mismatched == 0 && goOK
	return CheckReport{
		ConfigPath: confPath,
		LockPath:   lockPath,
		OK:         ok,
		Missing:    missing,
		Mismatched: mismatched,
		Extras:     extras,
		Tools:      rows,
		Go:         goRow,
	}, nil
}

func (r CheckReport) MarshalJSONStable() ([]byte, error) {
	return json.Marshal(r)
}
