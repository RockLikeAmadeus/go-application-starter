package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

func TestRenameFileRewritesAndMovesAFile(t *testing.T) {
	from := project.Names{Module: "example.com/you/notes", Brand: "Old Notes", Slug: "old-notes", App: "old-notes", EnvPrefix: "OLD_NOTES_"}
	to := project.Names{Module: "example.com/you/zenith", Brand: "Zenith", Slug: "zenith", App: "zenith", EnvPrefix: "ZENITH_"}
	dir := t.TempDir()
	old := filepath.Join(dir, "cmd", "old-notes", "run.sh")
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("# Runs Old Notes.\nDROP ME\nexec old-notes\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	var seen string
	extra := func(rel, text string) (string, error) {
		seen = rel + ": " + text
		return strings.ReplaceAll(text, "DROP ME\n", ""), nil
	}

	if err := project.RenameFile(dir, "cmd/old-notes/run.sh", from, to, extra); err != nil {
		t.Fatalf("RenameFile error = %v", err)
	}

	if want := "cmd/old-notes/run.sh: # Runs Old Notes.\nDROP ME\nexec old-notes\n"; seen != want {
		t.Errorf("extra saw %q, want the original file %q", seen, want)
	}
	renamed := filepath.Join(dir, "cmd", "zenith", "run.sh")
	data, err := os.ReadFile(renamed)
	if err != nil {
		t.Fatalf("the file was not moved to cmd/zenith: %v", err)
	}
	if want := "# Runs Zenith.\nexec zenith\n"; string(data) != want {
		t.Errorf("cmd/zenith/run.sh = %q, want %q", data, want)
	}
	if info, err := os.Stat(renamed); err != nil || info.Mode().Perm() != 0o755 {
		t.Errorf("cmd/zenith/run.sh mode = %v (err %v), want -rwxr-xr-x", info.Mode().Perm(), err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cmd", "old-notes")); !os.IsNotExist(err) {
		t.Errorf("cmd/old-notes still exists (err %v), want it removed once empty", err)
	}
}

func TestRenameFileKeepsGoImportsSorted(t *testing.T) {
	// The new module path sorts after github.com, where the old one sorted before.
	from := project.Names{Module: "example.com/you/notes", Brand: "Notes", Slug: "notes", App: "notes", EnvPrefix: "NOTES_"}
	to := project.Names{Module: "zenith.dev/notes", Brand: "Zenith", Slug: "zenith", App: "zenith", EnvPrefix: "ZENITH_"}
	dir := t.TempDir()
	in := "package main\n\nimport (\n" +
		"\t\"fmt\"\n\n" +
		"\t\"example.com/you/notes/internal/app\"\n" +
		"\t\"github.com/spf13/cobra\"\n" +
		")\n\nvar _, _, _ = fmt.Sprint, app.New, cobra.Command{}\n"
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	unchanged := func(_, text string) (string, error) { return text, nil }

	if err := project.RenameFile(dir, "main.go", from, to, unchanged); err != nil {
		t.Fatalf("RenameFile error = %v", err)
	}

	want := "package main\n\nimport (\n" +
		"\t\"fmt\"\n\n" +
		"\t\"github.com/spf13/cobra\"\n" +
		"\t\"zenith.dev/notes/internal/app\"\n" +
		")\n\nvar _, _, _ = fmt.Sprint, app.New, cobra.Command{}\n"
	data, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Errorf("main.go =\n%s\nwant\n%s", data, want)
	}
}

func TestRenameFileLeavesSkippedFilesUntouched(t *testing.T) {
	from := project.Names{Module: "example.com/you/notes", Brand: "Notes", Slug: "notes", App: "notes", EnvPrefix: "NOTES_"}
	to := project.Names{Module: "example.com/you/zenith", Brand: "Zenith", Slug: "zenith", App: "zenith", EnvPrefix: "ZENITH_"}
	dir := t.TempDir()
	// A lock file is regenerated after a rename, never rewritten.
	in := "example.com/you/notes v0.0.0 h1:abc=\n"
	if err := os.WriteFile(filepath.Join(dir, "go.sum"), []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	wasCalled := false
	extra := func(_, text string) (string, error) {
		wasCalled = true
		return text, nil
	}

	if err := project.RenameFile(dir, "go.sum", from, to, extra); err != nil {
		t.Fatalf("RenameFile error = %v", err)
	}

	if wasCalled {
		t.Error("extra was called for go.sum, want it skipped")
	}
	data, err := os.ReadFile(filepath.Join(dir, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != in {
		t.Errorf("go.sum = %q, want it unchanged: %q", data, in)
	}
}

func TestRenameFileNamesTheFileInItsErrors(t *testing.T) {
	from := project.Names{Module: "example.com/you/notes", Brand: "Notes", Slug: "notes", App: "notes", EnvPrefix: "NOTES_"}
	to := project.Names{Module: "example.com/you/zenith", Brand: "Zenith", Slug: "zenith", App: "zenith", EnvPrefix: "ZENITH_"}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "guide.md"), []byte("# Notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unbalanced := errors.New("unbalanced block")
	extra := func(_, _ string) (string, error) { return "", unbalanced }

	err := project.RenameFile(dir, "docs/guide.md", from, to, extra)

	if !errors.Is(err, unbalanced) {
		t.Fatalf("RenameFile error = %v, want it to wrap %v", err, unbalanced)
	}
	if !strings.Contains(err.Error(), "docs/guide.md") {
		t.Errorf("RenameFile error = %q, want it to name docs/guide.md", err)
	}
}
