package project

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitFiles lists the files git tracks in the current directory, as
// slash-separated paths.
func GitFiles() ([]string, error) {
	list, err := Git("ls-files", "-z")
	if err != nil {
		return nil, err
	}
	var files []string
	for f := range strings.SplitSeq(list, "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

// Git runs a git command in the current directory and returns its output.
func Git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return string(out), nil
}
