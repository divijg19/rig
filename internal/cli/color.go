package cli

import (
	"errors"
	"os"
)

type colorMode string

const (
	colorAuto   colorMode = "auto"
	colorAlways colorMode = "always"
	colorNever  colorMode = "never"
)

func resolveColorEnabled(mode string, out *os.File) (bool, error) {
	if mode == "" {
		mode = string(colorAuto)
	}
	switch colorMode(mode) {
	case colorAuto, colorAlways, colorNever:
		// valid
	default:
		return false, errors.New("error: invalid --color value (expected auto|always|never)")
	}
	if colorMode(mode) == colorNever {
		return false, nil
	}
	if out == nil {
		return false, nil
	}
	if !isTTY(out) {
		return false, nil
	}
	if os.Getenv("CI") != "" {
		return false, nil
	}
	if os.Getenv("NO_COLOR") != "" {
		return false, nil
	}
	if colorMode(mode) == colorAlways {
		return true, nil
	}
	return true, nil
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

const (
	ansiReset    = "\x1b[0m"
	ansiBoldCyan = "\x1b[1;36m"
	ansiYellow   = "\x1b[33m"
	ansiRed      = "\x1b[31m"
)
