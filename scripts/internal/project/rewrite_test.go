package project_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

func TestRewriteReplacesOneProjectsNamesWithAnothers(t *testing.T) {
	// The old deploy app name differs from its slug, and contains it.
	from := project.Names{
		Module:    "example.com/old/notes",
		Brand:     "Old Notes",
		Slug:      "old-notes",
		App:       "old-notes-app",
		EnvPrefix: "OLD_NOTES_",
	}
	to := project.Names{
		Module:    "example.com/zenith",
		Brand:     "Zenith",
		Slug:      "zenith",
		App:       "zenith",
		EnvPrefix: "ZENITH_",
	}
	in := "module example.com/old/notes\n" +
		`import "example.com/old/notes/internal/core/domain"` + "\n" +
		"app = 'old-notes-app-staging'\n" +
		`"name": "old-notes-web",` + "\n" +
		`getenv("OLD_NOTES_STORAGE", "jsonfile")` + "\n" +
		"<title>Old Notes</title>\n" +
		"go run ./cmd/old-notes tui\n" +
		"Old Notesy and old-notesy stay\n"
	want := "module example.com/zenith\n" +
		`import "example.com/zenith/internal/core/domain"` + "\n" +
		"app = 'zenith-staging'\n" +
		`"name": "zenith-web",` + "\n" +
		`getenv("ZENITH_STORAGE", "jsonfile")` + "\n" +
		"<title>Zenith</title>\n" +
		"go run ./cmd/zenith tui\n" +
		"Old Notesy and old-notesy stay\n"

	if got := project.Rewrite(in, from, to); got != want {
		t.Errorf("Rewrite(%q) =\n%q\nwant\n%q", in, got, want)
	}
}

func TestRewriteLeavesAnUnchangedModulePathIntact(t *testing.T) {
	// The module path stays, though it contains the old slug as a whole word.
	from := project.Names{
		Module:    "example.com/you/old-notes",
		Brand:     "Old Notes",
		Slug:      "old-notes",
		App:       "old-notes",
		EnvPrefix: "OLD_NOTES_",
	}
	to := project.Names{
		Module:    from.Module,
		Brand:     "Zenith",
		Slug:      "zenith",
		App:       "zenith",
		EnvPrefix: "ZENITH_",
	}
	in := "module example.com/you/old-notes\n" +
		`import "example.com/you/old-notes/internal/core/domain"` + "\n" +
		"go run ./cmd/old-notes tui\n"
	want := "module example.com/you/old-notes\n" +
		`import "example.com/you/old-notes/internal/core/domain"` + "\n" +
		"go run ./cmd/zenith tui\n"

	if got := project.Rewrite(in, from, to); got != want {
		t.Errorf("Rewrite(%q) =\n%q\nwant\n%q", in, got, want)
	}
}

func TestRewriteTreatsTheOldBrandLiterally(t *testing.T) {
	// "." in a pattern would match any character, so "NotesXio" must survive.
	from := project.Names{
		Module:    "example.com/you/notes",
		Brand:     "Notes.io",
		Slug:      "notes-io",
		App:       "notes-io",
		EnvPrefix: "NOTES_IO_",
	}
	to := project.Names{
		Module:    "example.com/you/zenith",
		Brand:     "Zenith",
		Slug:      "zenith",
		App:       "zenith",
		EnvPrefix: "ZENITH_",
	}
	in := "<title>Notes.io</title>\nNotesXio is someone else's product.\n"
	want := "<title>Zenith</title>\nNotesXio is someone else's product.\n"

	if got := project.Rewrite(in, from, to); got != want {
		t.Errorf("Rewrite(%q) =\n%q\nwant\n%q", in, got, want)
	}
}

func TestRewriteGivesTheAndroidAppAnIDWithoutHyphens(t *testing.T) {
	// An Android application ID is a Java package name: a hyphen makes it
	// invalid, so the slug's hyphens are dropped.
	from := project.Names{
		Module:    "example.com/you/old-notes",
		Brand:     "Old Notes",
		Slug:      "old-notes",
		App:       "old-notes",
		EnvPrefix: "OLD_NOTES_",
	}
	to := project.Names{
		Module:    "example.com/you/acme-tasks",
		Brand:     "Acme Tasks",
		Slug:      "acme-tasks",
		App:       "acme-tasks",
		EnvPrefix: "ACME_TASKS_",
	}
	in := `applicationId "com.example.oldnotes"` + "\n" +
		"adb shell am start -n com.example.oldnotes/com.wails.app.MainActivity\n"
	want := `applicationId "com.example.acmetasks"` + "\n" +
		"adb shell am start -n com.example.acmetasks/com.wails.app.MainActivity\n"

	if got := project.Rewrite(in, from, to); got != want {
		t.Errorf("Rewrite(%q) =\n%q\nwant\n%q", in, got, want)
	}
}

func TestRenamePathMovesTheCommandLineTool(t *testing.T) {
	from := project.Names{Slug: "old-notes"}
	to := project.Names{Slug: "zenith"}
	for path, want := range map[string]string{
		"cmd/old-notes/main.go":      "cmd/zenith/main.go",
		"cmd/old-notes/main_test.go": "cmd/zenith/main_test.go",
		"cmd/server/main.go":         "cmd/server/main.go",
		"cmd/old-notes-kit/main.go":  "cmd/old-notes-kit/main.go",
		"docs/old-notes/notes.md":    "docs/old-notes/notes.md",
	} {
		if got := project.RenamePath(path, from, to); got != want {
			t.Errorf("RenamePath(%q) = %q, want %q", path, got, want)
		}
	}
}
