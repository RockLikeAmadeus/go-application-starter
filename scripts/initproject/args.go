package main

import (
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"

	"github.com/spf13/pflag"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// Options are what the user asked init to do.
type Options struct {
	Module             string
	Name               string
	ShouldRemoveSample bool
	Description        string // one line about the app, for the desktop packages
	Publisher          string // who publishes the app
	Email              string // where to reach the publisher about the packages
}

// required are the flags every run must give, in the order they're checked.
var required = []string{"module", "name", "description", "publisher", "email"}

// ParseArgs reads and checks the command line. It rejects unknown flags so a
// typo fails loudly instead of being ignored.
func ParseArgs(args []string) (Options, error) {
	var options Options
	flags := pflag.NewFlagSet("initproject", pflag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.Module, "module", "", "Go module path for the new project, e.g. github.com/you/tasks")
	flags.StringVar(&options.Name, "name", "", `brand name people see, e.g. "Acme Tasks"`)
	flags.BoolVar(&options.ShouldRemoveSample, "remove-sample", false, "remove the sample Item capability")
	flags.StringVar(&options.Description, "description", "", `one line about the app, e.g. "Tasks for busy teams"`)
	flags.StringVar(&options.Publisher, "publisher", "", `person or company that publishes the app, e.g. "Acme Ltd"`)
	flags.StringVar(&options.Email, "email", "", "where people reach the publisher about the packages")
	if err := flags.Parse(args); err != nil {
		return Options{}, err
	}
	for _, name := range required {
		if flags.Lookup(name).Value.String() == "" {
			return Options{}, fmt.Errorf("--%s is required", name)
		}
	}
	if err := project.CheckModule(options.Module); err != nil {
		return Options{}, err
	}
	if err := checkDetails(options); err != nil {
		return Options{}, err
	}
	return options, nil
}

// checkDetails checks the package details. The description and publisher land
// in YAML, JSON, a desktop file, and makensis's arguments, so they must be one
// line with no double quotes or backslashes. The publisher also goes in the
// Linux maintainer, "<publisher> <<email>>", so it can't hold < or >, and the
// email must be a bare address.
func checkDetails(options Options) error {
	if err := checkLine("description", options.Description); err != nil {
		return err
	}
	if err := checkLine("publisher", options.Publisher); err != nil {
		return err
	}
	if strings.ContainsAny(options.Publisher, "<>") {
		return errors.New("publisher must not contain < or >")
	}
	if addr, err := mail.ParseAddress(options.Email); err != nil || addr.Name != "" || addr.Address != options.Email {
		return errors.New("invalid email address")
	}
	return nil
}

// checkLine refuses a value that would break a quoted string or a one-line
// field, naming it as field.
func checkLine(field, value string) error {
	if strings.ContainsAny(value, `"\`) {
		return fmt.Errorf("%s must not contain double quotes or backslashes", field)
	}
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s must be one line", field)
	}
	return nil
}
