# Starter template backlog (temporary)

Ideas for things every project built from this template is likely to need.
Work through them one at a time (each usually starts with `/feature`), tick them off, and delete this file when it's empty.

## Do first
- [x] **`/readyz` checks storage without the sample.** The server's only readiness check is `opts.Items.ListItems` (`internal/adapters/httpapi/server.go`), inside a sample block, so after `--remove-sample` `/readyz` always answers "ready", even when storage is broken. Give it a check that doesn't depend on an entity: for example a `Ready func(ctx) error` in `httpapi.Options` that `bootstrap` takes from the storage adapter (the jsonfile store, and later SQL's ping). Cover it with a test that runs without the sample, and run `task test:init`.
  - Done 2026-11-02: `httpapi.Options.Ready` from `bootstrap`'s `App.Ready` (jsonfile writes and removes a temp file in its data folder; memory is always ready), wired in `cmd/server` and `desktop.Open`. SPEC.md §4.10, `features/readiness.feature` on the new `server` driver.
- [x] **Move generic UI helpers out of the sample.** After `--remove-sample`, a template user rebuilding a screen has to recreate these:
  - the TUI's `danger` colour, `errorStyle` and `selectedStyle` (in a sample block in `internal/adapters/tui/styles.go`);
  - the web app's `ErrorMessage` alert (in `web/src/features/items/ItemsPage.tsx`), which belongs in a shared component.

  Keep them in template code, with the sample using them, and run `task test:init`. Watch the linters: with the sample removed, an unexported style nothing uses fails golangci-lint's `unused` check, which may be why they ended up in the sample block. (The HTTP client's `toError` was the other case found; it moved to `httpclient/client.go` in task 18 of the plans and limits feature.)
  - Done 2026-11-02: the TUI's `errorLine` and `selectedRow` helpers (with their styles) in `styles.go`, the web app's `ErrorMessage` in `web/src/components/`, and a template step that keeps `testPlans.define` in use. `task test:init` now also lints the project with the sample removed (`scripts/initproject/features/init.feature`, `docs/template.md`).

- [x] **Convert configuration from drawing from environment variables to a config file**. Use a well-established library for this, help me decide between top candidates given our architecture. On task completion or along the way, update docs to reflect this change. **If there are good reasons not to do this, including if this is non-standard in modern software, please tell me so**.
  - Done (2026-11-02): replacing environment variables would have been non-standard (servers, Docker, Fly.io, CI and secrets all use them), so the settings file is a layer instead: built in < a TOML settings file < env vars < CLI options, read with koanf. The CLI, TUI and desktop app share `~/.config/starter/config.toml` (or `STARTER_CONFIG_FILE`, `--config`); the server reads a file only when `STARTER_CONFIG_FILE` names one. Mistakes stop the app at start, and `starter info` shows where each setting came from. SPEC.md §4.12, [`features/settings-file.feature`](features/settings-file.feature), [docs/configuration.md](docs/configuration.md), [ADR 6](docs/adr/0006-layered-configuration.md).
