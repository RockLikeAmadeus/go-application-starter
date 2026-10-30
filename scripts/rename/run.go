package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// Run renames the project in dir as asked by args. files are the paths git
// tracks in dir, slash-separated; untracked and ignored files are never
// touched. porcelain is the output of git status --porcelain for dir. With
// --dry-run it only lists the changes on out, and needs no clean tree.
//
// Run returns the commands the caller must then run in dir to bring go.mod in
// line with the changed sources.
func Run(dir string, files []string, args []string, porcelain string, out io.Writer) ([][]string, error) {
	opts, err := ParseArgs(args)
	if err != nil {
		return nil, err
	}
	old, err := project.ReadRecord(dir)
	if err != nil {
		return nil, err
	}
	n, err := NewNames(opts, old)
	if err != nil {
		return nil, err
	}
	if opts.DryRun {
		return nil, previewAll(out, dir, files, old, n)
	}
	if err := project.CheckClean(porcelain); err != nil {
		return nil, err
	}
	if err := renameAll(dir, files, old, n); err != nil {
		return nil, err
	}
	if n.Module != old.Module {
		return [][]string{tidy}, nil
	}
	return nil, nil
}

// tidy follows a change of module path, so go.mod and go.sum agree with the
// renamed imports.
var tidy = []string{"go", "mod", "tidy"}

// previewAll lists on out what renameAll would do to files, without changing
// them.
func previewAll(out io.Writer, dir string, files []string, old, n project.Names) error {
	for _, rel := range files {
		if rel == project.RecordFile {
			continue
		}
		if err := preview(out, dir, rel, old, n); err != nil {
			return err
		}
	}
	return nil
}

// renameAll gives files the new names, then records them. The record is
// written afresh, not renamed.
func renameAll(dir string, files []string, old, n project.Names) error {
	keep := func(_, text string) (string, error) { return text, nil }
	for _, rel := range files {
		if rel == project.RecordFile {
			continue
		}
		if err := project.RenameFile(dir, rel, old, n, keep); err != nil {
			return err
		}
	}
	return project.WriteRecord(dir, n)
}

// preview lists on out what renaming the file rel from the old names to the
// new ones would do: its changed lines, then its move, if any.
func preview(out io.Writer, dir, rel string, old, n project.Names) error {
	content, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		return err
	}
	if project.ShouldRewrite(rel, content) {
		if text := project.Rewrite(string(content), old, n); text != string(content) {
			fmt.Fprintln(out, rel)
			printChangedLines(out, string(content), text)
		}
	}
	if newRel := project.RenamePath(rel, old, n); newRel != rel {
		fmt.Fprintln(out, rel+" -> "+newRel)
	}
	return nil
}

// printChangedLines prints each line that differs between before and after as
// "- old" then "+ new". Renaming never adds or removes lines, so the two texts
// line up.
func printChangedLines(out io.Writer, before, after string) {
	a := strings.Split(after, "\n")
	for i, line := range strings.Split(before, "\n") {
		if line != a[i] {
			fmt.Fprintln(out, "- "+line)
			fmt.Fprintln(out, "+ "+a[i])
		}
	}
}
