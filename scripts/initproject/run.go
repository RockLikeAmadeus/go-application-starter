package main

import (
	"fmt"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// Run makes the template in dir into a new project, as asked by args (the
// command line without the program name). files are the paths git tracks in
// dir, slash-separated; untracked and ignored files are never touched.
// porcelain is the output of git status --porcelain for dir. Every check runs
// before any file is touched, so a refused run leaves the project exactly as
// it was.
//
// Run returns the commands the caller must then run in dir to bring generated
// code and go.mod in line with the changed sources.
func Run(dir string, files []string, args []string, porcelain string) ([][]string, error) {
	opts, err := ParseArgs(args)
	if err != nil {
		return nil, err
	}
	n, err := NewNames(opts)
	if err != nil {
		return nil, err
	}
	if err := project.CheckClean(porcelain); err != nil {
		return nil, err
	}
	if opts.RemoveSample {
		if err := removeAll(dir, SampleFiles); err != nil {
			return nil, err
		}
	}
	if err := removeAll(dir, toolFiles); err != nil {
		return nil, err
	}
	for _, rel := range files {
		if isUnder(rel, toolFiles) || opts.RemoveSample && IsSample(rel) {
			continue
		}
		if err := renameFile(dir, rel, n, opts.RemoveSample); err != nil {
			return nil, err
		}
	}
	if opts.RemoveSample {
		return regenerate, nil
	}
	return [][]string{tidy}, nil
}

// tidy follows every run: it drops the init tool's own dependencies from
// go.mod, and the sample's too when that is removed.
var tidy = []string{"go", "mod", "tidy"}

// regenerate follows sample removal: the API code is regenerated from the
// trimmed spec first, so tidy sees the final imports.
var regenerate = [][]string{
	{"go", "tool", "-modfile=tools/go.mod", "oapi-codegen", "-config", "api/oapi-codegen.server.yaml", "api/openapi.yaml"},
	{"go", "tool", "-modfile=tools/go.mod", "oapi-codegen", "-config", "api/oapi-codegen.client.yaml", "api/openapi.yaml"},
	tidy,
	{"pnpm", "--dir", "web", "gen:api"},
}

// renameFile transforms one file's content (see transform) and moves it to its
// new path, keeping its permissions. Directories the move empties are removed.
func renameFile(dir, rel string, n project.Names, removeSample bool) error {
	return project.RewriteFile(dir, rel, RenamePath(rel, n), func(content []byte) ([]byte, error) {
		content, err := transform(rel, content, n, removeSample)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rel, err)
		}
		return content, nil
	})
}

// goimports are the settings gofmt and goimports use by default.
var goimports = &imports.Options{Comments: true, TabIndent: true, TabWidth: 8}

// transform gives a file's new content: template blocks (and, if asked, sample
// blocks) removed and the template's names replaced. Files project.ShouldRewrite skips
// come back unchanged. Go files that lost sample blocks are tidied as goimports
// would, so the sample's code can be fenced without fencing its imports.
func transform(rel string, content []byte, n project.Names, removeSample bool) ([]byte, error) {
	if !project.ShouldRewrite(rel, content) {
		return content, nil
	}
	text, err := RemoveBlocks(string(content), templateTag)
	if err != nil {
		return nil, err
	}
	if removeSample {
		if text, err = RemoveBlocks(text, sampleTag); err != nil {
			return nil, err
		}
	}
	out := []byte(Rewrite(text, n))
	// Only removal is tidied here. A new module path that sorts differently can
	// still leave imports out of order; project.RenameFile formats for that.
	if removeSample && strings.HasSuffix(rel, ".go") {
		return imports.Process(rel, out, goimports)
	}
	return out, nil
}
