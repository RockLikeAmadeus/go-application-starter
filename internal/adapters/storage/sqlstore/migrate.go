// Package sqlstore holds what the SQL storage adapters share: the goose
// migrations and the repositories, in SQL that both SQLite and Postgres run.
// The sqlite and postgres packages keep only what differs between them:
// opening the database, goose's dialect and readiness.
package sqlstore

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate brings db's schema up to date with the shared migrations. It
// refuses a database whose version is newer than the newest migration, and
// then changes nothing.
func Migrate(ctx context.Context, db *sql.DB, dialect goose.Dialect) error {
	sub, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("embedded migrations: %w", err)
	}
	provider, err := goose.NewProvider(dialect, db, sub)
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
