package main

import (
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
	strip := func(rel, text string) (string, error) { return stripBlocks(rel, text, opts.RemoveSample) }
	for _, rel := range files {
		if isUnder(rel, toolFiles) || opts.RemoveSample && IsSample(rel) {
			continue
		}
		if err := project.RenameFile(dir, rel, template, n, strip); err != nil {
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

// goimports are the settings gofmt and goimports use by default.
var goimports = &imports.Options{Comments: true, TabIndent: true, TabWidth: 8}

// stripBlocks removes a file's template blocks and, if asked, its sample
// blocks, before project.RenameFile gives it the new names. Go files that lost
// sample blocks are tidied as goimports would, so the sample's code can be
// fenced without fencing its imports.
func stripBlocks(rel, text string, removeSample bool) (string, error) {
	text, err := RemoveBlocks(text, templateTag)
	if err != nil {
		return "", err
	}
	if !removeSample {
		return text, nil
	}
	if text, err = RemoveBlocks(text, sampleTag); err != nil {
		return "", err
	}
	if !strings.HasSuffix(rel, ".go") {
		return text, nil
	}
	out, err := imports.Process(rel, []byte(text), goimports)
	return string(out), err
}
