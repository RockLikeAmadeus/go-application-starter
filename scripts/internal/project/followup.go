package project

import (
	"fmt"
	"io"
	"strings"
)

// optional is the program whose commands may fail without failing the run:
// regenerating the web client needs a working Node toolchain, which a fresh
// clone may not have yet.
const optional = "pnpm"

// FollowUp runs a tool's follow-up commands, in order, stopping at the first
// failure. A command whose program is not installed, or an optional command
// that fails, is skipped with a note on out, so the web toolchain never blocks
// the rest of the run.
func FollowUp(cmds [][]string, run func([]string) error, isInstalled func(string) bool, out io.Writer) error {
	for _, cmd := range cmds {
		if !isInstalled(cmd[0]) {
			fmt.Fprintf(out, "%s is not installed, so %q was skipped: run task gen once it is.\n", cmd[0], strings.Join(cmd, " "))
			continue
		}
		if err := run(cmd); err != nil {
			if cmd[0] != optional {
				return err
			}
			fmt.Fprintf(out, "%v, so it was skipped: run task gen once it is fixed.\n", err)
		}
	}
	return nil
}
