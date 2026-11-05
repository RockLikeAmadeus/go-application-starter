package main

import (
	"errors"
	"io"

	"github.com/spf13/pflag"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// Options are what the user asked rename to do.
type Options struct {
	Module   string
	Name     string
	IsDryRun bool
}

// ParseArgs reads and checks the command line. It rejects unknown flags so a
// typo fails loudly instead of being ignored.
func ParseArgs(args []string) (Options, error) {
	var options Options
	flags := pflag.NewFlagSet("rename", pflag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.Module, "module", "", "new Go module path; the current one is kept if omitted")
	flags.StringVar(&options.Name, "name", "", `new brand name people see, e.g. "Zenith Notes"`)
	flags.BoolVar(&options.IsDryRun, "dry-run", false, "list the changes without making them")
	if err := flags.Parse(args); err != nil {
		return Options{}, err
	}
	if options.Name == "" {
		return Options{}, errors.New("--name is required")
	}
	if options.Module != "" {
		if err := project.CheckModule(options.Module); err != nil {
			return Options{}, err
		}
	}
	return options, nil
}
