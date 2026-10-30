package main

import (
	"errors"
	"io"

	"github.com/spf13/pflag"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// Options are what the user asked init to do.
type Options struct {
	Module       string
	Name         string
	RemoveSample bool
}

// ParseArgs reads and checks the command line. It rejects unknown flags so a
// typo fails loudly instead of being ignored.
func ParseArgs(args []string) (Options, error) {
	var o Options
	fs := pflag.NewFlagSet("initproject", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&o.Module, "module", "", "Go module path for the new project, e.g. github.com/you/tasks")
	fs.StringVar(&o.Name, "name", "", `brand name people see, e.g. "Acme Tasks"`)
	fs.BoolVar(&o.RemoveSample, "remove-sample", false, "remove the sample Item capability")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	if o.Module == "" {
		return Options{}, errors.New("--module is required")
	}
	if o.Name == "" {
		return Options{}, errors.New("--name is required")
	}
	if err := project.CheckModule(o.Module); err != nil {
		return Options{}, err
	}
	return o, nil
}
