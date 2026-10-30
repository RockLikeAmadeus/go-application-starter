// Command initproject turns a fresh copy of this template into a new project.
// See docs/template.md.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "initproject:", err)
		os.Exit(1)
	}
}

// run asks git about the current directory, which must be the repository
// root, and hands the answers to Run.
func run() error {
	files, err := gitFiles()
	if err != nil {
		return err
	}
	status, err := git("status", "--porcelain")
	if err != nil {
		return err
	}
	cmds, err := Run(".", files, os.Args[1:], status)
	if err != nil {
		return err
	}
	return FollowUp(cmds, execute, installed, os.Stdout)
}

// execute runs a follow-up command in the current directory, showing its output.
func execute(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", strings.Join(args, " "), err)
	}
	return nil
}

// installed reports whether a program is on the PATH.
func installed(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// gitFiles lists the files git tracks, as slash-separated paths.
func gitFiles() ([]string, error) {
	list, err := git("ls-files", "-z")
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

// git runs a git command in the current directory and returns its output.
func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return string(out), nil
}
