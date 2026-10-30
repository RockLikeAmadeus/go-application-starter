package main_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

// writeFile creates dir/rel (a slash-separated path) with content, making any
// parent directories.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readFile returns the content of dir/rel (a slash-separated path).
func readFile(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// acme is the names of the project being renamed.
var acme = project.Names{Module: "example.com/acme/tasks", Brand: "Acme", Slug: "acme", App: "acme", EnvPrefix: "ACME_"}

// newAcmeProject returns a temporary directory whose record names it acme.
func newAcmeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := project.WriteRecord(dir, acme); err != nil {
		t.Fatal(err)
	}
	return dir
}
