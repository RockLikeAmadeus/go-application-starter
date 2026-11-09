package main_test

import (
	"os"
	"path/filepath"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

// repoRoot is the template's root, seen from this package's directory.
var repoRoot = filepath.Join("..", "..")

func TestSampleFiles(t *testing.T) {
	for _, entry := range initproject.SampleFiles {
		if _, err := os.Stat(filepath.Join(repoRoot, entry)); err != nil {
			t.Errorf("sample entry %q is not in the template: %v", entry, err)
		}
	}
	// One file from each layer the sample spans must be covered, by itself or by
	// a directory that contains it.
	for _, f := range []string{
		"internal/core/domain/item.go",
		"internal/core/app/item_service.go",
		"internal/adapters/storage/memory/item_repository.go",
		"internal/adapters/cli/item.go",
		"features/items.feature",
		"web/src/features/items/ItemsPage.tsx",
		"web/e2e/features/items.feature",
	} {
		if !initproject.IsSample(f) {
			t.Errorf("SampleFiles does not cover %s", f)
		}
	}
}

// goose refuses to run with no migrations, so removing the sample must leave
// the SQLite store at least one migration of the template's own.
func TestSampleFilesLeaveTheTemplateASQLiteMigration(t *testing.T) {
	dir := "internal/adapters/storage/sqlstore/migrations"
	entries, err := os.ReadDir(filepath.Join(repoRoot, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".sql" && !initproject.IsSample(dir+"/"+e.Name()) {
			return
		}
	}
	t.Errorf("every migration in %s is a sample file; the template needs one of its own", dir)
}
