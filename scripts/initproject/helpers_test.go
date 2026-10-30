package main_test

import (
	"os"
	"path/filepath"
	"testing"
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

// writeTemplateRecord gives dir the project.json the template ships with, so
// Run knows the names it is replacing.
func writeTemplateRecord(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, dir, "project.json", `{"module": "github.com/RockLikeAmadeus/go-application-starter", `+
		`"name": "Starter", "slug": "starter", "app": "go-application-starter"}`+"\n")
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
