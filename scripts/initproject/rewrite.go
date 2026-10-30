package main

import (
	"regexp"
	"strings"
)

// The template's own names, which Rewrite replaces.
const (
	templateModule = "github.com/RockLikeAmadeus/go-application-starter"
	templateApp    = "go-application-starter" // deploy app names
	templateWeb    = "starter-web"            // web/package.json name
	templateEnv    = "STARTER_"               // environment variable prefix
	templateCmdDir = "cmd/starter/"           // the command-line tool's main package
)

// The template's brand and command, matched as whole words so longer words
// such as "starters" are left alone.
var (
	templateBrand   = regexp.MustCompile(`\bStarter\b`)
	templateCommand = regexp.MustCompile(`\bstarter\b`)
)

// Names are the new project's names, derived from the user's Options.
type Names struct {
	Module    string // Go module path
	Brand     string // brand name people see, e.g. "Acme Tasks"
	Slug      string // machine name, e.g. "acme-tasks"; see Slug
	EnvPrefix string // environment variable prefix, e.g. "ACME_TASKS_"; see EnvPrefix
}

// Rewrite replaces the template's names in a file's text with the new project's.
//
// The compound names go first. At each position the replacer tries the pairs in
// order, so the module path wins over the app name it contains. The bare words
// "Starter" and "starter" go last, once the compound names are gone.
func Rewrite(text string, n Names) string {
	text = strings.NewReplacer(
		templateModule, n.Module,
		templateApp, n.Slug,
		templateWeb, n.Slug+"-web",
		templateEnv, n.EnvPrefix,
	).Replace(text)
	text = templateBrand.ReplaceAllLiteralString(text, n.Brand)
	return templateCommand.ReplaceAllLiteralString(text, n.Slug)
}

// RenamePath gives a file's new path: the command-line tool moves from
// cmd/starter to cmd/<slug>, so its binary is named after the project.
// Every other path is unchanged.
func RenamePath(path string, n Names) string {
	if rest, ok := strings.CutPrefix(path, templateCmdDir); ok {
		return "cmd/" + n.Slug + "/" + rest
	}
	return path
}
