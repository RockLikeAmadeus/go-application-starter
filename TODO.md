# Starter template backlog (temporary)

Ideas for things every project built from this template is likely to need.
Work through them one at a time (each usually starts with `/feature`), tick them off, and delete this file when it's empty.
`/prune-backlog` moves ticked items, once none of their sub-tasks are open, into [docs/backlog-history.md](docs/backlog-history.md).

**Before starting work on any feature below, discuss whether it's actually a valuable feature and what are the tradeoffs of including it or not including it. Many of these are Claude's ideas and probably very sensible, but some of them I've thought of myself and I want to be wary of template bloat.**

## Do first
- [ ] Start a **"What This Template Includes" guide** which describes the features we've included so far, broken down into appropriate sections and subsections, with links to further information about specific features where available/appropriate. Keep this document updated as we implement new features for the template. This will be very valuable so template users know both _what_ is already included and therefore doesn't need to be done from scratch, as well as _how_ to work with those features, add to them, change them, etc. The guide should also discuss at a high level the technologies involved and a very high-level description of the application architecture.
- [ ] **Consider how distributed workloads would function for projects using this template**. The hosted edition serving multiple users and relying on computationally intensive background tasks, for instance, is a future we haven't considered, if it even can be planned for ahead of deciding on a specific problem space. Even if the problem domain is relatively simplistic, a single server serving the web application to a theoretical large number of concurrent users isn't scalable, nor is a single database instance. Need to determine how much of this problem is worth considering at the template-level, if any.
- [ ] **Desktop and mobile remote mode:** let the desktop and Android apps work on cloud data instead of local data: `desktop.Open` (and so `mobile.Open`) chooses `bootstrap.NewRemote` behind the same in-process handler, and leaves `IsDataLocal` unset so the offline notice comes back. Choose the mode with an env var first (on the phone, a setting). Needs the accounts and authentication items (forwarding the user's session). No sync: see [ADR 4](docs/adr/0004-native-shells.md). This is step 10 of the [editions plan](docs/plans/editions-and-accounts.md).
- [ ] **Desktop packaging:** installers for the desktop app (AppImage or .deb, .dmg, .msi), probably with the `wails3` CLI in `tools/go.mod`, plus an app icon from `web/public/icon.svg`. Try the macOS and Windows builds for real; so far only Linux has been run.
  - Plan: [docs/plans/desktop-packaging.md](docs/plans/desktop-packaging.md). Linux AppImage, .deb and .rpm; Windows NSIS setup .exe instead of .msi; macOS .dmg built and tried on a Mac, and built in CI for releases; unsigned for now.
  - Linux and Windows done 2026-11-04: packages built and tried by hand, and attached to the GitHub Release by `release.yml` (tried with a throwaway `v0.0.1` tag). Left: the macOS `.dmg` (step 4, on the Mac) and the rest of step 5.
- [ ] **iOS app:** the Wails v3 mobile app on iPhone. It needs a Mac with full Xcode, and ideally a real iPhone. Its custom-scheme handler should pass request methods and bodies through, so it may not need the Android bridge (`docs/mobile.md`), but that's unverified. Add the generated Xcode project beside `mobile/android`, and cover the app ID in `task rename`.
- [ ] **Android release builds and Play Store signing:** a `mobile:android:release` task (release APK or AAB, `production` build tag), signing with a keystore from environment variables (the generated `app/build.gradle` already reads `ANDROID_KEYSTORE_*`), and a reversed-domain app ID instead of `com.example.<slug>` (see `docs/mobile.md`, "The app ID"). Consider building the APK in CI (it needs the NDK).
- [ ] **Android app icon:** the launcher icon is still Wails' default. Generate the `mipmap-*` icons (including round and adaptive ones) from `web/public/icon.svg`, alongside `pnpm --dir web icons`.
- [ ] **Android system bars follow dark mode:** the status and gesture bars always use the light background (`mobile/android/.../res/values/colors.xml`). In the app's dark mode they should go dark with light icons. The web app would tell Android its theme, through the bridge or a Java hook.
- [ ] **Android app on a real phone:** so far only the emulator has run it. Check the keyboard pushes the page up instead of covering the field (not checked on the emulator: its keyboard was stuck in stylus mode), and the layout on a phone with a notch or cut-out.
- [ ] **Desktop app command-line flags:** `starter-desktop` reads no arguments, so `--data-dir` (which `bin/starter` takes) is silently ignored and the app opens its default folder. Give it the same `--data-dir`, `--storage`, and `--config` flags as the CLI, layered as in ADR 6, and have it refuse a flag it doesn't know instead of ignoring it (agree the message's wording first).
- [ ] **Expand `~` in folder settings:** `data_dir` and the other paths are used as given, so `~/Dropbox/starter` from the settings file, `STARTER_DATA_DIR`, or a quoted `--data-dir` becomes a folder literally named `~`. Expand a leading `~` (and `~\` on Windows) to the user's home folder wherever a setting names a path, in `internal/config`, and say so in `docs/configuration.md`.
- [ ] Change the web app so that collapsing the sidebar (including the animation) is still possible even for wide displays. In those cases, rather than completely hiding the sidebar, it should be collapsed down to a narrow vertical strip with icons in place of the full text for each sidebar tab.
- [ ] **Authentication behind a port.** An `Authenticator` port with a simple session or OIDC adapter, wired through the REST API, the SPA, and the CLI's `--server` mode. Currently listed in SPEC.md §6 "Out of scope". This is step 8 of the [editions plan](docs/plans/editions-and-accounts.md): OIDC plus a dev-only fake login, with CLI and TUI login in step 9.
  - Part 8a done 2026-11-11: the sign-in seam (`ports.SignInService`), accounts and personal workspaces, sessions in an `HttpOnly` cookie and in storage (memory, SQLite, and Postgres), the dev sign-in in `devsignin` builds only, the web app's sign-in screen and Sign out button, and stored plans in the hosted edition. [ADR 9](docs/adr/0009-sign-in-and-sessions.md), [docs/sign-in.md](docs/sign-in.md), SPEC.md §4.18. Part 8b, the OIDC adapter, is next.
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
  - [ ] **A sample/placeholder home page** for the hosted edition SPA that provides feature info, pricing details, etc., along with a login button, unless this is more appropriate as a separate hosted site).
  - [ ] **Report a bug or suggest a feature** feature
  - [ ] **Other features along these same lines likely to be required by template users** that we brainstorm.

  

## Operability
- [ ] **Logging, metrics, and traces:**
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
- [ ] **Browser smoke runs against a deployment start no sign-in server.** With `E2E_BASE_URL` set, `web/playwright.config.ts` starts no server, but the `sign-in` project still runs `web/e2e/features/sign-in.feature` against its own base URL (`localhost`, the port after `PORT`), so it fails there. Skip that project when `E2E_BASE_URL` is set, or point it at a deployment with sign-in on.
- [x] **Mark the session cookie `Secure` behind an HTTPS proxy.** The server marks the cookie `Secure` only when its own connection is TLS (`r.TLS != nil` in `internal/adapters/httpapi/session.go`). Behind a proxy that handles HTTPS (Fly.io, a load balancer), it sees plain HTTP, so the cookie is never `Secure`. Trust `X-Forwarded-Proto` from a configured proxy only, or add a setting that always marks it `Secure`. Settle it in step 8b of the [editions plan](docs/plans/editions-and-accounts.md), before the first hosted deployment.
- [ ] **Release automation:** GoReleaser (or similar), a changelog convention, and Renovate or Dependabot.
- [ ] **Pre-release tags for the desktop packages.** The Taskfile's `PACKAGE_VERSION` accepts only `vX.Y.Z`, so a tag like `v0.1.0-rc.1` builds installers numbered `0.0.0` (GoReleaser's CLI downloads get the right version). Each format writes a pre-release its own way: `.deb` and `.rpm` want `0.1.0~rc.1` (so it sorts before `0.1.0`), while Windows' version info takes only numbers. Then mark such a Release as a pre-release (GoReleaser's `release.prerelease: auto`), and update docs/ci-cd.md's "Making a release".
- [ ] **Dockerized Mocking Infrastructure** such as a docker-compose stack to spin up auxiliary local services like mock authentication, testing databases, or local S3 object buckets seamlessly.
- [ ] **Load testing infrastructure** 
- [ ] **SQL injection and other security checks**
- [ ] **Create a Claude command to review open Dependabot PRs and branches on provide guidance on which to merge, or if any should be deleted.

## API and data conventions
- [ ] **Pagination, sorting, and filtering on list endpoints**, cursor-based. `GET /items` currently returns everything.
- [ ] **Concurrent-edit protection:** a `version` field or ETag on updates, so concurrent edits don't silently overwrite each other.
- [ ] **Seed data:** `task dev:seed` for realistic local data.

## Front-end polish
- [ ] **No flash of the wrong theme or style:** a tiny inline script in `web/index.html` that applies both before React loads.
- [ ] **Shared feedback patterns:** toasts, an error boundary, and a proper 404 page.
- [ ] **Installed-app polish** (follow-ups from SPEC.md §4.5):
  - precache the fonts and icons too (the service worker's default `globPatterns` cover only JS, CSS, and HTML, so offline the app falls back to system fonts until they're cached);
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
    - Install dependencies:
      - Install [Task](https://taskfile.dev), which runs the project's commands. It is installed with Go, so first install Go 1.27 or newer from <https://go.dev/dl>. Then run `go install github.com/go-task/task/v3/cmd/task@v3.54.0`, which is the same on Linux, macOS, and Windows and matches the version CI uses. Check it with `task --version`. If the command is not found, add the `bin` folder inside the path printed by `go env GOPATH` to your `PATH`, then open a new terminal.
      - Install Node 24, the version named in `.nvmrc`, with nvm. First uninstall any Node that was not installed with nvm, so that tools that do not load nvm cannot pick up the old version. Check with `node --version` in a new terminal that it prints v24.
        - Linux and macOS: [install nvm](https://github.com/nvm-sh/nvm#installing-and-updating), then from the project folder run `nvm install` (it reads `.nvmrc`) and `nvm alias default 24` so that new terminals use 24.
        - Windows: run `winget install CoreyButler.NVMforWindows`, open a new terminal (if your terminal runs inside an editor or IDE, close and reopen that application first, because its terminals keep the `PATH` it started with), then run `nvm install 24` and `nvm use 24`, approving the administrator prompt if one appears. nvm-windows does not read `.nvmrc`, so the version must be typed. `nvm use` applies to every terminal and stays in effect, so 24 is the default.
      - Install pnpm 10 with `npm i -g pnpm@10`, and check it with `pnpm --version`. nvm keeps global packages separately for each Node version, so run this again after changing to a new major version of Node. On Windows, if PowerShell says that running scripts is disabled, run `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned` and try again.
      - Install golangci-lint 2.15, the version CI uses, built with the Go version on the `go` line of `go.mod` (below, 1.27.3). It refuses to lint code that targets a newer Go than the one it was built with, and without `GOTOOLCHAIN` Go may build it with an older version. Check it with `golangci-lint version`, which also prints the Go version it was built with.
        - Linux and macOS: `GOTOOLCHAIN=go1.27.3 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15`
        - Windows (PowerShell): `$env:GOTOOLCHAIN='go1.27.3'; go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15; Remove-Item Env:GOTOOLCHAIN`
      - Optionally, install lefthook 2, which runs the git hooks: formatting and lint checks on each commit, and the unit tests on each push. Run `go install github.com/evilmartians/lefthook/v2@v2.2.2` and check it with `lefthook version`. Do this before `task setup`, which sets up the hooks and silently skips them if lefthook is missing.
      - Install a 64-bit C compiler, which `go test -race` needs. Check it with `go test -race ./internal/core/domain/`.
        - Linux: `sudo apt install build-essential`. macOS: `xcode-select --install`.
        - Windows: install [MSYS2](https://www.msys2.org), then in its terminal run `pacman -S mingw-w64-x86_64-gcc`. In PowerShell, run `go env -w CC=C:\msys64\mingw64\bin\gcc.exe`, so that Go uses this compiler even if an older `gcc` comes first on your `PATH`.
      - From the project folder, run `task setup`. It downloads the Go modules, installs the web packages, downloads the Playwright browsers (a few hundred megabytes), and sets up the git hooks. On Linux it may ask for your password, to install the system libraries the browsers need. Run it once after cloning.
      - Run `task check` to confirm that everything works. It does not yet pass on Windows (see "Make `task check` pass on Windows" below).
    - Replace the placeholder `Categories=Utility;` in the Linux desktop file (the `.desktop` file in `build/desktop/linux/`) with the app's own (from the [freedesktop menu spec](https://specifications.freedesktop.org/menu-spec/latest/category-registry.html)). The template needs some category, because the AppImage tool refuses a desktop file without one.
    - Choose where the server is hosted: follow the README's "One-time Fly.io setup" (or replace `deploy.yml` for another host), then set the repository variable `DEPLOY_ENABLED` to `true`. Until then, deploys are skipped.
    - For a hosted deployment, create a Postgres database (Fly Postgres or any other), and give the server `STARTER_STORAGE=postgres` and the database's connection URL as the secret `STARTER_DATABASE_URL` (on Fly.io, with `fly secrets set`).
    - Define the product's plans in `internal/adapters/plans/plans.jsonc`: its limits, its features (each with a lasting key, a display `name`, and `whenOff`; see `docs/features.md`), and the plans hosted workspaces will be on, with `newWorkspacePlan` naming the one a new account's workspace starts on (`docs/features.md`, "The starting plan"). The template ships `local` (no limits) and the examples `free` and `pro`, and its only feature is the sample's `item-export`. The local edition's plan is `local` unless a build picks another with `-ldflags "-X <module>/internal/bootstrap.LocalPlan=<plan>"`; `STARTER_PLANS_FILE` replaces the file at run time.
    - Decide how customers will pay: by hand, Stripe, or a merchant of record such as Paddle. `docs/billing.md` walks through each, and says what to do about tax, failed payments and a paid local licence. The template takes no payments, and needs an account with whichever provider is chosen, its keys and its webhook signing secret.
- [ ] **Make `task check` pass on Windows.** With a 64-bit C compiler in place, the remaining failures come from tests and a task that assume Unix:
  - `task e2e:terminal` is Unix only (`test/terminal` has `//go:build unix`), but `task check` always runs it, so `task check` stops there on Windows.
  - Tests that delete files the app still has open, which Windows refuses (`.lock`, `data.db`): `TestServerIsNotReadyOnceItsDataFolderIsRemoved`, `TestOpenIsNotReadyOnceTheDataFolderIsRemoved`, `TestStoreIsReadyUntilItsDataFolderIsRemoved`, `TestNewEmbedded_IsNotReadyOnceTheDataFolderIsRemoved`, and the three `server` acceptance scenarios about data that is gone or comes back. Decide what these should prove on Windows, since the behaviour is currently unchecked there.
  - `TestOpenUsesTheSettingsFilesPlansFile` leaves `.lock` open when its temporary folder is removed.
  - `TestWriteMarkReplacesTheOldFileRatherThanWritingOverIt` is a false alarm: on Windows, `os.SameFile` reads a file's identity when it is called, so both `os.Stat` results point at the new file. `WriteMark` does replace the file.
  - `TestRenameFileRewritesAndMovesAFile` expects a Unix executable permission.
  - `TestMainRunsInAGitRepository` builds `initproject` without the `.exe` extension, so it cannot start it, and `task init` is untested on Windows.
- [ ] Notate in CLAUDE.md that all bugfixes and new features should be functional for all clients: web, desktop, mobile, cli, and tui.
- [ ] Generate an architecture document, with diagrams, that explains the project directory structure and the purposes of each directory to give users of this template a foothold for starting new development. Also create a similar document with diagrams for each of the supported front-ends in place in their corresponding directory under `docs`, as well as the core Go library itself (just a stub for the template to be filled in later), and the REST API (also a stub).
- [ ] **Hosted and local editions:** let web, desktop, and mobile be commercial products, with data kept separate per account and features gated by paid plans, while the local edition (the CLI and TUI by default, the desktop and mobile apps in local mode, and a self-hosted server with sign-in off) stays a set of single-user tools on local data (plain text or SQLite, in a folder the user chooses, which can be a cloud-synced one), on the same core and data structures. The CLI and TUI can also log in to the hosted edition.
  - Plan: [docs/plans/editions-and-accounts.md](docs/plans/editions-and-accounts.md). Decisions: [ADR 5](docs/adr/0005-hosted-and-local-editions.md). Data belongs to a workspace; plans in config grant features and limits; billing is a driving adapter; OIDC authentication; synced folders are used from one device at a time, with no merging; no sync between local and cloud data.
- [ ] Make session expiration configurable (by admins only). Currently it's hardcoded to 30 days, which seems too long.
- [ ] Add release notes and update ./.claude/ commands to write them automatically
- [ ] Template-wide sweep to ensure developers using this template will be able to add domain-specific features while touching as few places as possible (Core, API, Docs, Storage adapters, CLI, TUI, and Web I think covers it but I could be missing some). Also make sure `/feature` commands prompt for decisions about the web, TUI, and CLI interfaces, since these cannot all be identical. The intention is that shared code is maximized with reason so that new feature development requires as little work as is feasible.
- [ ] Project cleanup: go through the project and look for potential opportunities to pair down unnecessary files and leave a clean-as-possible slate for users of the template.
  - [ ] Merge existing SQL migrations into a single one - backwards compatibility is not a consideration for this task.
