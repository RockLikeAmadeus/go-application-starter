# CLAUDE.md

Guidance for Claude Code in this repository. Product intent lives in `SPEC.md`; decisions live in `docs/adr/`.

## What this is

A walking skeleton for a multi-front-end application. It uses a hexagonal (ports and adapters) Go core with four driving adapters (REST API, React SPA via the API, Cobra CLI, Bubble Tea TUI) and swappable storage. A Wails desktop app runs the same SPA and REST API in-process, in a native window, and a Wails Android app does the same on a phone (ADR 4).
<!-- sample:start -->

The sample domain is `Item`; replace it when starting a real product.

The app comes in two **editions** (ADR 5), built step by step from `docs/plans/editions-and-accounts.md`. Most of it is still planned:
- **Local edition:** one user, one workspace, no sign-in. This is the CLI and TUI (embedded), the desktop and Android apps, and a self-hosted single-user server. All of these work today.
- **Hosted edition:** many users, each signing in to their own workspace, with plans gating features. This is the web app on a hosted server, the CLI and TUI after `starter login`, and the desktop and Android apps in remote mode. All planned.

Each copy uses one edition, and nothing syncs between them. Outside `internal/bootstrap` and `desktop.Open`, never assume local storage or a single user.

<!-- sample:end -->

## Workflow: BDD outside, TDD inside

1. `/feature <desc>` updates SPEC.md, writes `@wip` Gherkin in `features/`, and shows it failing. No production code.
2. `/red <behaviour>` writes one failing unit test.
3. `/green` writes the minimum code to pass it.
4. `/refactor` tidies the code on green, with no behaviour change.
5. Repeat 2–4 until `task accept` passes. Then remove the `@wip` tags.

`/iterate [behaviour]` runs one whole red → green → refactor cycle for the next open task (or the one given), stopping only to ask me something, and then once at the end. When the cycle goes cleanly, it commits and pushes it.

`/feature` saves its inner-loop task list to `docs/wip/<feature-slug>.md`, which is tracked by git, so it survives a cleared context. It also holds the feature's links, the decisions behind the plan, and a log of plan changes.
- When resuming a feature, read its `docs/wip/` file first.
- After each `/red`, `/green` or `/refactor`, tick off or annotate the item it advanced. If the plan changes, revise the list and note why.
- `/green` leaves a **For refactor:** note under its item, listing the refactoring it spotted. `/refactor` starts from that note and replaces it with what it did.
- Delete the file in the same change that removes the `@wip` tags.

When a change leaves something that a developer starting a new project from this template must set up or fill in themselves (a placeholder name, publisher or description, an account, a key), add a step for them in the same change. Until the template's permanent starter TODO list exists, these steps go under the TODO.md item that creates it ("Steps the new file must give the developer using the template"). They are tasks for template users, not for us.

<!-- template:start -->
This is still the template and has no users, so don't build backwards compatibility. Don't migrate data saved before a change, don't add defaults for old records, and don't hold a release for old data. Breaking changes are fine. Products built from the template handle their own compatibility. (`task init` removes this paragraph.)

The sample `Item` feature is removed by most projects, so work spent on it is mostly lost. When planning a feature, put its machinery in template code, then give it one thin sample use that runs it end to end: the worked example and the acceptance proof, not a place to cover every edge case. Template code must never depend on sample code, which includes a template helper living in a sample file. Run `task test:init` in any cycle that adds code next to sample code or moves it across a sample boundary, not only at the end of a feature. (`task init` removes this paragraph too.)
<!-- template:end -->

`TODO.md` is the backlog. When an item needs more than one `/feature`, its plan lives in `docs/plans/<item-slug>.md`, which is also tracked and temporary:
- The plan holds the item's decisions, an ordered list of steps (each one a `/feature`) with their status, and a log of plan changes. The TODO item links to it.
- When resuming the item, read the plan first, then the current step's `docs/wip/` file, if one exists.
- When a step's feature finishes, mark the step done in the plan, with a one-line summary and links, and log any plan changes.
- Delete the plan in the same change that ticks the TODO item. Move any follow-ups still open into `TODO.md` first.
- A plan made in plan mode lives outside the repo, so copy it into `docs/plans/` before starting the first step.