- [x] **Implement BDD layer for CLI and TUI clients** similar to the existing Playwright tests for the SPA, since these are the only two that aren't driven by the Web front-end. Use appropriate tools like `Netflix/go-expect` and `charmbracelet-vhs` or a alternatives after comparing commonly used existing tools
  - Done (2026-11-02): Godog + `creack/pty` + `charmbracelet/x/vt`, so the journeys are Go and Gherkin like `test/acceptance`, and run the built program in a real terminal. Rejected: go-expect + vt10x (unmaintained since 2022, predates Bubble Tea v2's escape sequences), VHS (for demo GIFs, not tests), Microsoft tui-test (TypeScript, no Gherkin), teatest (in-process, TUI only). Thin wiring journeys in `test/terminal/features/` (SPEC.md §4.13), run with `task e2e:terminal` in `task check` and CI; Unix only.
- [x] **"Make it mine" init script.** e.g. `task init -- --module github.com/you/app --name "Acme"`. It should:
  - rename the Go module, the "Starter" brand, the binaries and the Docker image;
  - optionally remove the sample `Item` domain.
- [x] **Create a "Rename project" script** based on the init script above. Allows for easy rebranding or changing from an initial "working title" later in development. Unlike the init script, it does not get deleted after running. Where logic can be shared between this and the init script and it makes sense to do so, avoid duplicating code.
- [x] In addition to the existing CLI, TUI, and web app, **create a matching PWA-based native Android and iOS app project, as well as a standalone desktop client** based on Electron or a similar technology to avoid duplicating the web UI. Help me think through the right frameworks/libraries for doing this that involves the least amount of duplicated code and extra work, and also help me figure out what software-based mobile emulators I can use on my Linux machine to test these out before loading onto physical devices.
  - Done (2026-10-30): a phone-width layout, an installable web app, a Wails v3 desktop app, and a Wails v3 Android app with its data on the phone. Decisions are in [ADR 4](docs/adr/0004-native-shells.md); emulator setup is in [docs/mobile.md](docs/mobile.md). iOS and the other follow-ups are the items below.
- [ ] **Desktop and mobile remote mode:** let the desktop and Android apps work on cloud data instead of local data: `desktop.Open` (and so `mobile.Open`) chooses `bootstrap.NewRemote` behind the same in-process handler, and leaves `LocalData` unset so the offline notice comes back. Choose the mode with an env var first (on the phone, a setting). Needs the accounts and authentication items (forwarding the user's session). No sync: see [ADR 4](docs/adr/0004-native-shells.md). This is step 10 of the [editions plan](docs/plans/editions-and-accounts.md).
- [ ] **Desktop packaging:** installers for the desktop app (AppImage or .deb, .dmg, .msi), probably with the `wails3` CLI in `tools/go.mod`, plus an app icon from `web/public/icon.svg`. Try the macOS and Windows builds for real; so far only Linux has been run.
  - Plan: [docs/plans/desktop-packaging.md](docs/plans/desktop-packaging.md). Linux AppImage, .deb and .rpm; Windows NSIS setup .exe instead of .msi; macOS .dmg built and tried on a Mac, and built in CI for releases; unsigned for now.
- [ ] **Lock the data folder on Windows.** The jsonfile adapter's lock uses `flock` and only builds on Unix (`internal/adapters/storage/jsonfile/lock_unix.go`); on Windows `lock_other.go` does nothing. So two copies of the desktop app, or the app and an embedded CLI, can open the same `%AppData%\starter` and overwrite each other's changes. Lock it there too (e.g. `LockFileEx` from `golang.org/x/sys/windows`), so a second copy fails fast with `ErrLocked`, as on Linux. Then drop the Windows caveat in [docs/data-locations.md](docs/data-locations.md) and check CLAUDE.md's Gotcha about the lock still holds.
- [ ] **iOS app:** the Wails v3 mobile app on iPhone. It needs a Mac with full Xcode, and ideally a real iPhone. Its custom-scheme handler should pass request methods and bodies through, so it may not need the Android bridge (`docs/mobile.md`), but that's unverified. Add the generated Xcode project beside `mobile/android`, and cover the app ID in `task rename`.
- [ ] **Android release builds and Play Store signing:** a `mobile:android:release` task (release APK or AAB, `production` build tag), signing with a keystore from environment variables (the generated `app/build.gradle` already reads `ANDROID_KEYSTORE_*`), and a reversed-domain app ID instead of `com.example.<slug>` (see `docs/mobile.md`, "The app ID"). Consider building the APK in CI (it needs the NDK).
- [ ] **Android app icon:** the launcher icon is still Wails' default. Generate the `mipmap-*` icons (including round and adaptive ones) from `web/public/icon.svg`, alongside `pnpm --dir web icons`.
- [ ] **Android system bars follow dark mode:** the status and gesture bars always use the light background (`mobile/android/.../res/values/colors.xml`). In the app's dark mode they should go dark with light icons. The web app would tell Android its theme, through the bridge or a Java hook.
- [ ] **Android app on a real phone:** so far only the emulator has run it. Check the keyboard pushes the page up instead of covering the field (not checked on the emulator: its keyboard was stuck in stylus mode), and the layout on a phone with a notch or cut-out.
- [ ] Change the web app so that collapsing the sidebar (including the animation) is still possible even for wide displays. In those cases, rather than completely hiding the sidebar, it should be collapsed down to a narrow vertical strip with icons in place of the full text for each sidebar tab.
- [x] **SQL storage adapter + migrations** (SQLite first). It must implement `ports.ItemRepository` and pass `storagetest.RunItemRepositoryContract`. It lifts the jsonfile adapter's single-instance limit (ADR 0003). This is step 6 of the [editions plan](docs/plans/editions-and-accounts.md); Postgres is step 7.
  - Done 2026-11-03: `storage = "sqlite"` (`internal/adapters/storage/sqlite`, [ADR 7](docs/adr/0007-sqlite-storage.md)): modernc.org/sqlite, goose migrations, WAL with no lock, so the desktop app and the CLI share one `data.db`. SPEC.md §4.14, `features/sqlite-storage.feature`, and every business scenario on the new `sqlite` driver.
- [ ] **Warn when SQLite's data folder looks synced.** A file-sync service (Dropbox, iCloud Drive, Google Drive, Syncthing, …) can corrupt a SQLite database (SPEC.md §4.14, [ADR 7](docs/adr/0007-sqlite-storage.md)). On `storage = "sqlite"`, warn when the data folder looks synced: perhaps by its path, or by the sync service's marker files, and perhaps with step 5's conflict-copy detection in the [editions plan](docs/plans/editions-and-accounts.md). Then update SPEC.md §4.14's "is planned" and docs/data-locations.md.
- [ ] **Authentication behind a port.** An `Authenticator` port with a simple session or OIDC adapter, wired through the REST API, the SPA, and the CLI's `--server` mode. Currently listed in SPEC.md §6 "Out of scope". This is step 8 of the [editions plan](docs/plans/editions-and-accounts.md): OIDC plus a dev-only fake login, with CLI and TUI login in step 9.
- [x] **Phone-width SPA layout.** The sidebar is a fixed `w-60`, so the app breaks on phones. Add a collapsible sidebar or bottom navigation, and check it in all three styles and both themes.
  - [x] Run `task test:init` to confirm `--remove-sample` still works now that the sidebar's content is shared with the phone menu (`SidebarContent` in `web/src/routes/Layout.tsx`). Passed on 2026-10-30.
- [ ] **Universal application user stories that should survive `--remove-sample`** (each a separate /feature and should be designed such that template users can expand on them easily).  The idea is to have to do as little as possible from scratch for features that are common to utilities or consumer-focused software, so that template users can focus development primarily on their respective problem space. Each should be well-documented, and this documentation should be pointed to by a top-level guide for template users.
  - [ ] **Change application settings (e.g. theme, style, other domain-specific settings) from a settings menu in all clients**
  - [ ] **Stub for application documentation/help section.** Could also lay the groundwork for easy tooltips the for template users, that potentially could themselves link to documentation pages?
  - [ ] **Notifications feature**. This might not be relevant for all applications based on this template, but likely common enough to build in, and template users can remove if it's not applicable. This could be in-app notifications as well as boilerplate for email/mobile/etc. notifications.
  - [ ] If possible, **basic record searching, sorting, and filtering functionality** that can be reused for as-yet-unimplemented and domain-specific data models in the application.
  - [ ] **Placeholder functionality for exporting workspace data to a file, importing workspace data from a file**--especially useful for converting from the local only version to the hosted version, or possibly for the data backups task under `## Operability`.
  - [ ] **Form validation helpers**
  - [ ] **Multi-provider auth** for traditional email/password and standard OAuth providers.
  - [ ] **Session management**: secure token refresh cycles, automatic logout on expiration, and multi-device session invalidation?
  - [ ] **User account** flows for common tasks like password changes, account deletion, etc. **ONLY FOR HOSTED EDITIONS**.
  - [ ] **Functional feature-flagging support** to toggle unreleased features or modify runtime behavior dynamically without redeploying.

  - [ ] **Other features along these same lines likely to be required by template users** that we brainstorm.

  

## Operability
- [ ] **Logging, metrics and traces:**
  - structured `slog` logs with a request ID on every request;
  - a `/metrics` endpoint;
  - optional OpenTelemetry tracing.
- [ ] **Automatic periodic and manual data backups?**
- [ ] **Crash & Error Reporting**: pre-integrated boundaries (like Global Exception Handlers or React Error Boundaries) that intercept crashes and prepare them for upstream analytics.
- [ ] **Telemetry & Analytics**: a unified tracking hook or service layer to measure core user events, app launch performance, and page/view transitions.
- [ ] **Security basics:**
  - security headers and a CSP on the server;
  - request size limits and rate limiting on the API;
  - `govulncheck` and `pnpm audit` in CI.
- [ ] **Release automation:** GoReleaser (or similar), a changelog convention, and Renovate or Dependabot.
- [ ] **Pre-release tags for the desktop packages.** The Taskfile's `PACKAGE_VERSION` accepts only `vX.Y.Z`, so a tag like `v0.1.0-rc.1` builds installers numbered `0.0.0` (GoReleaser's CLI downloads get the right version). Each format writes a pre-release its own way: `.deb` and `.rpm` want `0.1.0~rc.1` (so it sorts before `0.1.0`), while Windows' version info takes only numbers. Then mark such a Release as a pre-release (GoReleaser's `release.prerelease: auto`), and update docs/ci-cd.md's "Making a release".
- [ ] **Dockerized Mocking Infrastructure** such as a docker-compose stack to spin up auxiliary local services like mock authentication, testing databases, or local S3 object buckets seamlessly.

