package project

import (
	"regexp"
	"strings"
)

// Rewrite replaces the names from goes by in a file's text with the names to goes by.
//
// The compound names go first: the module path, the deploy app name, the web
// package, the environment variable prefix, and the app ID. At each
// position the replacer tries the pairs in order, so the module path wins over
// the app name it may contain. The bare brand and slug go last, as whole
// words, once the compound names are gone.
func Rewrite(text string, from, to Names) string {
	text = strings.NewReplacer(
		from.Module, to.Module,
		from.App, to.App,
		from.Slug+"-web", to.Slug+"-web",
		from.EnvPrefix, to.EnvPrefix,
		appID(from.Slug), appID(to.Slug),
	).Replace(text)
	brand, slug := wholeWord(from.Brand), wholeWord(from.Slug)
	// The module path may contain the old slug even when it does not change.
	return outside(text, to.Module, func(s string) string {
		s = brand.ReplaceAllLiteralString(s, to.Brand)
		return slug.ReplaceAllLiteralString(s, to.Slug)
	})
}

// wholeWord matches s only where it stands as a whole word.
func wholeWord(s string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(s) + `\b`)
}

// outside applies f to every part of text that is not an occurrence of keep.
func outside(text, keep string, f func(string) string) string {
	pieces := strings.Split(text, keep)
	for i, piece := range pieces {
		pieces[i] = f(piece)
	}
	return strings.Join(pieces, keep)
}

// RenamePath gives a file's new path: the command-line tool moves from
// cmd/<old slug> to cmd/<new slug>, so its binary is named after the project.
// Every other path is unchanged.
func RenamePath(path string, from, to Names) string {
	if rest, ok := strings.CutPrefix(path, "cmd/"+from.Slug+"/"); ok {
		return "cmd/" + to.Slug + "/" + rest
	}
	return path
}
