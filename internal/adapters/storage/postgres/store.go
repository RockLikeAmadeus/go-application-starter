// Package postgres is a storage adapter that keeps data in a Postgres
// database, through database/sql and the pure-Go pgx driver. Open brings the
// schema up to date with sqlstore's shared migrations.
package postgres

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
	"github.com/pressly/goose/v3"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
)

// Store owns the database connections.
type Store struct {
	db *sql.DB
}

// Open connects to the database at databaseURL and applies the migrations.
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := sqlstore.Migrate(ctx, db, goose.DialectPostgres); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// DB is the store's database, for the shared sqlstore repositories. It stays
// the store's: close the store, not the database.
func (s *Store) DB() *sql.DB {
	return s.db
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}