## API and data conventions
- [ ] **Pagination, sorting and filtering on list endpoints**, cursor-based. `GET /items` currently returns everything.
- [ ] **Concurrent-edit protection:** a `version` field or ETag on updates, so concurrent edits don't silently overwrite each other.
- [ ] **Seed data:** `task dev:seed` for realistic local data.

## Front-end polish
- [ ] **No flash of the wrong theme or style:** a tiny inline script in `web/index.html` that applies both before React loads.
- [ ] **Shared feedback patterns:** toasts, an error boundary and a proper 404 page.
- [ ] **Installed-app polish** (follow-ups from SPEC.md §4.5):
  - precache the fonts and icons too (the service worker's default `globPatterns` cover only JS, CSS and HTML, so offline the app falls back to system fonts until they're cached);
  - a dark-theme `theme-color` meta, so the browser bar matches the dark theme;
  - offline, the Items page shows "Loading…" under the offline notice, because TanStack Query pauses its requests. Show nothing, or a quieter placeholder.
- [ ] **Automated accessibility checks:** axe in the Playwright e2e tests, across every style × theme.

## Maybe (depends on the products built from it)
- [ ] Translations (i18n)
- [ ] Background jobs or domain events (outbox pattern) - (this very likely will be relevant for my first intended use of the template)
- [ ] Audit log (who changed what, and when)

