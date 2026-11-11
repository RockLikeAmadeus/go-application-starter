// Package sqlstore holds what the SQL storage adapters share: the goose
// migrations and the repositories, in SQL that both SQLite and Postgres run.
// The sqlite and postgres packages keep only what differs between them:
// opening the database, goose's dialect, and readiness.
package sqlstore

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed migrations/*.sql
var migrations embed.FS

// A server waiting for another's migration lock tries again every second, for
// up to five minutes: goose's limit, at a shorter interval than its default of
// five seconds, so that a waiting server starts sooner.
const (
	lockPollSeconds = 1
	lockPollCount   = 300
)

// Migrate brings db's schema up to date with the shared migrations. It
// refuses a database whose version is newer than the newest migration, and
// then changes nothing. On Postgres it holds a session lock while it works, so
// that servers starting at once don't race.
func Migrate(ctx context.Context, db *sql.DB, dialect goose.Dialect) error {
	sub, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("embedded migrations: %w", err)
	}
	options, err := lockOptions(dialect)
	if err != nil {
		return fmt.Errorf("prepare migration lock: %w", err)
	}
	provider, err := goose.NewProvider(dialect, db, sub, options...)
	if err != nil {
		return fmt.Errorf("prepare migrations: %w", err)
	}
	if err := refuseNewer(ctx, provider); err != nil {
		return err
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// lockOptions makes servers starting at once on one Postgres database take
// turns to migrate it. SQLite needs no lock: its busy timeout makes writers
// wait.
func lockOptions(dialect goose.Dialect) ([]goose.ProviderOption, error) {
	if dialect != goose.DialectPostgres {
		return nil, nil
	}
	locker, err := lock.NewPostgresSessionLocker(lock.WithLockTimeout(lockPollSeconds, lockPollCount))
	if err != nil {
		return nil, err
	}
	return []goose.ProviderOption{goose.WithSessionLocker(locker)}, nil
}

// refuseNewer fails if a newer app has migrated the database past the newest
// migration this one embeds: its schema may not be one this app can use, so
// Migrate changes nothing and stops.
func refuseNewer(ctx context.Context, provider *goose.Provider) error {
	current, err := provider.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("read database version: %w", err)
	}
	sources := provider.ListSources() // never empty: goose refuses to start with no migrations
	if newest := sources[len(sources)-1].Version; current > newest {
		return fmt.Errorf("the data was saved by a newer version of the app (database version %d, this app knows up to %d)", current, newest)
	}
	return nil
}
