package main_test

import (
	"os"
	"path/filepath"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

// details are the package details every valid command line must give.
var details = []string{"--description", "Tasks for busy teams", "--publisher", "Acme Ltd", "--email", "dev@acme.example"}

// withDetails returns the command line args followed by details.
func withDetails(args ...string) []string { return append(args, details...) }

// wantDetails is o with the package details from details.
func wantDetails(o initproject.Options) initproject.Options {
	o.Description, o.Publisher, o.Email = "Tasks for busy teams", "Acme Ltd", "dev@acme.example"
	return o
}

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
