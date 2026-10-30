package project_test

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

func TestRecordRoundTrip(t *testing.T) {
	dir := t.TempDir()
	n := project.Names{
		Module:    "example.com/you/notes",
		Brand:     "Old Notes",
		Slug:      "old-notes",
		App:       "old-notes-app",
		EnvPrefix: "OLD_NOTES_",
	}

	if err := project.WriteRecord(dir, n); err != nil {
		t.Fatalf("WriteRecord error = %v", err)
	}
	got, err := project.ReadRecord(dir)
	if err != nil {
		t.Fatalf("ReadRecord error = %v", err)
	}
	if got != n {
		t.Errorf("ReadRecord = %+v, want %+v", got, n)
	}

	// The environment variable prefix is derived from the slug, not stored.
	data, err := os.ReadFile(filepath.Join(dir, "project.json"))
	if err != nil {
		t.Fatalf("project.json was not written: %v", err)
	}
	var stored map[string]string
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("project.json is not a JSON object of strings: %v\n%s", err, data)
	}
	want := map[string]string{"module": n.Module, "name": n.Brand, "slug": n.Slug, "app": n.App}
	if !maps.Equal(stored, want) {
		t.Errorf("project.json holds %v, want %v", stored, want)
	}
}
