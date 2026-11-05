// Command initproject turns a fresh copy of this template into a new project.
// See docs/template.md.
package main

import (
	"fmt"
	"os"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
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
	files, err := project.GitFiles()
	if err != nil {
		return err
	}
	status, err := project.Git("status", "--porcelain")
	if err != nil {
		return err
	}
	cmds, err := Run(".", files, os.Args[1:], status)
	if err != nil {
		return err
	}
	return project.FollowUp(cmds, project.Execute, project.IsInstalled, os.Stdout)
}
