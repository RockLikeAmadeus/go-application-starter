package bootstrap_test

// The tests in this file use the sample Item API, so `task init --remove-sample`
// deletes the file (see scripts/initproject/manifest.go).

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres/postgrestest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// Items saved on SQLite are still there for the next program to open the
// same data folder.
func TestNewEmbedded_OnSQLiteKeepsItemsAfterClosingAndReopening(t *testing.T) {
	assertKeepsItemsAfterClosingAndReopening(t, config.Config{Storage: config.StorageSQLite, DataDir: t.TempDir()})
}

// Items saved on Postgres are still there for the next program to open the
// same database.
func TestNewEmbedded_OnPostgresKeepsItemsAfterClosingAndReopening(t *testing.T) {
	assertKeepsItemsAfterClosingAndReopening(t, config.Config{Storage: config.StoragePostgres, DatabaseURL: postgrestest.NewDatabase(t)})
}

func assertKeepsItemsAfterClosingAndReopening(t *testing.T, cfg config.Config) {
	t.Helper()

	first, err := bootstrap.NewEmbedded(cfg)
	if err != nil {
		t.Fatalf("first NewEmbedded: %v", err)
	}
	if _, err := first.Items.CreateItem(t.Context(), "Widget"); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := bootstrap.NewEmbedded(cfg)
	if err != nil {
		t.Fatalf("second NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })
	items, err := second.Items.ListItems(t.Context())
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}
	if len(items) != 1 || items[0].Name != "Widget" {
		t.Errorf("items after reopening = %+v, want just Widget", items)
	}
}
