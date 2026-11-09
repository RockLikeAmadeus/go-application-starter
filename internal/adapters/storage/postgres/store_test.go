package postgres_test

import (
	"database/sql"
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

	// Reads the database the way another program would.
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open the database: %v", err)
	}
	defer db.Close()
	var version int
	if err := db.QueryRow("SELECT MAX(version_id) FROM goose_db_version").Scan(&version); err != nil || version < 1 {
		t.Errorf("goose version = %d, %v; want migrations applied", version, err)
	}
}
