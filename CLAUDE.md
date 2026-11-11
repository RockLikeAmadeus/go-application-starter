# CLAUDE.md

Guidance for Claude Code in this repository. Product intent lives in `SPEC.md`; decisions live in `docs/adr/`. Topic-specific rules live in `.claude/rules/` and load automatically when you work on matching files (index at the end).

## What this is

A walking skeleton for a multi-front-end application. It uses a hexagonal (ports and adapters) Go core with four driving adapters (REST API, React SPA via the API, Cobra CLI, Bubble Tea TUI) and swappable storage. A Wails desktop app runs the same SPA and REST API in-process, in a native window, and a Wails Android app does the same on a phone (ADR 4).
<!-- sample:start -->

The sample domain is `Item`; replace it when starting a real product.

<!-- sample:end -->

The app comes in two **editions** (ADR 5), built step by step from `docs/plans/editions-and-accounts.md`. Most of it is still planned:
- **Local edition:** one user, one workspace, no sign-in. This is the CLI and TUI (embedded), the desktop and Android apps, and a self-hosted single-user server. All of these work today.
- **Hosted edition:** many users, each signing in to their own workspace, with plans gating features. The web app on a server with `sign_in` on works today with the dev sign-in (ADR 9, `docs/sign-in.md`). Real providers (OIDC) come in 8b. The CLI and TUI after `starter login`, and the desktop and Android apps in remote mode, are planned.

Each copy uses one edition, and nothing syncs between them. Outside `internal/bootstrap` and `desktop.Open`, never assume local storage or a single user.

## Workflow: BDD outside, TDD inside

1. `/feature <desc>` checks that we're on an up-to-date, clean `main`, creates the branch `feature/<feature-slug>`, updates SPEC.md, writes `@wip` Gherkin in `features/`, shows it failing, and commits and pushes the spec, scenarios, and task list. No production code. The task list ends with a docs-and-finish task that removes the `@wip` tags, deletes the `docs/wip/` file, and updates the plan.
2. `/red <behaviour>` writes one failing unit test.
3. `/green` writes the minimum code to pass it.
4. `/refactor` tidies the code on green, with no behaviour change.
5. Repeat 2–4 until `task accept` passes. The last task removes the `@wip` tags.

`/iterate [behaviour]` runs one whole red → green → refactor cycle for the next open task (or the one given), stopping only to ask me something, and then once at the end. When the cycle goes cleanly, it commits and pushes it, and after the feature's last task it opens the branch's pull request and gives me its link. `/iterate-all [feature]` runs an `/iterate` cycle for every open task of a `docs/wip/` file, each in a fresh subagent, one commit per task, stopping only when the tasks are done or a cycle can't finish cleanly. When the tasks are done, it opens the pull request. Questions pause it, and it carries on with my answer.

`TODO.md` is the backlog, and `/prune-backlog` moves its finished items to `docs/backlog-history.md`. A feature's task list lives in `docs/wip/<feature-slug>.md`; a TODO item needing more than one `/feature` has a plan in `docs/plans/<item-slug>.md`. Both are tracked and temporary. **When starting or resuming work, run `task wip:next` first, then read the parts of the plan the task needs.** How to read and update them is in `.claude/rules/work-tracking.md`.

Keep tool output small, as every cycle starts from a cleared context. Search narrowly: `grep -l` or `-m`, a specific path, and `head` on `find`. When output was too long and was saved to a file, `grep` or `tail` that file instead of reading all of it. Read a large file by section, not whole.

When a change leaves something that a developer starting a new project from this template must set up or fill in themselves (a placeholder name, publisher, or description, an account, a key), add a step for them in the same change. Until the template's permanent starter TODO list exists, these steps go under the TODO.md item that creates it ("Steps the new file must give the developer using the template"). They are tasks for template users, not for us.

<!-- template:start -->
This is still the template and has no users, so don't build backwards compatibility. Don't migrate data saved before a change, don't add defaults for old records, and don't hold a release for old data. Breaking changes are fine. Products built from the template handle their own compatibility. (`task init` removes this paragraph.)

The sample `Item` feature is removed by most projects, so work spent on it is mostly lost. When planning a feature, put its machinery in template code, then give it one thin sample use that runs it end to end: the worked example and the acceptance proof, not a place to cover every edge case. Template code must never depend on sample code, which includes a template helper living in a sample file. Run `task test:init` in any cycle that adds code next to sample code or moves it across a sample boundary, not only at the end of a feature. (`task init` removes this paragraph too.)
<!-- template:end -->

Never skip a step or merge steps unless I ask. Never write production code without a failing test that demands it.

