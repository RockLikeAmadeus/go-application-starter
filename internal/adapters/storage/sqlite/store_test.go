package sqlite_test

import (
	"database/sql"
	"path/filepath"
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
