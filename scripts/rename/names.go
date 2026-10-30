package main

import (
	"errors"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// NewNames derives the project's new names from what the user asked for and
// the names it goes by now: the module path stays unless a new one is given.
// It fails if the name gives no usable slug, or if nothing would change.
func NewNames(opts Options, current project.Names) (project.Names, error) {
	module := current.Module
	if opts.Module != "" {
		module = opts.Module
	}
	n, err := project.NewNames(module, opts.Name)
	if err != nil {
		return project.Names{}, err
	}
	if n == current {
		return project.Names{}, errors.New("nothing to rename")
	}
	return n, nil
}
