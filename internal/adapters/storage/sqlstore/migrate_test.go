package sqlstore_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // registers the "sqlite" driver

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
)

// openSQLite opens an empty SQLite database, closed when the test ends.
func openSQLite(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMigrateAppliesTheSharedMigrations(t *testing.T) {
	db := openSQLite(t)

	if err := sqlstore.Migrate(t.Context(), db, goose.DialectSQLite3); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	var version int
	if err := db.QueryRow("SELECT MAX(version_id) FROM goose_db_version").Scan(&version); err != nil || version < 1 {
		t.Errorf("goose version = %d, %v; want migrations applied", version, err)
	}
}

func TestMigrateRefusesADatabaseSavedByANewerVersionOfTheApp(t *testing.T) {
	db := openSQLite(t)
	if err := sqlstore.Migrate(t.Context(), db, goose.DialectSQLite3); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	// A newer app applied a migration this one doesn't know.
	if _, err := db.Exec("INSERT INTO goose_db_version (version_id, is_applied) VALUES (99999, 1)"); err != nil {
		t.Fatalf("record a newer migration: %v", err)
	}

	err := sqlstore.Migrate(t.Context(), db, goose.DialectSQLite3)
	if err == nil {
		t.Fatal("Migrate succeeded; want it to refuse a newer database")
	}
	if !strings.Contains(err.Error(), "newer version of the app") {
		t.Errorf("Migrate error = %q; want it to say a newer version of the app saved the data", err)
	}
}
