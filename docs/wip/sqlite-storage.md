# WIP: Keep data in SQLite (SQLite storage adapter + migrations)

Temporary: delete this file in the same change that removes the `@wip` tags.

## Links
- Spec: [SPEC.md §4.14 "Keep data in SQLite"](../../SPEC.md) (also §4.6 and §4.11 now limit "one at a time" to plain text, and §5–6 are updated)
- Scenarios: [`features/sqlite-storage.feature`](../../features/sqlite-storage.feature); steps in `test/acceptance/sqlite_steps.go`
- Item plan: [docs/plans/editions-and-accounts.md](../plans/editions-and-accounts.md), step 6. TODO.md: "SQL storage adapter + migrations"
- Context: [ADR 3](../adr/0003-plain-text-storage.md) (jsonfile's single-instance limit), [ADR 5](../adr/0005-hosted-and-local-editions.md), [docs/workspaces.md](../workspaces.md)

## Decisions (2026-11-03)
- **Driver: `modernc.org/sqlite`** (pure Go). The Windows build and package tasks set `CGO_ENABLED=0`, and cgo would make Android harder too. `mattn/go-sqlite3` was ruled out for that reason.
- **Migrations: `pressly/goose` v3**, run as a library when the store opens, from SQL files embedded with `embed.FS`. Step 7 (Postgres) reuses it and shares the schema where it can.
- **Newer data is refused:** if the database's goose version is newer than the newest embedded migration, `Open` fails with a clear error ("saved by a newer version of the app") and changes nothing.
- **No lock: several programs at once.** WAL mode plus a busy timeout lets the desktop app, the CLI/TUI, and several servers on one disk share the database. This lifts ADR 3's single-instance limit for SQLite. jsonfile keeps its `flock`. §4.6 and §4.11's "one at a time" now applies only to plain text.
- **One file in the data folder:** `<data dir>/data.db` (with its `-wal`/`-shm` files). The name is generic, so `task rename` has nothing to rewrite.
- **Setting:** `storage = "sqlite"` (`STARTER_STORAGE=sqlite`), next to `memory` and `jsonfile`. The default stays `jsonfile`. Server, CLI/TUI and desktop app honour it. `desktop.Open` stops forcing jsonfile: it takes `jsonfile` or `sqlite` from the settings and still never uses memory. The mobile app stays on jsonfile, because it has no settings file.
- **Readiness:** `Store.Ready` runs a query (a ping) **and** checks that the database file at its path is still the one it opened (`os.SameFile`). On Unix, an open handle keeps a removed file alive, so a ping alone would say "ready" while writes go nowhere.
- **Workspace isolation:** every query filters by `workspace_id`. Item's repository runs `storagetest.RunItemRepositoryContract`, which includes `RunWorkspaceIsolation`. Plan limits count through `FindAll`, as on the other adapters, so there's no extra query to isolate.
- **Template vs sample:** `store.go` (open, pragmas, goose, Ready) and the migration runner are template code. The items migration and `item_repository.go` are sample files and go in `SampleFiles` (`scripts/initproject/manifest.go`). The template's embedded migrations must still compile, and `Open` must still work, with the sample migration gone. That may need a template baseline migration; decide in that cycle. Run `task test:init` there.
- **Deferred:** the warning when SQLite's folder looks synced (Dropbox, iCloud, Syncthing, …). It's a follow-up in the editions plan and goes into TODO.md when this step finishes. Step 5 (sync-safe plain text) is still not started. This step comes first because you asked for it; neither depends on the other.
- **Acceptance:** a new `sqlite` driver runs every business scenario (same tags as jsonfile: workspace and plan scenarios included). The `@wip` scenarios here run on the desktop driver (settings file → SQLite → reopen) and on the server driver (`testServer.restart` on SQLite). The server driver can now restart on another storage. That's a test-only change, already made.

## Tasks
1. [x] **config:** `storage = "sqlite"` is accepted (`config.StorageSQLite`), and the error for an unknown storage names all three. Red: `undefined: config.StorageSQLite` (one test, `TestStorageAcceptsSQLite…`). Green: constant, `validate` case and error text. For refactor: none (`cli.keepsDataInAFolder` and `desktop.Open` need sqlite in tasks 7–8).
2. [x] **sqlite.Open(dir)** creates `<dir>/data.db`, turns on WAL, a busy timeout and foreign keys, and applies the embedded goose migrations. Opening it a second time in the same directory works (no lock), and `Close` releases it. Red: package didn't exist (one test: WAL, goose version, second Open). Green: `store.go` (pragmas in the DSN, goose provider over embedded `migrations/`) and a template baseline migration `00001_baseline.sql` (`SELECT 1`), because goose fails with no migrations; task 10 must decide whether it stays. For refactor: none (imports tidied). `foreign_keys`/`busy_timeout` aren't asserted: they're per-connection, so only task 6 can show them.
3. [x] **Newer data refused:** `Open` fails, naming a newer version of the app, when the database's migration version is ahead of the newest embedded migration. Red: `TestOpenRefusesADatabaseSavedByANewerVersionOfTheApp`, "Open succeeded" (goose ignores an unknown newer version). Green: `migrate` compares `GetDBVersion` with the newest of `ListSources` before `Up`.
   Refactor: the check moved into `refuseNewer`. Its `ListSources()[len-1]` relies on at least one migration (goose refuses none), which task 10 must keep in mind.
4. [x] **Store.Ready** is nil on an open store, and fails once the data folder (and so the database file) is removed. It checks afresh each time. Red: `TestStoreIsReadyUntilItsDataFolderIsRemoved`, "Ready error = nil" after removal (stub returned nil). Green: `Ready` pings, then `os.SameFile` against the `FileInfo` that `Open` took of `data.db`.
   Refactor: the DSN moved into `dsn(path)`, so `Open` reads as steps. A test for a *replaced* file (SameFile) was left out: removal already shows it, and a separate guard would add little.
5. [ ] **(sample) ItemRepository passes `storagetest.RunItemRepositoryContract`**, workspace isolation included: an `items` migration plus `Save`/`FindByID`/`FindAll`/`Delete`. Expect two cycles: Save + FindByID, then FindAll ordering + Delete. Times round-trip in UTC, at full precision.
   - Cycle 1 (Save + FindByID). Red: `TestItemRepositoryFindsASavedItemWithItsTimesInUTCAtFullPrecision`, zero Item back (stubs).
     Green: `00002_items.sql` (key `workspace_id, id`; times as fixed-width UTC text), `Save` (plain INSERT), `FindByID` (filters by workspace). The three files went into `SampleFiles` now, not in task 10, so `task test:init` holds.
     Refactor: `scanItem` (for `FindAll` next), `formatTime`, and the shared `itemColumns`.
     `task test:init` is red only because its plain `go test ./test/...` runs the `@wip` scenarios (until task 9).
6. [ ] **(sample) Two stores on one data folder** (two programs) each see what the other saved, and writing from both doesn't fail with "database is locked".
7. [ ] **bootstrap:** `NewEmbedded` with `StorageSQLite` wires the store: items saved survive a reopen, `Close` closes it, and `Ready` fails once the data folder is removed.
8. [ ] **desktop.Open** honours `storage = "sqlite"` from the settings file (jsonfile otherwise, never memory). Two `desktop.Open`s on SQLite in one folder both work.
9. [ ] **Acceptance `sqlite` driver** in `test/acceptance/drivers.go` (tags like jsonfile's). Run all of `task accept`, and get the `@wip` scenarios green.
10. [ ] **Init:** add the items migration and `item_repository*.go` to `SampleFiles`. Make sure the template's migrations embed and `Open` still work without them. Run `task test:init`.
11. [ ] **Docs and finish:**
   - `docs/configuration.md` (the storage values);
   - a new ADR 7 (SQLite: modernc, goose, WAL, no lock, `data.db`), and ADR 3's consequences pointing to it;
   - CLAUDE.md: the `storage/*` line, and Gotchas (pure-Go driver, SQLite has no lock, readiness);
   - `docs/data-locations.md` and the README if they list storage;
   - `task check`, remove the `@wip` tags, delete this file;
   - mark step 6 done in the editions plan, and tick the TODO item (adding the synced-folder warning follow-up to TODO.md).

## Changes to the plan
