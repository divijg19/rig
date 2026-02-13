//go:build !linux

package cli

import "golang.org/x/term"

func setDevInputMode(fd int) (func(), error) {
	state, err := term.MakeRaw(fd)
	if err != nil {
		return func() {}, err
	}
	return func() {
		_ = term.Restore(fd, state)
	}, nil
}
