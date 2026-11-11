---
paths:
  - "internal/adapters/storage/**"
  - "internal/adapters/syncedfolder/**"
  - "internal/adapters/desktop/**"
  - "internal/bootstrap/**"
  - "internal/config/**"
  - "docs/data-locations.md"
---

# Storage

## Adding a backend

A new storage backend implements the `…Repository` ports and **must** run each one's contract suite from `storagetest` in its tests. Wire it in `internal/bootstrap`, including its readiness check (`App.Ready`, which answers the server's `/readyz`), as jsonfile's, sqlite's, and postgres's `Store.Ready` are. A SQL backend shares `sqlstore` (below) rather than writing its own repositories. Readiness comes from storage, never from an entity's data.

## jsonfile: one file per record (ADR 3)

Each record is `<data dir>/<table>/<key>.json`, holding `{"version": 2, "key": …, "record": …}`, written through `writeRecord` (temp file, fsync, rename) and read with `readRecord` (one file, for a lookup) or `readRecords` (the whole table). Use these for a new entity, never a file of your own. A file whose key doesn't match its name is a sync service's conflict copy: reading skips it, `Store.ConflictCopies` lists it, and `bootstrap` warns about it. While a folder looks synced (and `ignore_synced_folder` is off), `bootstrap` keeps this device's mark in `.devices/` with `syncedfolder.KeepMark`, and warns about other devices' recent marks (`syncedfolder.OtherDevices`); `App.Close` removes the mark. Plain text is local edition only, so a data folder holds one workspace.

## jsonfile: the data directory lock

The jsonfile adapter locks its data directory (`flock` on Unix, `LockFileEx` on Windows, nothing elsewhere; the `go-windows` CI job runs its tests on Windows, which nothing local can); sqlite doesn't (see below). The embedded CLI and TUI share the desktop app's data folder by default (`config.LocalDataDir`: `STARTER_DATA_DIR`, else the settings file's `data_dir`, else `~/.config/starter` on Linux; `starter info` prints it), so while the desktop app is open they fail fast with "Your data folder is in use by another program on this computer ({path}).", as does a second copy of the app. The server keeps `./data` (in the repository under `task dev`), so it doesn't clash with them; to reach its data, use `--server http://localhost:8080`. `task run:cli` and `task run:tui` use your real data folder, as `task run:desktop` does.

## SQLite (ADR 7)

SQLite storage (`storage = "sqlite"`) is `<data dir>/data.db`, opened with `modernc.org/sqlite`, a pure-Go driver: keep it that way, because the Windows tasks set `CGO_ENABLED=0`. It takes no lock: WAL and a busy timeout let several programs share it, and each store has one connection (`SetMaxOpenConns(1)`), so a query must close its rows before the next one starts. `sqlite.Open` applies the goose migrations embedded from `internal/adapters/storage/sqlstore/migrations/` (`sqlstore.Migrate`), and refuses a database whose version is newer than the newest one. A schema change is a new numbered migration, never an edit to an old one. `00001_baseline.sql` is template code and must stay: goose won't start with no migrations. `Store.Ready` pings and checks `data.db` is still the file it opened, because on Unix an open handle keeps a removed file alive. `desktop.Open` takes `jsonfile` or `sqlite` from the settings (never memory), and refuses `postgres` (below); the Android app stays on jsonfile. On SQLite, when `syncedfolder.Detect` says the data folder looks synced, `bootstrap` puts a warning in `App.Warnings`, which every front end shows (`GET /api/v1/warnings`, stderr, the TUI's warning line, the server's log); `ignore_synced_folder` turns it off.

## The shared SQL layer: `sqlstore` (ADR 8)

`internal/adapters/storage/sqlstore` holds what SQLite and Postgres share: the goose migrations (`sqlstore/migrations/`), `sqlstore.Migrate(ctx, db, dialect)`, and the `…Repository` implementations on a `*sql.DB` (`sqlstore.NewItemRepository(store.DB())`). `sqlite` and `postgres` keep only opening, goose's dialect and readiness. Write SQL that both engines run, and never two copies of a query: `$1` placeholders, `INSERT … ON CONFLICT (…) DO UPDATE SET …` for upserts, times as fixed-width UTC `TEXT`, IDs from `ports.IDGenerator` (no auto-increment), and migrations within what SQLite can do (rebuild a table rather than `ALTER COLUMN`). Each repository's contract test runs on SQLite and, with a test database, on Postgres. The Item repository and `00002_items.sql` are sample code; the runner, the baseline and the plan repository are template code.

## Postgres (ADR 8)

Postgres storage (`storage = "postgres"`) connects to `database_url` (`STARTER_DATABASE_URL`) through `github.com/jackc/pgx/v5/stdlib`, which is pure Go. A missing `database_url` is refused at start, and `starter info` shows a set one as `(hidden)`. `postgres.Open` runs `sqlstore.Migrate`, which takes goose's session lock on Postgres, so servers starting together don't race. `Store.Ready` pings, and pings again on `driver.ErrBadConn` (at most 3 attempts), so the first probe after an outage answers ready. The server, the CLI and the TUI honour `postgres` (`bootstrap.NewEmbedded`); `desktop.Open`, and with it the Android app, refuses it with "The storage setting is "postgres", but this application keeps its data on this device. Please use "jsonfile" or "sqlite"." before choosing a data folder: those apps reach hosted data through a server, never the database. Tests use `STARTER_TEST_DATABASE_URL` (`task db:up` starts a server; `task db:down` stops it) and `postgrestest.NewDatabase(t)`, a fresh database per test, dropped on cleanup; without the URL they skip.
