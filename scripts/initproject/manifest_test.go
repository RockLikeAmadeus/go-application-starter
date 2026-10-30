package main_test

import (
	"os"
	"path/filepath"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

func TestSampleFiles(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, entry := range initproject.SampleFiles {
		if _, err := os.Stat(filepath.Join(root, entry)); err != nil {
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
