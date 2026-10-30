package main

import "github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"

// The template's own names, which Rewrite and RenamePath replace.
var template = project.Names{
	Module:    "github.com/RockLikeAmadeus/go-application-starter",
	Brand:     "Starter",
	Slug:      "starter",
	App:       "go-application-starter", // deploy app names
	EnvPrefix: "STARTER_",
}

// Rewrite replaces the template's names in a file's text with the new project's.
func Rewrite(text string, n project.Names) string {
	return project.Rewrite(text, template, n)
}

// RenamePath gives a file's new path: the command-line tool moves from
// cmd/starter to cmd/<slug>. Every other path is unchanged.
func RenamePath(path string, n project.Names) string {
	return project.RenamePath(path, template, n)
}