## Other
- [ ] **Generate a new TODO.md file that lives in the template permanently.** It's purpose is to list out tasks for a new project that a developer should start with when using this template, including things like running the `init` script, defining the vision and the definition of MVP in an appropriate place for Claude to reference, and swapping out the mockups for new styles, among others. It should link to referenced files where appropriate. Once these initial tasks are complete, the TODO.md file can stick around as a place to keep new ideas for the application that have yet to be implemented. Add a note to CLAUDE.md that feature and bugfix ideas live here. Its setup tasks should point to the README's Prerequisites (tools, plus system packages such as the desktop app's GTK/WebKit libraries and the Linux and Windows packaging tools) rather than repeat them.
  - Steps the new file must give **the developer using the template** (the template deliberately leaves these unset; they are not tasks for us):
    - Replace the placeholder `Categories=Utility;` in the Linux desktop file (the `.desktop` file in `build/desktop/linux/`) with the app's own (from the [freedesktop menu spec](https://specifications.freedesktop.org/menu-spec/latest/category-registry.html)). The template needs some category, because the AppImage tool refuses a desktop file without one.
    - Choose where the server is hosted: follow the README's "One-time Fly.io setup" (or replace `deploy.yml` for another host), then set the repository variable `DEPLOY_ENABLED` to `true`. Until then, deploys are skipped.
    - Define the product's plans in `internal/adapters/plans/plans.jsonc`: its limits, and the plans hosted workspaces will be on. The template ships `local` (no limits) and the examples `free` and `pro`. The local edition's plan is `local` unless a build picks another with `-ldflags "-X <module>/internal/bootstrap.LocalPlan=<plan>"`; `STARTER_PLANS_FILE` replaces the file at run time.
- [x] **Init asks for the app's description and publisher.** The desktop packages' descriptions, publisher, maintainer and copyright (the template-user steps above) are filled in by hand for now. `task init` could take them as flags or prompts and write them into `build/desktop/`, so a new project starts with none of the template's placeholders.
  - Done 2026-11-03: `--description`, `--publisher` and `--email` are required flags; init writes them into `nfpm.yaml`, `app.desktop` and `info.json` (see [docs/template.md](docs/template.md)).
- [ ] Notate in CLAUDE.md that all bugfixes and new features should be functional for all clients: web, desktop, mobile, cli, and tui.
- [ ] Generate an architecture document, with diagrams, that explains the project directory structure and the purposes of each directory to give users of this template a foothold for starting new development. Also create a similar document with diagrams for each of the supported front-ends in place in their corresponding directory under `docs`, as well as the core Go library itself (just a stub for the template to be filled in later), and the REST API (also a stub).
- [ ] **Hosted and local editions:** let web, desktop and mobile be commercial products, with data kept separate per account and features gated by paid plans, while the local edition (the CLI and TUI by default, the desktop and mobile apps in local mode, and a self-hosted server with sign-in off) stays a set of single-user tools on local data (plain text or SQLite, in a folder the user chooses, which can be a cloud-synced one), on the same core and data structures. The CLI and TUI can also log in to the hosted edition.
  - Plan: [docs/plans/editions-and-accounts.md](docs/plans/editions-and-accounts.md). Decisions: [ADR 5](docs/adr/0005-hosted-and-local-editions.md). Data belongs to a workspace; plans in config grant features and limits; billing is a driving adapter; OIDC authentication; synced folders are used from one device at a time, with no merging; no sync between local and cloud data.
- [x] **Fix `task accept -- -run 'TestFeatures/<driver>'`.** CLAUDE.md documents it for running one acceptance driver, but on 2026-10-30 `-run 'TestFeatures/memory$'` still ran all five drivers.
  - Done 2026-11-03: `godog.BindCommandLineFlags` adds Go's `flag.CommandLine` to pflag, so `pflag.Parse()` in `TestMain` marked it parsed with no arguments and `m.Run` never read `-run` or `-v`. The four Godog suites now bind their flags with `godog.BindFlags("godog.", flag.CommandLine, &opts)` and let `go test` parse everything.
- [ ] Add release notes and update commands to write them automatically
- [ ] Template-wide sweep to ensure developers using this template will be able to add domain-specific features while touching as few places as possible (Core, API, Docs, Storage adapters, CLI, TUI, and Web I think covers it but I could be missing some). Also make sure `/feature` commands prompt for decisions about the web, TUI, and CLI interfaces, since these cannot all be identical. The intention is that shared code is maximized with reason so that new feature development requires as little work as is feasible.
- [ ] Project cleanup: go through the project and look for potential opportunities to pair down unnecessary files and leave a clean-as-possible slate for users of the template.