Never stage, commit, push, or pull unless I explicitly ask. That includes `git add`, `git mv`, and `git rm`, which stage changes. Running `/iterate` is such a request, for its own cycle only (see `.claude/commands/iterate.md`, whose `allowed-tools` lets it run `git add`, `git commit`, and `git push` without a prompt). Running `/iterate-all` is the same request for each of its cycles. Both open the feature's pull request after its last task (`.claude/rules/work-tracking.md`), and `/feature` creates the feature's branch; nothing else opens a pull request or changes branch unless I ask. Running `/feature` is the same request for its own spec, scenarios, and task list. `/red`, `/green`, and `/refactor` leave everything unstaged. Move or delete files with plain `mv` and `rm`. I review and commit the changes myself.

## Writing and asking

Use the Oxford comma in everything written in this repository: docs, rules, commands, Gherkin, code comments, and messages. Docs for developers (README, SPEC.md, the `docs/` guides, and the ADRs) favour clarity over brevity, as set out in `.claude/rules/documentation.md`.

When you ask me a question, don't assume I know the tools, terms, or patterns behind it, including design, devops, and industry practices that might seem standard. I'm new to much of this stack, and I often pick your recommendation only because I can't judge the other options. So explain what is being decided and why it matters. For each option, say what it would mean in practice, and its trade-offs: what it costs, and what it makes harder later. Link to a tool's documentation where that helps, and draw a small diagram (plain text in the terminal) when the choice is about how parts fit together.

## Commands

| Task | Command |
|---|---|
| Unit tests (Go race + Vitest) | `task test` (`task test:go`, `task test:web`) |
| One Go test | `go test ./internal/core/app/ -run TestName` |
| Go benchmarks | `task bench` · narrow it: `task bench -- -bench ConflictCopies` (never run by `task check` or CI) |
| Test database (Postgres) | `task db:up` (Postgres on :5432 for `STARTER_TEST_DATABASE_URL`; safe to re-run) · `task db:down`. Without the URL, Postgres tests skip |
| Acceptance (Godog) | `task accept` · only WIP: `task accept -- --godog.tags=@wip` · one driver: `task accept -- -run 'TestFeatures/memory'` |
| Browser e2e (playwright-bdd) | `task e2e` (builds the SPA and starts the Go server itself) |
| Terminal journeys (Godog) | `task e2e:terminal` (builds `starter` and `server`, runs them in a pseudo-terminal; Unix only) · only WIP: `task e2e:terminal -- --godog.tags=@wip` |
| Lint / format | `task lint` · `task fmt` |
| Regenerate API code | `task gen` (after editing `api/openapi.yaml`) |
| Everything CI runs | `task check` |
| Rename the project | `task rename -- --name "New Name" [--module …] [--dry-run]` (see `docs/rename.md`) |
| A cycle's view of the feature in progress | `task wip:next` · one feature: `task wip:next -- docs/wip/<slug>.md` |
| Run locally | `task dev` (API on :8080, restarted by Air on Go changes + Vite on :5173 with HMR, opened in your browser) · `task run:cli -- --help` · `task run:tui` · `task run:desktop` |
| Build | `task build` (`bin/server` with the SPA embedded, `bin/starter`) |

Desktop builds and packaging, the Android app, and the app icons have their own commands, listed in `desktop.md`, `desktop-packaging.md`, `mobile.md` and `web.md` under `.claude/rules/`.

## Architecture

```
cmd/server, cmd/starter          entrypoints, which only call bootstrap + an adapter
cmd/desktop                      Wails v3 window around desktop.Open's handler (build tag `desktop`)
build/desktop/{linux,windows}    the desktop app's packaging (.deb/.rpm/AppImage; NSIS setup program)
cmd/mobile                       Wails v3 Android app around mobile.Open, plus the Bridge service (build tag `android`)
internal/core/domain             entities, value rules, domain errors (stdlib only)
internal/core/ports              interfaces: driving (…Service) + driven (…Repository, Clock, IDGenerator)
internal/core/app                use cases implementing driving ports (orchestration only)
internal/adapters/httpapi        REST server (generated strict interface from OpenAPI), SPA serving, and sign-in under `/auth/` (session cookie, dev sign-in in `devsignin` builds)
internal/adapters/httpclient     driving ports over HTTP, used by CLI/TUI in --server mode (later, desktop and mobile remote mode too)
internal/adapters/cli, tui       Cobra / Bubble Tea + Lip Gloss
internal/adapters/desktop        desktop app wiring with no Wails import: data folder, storage, in-process handler
internal/adapters/mobile         Android wiring on top of desktop's, no Wails import: the page handler and Serve, the API bridge
internal/adapters/storage/*      memory, jsonfile, sqlite, postgres; sqlstore (the SQL both share: goose migrations, repositories); storagetest contract suite
internal/adapters/workspace      ports.WorkspaceResolver: Local (the local edition's one workspace), FromContext (set with WithID)
internal/adapters/plans          ports.EntitlementsProvider and ports.PlanCatalog from the plans file (plans.jsonc, built in; STARTER_PLANS_FILE replaces it); plans.StoredPlanOf reads stored plans
internal/bootstrap               composition root: the only place that picks adapters
web/                             Vite + React + TS + Tailwind + TanStack Query; components every screen can use (e.g. `ErrorMessage`) go in web/src/components
mobile/android                   Wails' generated Gradle project for the Android app (see docs/mobile.md)
features/ + test/acceptance      Godog business scenarios, run against memory, jsonfile, sqlite, postgres, http, desktop, mobile, and server drivers (tags pick drivers; see .claude/rules/testing.md)
web/e2e                          playwright-bdd browser journeys
test/terminal                    Godog terminal journeys: the built CLI and TUI (and a server for --server) in a pseudo-terminal; term is the terminal helper (Unix only)
test/packaging                   package check: builds the Linux packages and the Windows setup program and inspects them (build tag `packaging`, `task test:package`)
```

