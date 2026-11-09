// Package sqlite is a storage adapter that keeps all data in one SQLite
// database file, <dir>/data.db, using the pure-Go modernc.org/sqlite driver.
//
// Write-ahead logging and a busy timeout let several programs share one
// database, so unlike jsonfile there is no lock on the data directory.
// Open brings the schema up to date with sqlstore's shared migrations.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // registers the "sqlite" driver

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
)

// Store owns the database connection. Repositories for individual domain
// types are created from it.
type Store struct {
	db     *sql.DB
	path   string
	opened os.FileInfo
}

// Open creates dir if needed, opens <dir>/data.db and applies the migrations.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join(dir, "data.db")
	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	// SQLite lets one connection write at a time. With one connection per
	// store, this program's writes queue here rather than in SQLite's busy
	// wait, where many at once outlast the timeout and fail with "database
	// is locked"; only other programs then contend for the lock. A query
	// must therefore finish (close its rows) before the next one starts.
	db.SetMaxOpenConns(1)
	if err := sqlstore.Migrate(context.Background(), db, goose.DialectSQLite3); err != nil {
		db.Close()
		return nil, err
	}
	opened, err := os.Stat(path)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("database file: %w", err)
	}
	return &Store{db: db, path: path, opened: opened}, nil
}

// dsn names the database file with the pragmas every connection needs.
// Pragmas in the DSN apply to every connection the pool opens, not just the
// first.
func dsn(path string) string {
	q := url.Values{}
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "foreign_keys(1)")
	return "file:" + filepath.ToSlash(path) + "?" + q.Encode()
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Ready reports whether the store can still use its database: a query
// succeeds and the file at its path is still the one Open opened. On Unix an
// open handle keeps a removed file alive, so a ping alone would say "ready"
// while writes go nowhere. It checks afresh each time.
func (s *Store) Ready(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return err
	}
	current, err := os.Stat(s.path)
	if err != nil {
		return err
	}
	if !os.SameFile(s.opened, current) {
		return fmt.Errorf("database file %s was replaced", s.path)
	}
	return nil
}
