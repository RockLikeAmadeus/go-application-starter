package main

import "errors"

// CheckClean refuses to run on a working tree with uncommitted changes, given
// the output of git status --porcelain. The tool rewrites the project in place,
// so starting clean means git status shows exactly what it changed, and
// git reset --hard && git clean -fd undoes it.
func CheckClean(porcelain string) error {
	if porcelain != "" {
		return errors.New("commit or stash your changes first")
	}
	return nil
}