### Rules (enforced by depguard in `.golangci.yml`)

- `core/domain` imports only the stdlib. `core/app` imports `domain` and `ports`, never adapters, `bootstrap`, `config`, `net/http`, `os`, or `database/sql`.
- Business rules live in `domain`. Use cases in `app` only orchestrate.
- UIs call the core **only** through driving ports (the `…Service` interfaces in `ports`). Never put business logic in handlers, Cobra commands, Bubble Tea models, or React components.
- A new storage backend implements the `…Repository` ports, runs each one's `storagetest` contract suite, and is wired (with its readiness check) in `internal/bootstrap` (`.claude/rules/storage.md`).
- Data belongs to a workspace, passed **explicitly** to every repository call; plans set a workspace's limits, checked before saving and never on delete; domain errors are the shared language across HTTP. Details are in `.claude/rules/workspaces-and-plans.md`. **Adding an entity: follow the checklist in `docs/workspaces.md`. Gating a feature by plan: `docs/features.md`. Connecting a payment provider: `docs/billing.md`.**
- Time and IDs come from `ports.Clock` / `ports.IDGenerator`. Use `internal/testsupport` fakes in tests.
- The core's tests are unit tests, with no database, network, or files: they may import only the standard library, the core, `storage/memory`, and `internal/testsupport`, and never `database/sql`, `net/...`, or `os`.

### API contract

`api/openapi.yaml` is the source of truth. Edit it, then run `task gen`. **Never hand-edit `*.gen.go` or `*.gen.ts`.** CI fails if generated code drifts.

### Testing in brief

Go: standard `testing`, black-box `_test` packages, no assertion library. Godog steps are thin. Gherkin in `features/` uses domain language and is UI-agnostic. Full conventions are in `.claude/rules/testing.md`.

## Topic rules (`.claude/rules/`)

Each loads when you read a file matching its paths. Read one directly when a task touches its topic before you've opened such a file.

| File | Covers |
|---|---|
| `naming.md` | naming types after what their values are, the "is a" test, naming files, renaming |
| `documentation.md` | developer docs: clarity over brevity, explaining tools and terms, examples, diagrams |
| `user-facing-text.md` | agreeing wording with me first, formal style, the Oxford comma |
| `work-tracking.md` | keeping `docs/wip/` task lists and `docs/plans/` plans |
| `workspaces-and-plans.md` | workspace scoping, entitlements, the local plan and plans file, changing a plan, domain errors |
| `storage.md` | new backends, the jsonfile lock and shared data folder, SQLite (ADR 7) |
| `configuration.md` | layered settings (ADR 6), adding a setting |
| `server.md` | SPA embedding, new server routes, `/api/v1/connection`, the `/auth/` routes |
| `tui.md` | vim keys, shared TUI styles |
| `web.md` | Node/TS versions, icons, installable app, offline behaviour, Android bridge client |
| `web-styles.md` | styles, palette, focus, phone width, entry transitions |
| `web-e2e.md` | Playwright projects, `@phone`, WebKit deps |
| `testing.md` | Go, Godog, terminal, and web test conventions; acceptance driver tags |
| `desktop.md` | Wails desktop build, GTK/WebKitGTK deps, local data |
| `desktop-packaging.md` | Linux packages and the Windows setup program |
| `mobile.md` | Android build, lint, the API bridge |
| `dependencies.md` | dev tools in `tools/go.mod`, Wails version lockstep, Node/TS pins |
