// Package project holds what the init and rename tools share: deriving a
// project's names, replacing them in its files, and the git and follow-up
// plumbing around a run. It survives init, so it must never mention the
// template's own names: a rename rewrites every tracked file, this one included.
package project

import (
	"errors"
	"strings"
)

// Names are the names a project goes by.
type Names struct {
	Module    string // Go module path
	Brand     string // brand name people see, e.g. "Acme Tasks"
	Slug      string // machine name, e.g. "acme-tasks"; see Slug
	App       string // deploy app name: the slug, except in the template
	EnvPrefix string // environment variable prefix, e.g. "ACME_TASKS_"; see EnvPrefix
}

// Slug derives the project's machine name from its brand name: lower case,
// with each run of anything other than ASCII letters and digits turned into a
// single hyphen. It names the command, the Docker image and the deploy configs,
// so other characters are dropped ("Café" becomes "caf"). Names with double
// quotes or backslashes are refused: the brand lands in quoted strings, which
// those would break.
func Slug(name string) (string, error) {
	if strings.ContainsAny(name, `"\`) {
		return "", errors.New("name must not contain double quotes or backslashes")
	}
	words := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	if len(words) == 0 {
		return "", errors.New("name must contain a letter or digit")
	}
	return strings.Join(words, "-"), nil
}

// EnvPrefix derives the prefix for the project's environment variables from
// its slug: "acme-tasks" reads ACME_TASKS_STORAGE and friends.
func EnvPrefix(slug string) string {
	return strings.ToUpper(strings.ReplaceAll(slug, "-", "_")) + "_"
}

// NewNames derives a project's names from its module path and brand name. It
// fails if the brand name gives no usable slug.
func NewNames(module, name string) (Names, error) {
	slug, err := Slug(name)
	if err != nil {
		return Names{}, err
	}
	return Names{Module: module, Brand: name, Slug: slug, App: slug, EnvPrefix: EnvPrefix(slug)}, nil
}
