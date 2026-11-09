package postgres_test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres/postgrestest"
)

func TestOpenBringsTheSchemaUpToDate(t *testing.T) {
	databaseURL := postgrestest.NewDatabase(t)

	store, err := postgres.Open(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()

	db := openDirectly(t, databaseURL)
	var version int
	if err := db.QueryRow("SELECT MAX(version_id) FROM goose_db_version").Scan(&version); err != nil || version < 1 {
		t.Errorf("goose version = %d, %v; want migrations applied", version, err)
	}
}

func TestOpenRefusesADatabaseSavedByANewerVersionOfTheApp(t *testing.T) {
	databaseURL := postgrestest.NewDatabase(t)
	store, err := postgres.Open(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	store.Close()

	// A newer app applied a migration this one doesn't know.
	db := openDirectly(t, databaseURL)
	if _, err := db.Exec("INSERT INTO goose_db_version (version_id, is_applied) VALUES (99999, true)"); err != nil {
		t.Fatalf("record a newer migration: %v", err)
	}
	before := migrationRecords(t, db)

	store, err = postgres.Open(t.Context(), databaseURL)
	if err == nil {
		store.Close()
		t.Fatal("Open succeeded; want it to refuse a newer database")
	}
	if !strings.Contains(err.Error(), "newer version of the app") {
		t.Errorf("Open error = %q; want it to say a newer version of the app saved the data", err)
	}
	if after := migrationRecords(t, db); after != before {
		t.Errorf("migration records = %d after the refusal, %d before; want no change", after, before)
	}
}

func TestStoreIsReadyOnlyWhileItsDatabaseCanBeReached(t *testing.T) {
	databaseURL := postgrestest.NewDatabase(t)
	store, err := postgres.Open(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()

	if err := store.Ready(t.Context()); err != nil {
		t.Fatalf("Ready error = %v while the database can be reached, want nil", err)
	}

	postgrestest.BlockConnections(t, databaseURL)
	if err := store.Ready(t.Context()); err == nil {
		t.Error("Ready error = nil while connections are blocked, want an error")
	}

	postgrestest.AllowConnections(t, databaseURL)
	if err := store.Ready(t.Context()); err != nil {
		t.Errorf("Ready error = %v once connections are allowed again, want nil", err)
	}
}

// openDirectly opens the database the way another program would, closed when
// the test ends.
func openDirectly(t *testing.T, databaseURL string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open the database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// migrationRecords counts the migrations goose has recorded in db.
func migrationRecords(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM goose_db_version").Scan(&count); err != nil {
		t.Fatalf("count migration records: %v", err)
	}
	return count
}