Never skip a step or merge steps unless I ask. Never write production code without a failing test that demands it.

Never stage, commit, push or pull unless I explicitly ask. That includes `git add`, `git mv` and `git rm`, which stage changes. Running `/iterate` is such a request, for its own cycle only (see `.claude/commands/iterate.md`, whose `allowed-tools` lets it run `git add`, `git commit` and `git push` without a prompt). `/red`, `/green`, `/refactor` and `/feature` leave everything unstaged. Move or delete files with plain `mv` and `rm`. I review and commit the changes myself.

## Commands

| Task | Command |
|---|---|
| Unit tests (Go race + Vitest) | `task test` (`task test:go`, `task test:web`) |
| One Go test | `go test ./internal/core/app/ -run TestName` |
| Acceptance (Godog) | `task accept` · only WIP: `task accept -- --godog.tags=@wip` · one driver: `task accept -- -run 'TestFeatures/memory'` |
| Browser e2e (playwright-bdd) | `task e2e` (builds the SPA and starts the Go server itself) |
| Terminal journeys (Godog) | `task e2e:terminal` (builds `starter` and `server`, runs them in a pseudo-terminal; Unix only) · only WIP: `task e2e:terminal -- --godog.tags=@wip` |
| Lint / format | `task lint` · `task fmt` |
| Regenerate API code | `task gen` (after editing `api/openapi.yaml`) |
| Everything CI runs | `task check` |
| Regenerate the app icons | `pnpm --dir web icons` (after editing `web/public/icon.svg`; commit the PNGs and `build/desktop/windows/icon.ico`, which it also writes). The desktop app uses `pwa-512x512.png` too |
| Rename the project | `task rename -- --name "New Name" [--module …] [--dry-run]` (see `docs/rename.md`) |
| Run locally | `task dev` (API on :8080, restarted by Air on Go changes + Vite on :5173 with HMR, opened in your browser) · `task run:cli -- --help` · `task run:tui` · `task run:desktop` |
| Build | `task build` (`bin/server` with the SPA embedded, `bin/starter`) · `task build:desktop` (`bin/starter-desktop`; needs the desktop build deps, see Gotchas) · `task build:desktop:windows` (`bin/starter-desktop.exe`, cross-compiled; needs no desktop build deps) · `task mobile:android:build` (`bin/starter.apk`; needs the Android tools, see `docs/mobile.md`) |
| Package the desktop app | `task package:desktop:linux` (`.deb`, `.rpm` and AppImage in `bin/`; one format: `…:deb`, `…:rpm`, `…:appimage`) · `task package:desktop:windows` (`bin/starter-desktop-<version>-setup.exe`; needs `makensis`) · check them: `task test:package` (slow; not part of `task check`) |
| Run the Android app | `task mobile:android:run` (builds, installs on the connected phone or emulator, starting the first AVD if none is, and opens it) |

## Architecture

```
cmd/server, cmd/starter          entrypoints, which only call bootstrap + an adapter
cmd/desktop                      Wails v3 window around desktop.Open's handler (build tag `desktop`)
build/desktop/linux              the desktop app's Linux packaging: nfpm.yaml (.deb, .rpm), app.desktop, install script
build/desktop/windows            the desktop app's Windows packaging: icon.ico, app.manifest, info.json (version info), installer.nsi + wails_tools.nsh (NSIS)
cmd/mobile                       Wails v3 Android app around mobile.Open, plus the Bridge service (build tag `android`)
internal/core/domain             entities, value rules, domain errors (stdlib only)
internal/core/ports              interfaces: driving (…Service) + driven (…Repository, Clock, IDGenerator)
internal/core/app                use cases implementing driving ports (orchestration only)
internal/adapters/httpapi        REST server (generated strict interface from OpenAPI) + SPA serving
internal/adapters/httpclient     driving ports over HTTP, used by CLI/TUI in --server mode (later, desktop and mobile remote mode too)
internal/adapters/cli, tui       Cobra / Bubble Tea + Lip Gloss
internal/adapters/desktop        desktop app wiring with no Wails import: data folder, storage, in-process handler
internal/adapters/mobile         Android wiring on top of desktop's, no Wails import: the page handler and Serve, the API bridge
internal/adapters/storage/*      memory, jsonfile, sqlite (goose migrations in sqlite/migrations), + storagetest contract suite
internal/adapters/workspace      ports.WorkspaceResolver: Local (the local edition's one workspace), FromContext (set with WithID)
internal/adapters/plans          ports.Entitlements from the plans file (plans.jsonc, built in; STARTER_PLANS_FILE replaces it)
internal/bootstrap               composition root: the only place that picks adapters
web/                             Vite + React + TS + Tailwind + TanStack Query; components every screen can use (e.g. `ErrorMessage`) go in web/src/components
mobile/android                   Wails' generated Gradle project for the Android app (see docs/mobile.md)
features/ + test/acceptance      Godog business scenarios, run against memory, jsonfile, sqlite, http, desktop and mobile drivers (`@desktop` and `@mobile` scenarios run only on their own driver; `@workspaces` ones only on memory, jsonfile and sqlite; `@plans` ones only on memory, jsonfile, sqlite and http). `@server` ones (running the server, e.g. readiness) run only on the server driver, which runs nothing else
web/e2e                          playwright-bdd browser journeys
test/terminal                    Godog terminal journeys: the built CLI and TUI (and a server for --server) in a pseudo-terminal; term is the terminal helper (Unix only)
test/packaging                   package check: builds the Linux packages and the Windows setup program and inspects them (build tag `packaging`, `task test:package`)
```

