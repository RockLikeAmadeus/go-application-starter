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

A new storage backend implements the `…Repository` ports and **must** run each one's contract suite from `storagetest` in its tests. Wire it in `internal/bootstrap`, including its readiness check (`App.Ready`, which answers the server's `/readyz`), as jsonfile's and sqlite's `Store.Ready` are. Readiness comes from storage, never from an entity's data.

## jsonfile: the data directory lock

The jsonfile adapter locks its data directory (`flock` on Unix, `LockFileEx` on Windows, nothing elsewhere; the `go-windows` CI job runs its tests on Windows, which nothing local can); sqlite doesn't (see below). The embedded CLI and TUI share the desktop app's data folder by default (`config.LocalDataDir`: `STARTER_DATA_DIR`, else the settings file's `data_dir`, else `~/.config/starter` on Linux; `starter info` prints it), so while the desktop app is open they fail fast with "data directory is locked", as does a second copy of the app. The server keeps `./data` (in the repository under `task dev`), so it doesn't clash with them; to reach its data, use `--server http://localhost:8080`. `task run:cli` and `task run:tui` use your real data folder, as `task run:desktop` does.

## SQLite (ADR 7)

SQLite storage (`storage = "sqlite"`) is `<data dir>/data.db`, opened with `modernc.org/sqlite`, a pure-Go driver: keep it that way, because the Windows tasks set `CGO_ENABLED=0`. It takes no lock: WAL and a busy timeout let several programs share it, and each store has one connection (`SetMaxOpenConns(1)`), so a query must close its rows before the next one starts. `sqlite.Open` applies the goose migrations embedded from `internal/adapters/storage/sqlite/migrations/`, and refuses a database whose version is newer than the newest one. A schema change is a new numbered migration, never an edit to an old one. `00001_baseline.sql` is template code and must stay: goose won't start with no migrations. `Store.Ready` pings and checks `data.db` is still the file it opened, because on Unix an open handle keeps a removed file alive. `desktop.Open` takes `jsonfile` or `sqlite` from the settings (never memory); the Android app stays on jsonfile. On SQLite, when `syncedfolder.Detect` says the data folder looks synced, `bootstrap` puts a warning in `App.Warnings`, which every front end shows (`GET /api/v1/warnings`, stderr, the TUI's warning line, the server's log); `ignore_synced_folder` turns it off.
