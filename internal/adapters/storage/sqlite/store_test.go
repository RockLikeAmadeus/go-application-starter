package sqlite_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite" // reads the database file the way another program would

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlite"
)

func TestOpenCreatesAMigratedWALDatabaseThatSeveralStoresShare(t *testing.T) {
	dir := t.TempDir()
	first, err := sqlite.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	second, err := sqlite.Open(dir)
	if err != nil {
		t.Fatalf("second Open in the same directory: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, "data.db"))
	if err != nil {
		t.Fatalf("open data.db: %v", err)
	}
	defer db.Close()
	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil || mode != "wal" {
		t.Errorf("journal_mode = %q, %v; want wal", mode, err)
	}
	var version int
	if err := db.QueryRow("SELECT MAX(version_id) FROM goose_db_version").Scan(&version); err != nil || version < 1 {
		t.Errorf("goose version = %d, %v; want migrations applied", version, err)
	}

	if err := first.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestOpenRefusesADatabaseSavedByANewerVersionOfTheApp(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// A newer app applied a migration this one doesn't know.
	db, err := sql.Open("sqlite", filepath.Join(dir, "data.db"))
	if err != nil {
		t.Fatalf("open data.db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec("INSERT INTO goose_db_version (version_id, is_applied) VALUES (99999, 1)"); err != nil {
		t.Fatalf("record a newer migration: %v", err)
	}

	store, err = sqlite.Open(dir)
	if err == nil {
		store.Close()
		t.Fatal("Open succeeded; want it to refuse a newer database")
	}
	if !strings.Contains(err.Error(), "newer version of the app") {
		t.Errorf("Open error = %q; want it to say a newer version of the app saved the data", err)
	}
}

func TestStoreIsReadyUntilItsDataFolderIsRemoved(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()

	if err := store.Ready(t.Context()); err != nil {
		t.Errorf("Ready error = %v, want nil while the database is there", err)
	}
	// An open handle keeps a removed file alive on Unix, so a ping alone
	// would still succeed here.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := store.Ready(t.Context()); err == nil {
		t.Error("Ready error = nil, want an error once the data folder is gone")
	}
}
