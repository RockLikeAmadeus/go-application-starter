package main

import (
	"errors"
	"io"

	"github.com/spf13/pflag"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// Options are what the user asked rename to do.
type Options struct {
	Module string
	Name   string
	DryRun bool
}

// ParseArgs reads and checks the command line. It rejects unknown flags so a
// typo fails loudly instead of being ignored.
func ParseArgs(args []string) (Options, error) {
	var o Options
	fs := pflag.NewFlagSet("rename", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&o.Module, "module", "", "new Go module path; the current one is kept if omitted")
	fs.StringVar(&o.Name, "name", "", `new brand name people see, e.g. "Zenith Notes"`)
	fs.BoolVar(&o.DryRun, "dry-run", false, "list the changes without making them")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	if o.Name == "" {
		return Options{}, errors.New("--name is required")
	}
	if o.Module != "" {
		if err := project.CheckModule(o.Module); err != nil {
			return Options{}, err
		}
	}
	return o, nil
}
