# 8. Postgres storage: one SQL layer shared with SQLite

- Status: accepted
- Date: 2026-11-10

## Context

The hosted edition (ADR 5) runs several copies of the server, on different machines, against one store of data. jsonfile (ADR 3) is one program per folder, and SQLite (ADR 7) is several programs on one disk, so neither can be shared between machines. The editions plan (step 7) adds Postgres, the hosted edition's database.

SQLite and Postgres both speak SQL. Writing each entity's migrations and repository twice, once per engine, would double the work of every new entity and let the two drift apart.

## Decision

- **One SQL layer for both engines: `internal/adapters/storage/sqlstore`.** It holds the goose migrations (`sqlstore/migrations/`), the migration runner (`sqlstore.Migrate(ctx, db, dialect)`) and the `…Repository` implementations, all on a `*sql.DB`. The `sqlite` and `postgres` packages keep only what really differs: opening the database, goose's dialect, and readiness. A new entity is written once (`docs/workspaces.md`).
- **Portable SQL.** Queries use `$1` placeholders, which both drivers bind, and `INSERT … ON CONFLICT (…) DO UPDATE SET …` for upserts. Times stay fixed-width UTC `TEXT`, and IDs come from `ports.IDGenerator`, so nothing relies on an engine's types or auto-increment. Migrations stay within what SQLite can do (rebuild a table rather than `ALTER COLUMN`); if that ever bites, a goose Go migration can branch on the dialect for that one change.
- **Driver: `github.com/jackc/pgx/v5/stdlib`**, through `database/sql`. It is pure Go, so `CGO_ENABLED=0` still builds. Connection counts keep pgx's defaults.
- **Opening** (`postgres.Open(ctx, url)`) connects and runs `sqlstore.Migrate`, as `sqlite.Open` does, so it refuses a database saved by a newer version of the app and changes nothing. On Postgres, `Migrate` takes goose's session lock (polled every second, for up to 5 minutes), so servers starting at once on an unmigrated database don't race. SQLite needs no lock: one disk, one migration at a time.
- **Readiness** (`Store.Ready`, behind `/readyz`) pings the database with the request's context. The pool reconnects by itself, but a ping can land on a connection that died during an outage, so `Ready` pings again on `driver.ErrBadConn`, at most 3 attempts: the first probe after an outage answers ready.
- **The settings:** `storage = "postgres"` with `database_url` (`STARTER_DATABASE_URL`), a server-oriented setting that `starter info` shows as `(hidden)`, since a URL usually holds a password. Postgres without a `database_url` is refused at start. The server, the CLI, and the TUI honour it (through `bootstrap.NewEmbedded`), so an operator can run `starter admin set-plan` on the hosted data. `desktop.Open`, and so the Android app, refuses it: those apps reach hosted data through a server (remote mode, step 10), never the database.
- **The test database: a URL plus compose.** Tests read `STARTER_TEST_DATABASE_URL`. `task db:up` starts Postgres from the `postgres` service in `docker-compose.yml` (data in memory, gone with the container), and CI runs a Postgres service container in the Go job. `postgrestest.NewDatabase(t)` gives each test, and each acceptance scenario, a fresh database with a random name, dropped on cleanup. With the URL unset, Postgres tests skip with a message, and the acceptance `postgres` driver and the server driver's `@postgres` scenarios are left out with a note. With it set but unreachable, they fail. CI always sets it.
- **An outage in tests** is `postgrestest.BlockConnections` (`ALLOW_CONNECTIONS false` plus `pg_terminate_backend`) and `AllowConnections`, so readiness can fail and recover without stopping the shared container.

## Consequences

- Several server machines can share one database, so a hosted deployment can scale out. It needs a Postgres database and `STARTER_DATABASE_URL` as a secret (Fly Postgres or any other).
- A new entity needs one migration and one repository, in `sqlstore`, which runs its contract suite on SQLite and, with a test database, on Postgres. SQL that only one engine accepts fails one of the two runs.
- Running every test locally needs Docker (`task db:up`). Without it, the Postgres tests skip and CI still runs them.
- With sign-in off, a server on Postgres is a self-hosted local edition with one workspace, as on any other storage. With sign-in on (ADR 9), it keeps accounts and sessions in Postgres too, so every server copy on the database knows every session.
