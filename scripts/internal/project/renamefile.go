package project

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

// RenameFile renames the project's names in one file and moves it to its new
// path (see RenamePath), keeping its permissions. extra transforms the original
// text first. Directories the move empties are removed.
func RenameFile(dir, rel string, from, to Names, extra func(rel, text string) (string, error)) error {
	return RewriteFile(dir, rel, RenamePath(rel, from, to), func(content []byte) ([]byte, error) {
		if !ShouldRewrite(rel, content) {
			return content, nil
		}
		text, err := extra(rel, string(content))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rel, err)
		}
		out := []byte(Rewrite(text, from, to))
		if strings.HasSuffix(rel, ".go") {
			return format.Source(out)
		}
		return out, nil
	})
}

// RewriteFile transforms the content of the file rel and writes it to newRel
// (both slash-separated, under dir), keeping its permissions. When the paths
// differ, the old file goes, along with any directories that leaves empty.
func RewriteFile(dir, rel, newRel string, transform func([]byte) ([]byte, error)) error {
	old := filepath.Join(dir, filepath.FromSlash(rel))
	info, err := os.Stat(old)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(old)
	if err != nil {
		return err
	}
	content, err = transform(content)
	if err != nil {
		return err
	}
	renamed := filepath.Join(dir, filepath.FromSlash(newRel))
	if err := os.MkdirAll(filepath.Dir(renamed), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(renamed, content, info.Mode().Perm()); err != nil {
		return err
	}
	if renamed == old {
		return nil
	}
	if err := os.Remove(old); err != nil {
		return err
	}
	RemoveEmptyParents(dir, old)
	return nil
}

// RemoveEmptyParents removes the directories above path, from the nearest up,
// until one is not empty. It stops at dir, which path must lie inside.
func RemoveEmptyParents(dir, path string) {
	dir = filepath.Clean(dir)
	for parent := filepath.Dir(path); parent != dir; parent = filepath.Dir(parent) {
		if os.Remove(parent) != nil { // fails once a directory is not empty
			return
		}
	}
}