### Rules (enforced by depguard in `.golangci.yml`)

- `core/domain` imports only the stdlib. `core/app` imports `domain` and `ports`, never adapters, `bootstrap`, `config`, `net/http`, `os` or `database/sql`.
- Business rules live in `domain`. Use cases in `app` only orchestrate.
- UIs call the core **only** through driving ports (the `…Service` interfaces in `ports`). Never put business logic in handlers, Cobra commands, Bubble Tea models or React components.
- A new storage backend implements the `…Repository` ports and **must** run each one's contract suite from `storagetest` in its tests. Wire it in `internal/bootstrap`, including its readiness check (`App.Ready`, which answers the server's `/readyz`), as jsonfile's and sqlite's `Store.Ready` are. Readiness comes from storage, never from an entity's data.
- Data belongs to a workspace (ADR 5). Use cases get the current one from `ports.WorkspaceResolver` and pass it **explicitly** to every repository call. A workspace-scoped entity's repository port embeds `ports.WorkspaceRepository[T]`, treats another workspace's record as missing (`ErrNotFound`), and every adapter's contract tests run `storagetest.RunWorkspaceIsolation`, plus an isolation case for each extra query. With no workspace, a use case returns `domain.ErrNoWorkspace` before touching storage. `bootstrap` wires `workspace.Local`; tests use `testsupport.FixedWorkspace` or `NoWorkspace`. **Adding an entity: follow the checklist in `docs/workspaces.md`.**
- Plans set a workspace's limits (ADR 5). A use case that a limit gates asks `ports.Entitlements` for the workspace's plan, counts, and checks with the domain's rules before saving: `Entitlements.CheckCreate` before adding one, `CheckChange` before changing one. Deleting is never limited, so a workspace over a read-only limit can get back within it. If the plan or the count can't be looked up, refuse the change: a failed lookup must never mean "no limits". Front ends read the plan through `ports.EntitlementsService` (`GET /api/v1/entitlements`). `bootstrap` wires the local plan; tests use `testsupport.FixedEntitlements`.
- Domain errors (`domain.ErrNotFound`, `*domain.ValidationError`, `*domain.LimitError`, `domain.ErrNoWorkspace`) are the shared language. `httpapi/problem.go` maps them to HTTP, and `httpclient` maps them back. (`ErrNoWorkspace` isn't mapped yet: nothing over HTTP can lack a workspace until sign-in exists.)
- Time and IDs come from `ports.Clock` / `ports.IDGenerator`. Use `internal/testsupport` fakes in tests.

### API contract

`api/openapi.yaml` is the source of truth. Edit it, then run `task gen`. **Never hand-edit `*.gen.go` or `*.gen.ts`.** CI fails if generated code drifts.

## TUI conventions

- Vim navigation is required (see SPEC.md §5). Every list or selectable view binds `j`/`k` alongside `down`/`up`. Bind them only while browsing, never while a text input has focus, so users can still type those letters.
- Cover the `j`/`k` bindings in each new view's tests.
- Shared colours, styles and the helpers every screen uses (`errorLine`, `selectedRow`) live in `internal/adapters/tui/styles.go`, outside sample blocks. Give a style a small helper with its own test (`styles_test.go`), so it stays in use, and passes golangci-lint's `unused` check, when the sample is removed.
<!-- sample:start -->
- The Items screen (`internal/adapters/tui/items.go`) and its tests (`items_test.go`) show both.
<!-- sample:end -->

## Web styles

- A **style** (SPEC.md §4.3) is a named look for the SPA: layout, fonts, shape and density. It is separate from the dark/light **theme**, and every style must work in both.
- Styles and the shared colour palette live in `web/src/styles/`. The active style is the `data-style` attribute on `<html>`.
- Every page must work at phone width (SPEC.md §4.4), with no sideways scrolling in any style or theme. Below `md`, the sidebar becomes a Menu drawer (`MobileNav` in `web/src/routes/Layout.tsx`). In a flex row, give inputs `min-w-0` so they can shrink: WebKit's default input width is wider than Chromium's.
- Tag phone browser scenarios `@phone`. They run on an emulated Pixel (Chromium), an emulated iPhone (WebKit, the local stand-in for Safari) and a phone-sized Firefox window. Untagged scenarios run on desktop Chromium.
- Colours: use only the semantic utilities from `palette.css` (`bg-panel`, `text-muted`, `border-line`, `bg-sidebar`, …). Tailwind's default colours are switched off, so `slate-500` and friends won't compile to anything.
- Keyboard focus: a base `:focus-visible` rule in `web/src/index.css` gives every control and link the palette's `focus` ring. Don't add per-control focus classes, and never remove the outline (`outline-none`) without putting an equally visible focus style in its place. Use real `<button>`, `<a>` and form elements, so they can take focus and get the ring.
- Style tokens: `font-sans`, `rounded-panel`, `rounded-control`, `py-row` and the `type-label` utility (headings, buttons, table headers) change per style. Reach for the `crisp:`, `terminal:` and `soft:` variants only where markup really must differ.
- Entry transitions (`starting:`): write the start value plainly, e.g. `starting:open:[translate:-100%_0]`, not with Tailwind's `translate-*`, `scale-*` or `rotate-*` utilities. Those read registered `--tw-*` variables, and WebKitGTK (the desktop app's engine) drops such a declaration inside `@starting-style`, so nothing animates in. Playwright's WebKit doesn't show this.
- To add a style: add it to `styles` in `web/src/features/style/style.ts`, create `web/src/styles/<id>.css` setting the `--style-*` variables under `:root[data-style="<id>"]` (self-host fonts with `@fontsource`), and import it in `web/src/index.css`.
- `web/design-inspiration/` holds the reference palette (`.hex`) and mockups for the current styles. A new project replaces it with its own. Colours come from the palette, not the mockups.

