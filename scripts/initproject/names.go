package main

import (
	"errors"
	"strings"
)

// Slug derives the project's machine name from its brand name: lower case,
// with each run of anything other than ASCII letters and digits turned into a
// single hyphen. It names the command, the Docker image and the deploy configs,
// so other characters are dropped ("Café" becomes "caf").
func Slug(name string) (string, error) {
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

// NewNames derives all of the new project's names from what the user asked
// for. It fails if the brand name gives no usable slug.
func NewNames(opts Options) (Names, error) {
	slug, err := Slug(opts.Name)
	if err != nil {
		return Names{}, err
	}
	return Names{Module: opts.Module, Brand: opts.Name, Slug: slug, EnvPrefix: EnvPrefix(slug)}, nil
}
