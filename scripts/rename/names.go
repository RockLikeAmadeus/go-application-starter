package main

import "github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"

// NewNames derives the project's new names from what the user asked for and
// the names it goes by now.
func NewNames(opts Options, current project.Names) (project.Names, error) {
	slug, err := project.Slug(opts.Name)
	if err != nil {
		return project.Names{}, err
	}
	return project.Names{Module: current.Module, Brand: opts.Name, Slug: slug, App: slug, EnvPrefix: project.EnvPrefix(slug)}, nil
}
