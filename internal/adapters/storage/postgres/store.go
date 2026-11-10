// Package postgres is a storage adapter that keeps data in a Postgres
// database, through database/sql and the pure-Go pgx driver. Open brings the
// schema up to date with sqlstore's shared migrations.
package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"

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

// readyPingAttempts bounds how many pings Ready makes while they land on dead
// pooled connections.
const readyPingAttempts = 3

// Ready reports whether the database can be reached. After an outage the
// pool may still hold connections the database ended, and a ping on one
// answers driver.ErrBadConn without trying another, so Ready pings again,
// up to readyPingAttempts times.
func (s *Store) Ready(ctx context.Context) error {
	var err error
	for range readyPingAttempts {
		err = s.db.PingContext(ctx)
		if !errors.Is(err, driver.ErrBadConn) {
			return err
		}
	}
	return err
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}
