# 7. SQLite storage: one database file, shared by several programs

- Status: accepted
- Date: 2026-11-03

## Context

The `jsonfile` adapter (ADR 3) lets one program at a time use a data folder. With the desktop app open, the embedded CLI and TUI can't touch the data, and a server can't run more than one copy on a volume. The editions plan (ADR 5, step 6) gives the local edition a second format, SQLite, which the user chooses, and the SQL adapter is the pattern Postgres (step 7) will follow.

The app is built for Windows with `CGO_ENABLED=0` (the Windows build and setup tasks), and the Android app is cross-compiled with the NDK, so a driver that needs cgo would make both harder.

## Decision

- **Driver: `modernc.org/sqlite`**, a pure-Go translation of SQLite. It builds wherever Go does, with `CGO_ENABLED=0`. `mattn/go-sqlite3` was ruled out because it needs cgo.
- **Migrations: `pressly/goose` v3**, run as a library by `sqlite.Open`, from SQL files embedded in the binary (`internal/adapters/storage/sqlite/migrations/`). Opening the data brings its schema up to date; there's no separate migrate command. Postgres will reuse goose and share the schema where it can.
- **Newer data is refused.** If the database's goose version is ahead of the newest embedded migration, `Open` fails with "the data was saved by a newer version of the app" and changes nothing.
- **A template baseline migration** (`00001_baseline.sql`, a no-op) stays when the sample is removed: goose won't start with no migrations, and the newer-data check needs a newest one.
- **One file:** `<data dir>/data.db`, plus SQLite's `-wal` and `-shm` files beside it. The name is generic, so `task rename` has nothing to rewrite.
- **No lock: several programs at once.** Write-ahead logging (WAL) and a 5-second busy timeout let the desktop app, the CLI and TUI, and several servers on one disk share the database. Each store keeps one connection (`SetMaxOpenConns(1)`), so a program's own writes queue in Go and only separate programs contend for SQLite's write lock. This lifts ADR 3's single-program limit for SQLite; jsonfile keeps its `flock`.
- **Readiness** (`Store.Ready`, behind `/readyz`) pings the database **and** checks that the file at its path is still the one `Open` opened (`os.SameFile`). On Unix an open handle keeps a removed file alive, so a ping alone would say "ready" while writes go nowhere.
- **Workspaces:** every table has a `workspace_id` column, part of the primary key, and every query filters by it. The adapter passes the same contract suites, isolation included, as memory and jsonfile.
- **The setting:** `storage = "sqlite"` (`STARTER_STORAGE=sqlite`, `--storage sqlite`). The default stays `jsonfile`. The server, the CLI and TUI, and the desktop app honour it; the Android app stays on jsonfile, since it has no settings file.

## Consequences

- With `storage = "sqlite"`, the CLI and TUI work while the desktop app is open, and several server processes can share one disk.
- Several **machines** still can't share one database: SQLite's locking needs one disk, not a network share or a Fly volume per machine. Scaling a server out needs Postgres (step 7).
- A data folder in a file-sync service (Dropbox, iCloud Drive, Syncthing, …) can corrupt a SQLite database, which plain text survives. When the folder looks synced (its path, or a sync service's marker files in it or above it: `internal/adapters/syncedfolder`), every front end warns, naming the service, and keeps working. The `ignore_synced_folder` setting turns the warning off.
- A new entity on SQLite needs a migration and a repository, and its own isolation cases for any extra query (`docs/workspaces.md`). A schema change is a new migration file, never an edit to an old one.
- The data isn't human-readable any more; `sqlite3 data.db` opens it.
