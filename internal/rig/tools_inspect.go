package rig

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type ManagedToolInfo struct {
	Name      string
	Requested string
	Resolved  string
	Path      string
	Status    ToolState
}

type ToolWhyInfo struct {
	Name      string
	Requested string
	Resolved  string
	SHA256    string
	Path      string
}

type ToolDoctorReport struct {
	Name         string
	Path         string
	Exists       bool
	Executable   bool
	SHAExpected  string
	SHAActual    string
	SHAMatch     bool
	ResolvedPath string
	ResolvedOK   bool
	Status       ToolState
	Error        string
}

func ToolsLS(startDir string) ([]ManagedToolInfo, error) {
	conf, confPath, err := LoadConfig(startDir)
	if err != nil {
		return nil, err
	}
	lock, err := ReadRigLockForConfig(confPath)
	if err != nil {
		return nil, err
	}
	rows, _, _, _, err := CheckInstalledTools(conf.Tools, lock, confPath)
	if err != nil {
		return nil, err
	}

	byName := map[string]LockedTool{}
	for _, lt := range lock.Tools {
		name, _, perr := ParseRequested(lt.Requested)
		if perr != nil {
			return nil, perr
		}
		byName[name] = lt
	}

	out := make([]ManagedToolInfo, 0, len(rows))
	for _, row := range rows {
		lt, ok := byName[row.Name]
		if !ok {
			return nil, fmt.Errorf("tool %q missing from rig.lock", row.Name)
		}
		out = append(out, ManagedToolInfo{
			Name:      row.Name,
			Requested: lt.Requested,
			Resolved:  lt.Resolved,
			Path:      ToolBinPath(confPath, lt.Bin),
			Status:    ToolState(row.Status),
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func ToolPath(startDir, name string) (string, error) {
	conf, confPath, err := LoadConfig(startDir)
	if err != nil {
		return "", err
	}
	lock, err := ReadRigLockForConfig(confPath)
	if err != nil {
		return "", err
	}
	if err := LockMatchesTools(lock, conf.Tools); err != nil {
		return "", err
	}
	lt, err := findLockedToolByName(lock, name)
	if err != nil {
		return "", err
	}

	bin := strings.TrimSpace(lt.Bin)
	if bin == "" {
		bin = ResolveToolIdentity(name).Bin
	}
	binPath := ToolBinPath(confPath, bin)
	if err := ensureExecutable(binPath); err != nil {
		return "", fmt.Errorf("tool %q not found in .rig/bin\nhint: run `rig sync`", name)
	}
	actual, err := ComputeFileSHA256(binPath)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(actual) != strings.TrimSpace(lt.SHA256) {
		return "", fmt.Errorf("tool %q checksum mismatch\nhint: run `rig sync`", name)
	}
	return binPath, nil
}

func ToolWhy(startDir, name string) (ToolWhyInfo, error) {
	conf, confPath, err := LoadConfig(startDir)
	if err != nil {
		return ToolWhyInfo{}, err
	}
	lock, err := ReadRigLockForConfig(confPath)
	if err != nil {
		return ToolWhyInfo{}, err
	}
	if err := LockMatchesTools(lock, conf.Tools); err != nil {
		return ToolWhyInfo{}, err
	}
	lt, err := findLockedToolByName(lock, name)
	if err != nil {
		return ToolWhyInfo{}, err
	}
	return ToolWhyInfo{
		Name:      name,
		Requested: firstNonEmptyString(conf.Tools[name], lt.Requested),
		Resolved:  lt.Resolved,
		SHA256:    lt.SHA256,
		Path:      ToolBinPath(confPath, firstNonEmptyString(lt.Bin, ResolveToolIdentity(name).Bin)),
	}, nil
}

func ToolsDoctor(startDir, name string) ([]ToolDoctorReport, error) {
	conf, confPath, err := LoadConfig(startDir)
	if err != nil {
		return nil, err
	}
	lock, err := ReadRigLockForConfig(confPath)
	if err != nil {
		return nil, err
	}
	if err := LockMatchesTools(lock, conf.Tools); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(conf.Tools))
	for toolName := range conf.Tools {
		if toolName == "go" {
			continue
		}
		names = append(names, toolName)
	}
	sort.Strings(names)
	if strings.TrimSpace(name) != "" {
		if _, ok := conf.Tools[name]; !ok {
			return nil, fmt.Errorf("tool %q not declared in rig.toml", name)
		}
		names = []string{name}
	}

	reports := make([]ToolDoctorReport, 0, len(names))
	for _, toolName := range names {
		lt, lerr := findLockedToolByName(lock, toolName)
		if lerr != nil {
			return nil, lerr
		}
		bin := strings.TrimSpace(lt.Bin)
		if bin == "" {
			bin = ResolveToolIdentity(toolName).Bin
		}
		p := ToolBinPath(confPath, bin)
		r := ToolDoctorReport{
			Name:         toolName,
			Path:         p,
			SHAExpected:  lt.SHA256,
			ResolvedPath: filepath.Clean(p),
			ResolvedOK:   filepath.Clean(p) == filepath.Clean(ToolBinPath(confPath, bin)),
			Status:       ToolOK,
		}

		if err := ensureExecutable(p); err != nil {
			r.Exists = false
			r.Executable = false
			r.SHAMatch = false
			r.Status = ToolMissing
			r.Error = err.Error()
			reports = append(reports, r)
			continue
		}
		r.Exists = true
		r.Executable = true

		sum, err := ComputeFileSHA256(p)
		if err != nil {
			r.Status = ToolMismatch
			r.Error = err.Error()
			reports = append(reports, r)
			continue
		}
		r.SHAActual = sum
		r.SHAMatch = strings.TrimSpace(sum) == strings.TrimSpace(lt.SHA256)
		if !r.SHAMatch {
			r.Status = ToolMismatch
			r.Error = "sha256 mismatch"
		}
		reports = append(reports, r)
	}

	return reports, nil
}

func findLockedToolByName(lock Lockfile, toolName string) (LockedTool, error) {
	for _, lt := range lock.Tools {
		name, _, err := ParseRequested(lt.Requested)
		if err != nil {
			return LockedTool{}, err
		}
		if name == toolName {
			return lt, nil
		}
	}
	return LockedTool{}, fmt.Errorf("tool %q not found in rig.lock", toolName)
}

func firstNonEmptyString(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return strings.TrimSpace(a)
	}
	return strings.TrimSpace(b)
}
