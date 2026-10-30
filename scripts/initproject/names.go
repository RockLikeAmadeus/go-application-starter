package main

import "github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"

// NewNames derives all of the new project's names from what the user asked
// for. It fails if the brand name gives no usable slug.
func NewNames(opts Options) (project.Names, error) {
	slug, err := project.Slug(opts.Name)
	if err != nil {
		return project.Names{}, err
	}
	return project.Names{Module: opts.Module, Brand: opts.Name, Slug: slug, EnvPrefix: project.EnvPrefix(slug)}, nil
}
