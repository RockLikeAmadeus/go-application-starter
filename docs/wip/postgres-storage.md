# Keep data in Postgres

Step 7 of [the editions plan](../plans/editions-and-accounts.md).

- Spec: [SPEC.md §4.17](../../SPEC.md#417-keep-data-in-postgres)
- Scenarios: [`features/postgres-storage.feature`](../../features/postgres-storage.feature) (`@server @postgres`, 4 scenarios, each tagged `@wip`; the last is sample)
- Decisions: [ADR 7](../adr/0007-sqlite-storage.md) (to update), a new ADR 8 for Postgres (task 13), [ADR 5](../adr/0005-hosted-and-local-editions.md)

## Decisions

Agreed with the user:
- **One source for the SQL, shared by both engines.** One package, `internal/adapters/storage/sqlstore`, holds the goose migrations and the `…Repository` implementations, in SQL that SQLite and Postgres both run. `sqlite` and `postgres` keep only what really differs: opening the database, goose's dialect, and readiness. A new entity is written once (`docs/workspaces.md` changes to match).
  - Portable SQL: `INSERT … ON CONFLICT (…) DO UPDATE SET …` instead of `INSERT OR REPLACE`; times stay fixed-width UTC `TEXT`; IDs come from `ports.IDGenerator`, so there's no auto-increment; one placeholder style, `$1` (task 2 checks modernc binds it; if not, a small rebind helper in `sqlstore`, never two copies of a query).
  - Migrations stay within what SQLite can do: rebuild a table rather than `ALTER COLUMN`. If that ever bites, a goose Go migration can branch on the dialect for that one change.
- **Test database: a URL plus compose.** Tests read `STARTER_TEST_DATABASE_URL`. `task db:up` starts Postgres from a `postgres` service in `docker-compose.yml`, and CI uses a Postgres service container. With the URL unset, Postgres tests skip with a message (`t.Skip`, and the acceptance drivers leave them out with a note). With it set but unreachable, they fail. CI always sets it.
- **Who honours `storage = "postgres"`:** the server, the CLI and the TUI (all through `bootstrap.NewEmbedded`), so an operator can run `starter admin set-plan` on the hosted data. `desktop.Open` refuses it: the desktop and mobile apps reach hosted data through a server (remote mode, step 10), never the database.
- **The connection is a new setting, `database_url`** (`STARTER_DATABASE_URL`), server-oriented. Postgres without it fails at start.

Decided in planning:
- **Driver: `github.com/jackc/pgx/v5/stdlib`** through `database/sql`, so `sqlstore` takes a `*sql.DB`. It's pure Go, so `CGO_ENABLED=0` still builds.
- **Per-test isolation:** a test helper (`postgres/postgrestest`) creates a fresh database with a random name from the test URL, and drops it on cleanup. Each contract run and each acceptance scenario gets its own.
- **Unreachable database in the scenarios:** the helper can block connections to a scenario's database (`ALTER DATABASE … ALLOW_CONNECTIONS false` plus `pg_terminate_backend`) and allow them again, so readiness can fail and recover without stopping the shared container.
- **Readiness:** `postgres.Store.Ready` pings with the request's context. A pool reconnects by itself, so recovery needs no code beyond the ping.
- **No connection-count tuning** (`SetMaxOpenConns`) beyond pgx's defaults for now. SQLite keeps its single connection.
- **The acceptance `postgres` driver** runs what the `sqlite` driver runs (`~@desktop && ~@mobile && ~@server`). The server driver adds `~@postgres` when the URL is unset.
- **Sample boundary:** the Item repository and `00002_items.sql` are sample code in `sqlstore`. The migration runner, the baseline and the plan repository are template code. Run `task test:init` in tasks 1–3 and 10–11.
- **Template-user step:** a hosted deployment needs a Postgres database and `STARTER_DATABASE_URL` as a secret (Fly Postgres or any other). It goes under the TODO.md item for the starter TODO list in task 7.

Prerequisite: Docker, usable without `sudo`, from task 4 on. It's installed here, but this user can't reach `/var/run/docker.sock` yet. Task 4 records it in README Prerequisites.

## Tasks

Each is one `/red` → `/green` → `/refactor` cycle.

The shared SQL layer (on SQLite, so it runs without Postgres):
- [x] 1. `sqlstore.Migrate(ctx, db, dialect)` applies the shared migrations (moved from `sqlite/migrations/` to `sqlstore/migrations/`) and refuses a database whose version is newer than the newest one. `sqlite.Open` calls it. The migration tests move from `sqlite/store_test.go`.
  - Red: `TestMigrateAppliesTheSharedMigrations` failed: no `goose_db_version` table (stub `Migrate` applied nothing).
  - Green: migrations and `migrate`/`refuseNewer` moved from `sqlite` to `sqlstore.Migrate`; `sqlite.Open` calls it; init manifest path updated.
  - Refactor: the refuse-newer test moved from `sqlite/store_test.go` to `sqlstore` (`TestMigrateRefusesADatabaseSavedByANewerVersionOfTheApp`); sqlite keeps its WAL and migrated-on-open test.
  - `task test:init` failed on the undefined `@wip` scenarios; its "healthy" step now runs the new project's acceptance tests with `--godog.tags=~@wip` (agreed with the user).
- [x] 2. The Item repository moves to `sqlstore.NewItemRepository(db)` in portable SQL (`ON CONFLICT` upsert, `$n` placeholders), passing the Item contract and `RunWorkspaceIsolation` on SQLite. `sqlite.NewItemRepository` goes, and its callers use the new one. Sample: `task test:init`.
  - Red: `TestItemRepositoryContractOnSQLite` (sqlstore) failed every contract and isolation check against a stub repository.
  - Green: `sqlstore.ItemRepository` in portable SQL (`ON CONFLICT` upsert, `$n`, which modernc binds).
  - Refactor: `sqlite.ItemRepository` removed; `sqlite.Store.DB()` added, and bootstrap and the acceptance sqlite driver use `sqlstore.NewItemRepository(store.DB())`; UTC test moved to sqlstore, the two-stores test stays in sqlite; init manifest updated.
- [x] 3. The workspace plan repository moves to `sqlstore.NewWorkspacePlanRepository(db)` in the same way, passing its contract suite on SQLite.
  - Red: `TestWorkspacePlanRepositoryContractOnSQLite` (sqlstore) failed every contract check against a stub repository.
  - Green: `sqlstore.WorkspacePlanRepository` in portable SQL (`ON CONFLICT` upsert, `$n`).
  - Refactor: `sqlite.WorkspacePlanRepository` and its contract test removed; the acceptance sqlite driver uses `sqlstore.NewWorkspacePlanRepository(store.DB())`; the reopen test stays in sqlite on the shared repository.

The Postgres adapter:
- [x] 4. `postgres.Open(ctx, url)` connects through pgx and brings the schema up to date with `sqlstore.Migrate`, and `postgrestest.NewDatabase(t)` gives each test a fresh database (skipping without `STARTER_TEST_DATABASE_URL`). Add the compose `postgres` service, `task db:up` and `task db:down`, and Docker in README Prerequisites (`task setup`, Gotchas).
  - Red: `TestOpenBringsTheSchemaUpToDate` failed: `relation "goose_db_version" does not exist`.
  - Green: `postgres.Open` opens through pgx and runs `sqlstore.Migrate`.
  - Refactor: nothing worth changing. `task setup` is unchanged, since it can't install Docker and Postgres tests skip without it; CLAUDE.md has no Gotchas section now.
- [x] 5. `postgres.Open` refuses a database saved by a newer version of the app, and changes nothing.
  - Red: no red: `TestOpenRefusesADatabaseSavedByANewerVersionOfTheApp` passed at once, as `sqlstore.Migrate` already refuses (task 1). Kept as a guard; it fails with the refusal taken out.
  - Green: skipped, no production code demanded.
  - Refactor: the tests' direct connection moved to an `openDirectly` helper.
- [x] 6. The `sqlstore` Item repository passes its contract suite, isolation included, on Postgres.
  - Red: `TestItemRepositoryContractOnPostgres` (sqlstore) panicked on the first `Save`: nil `*sql.DB` from the stub `Store.DB`.
  - Green: `postgres.Store.DB()` returns the store's database; the SQL was already portable.
  - Refactor: nothing worth changing; any shared Postgres set-up waits for 6b, which is the first repeat.
- [x] 6b. The `sqlstore` workspace plan repository passes its contract suite on Postgres.
  - Red: no red: `TestWorkspacePlanRepositoryContractOnPostgres` passed at once, as task 6 added `Store.DB` and the SQL is portable. Kept as a guard; it fails with the upsert broken.
  - Green: skipped, no production code demanded.
  - Refactor: the repeated Postgres set-up moved to an `openPostgres` helper in `migrate_test.go`, which is template code, so the plan test doesn't depend on the sample Item test.
- [x] 7. `postgres.Store.Ready` is nil while the database is reachable, errors while connections are blocked, and is nil again once they're allowed.
  - Red: `TestStoreIsReadyOnlyWhileItsDatabaseCanBeReached` failed with "Ready error = nil while connections are blocked".
  - Green: `Ready` pings with the caller's context; `postgrestest` gained `BlockConnections` and `AllowConnections`.
  - Refactor: `NewDatabase` and `openServerOf` share `openServer`; production `Ready` is one line, so nothing else was worth changing.
- [x] 7a. On Postgres, `sqlstore.Migrate` takes goose's session lock, so servers starting at once on one unmigrated database don't race: several concurrent `postgres.Open` calls on a fresh database all succeed. SQLite stays as it is.
  - Red: `TestOpenSucceedsForSeveralServersStartingAtOnceOnAFreshDatabase` failed with "partial migration error … relation \"items\" already exists" (16 servers: 30 of 30 runs; 8 servers failed only 14 of 20).
  - Green: `Migrate` adds goose's Postgres session locker for `goose.DialectPostgres`.
  - Refactor: the dialect branch moved to `lockOptions`; the lock is polled every second, keeping goose's 5-minute limit (the test went from 75s to 15s).

Settings and wiring:
- [x] 8. `config.StoragePostgres`: `storage = "postgres"` loads.
  - Red: `TestStorageAcceptsPostgres` failed with "The storage setting "postgres" is not recognized."
  - Green: `Config.Validate` accepts `StoragePostgres`.
  - Refactor: nothing worth changing. Left the error's list for 8c.
- [x] 8b. The `database_url` setting (`STARTER_DATABASE_URL`): in `envVars`, `Load`, `Config.Settings`, the `serverOnly` list, and `docs/configuration.md` (which also adds `postgres` to the `storage` row). Add the template-user step to TODO.md.
  - Red: `TestDatabaseURLIsASettingReadFromTheEnvironment` failed with `DatabaseURL = ""` and no `database_url` in `Settings()`.
  - Green: `database_url` is in `envVars`, `Load` and `Config.Settings`, and in the `serverOnly` list.
  - Refactor: no code worth changing. Wrote the `docs/configuration.md` rows and example, and the TODO.md step.
- [x] 8b2. `Config.Settings` replaces a set `database_url` with `(hidden)`, so `starter info` prints `database_url = "(hidden)"  # environment (STARTER_DATABASE_URL)` (wording agreed with the user). An unset one stays `""`, and `Config.DatabaseURL` keeps the real URL. Update the `starter info` example if it shows one.
  - Red: `TestSettingsHidesTheDatabaseURLOnlyWhenItIsSet` failed: `Settings()` listed the full URL.
  - Green: `Config.Settings` lists `(hidden)` for a set `database_url`.
  - Refactor: moved the rule into `hiddenIfSet`. The `starter info` example shows an unset URL, which stays right.
- [x] 8c. The unrecognized-storage error lists `postgres` (wording agreed with the user: `Please use "memory", "jsonfile", "sqlite", or "postgres".`).
  - Red: `TestAnUnknownStorageErrorReadsAsASentence` failed: the error listed only three storages.
  - Green: `UnrecognizedStorageError` names `postgres`; the three other tests pinning the sentence follow.
  - Refactor: nothing worth changing. Left the list spelled out in the one sentence, as the log-format error does.
- [x] 9. `bootstrap.NewEmbedded` on `postgres` opens `postgres.Open(cfg.DatabaseURL)`, with its readiness and `Close`.
  - Red: `TestNewEmbedded_OnPostgresIsReadyUntilClosed` failed with "The storage setting "postgres" is not recognized."
  - Green: a `StoragePostgres` case opens the store and sets `a.close` and `a.ready`; `Items` has no repository until 9b.
  - Refactor: nothing worth changing. The two lines shared with the sqlite case are clearer repeated than wrapped.
- [x] 9b. `bootstrap.NewEmbedded` on `postgres` uses the `sqlstore` repositories: items are still there after closing and reopening. Sample.
  - Red: `TestNewEmbedded_OnPostgresKeepsItemsAfterClosingAndReopening` panicked on the first `CreateItem`: nil `ItemRepository` on `postgres`.
  - Green: the `StoragePostgres` case sets `itemRepo` to `sqlstore.NewItemRepository(store.DB())`, in a sample block.
  - Refactor: the SQLite and Postgres reopen tests share `assertKeepsItemsAfterClosingAndReopening`; production unchanged, as the sample block matches the sqlite case's.
- [x] 9c. `bootstrap.NewEmbedded` refuses `postgres` with no `database_url` (wording agreed with the user: `The storage setting is "postgres", but the database_url setting is missing. Please set database_url in the settings file or the STARTER_DATABASE_URL environment variable.`).
  - Red: `TestNewEmbedded_RefusesPostgresWithNoDatabaseURL` failed: `read database version: failed to connect to` a default local socket.
  - Green: the `StoragePostgres` case returns the agreed error when `cfg.DatabaseURL` is empty, before opening anything.
  - Refactor: nothing worth changing. One caller, so the sentence stays beside its check.
- [ ] 9d. `starter info` on `postgres` says where the data is, rather than a folder: `dataLocation` in `internal/adapters/cli/info.go` knows only folders, a server and memory, so it would print `Data folder: ./data` (user-facing: agree the wording first).
- [ ] 10. `desktop.Open` refuses `storage = "postgres"` (user-facing: agree the wording first).

Acceptance and CI:
- [ ] 11. A `postgres` acceptance driver runs every business scenario on Postgres, with a fresh database per scenario, and is left out with a note when the URL is unset. The server driver leaves out `@postgres` then, too. Steps for the three readiness scenarios turn them green.
- [ ] 12. Steps for "Servers sharing a Postgres database see each other's items": a second test server on the same database, items driven over each server's HTTP. Sample: `task test:init`.
- [ ] 13. CI: a Postgres service container in the Go job, setting `STARTER_TEST_DATABASE_URL` for the unit and acceptance tests.

- [ ] 14. Docs and finish:
  - Write a new ADR 8 (Postgres, the shared `sqlstore`, the test database), and update ADR 7 (migrations now in `sqlstore`) and ADR 5 if it names the hosted storage.
  - Update `.claude/rules/storage.md`, `.claude/rules/testing.md` (the `postgres` driver, `@postgres`), `docs/workspaces.md` (the `sqlstore` step instead of the sqlite one), the CLAUDE.md architecture block and Commands table (`task db:up`), and README.
  - Check `task check`, `task test:init`, and `task accept` with a database.
  - Remove the `@wip` tags, delete this file, and mark step 7 done in the plan, with a summary and links. Log the shared-SQL decision under "Plan changes".

## Changes to the plan
- 2026-11-09: split task 6 into 6 (Item) and 6b (workspace plan), one contract test each.
- 2026-11-09: added task 7a, agreed with the user. The task 4 cycle found that `sqlstore.Migrate` takes no lock, so servers starting together on an unmigrated database could race, and §4.17 promises several servers can share one.
- 2026-11-09: split task 8 into 8 (accept `postgres`), 8b (`database_url`) and 8c (the error's wording), one test each.
- 2026-11-09: split 8b into 8b (read the setting) and 8b2 (hide its value), agreed with the user. `starter info` prints every setting in full, and a database URL usually holds a password.
- 2026-11-09: split task 9 into 9 (open, readiness, `Close`: template code), 9b (the repositories: sample) and 9c (refuse a missing `database_url`), one test each.
- 2026-11-09: split 9c into 9c (refuse a missing `database_url`) and 9d (`starter info` on Postgres, its note until now), agreed with the user. Each needs its own test, and 9d's wording is still to agree.
