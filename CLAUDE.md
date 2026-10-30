# CLAUDE.md

Guidance for Claude Code in this repository. Product intent lives in `SPEC.md`; decisions live in `docs/adr/`.

## What this is

A walking skeleton for a multi-front-end application. It uses a hexagonal (ports and adapters) Go core with four driving adapters (REST API, React SPA via the API, Cobra CLI, Bubble Tea TUI) and swappable storage.
<!-- sample:start -->

The sample domain is `Item`; replace it when starting a real product.
<!-- sample:end -->

## Workflow: BDD outside, TDD inside

1. `/feature <desc>` updates SPEC.md, writes `@wip` Gherkin in `features/`, and shows it failing. No production code.
2. `/red <behaviour>` writes one failing unit test.
3. `/green` writes the minimum code to pass it.
4. `/refactor` tidies the code on green, with no behaviour change.
5. Repeat 2–4 until `task accept` passes. Then remove the `@wip` tags.

`/feature` saves its inner-loop task list to `docs/wip/<feature-slug>.md`, which is tracked by git, so it survives a cleared context. It also holds the feature's links, the decisions behind the plan, and a log of plan changes.
- When resuming a feature, read its `docs/wip/` file first.
- After each `/red`, `/green` or `/refactor`, tick off or annotate the item it advanced. If the plan changes, revise the list and note why.
- `/green` leaves a **For refactor:** note under its item, listing the refactoring it spotted. `/refactor` starts from that note and replaces it with what it did.
- Delete the file in the same change that removes the `@wip` tags.

Never skip a step or merge steps unless I ask. Never write production code without a failing test that demands it.

Never stage, commit, push or pull unless I explicitly ask. That includes `git add`, `git mv` and `git rm`, which stage changes. Move or delete files with plain `mv` and `rm`. I review and commit the changes myself.

## Commands

| Task | Command |
|---|---|
| Unit tests (Go race + Vitest) | `task test` (`task test:go`, `task test:web`) |
| One Go test | `go test ./internal/core/app/ -run TestName` |
| Acceptance (Godog) | `task accept` · only WIP: `task accept -- --godog.tags=@wip` · one driver: `task accept -- -run 'TestFeatures/memory'` |
| Browser e2e (playwright-bdd) | `task e2e` (builds the SPA and starts the Go server itself) |
| Lint / format | `task lint` · `task fmt` |
| Regenerate API code | `task gen` (after editing `api/openapi.yaml`) |
| Everything CI runs | `task check` |
| Rename the project | `task rename -- --name "New Name" [--module …] [--dry-run]` (see `docs/rename.md`) |
| Run locally | `task dev` (API on :8080, restarted by Air on Go changes + Vite on :5173 with HMR, opened in your browser) · `task run:cli -- --help` · `task run:tui` |

## Architecture

```
cmd/server, cmd/starter          entrypoints, which only call bootstrap + an adapter
internal/core/domain             entities, value rules, domain errors (stdlib only)
internal/core/ports              interfaces: driving (…Service) + driven (…Repository, Clock, IDGenerator)
internal/core/app                use cases implementing driving ports (orchestration only)
internal/adapters/httpapi        REST server (generated strict interface from OpenAPI) + SPA serving
internal/adapters/httpclient     driving ports over HTTP, used by CLI/TUI in --server mode
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
- UIs call the core **only** through driving ports (the `…Service` interfaces in `ports`). Never put business logic in handlers, Cobra commands, Bubble Tea models or React components.
- A new storage backend implements the `…Repository` ports and **must** run each one's contract suite from `storagetest` in its tests. Wire it in `internal/bootstrap`.
- Domain errors (`domain.ErrNotFound`, `*domain.ValidationError`) are the shared language. `httpapi/problem.go` maps them to HTTP, and `httpclient` maps them back.
- Time and IDs come from `ports.Clock` / `ports.IDGenerator`. Use `internal/testsupport` fakes in tests.

### API contract

`api/openapi.yaml` is the source of truth. Edit it, then run `task gen`. **Never hand-edit `*.gen.go` or `*.gen.ts`.** CI fails if generated code drifts.

## TUI conventions

- Vim navigation is required (see SPEC.md §5). Every list or selectable view binds `j`/`k` alongside `down`/`up`. Bind them only while browsing, never while a text input has focus, so users can still type those letters.
- Cover the `j`/`k` bindings in each new view's tests.
<!-- sample:start -->
- The Items screen (`internal/adapters/tui/items.go`) and its tests (`items_test.go`) show both.
<!-- sample:end -->

## Web styles

- A **style** (SPEC.md §4.3) is a named look for the SPA: layout, fonts, shape and density. It is separate from the dark/light **theme**, and every style must work in both.
- Styles and the shared colour palette live in `web/src/styles/`. The active style is the `data-style` attribute on `<html>`.
- Colours: use only the semantic utilities from `palette.css` (`bg-panel`, `text-muted`, `border-line`, `bg-sidebar`, …). Tailwind's default colours are switched off, so `slate-500` and friends won't compile to anything.
- Style tokens: `font-sans`, `rounded-panel`, `rounded-control`, `py-row` and the `type-label` utility (headings, buttons, table headers) change per style. Reach for the `crisp:`, `terminal:` and `soft:` variants only where markup really must differ.
- To add a style: add it to `styles` in `web/src/features/style/style.ts`, create `web/src/styles/<id>.css` setting the `--style-*` variables under `:root[data-style="<id>"]` (self-host fonts with `@fontsource`), and import it in `web/src/index.css`.
- `web/design-inspiration/` holds the reference palette (`.hex`) and mockups for the current styles. A new project replaces it with its own. Colours come from the palette, not the mockups.

## Testing conventions

- Go: standard `testing`, table tests where useful, black-box `_test` packages by default. No assertion library.
- Godog steps (`test/acceptance/steps.go`) are thin: they call the driving ports and assert. No business logic.
- Gherkin in `features/` uses domain language and is UI-agnostic. `web/e2e/features/` holds only a few journeys that prove the wiring.
- Web: Vitest + Testing Library, querying by role or label. Stub `fetch` rather than mocking modules.
- E2E scenarios run in parallel against one server, so use unique names and don't assert on whole lists.

## Gotchas

- The jsonfile adapter locks its data directory. Running the server and an embedded CLI against the same `./data` fails fast with "data directory is locked". Use `--server http://localhost:8080` for the CLI instead.
- The server embeds the SPA only when built with `-tags embedweb` after `pnpm build` (`task build` does both). Otherwise `/` shows a hint page.
- Dev-only Go tools (oapi-codegen for `task gen`, Air for `task dev`) live in `tools/go.mod` so their dependencies stay out of the app's `go.mod`. Add one with `cd tools && go get -tool <pkg>@<version>`, and run it with `go tool -modfile=tools/go.mod <name>`.
- Node ≥ 24 is required (see `.nvmrc`). TypeScript is pinned to 6.0.x until typescript-eslint supports 7.
