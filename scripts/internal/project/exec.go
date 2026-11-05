package project

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Execute runs a follow-up command in the current directory, showing its
// output. It is FollowUp's usual runner.
func Execute(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", strings.Join(args, " "), err)
	}
	return nil
}

// IsInstalled reports whether a program is on the PATH.
func IsInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