## Testing conventions

- Go: standard `testing`, table tests where useful, black-box `_test` packages by default. No assertion library.
- Godog steps (`test/acceptance/steps.go`) are thin: they call the driving ports and assert. No business logic.
- Gherkin in `features/` uses domain language and is UI-agnostic. `web/e2e/features/` (browser) and `test/terminal/features/` (CLI and TUI) hold only a few journeys that prove the wiring.
- Terminal journeys run the built program as a user would, each scenario on its own computer: a temp `HOME`, no `STARTER_*` variables from your environment, and a server on memory storage when it needs one. Steps wait for what the screen shows (`term.WaitFor`) rather than pausing for a fixed time, and press keys by name (`term.Press`).
- Web: Vitest + Testing Library, querying by role or label. Stub `fetch` rather than mocking modules.
- E2E scenarios run in parallel against one server, so use unique names and don't assert on whole lists.

## Gotchas

- The local edition's plan is `bootstrap.LocalPlan`: `local`, unless a build sets another with `-ldflags "-X <module>/internal/bootstrap.LocalPlan=<plan>"`. The plans come from `internal/adapters/plans/plans.jsonc`, built into the binary, unless `STARTER_PLANS_FILE` names another file. An unknown local plan or a bad plans file fails at start. The file is JSON with full-line `//` comments only: a comment after a value is a parse error.
- The jsonfile adapter locks its data directory (on Unix only: the lock is `flock`); sqlite doesn't (see the next point). The embedded CLI and TUI share the desktop app's data folder by default (`config.LocalDataDir`: `STARTER_DATA_DIR`, else the settings file's `data_dir`, else `~/.config/starter` on Linux; `starter info` prints it), so while the desktop app is open they fail fast with "data directory is locked", as does a second copy of the app. The server keeps `./data` (in the repository under `task dev`), so it doesn't clash with them; to reach its data, use `--server http://localhost:8080`. `task run:cli` and `task run:tui` use your real data folder, as `task run:desktop` does.
- SQLite storage (`storage = "sqlite"`, ADR 7) is `<data dir>/data.db`, opened with `modernc.org/sqlite`, a pure-Go driver: keep it that way, because the Windows tasks set `CGO_ENABLED=0`. It takes no lock: WAL and a busy timeout let several programs share it, and each store has one connection (`SetMaxOpenConns(1)`), so a query must close its rows before the next one starts. `sqlite.Open` applies the goose migrations embedded from `internal/adapters/storage/sqlite/migrations/`, and refuses a database whose version is newer than the newest one. A schema change is a new numbered migration, never an edit to an old one. `00001_baseline.sql` is template code and must stay: goose won't start with no migrations. `Store.Ready` pings and checks `data.db` is still the file it opened, because on Unix an open handle keeps a removed file alive. `desktop.Open` takes `jsonfile` or `sqlite` from the settings (never memory); the Android app stays on jsonfile. On SQLite, when `syncedfolder.Detect` says the data folder looks synced, `bootstrap` puts a warning in `App.Warnings`, which every front end shows (`GET /api/v1/warnings`, stderr, the TUI's warning line, the server's log); `ignore_synced_folder` turns it off.
- Settings come in layers (ADR 6, `docs/configuration.md`): built in < a TOML settings file < `STARTER_*` env vars (and `PORT`) < CLI options. The CLI, TUI and desktop app read `config.toml` in the app's settings folder (`~/.config/starter/config.toml` on Linux; `config.LocalConfigFile`), or the file `STARTER_CONFIG_FILE` or `--config` names. The server reads a file only when `STARTER_CONFIG_FILE` names one, so your own settings file never changes `task dev`. An unknown key, bad TOML or a missing named file fails at start; a missing default file is fine. `internal/config` never reads `os`: pass `config.Env` (`os.Getenv`, `os.UserConfigDir` in `cmd/*`, fakes in tests). A new setting goes in `envVars`, `Load` and `Config.Settings` (`internal/config/config.go`), and in `docs/configuration.md`.
- The server embeds the SPA only when built with `-tags embedweb` after `pnpm build` (`task build` does both). Otherwise `/` shows a hint page.
- Dev-only Go tools (oapi-codegen for `task gen`, Air for `task dev`) live in `tools/go.mod` so their dependencies stay out of the app's `go.mod`. Add one with `cd tools && go get -tool <pkg>@<version>`, and run it with `go tool -modfile=tools/go.mod <name>`.
- The desktop app (`cmd/desktop`, build tag `desktop`) uses Wails v3. On Linux it needs GTK 4 and WebKitGTK 6.0: `sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev` in a real terminal. Without them, the build fails with "Package 'gtk4' … not found". Wails' older GTK 3 stack (`-tags gtk3`) is dropped in Wails v3.1, so don't rely on it. Plain `go build ./...` and `task check` skip `cmd/desktop`, so they don't need these packages.
- Linux packages (`build/desktop/linux/`, `task package:desktop:linux`): the desktop file and icons are installed under the app ID (`com.example.starter`, set in `cmd/desktop`), so the desktop can match the window to them, but their sources keep generic names, because `task rename` rewrites names inside files and moves only `cmd/<slug>`. Package versions come from `VERSION`: a release tag (`v1.2.3`) loses its `v`, anything else becomes `0.0.0`. The AppImage bundles every WebKitGTK on the build machine, so build it where only WebKitGTK 6.0 is installed (not `libwebkit2gtk-4.1-dev`). `task test:package` runs the package check (`test/packaging`, build tag `packaging`); it takes a few minutes, mostly the AppImage.
- Windows setup program (`build/desktop/windows/`, `task package:desktop:windows`): built on Linux. Both tasks set `CGO_ENABLED=0`: the program needs no C compiler, and it also keeps `go tool wails3` from linking GTK, so neither needs the desktop build deps (check with `PKG_CONFIG_LIBDIR=/nonexistent`, which hides them). Only `go build` gets `GOOS=windows`: set for the whole task, `go tool wails3` would build a Windows `wails3.exe`. Its icon, manifest and version info go in through `cmd/desktop/rsrc_windows_amd64.syso`, which the task writes and deletes (gitignored); the WebView2 bootstrapper is written beside `installer.nsi` for `makensis` the same way. `info.json` is a template (`@VERSION@`) and the one source of the product name, publisher and copyright for both the program and the setup program. It installs for the current user (`WAILS_INSTALL_SCOPE=user`: `%LOCALAPPDATA%\Programs`, no admin prompt) and keeps the items in `%AppData%\starter`; its uninstaller deletes only WebView2's `%AppData%\starter-desktop.exe`. `wails_tools.nsh` is Wails' generated file: replace it, don't edit it, on a Wails upgrade. The package check reads what the setup program does from `bin/windows/makensis.log` (`makensis -V4`), because the script inside it is compressed. It's unsigned, so Windows SmartScreen warns before it runs ("More info" → "Run anyway").
- In the local edition, the desktop app keeps its data on the computer (`httpapi.Options.LocalData`; remote mode, which is planned, will leave it unset), so `GET /api/v1/connection` answers `{"required": false}`. The SPA then hides the offline notice and keeps TanStack Query online (`useConnectionRequired`, `useLocalDataOnline`). New offline behaviour in the SPA must respect that answer. The desktop handler answers `/sw.js` with 404, so the service worker never runs there.
- The Android app (`cmd/mobile`, build tag `android`) uses Wails v3 mobile, which is experimental. Building it needs a JDK, the Android SDK and the NDK (`docs/mobile.md`); `go build ./...` and `task check` skip it. Lint it with `GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=<NDK clang> golangci-lint run --build-tags android ./cmd/mobile/...`.
- On Android, the SPA's `/api` requests go over the Wails bridge (`apiFetch` → `bridgeFetch` → `main.Bridge.Serve` → `mobile.Serve`), because Wails' Android page path turns every request into a body-less `GET` and drops non-200 statuses. `mobile.AssetHandler` answers `/api/*` on the page path with 404, so a request that misses the bridge fails instead of silently succeeding. Don't rename `Bridge` or move it out of `cmd/mobile`'s package `main`: `client.ts` calls it by that name. `apiFetch` checks for `window.wails.invokeAsync` before importing `@wailsio/runtime`, because the import replaces `window.wails`.
- Keep `@wailsio/runtime` in `web/package.json` (pinned exactly), the Wails module in `go.mod` and wails3 in `tools/go.mod` on the same version. Dependabot ignores all three (`.github/dependabot.yml`), so upgrade them together by hand. `mobile/android` is generated by wails3 with a few hand edits; `docs/mobile.md` says how to regenerate it on upgrade. Its Java package must stay `com.wails.app` (the Go library's JNI names depend on it); only `applicationId` is ours.
- Playwright's WebKit (the `iphone` e2e project) needs system libraries that need sudo. `task setup` installs them with `--with-deps`. If it was skipped, run `sudo pnpm --dir web exec playwright install-deps webkit` in a real terminal: Claude Code's `!` can't answer sudo's password prompt.
- The SPA is an installable web app (`VitePWA` in `web/vite.config.ts`, SPEC.md §4.5). Its service worker precaches only the built app and must never cache `/api`: data always comes from the server. Offline, page loads get the cached app. That includes any path the Go server serves outside `/api`, so add new server routes (like `/healthz` and `/readyz`) to `navigateFallbackDenylist`. A deploy reaches an open or installed app the next time it's opened. The service worker is off under `task dev`; test it with `task e2e` or a built server (see `docs/installable-app.md`).
- Node ≥ 24 is required (see `.nvmrc`). TypeScript is pinned to 6.0.x until typescript-eslint supports 7. Dependabot ignores TypeScript 7 until then; drop that rule in `.github/dependabot.yml` when unpinning.
