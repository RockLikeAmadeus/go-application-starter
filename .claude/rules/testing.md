---
paths:
  - "**/*_test.go"
  - "features/**"
  - "test/**"
  - "internal/testsupport/**"
  - "web/src/**/*.test.ts"
  - "web/src/**/*.test.tsx"
---

# Testing conventions

- Go: standard `testing`, table tests where useful, black-box `_test` packages by default. No assertion library.
- Godog steps (`test/acceptance/steps.go`) are thin: they call the driving ports and assert. No business logic.
- Gherkin in `features/` uses domain language and is UI-agnostic. `web/e2e/features/` (browser) and `test/terminal/features/` (CLI and TUI) hold only a few journeys that prove the wiring.
- Acceptance drivers: scenarios run against memory, jsonfile, sqlite, postgres, http, oidc, desktop, and mobile drivers. `@desktop` and `@mobile` scenarios run only on their own driver; `@workspaces` ones only on memory, jsonfile, sqlite, postgres, and http (as the hosted edition, signing in to each workspace's account); `@plans` ones only on memory, jsonfile, sqlite, postgres, and http; `@sign-in` ones only on memory, sqlite, postgres, and http, each started as the hosted edition (`startHosted`), as plain text keeps no accounts. The oidc driver (the hosted edition over HTTP, with the real `oidc` adapter against `oidctest.Provider`, following redirects with a cookie jar) runs only `@sign-in` and `@sign-in-provider`, and `@sign-in-provider` runs on nothing else. `@server` ones (running the server, e.g. readiness) run only on the server driver, which runs nothing else.
- Postgres tests need `STARTER_TEST_DATABASE_URL` (`task db:up` starts a server for it; CI always sets it). Use `postgrestest.NewDatabase(t)`, which gives each test, and each acceptance scenario, a fresh database and skips the test without the URL; with it set but unreachable, tests fail. Without the URL, the `postgres` driver is left out with a note, and the server driver leaves out `@postgres` scenarios (server scenarios that need a database). `postgrestest.BlockConnections` and `AllowConnections` make an outage without stopping the server.
- Terminal journeys run the built program as a user would, each scenario on its own computer: a temp `HOME`, no `STARTER_*` variables from your environment, and a server on memory storage when it needs one. Steps wait for what the screen shows (`term.WaitFor`) rather than pausing for a fixed time, and press keys by name (`term.Press`).
- Web: Vitest + Testing Library, querying by role or label. Stub `fetch` rather than mocking modules.
- E2E scenarios run in parallel against one server, so use unique names and don't assert on whole lists.
- Time and IDs come from `ports.Clock` / `ports.IDGenerator`. Use `internal/testsupport` fakes in tests (also `testsupport.FixedWorkspace`, `NoWorkspace`, `FixedEntitlements`).
- Tests in `internal/core/` are unit tests: no database, network, or files. Lint (depguard's `core-tests` list) lets them import only the standard library, the core, `storage/memory`, and `internal/testsupport`, and refuses `database/sql`, `net/...`, and `os`. Test an adapter in its own package.
