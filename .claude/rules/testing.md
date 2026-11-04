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
- Acceptance drivers: scenarios run against memory, jsonfile, sqlite, http, desktop and mobile drivers. `@desktop` and `@mobile` scenarios run only on their own driver; `@workspaces` ones only on memory, jsonfile and sqlite; `@plans` ones only on memory, jsonfile, sqlite and http. `@server` ones (running the server, e.g. readiness) run only on the server driver, which runs nothing else.
- Terminal journeys run the built program as a user would, each scenario on its own computer: a temp `HOME`, no `STARTER_*` variables from your environment, and a server on memory storage when it needs one. Steps wait for what the screen shows (`term.WaitFor`) rather than pausing for a fixed time, and press keys by name (`term.Press`).
- Web: Vitest + Testing Library, querying by role or label. Stub `fetch` rather than mocking modules.
- E2E scenarios run in parallel against one server, so use unique names and don't assert on whole lists.
- Time and IDs come from `ports.Clock` / `ports.IDGenerator`. Use `internal/testsupport` fakes in tests (also `testsupport.FixedWorkspace`, `NoWorkspace`, `FixedEntitlements`).
