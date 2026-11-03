package main

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// Options are what the user asked init to do.
type Options struct {
	Module       string
	Name         string
	RemoveSample bool
	Description  string // one line about the app, for the desktop packages
	Publisher    string // who publishes the app
	Email        string // where to reach the publisher about the packages
}

// required are the flags every run must give, in the order they're checked.
var required = []string{"module", "name", "description", "publisher", "email"}

// ParseArgs reads and checks the command line. It rejects unknown flags so a
// typo fails loudly instead of being ignored.
func ParseArgs(args []string) (Options, error) {
	var o Options
	fs := pflag.NewFlagSet("initproject", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&o.Module, "module", "", "Go module path for the new project, e.g. github.com/you/tasks")
	fs.StringVar(&o.Name, "name", "", `brand name people see, e.g. "Acme Tasks"`)
	fs.BoolVar(&o.RemoveSample, "remove-sample", false, "remove the sample Item capability")
	fs.StringVar(&o.Description, "description", "", `one line about the app, e.g. "Tasks for busy teams"`)
	fs.StringVar(&o.Publisher, "publisher", "", `person or company that publishes the app, e.g. "Acme Ltd"`)
	fs.StringVar(&o.Email, "email", "", "where people reach the publisher about the packages")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	for _, name := range required {
		if fs.Lookup(name).Value.String() == "" {
			return Options{}, fmt.Errorf("--%s is required", name)
		}
	}
	if err := project.CheckModule(o.Module); err != nil {
		return Options{}, err
	}
	return o, nil
}
