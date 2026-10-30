# CLAUDE.md

Guidance for Claude Code in this repository. Product intent lives in `SPEC.md`; decisions live in `docs/adr/`.

## What this is

A walking skeleton for a multi-front-end application. It uses a hexagonal (ports and adapters) Go core with four driving adapters (REST API, React SPA via the API, Cobra CLI, Bubble Tea TUI) and swappable storage. The sample domain is `Item`; replace it when starting a real product.

## Workflow: BDD outside, TDD inside

1. `/feature <desc>` updates SPEC.md, writes `@wip` Gherkin in `features/`, and shows it failing. No production code.
2. `/red <behaviour>` writes one failing unit test.
3. `/green` writes the minimum code to pass it.
4. `/refactor` tidies the code on green, with no behaviour change.
5. Repeat 2–4 until `task accept` passes. Then remove the `@wip` tags.

Never skip a step or merge steps unless I ask. Never write production code without a failing test that demands it.

## Commands

| Task | Command |
|---|---|
| Unit tests (Go race + Vitest) | `task test` (`task test:go`, `task test:web`) |
| One Go test | `go test ./internal/core/app/ -run TestCreateItem` |
| Acceptance (Godog) | `task accept` · only WIP: `task accept -- --godog.tags=@wip` · one driver: `task accept -- -run 'TestFeatures/memory'` |
| Browser e2e (playwright-bdd) | `task e2e` (builds the SPA and starts the Go server itself) |
| Lint / format | `task lint` · `task fmt` |
| Regenerate API code | `task gen` (after editing `api/openapi.yaml`) |
| Everything CI runs | `task check` |
| Run locally | `task dev` (API on :8080, restarted by Air on Go changes + Vite on :5173 with HMR) · `task run:cli -- item list` · `task run:tui` |

## Architecture

```
cmd/server, cmd/starter          entrypoints, which only call bootstrap + an adapter
internal/core/domain             entities, value rules, domain errors (stdlib only)
internal/core/ports              interfaces: driving (ItemService) + driven (ItemRepository, Clock, IDGenerator)
internal/core/app                use cases implementing driving ports (orchestration only)
internal/adapters/httpapi        REST server (generated strict interface from OpenAPI) + SPA serving
internal/adapters/httpclient     ItemService over HTTP, used by CLI/TUI in --server mode
internal/adapters/cli, tui       Cobra / Bubble Tea + Lip Gloss
internal/adapters/storage/*      memory, jsonfile, + storagetest contract suite
internal/bootstrap               composition root: the only place that picks adapters
web/                             Vite + React + TS + Tailwind + TanStack Query
features/ + test/acceptance      Godog business scenarios, run against memory, jsonfile and http drivers
web/e2e                          playwright-bdd browser journeys
```

### Rules (enforced by depguard in `.golangci.yml`)

- `core/domain` imports only the stdlib. `core/app` imports `domain` and `ports`, never adapters, `bootstrap`, `config`, `net/http`, `os` or `database/sql`.
- Business rules live in `domain`. Use cases in `app` only orchestrate.
- UIs call the core **only** through driving ports (`ports.ItemService`). Never put business logic in handlers, Cobra commands, Bubble Tea models or React components.
- A new storage backend implements `ports.ItemRepository` and **must** call `storagetest.RunItemRepositoryContract` from its tests. Wire it in `internal/bootstrap`.
- Domain errors (`domain.ErrNotFound`, `*domain.ValidationError`) are the shared language. `httpapi/problem.go` maps them to HTTP, and `httpclient` maps them back.
- Time and IDs come from `ports.Clock` / `ports.IDGenerator`. Use `internal/testsupport` fakes in tests.

### API contract

`api/openapi.yaml` is the source of truth. Edit it, then run `task gen`. **Never hand-edit `*.gen.go` or `*.gen.ts`.** CI fails if generated code drifts.

## TUI conventions

- Vim navigation is required (see SPEC.md §5). Every list or selectable view binds `j`/`k` alongside `down`/`up`, as `internal/adapters/tui/model.go` does. Bind them only while browsing, never while a text input has focus, so users can still type those letters.
- Cover the `j`/`k` bindings in `model_test.go` for each new view.

## Web styles

- A **style** (SPEC.md §4.3) is a named look for the SPA: layout, fonts, shape and density. It is separate from the dark/light **theme**, and every style must work in both.
- Styles and the shared colour palette live in `web/src/styles/`. The active style is the `data-style` attribute on `<html>`.
- `web/design-inspiration/` holds the reference palette (`.hex`) and mockups for the starter's styles. A new project replaces it with its own. Colours come from the palette, not the mockups.

## Testing conventions

- Go: standard `testing`, table tests where useful, black-box `_test` packages by default. No assertion library.
- Godog steps (`test/acceptance/steps.go`) are thin: they call `ports.ItemService` and assert. No business logic.
- Gherkin in `features/` uses domain language and is UI-agnostic. `web/e2e/features/` holds only a few journeys that prove the wiring.
- Web: Vitest + Testing Library, querying by role or label. Stub `fetch` rather than mocking modules.
- E2E scenarios run in parallel against one server, so use unique names and don't assert on whole lists.

## Gotchas

- The jsonfile adapter locks its data directory. Running the server and an embedded CLI against the same `./data` fails fast with "data directory is locked". Use `--server http://localhost:8080` for the CLI instead.
- The server embeds the SPA only when built with `-tags embedweb` after `pnpm build` (`task build` does both). Otherwise `/` shows a hint page.
- Dev-only Go tools (oapi-codegen for `task gen`, Air for `task dev`) live in `tools/go.mod` so their dependencies stay out of the app's `go.mod`. Add one with `cd tools && go get -tool <pkg>@<version>`, and run it with `go tool -modfile=tools/go.mod <name>`.
- Node ≥ 24 is required (see `.nvmrc`). TypeScript is pinned to 6.0.x until typescript-eslint supports 7.
