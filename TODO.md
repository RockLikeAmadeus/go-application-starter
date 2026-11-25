# Starter template backlog (temporary)

Ideas for things every project built from this template is likely to need.
Work through them one at a time (each usually starts with `/feature`), tick them off, and delete this file when it's empty.
`/prune-backlog` moves ticked items, once none of their sub-tasks are open, into [docs/backlog-history.md](docs/backlog-history.md).

**Before starting work on any feature below, discuss whether it's actually a valuable feature and what are the tradeoffs of including it or not including it. Many of these are Claude's ideas and probably very sensible, but some of them I've thought of myself and I want to be wary of template bloat.**

## How to read this list

Everything here is done before the template is "version 1 complete" and products are built on it, so that no work is done twice. The list is ordered by what depends on what: an item comes after the items and decisions it needs, and before the ones that need it.

Some items are still undecided: whether they belong in the template at all. Each one sits just before the first item its outcome shapes, rather than all at the start, so it's discussed when there's enough built to judge it (see the note above). A "yes" gives it a size, and it's built where it stands. A "no" moves it to the history with a line on why.

The items sit in five sections. Work through them in order. Within a section, the order is a suggestion, apart from the dependencies listed.
1. **Foundations with no open decisions:** work that later items build on, or conventions they follow, which can start straight away.
2. **Decisions that shape foundations, then the foundations they shape.**
3. **Decide, then build:** the remaining undecided items, mostly features shared by most products.
4. **Independent work:** items that nothing else waits on.
5. **Finishing:** items that describe or tidy everything else, so they come last.

Tags at the start of each item:
- **Size.** `S`: a few TDD cycles or a docs change, no `/feature` of its own. `M`: one `/feature`. `L`: several `/feature`s, so it needs a plan in `docs/plans/`. `?`: can't be sized until it's been discussed.
- **`decide`:** whether it belongs in the template is still open.
- **`waits on hardware`:** needs a Mac or an iPhone. It still counts towards version 1, but it doesn't hold up the other items while the hardware isn't there.

Lines under an item:
- **Shapes:** what the item's outcome changes.
- **Depends on:** what should be settled or built first. Only real dependencies are listed.

A new idea goes just before the first item that needs it, and after everything it needs. If it's not yet agreed, it's tagged `decide`.

## 1. Foundations with no open decisions

- [ ] `S` Notate in CLAUDE.md that all bugfixes and new features should be functional for all clients: web, desktop, mobile, cli, and tui.
  - First, as it changes how every later item is built.
- [ ] `M` **Make `task check` pass on Windows.** With a 64-bit C compiler in place, the remaining failures come from tests and a task that assume Unix:
  - `task e2e:terminal` is Unix only (`test/terminal` has `//go:build unix`), but `task check` always runs it, so `task check` stops there on Windows.
  - Tests that delete files the app still has open, which Windows refuses (`.lock`, `data.db`): `TestServerIsNotReadyOnceItsDataFolderIsRemoved`, `TestOpenIsNotReadyOnceTheDataFolderIsRemoved`, `TestStoreIsReadyUntilItsDataFolderIsRemoved`, `TestNewEmbedded_IsNotReadyOnceTheDataFolderIsRemoved`, and the three `server` acceptance scenarios about data that is gone or comes back. Decide what these should prove on Windows, since the behaviour is currently unchecked there.
  - `TestOpenUsesTheSettingsFilesPlansFile` leaves `.lock` open when its temporary folder is removed.
  - `TestWriteMarkReplacesTheOldFileRatherThanWritingOverIt` is a false alarm: on Windows, `os.SameFile` reads a file's identity when it is called, so both `os.Stat` results point at the new file. `WriteMark` does replace the file.
  - `TestRenameFileRewritesAndMovesAFile` expects a Unix executable permission.
  - `TestMainRunsInAGitRepository` builds `initproject` without the `.exe` extension, so it cannot start it, and `task init` is untested on Windows.
  - Early, because every later cycle runs these checks, and they're developed on Windows.
- [ ] `S`–`M` **Fix the occasional failures of `task test:init`.** During the cli-login feature it failed twice (in tasks 26b and 37) and passed when run again straight after, with no change in between. Find what makes it fail now and then, such as shared temporary folders, a port or file another run still holds, or tests running in parallel, and make it pass every time. Save the output of the next failing run, since the cause wasn't recorded.
  - Early, for the same reason.
- [ ] `M` **Split the core into `platform` and `product`.** Keep the general application machinery (accounts, sessions, sign-in codes, workspaces, plans and entitlements, billing, and domain errors) apart from the code for the problem the product exists to solve (today, the sample's items), so that each can be thought about without the other. Domain-driven design calls these the generic subdomains and the core domain.
  - Today `internal/core` is split by layer only, and each layer mixes both. Split it by kind first, then by layer: `internal/core/platform/{domain,ports,app}` and `internal/core/product/{domain,ports,app}`.
  - `product` may import `platform`, and `platform` never imports `product`. Enforce it with depguard in `.golangci.yml`, beside the layer rules, so it's checked on every lint run rather than only by `task test:init`.
  - Anything both use, such as the domain errors and workspaces, lives in `platform`. There's no third, shared tier: `product` depends on `platform` anyway, and a `common` package tends to fill with whatever is hard to place. If `platform` grows large, split it inside itself (for example `platform/accounts`).
  - Update CLAUDE.md's architecture and rules, `.claude/rules/`, `docs/workspaces.md`'s checklist for adding an entity, and the `--remove-sample` manifest (which could then empty `product/`). Run `task test:init`.
  - Early, as it decides where nearly all later code lives.
- [ ] `M`–`L` **Take the `platform` and `product` split out from the core** to the adapters and the web app: the REST API's handlers (sign-in routes and item handlers sit side by side in `httpapi`), the storage repositories, the CLI commands, the TUI, and the web app's screens. Decide for each how the split shows (packages, folders, or file names), as Go packages and the web app's folders work differently.
  - Depends on: splitting the core. Straight after it, and before the template-wide sweep, which builds on both.
- [ ] `M` **Give the TUI the product's plain name.** Someone who only uses the TUI (from a terminal, and the desktop app otherwise) should see it as an application of its own, not as a sub-command of the CLI (`starter tui` today). Make `starter` open the TUI, and rename the CLI to `starter-cli`, following `starter-desktop`. Both stay one set of code: add a small entry point beside `cmd/starter` that calls the same `cli` and `tui` code, so the TUI keeps every option the CLI gives it (`--server`, `--data-dir`, and so on), and its `--help` and messages use its own name. They stay separate from the desktop app, which needs GTK and WebKit on Linux, and is a window program rather than a console one on Windows; the three share the product's display name (`Starter`) but each has its own command name. Cover the new name in `task rename`, the Taskfile's `run:` tasks, GoReleaser's builds and archives, the terminal journeys, and the docs.
  - Early, as every later item that writes docs or messages names the CLI's command.
- [ ] `S` **Release notes and a changelog** (two items merged):
  - Add release notes and update ./.claude/ commands to write them automatically
  - **Release automation:** GoReleaser (or similar), a changelog convention, and Renovate or Dependabot. Partly done: `.goreleaser.yaml` and `release.yml` build releases, and `.github/dependabot.yml` exists. Left: the changelog convention.
  - Early, so that every later feature gets its notes as it's built.
- [ ] `M`–`L` **Searching, sorting, filtering, and pagination of records** (two items merged):
  - If possible, **basic record searching, sorting, and filtering functionality** that can be reused for as-yet-unimplemented and domain-specific data models in the application.
  - **Pagination, sorting, and filtering on list endpoints**, cursor-based. `GET /items` currently returns everything.
  - A convention every list endpoint and list screen follows, so it comes before more of them exist.
- [ ] `S` **Finish the hosted and local editions plan:** let web, desktop, and mobile be commercial products, with data kept separate per account and features gated by paid plans, while the local edition (the CLI and TUI by default, the desktop and mobile apps in local mode, and a self-hosted server with sign-in off) stays a set of single-user tools on local data (plain text or SQLite, in a folder the user chooses, which can be a cloud-synced one), on the same core and data structures. The CLI and TUI can also log in to the hosted edition.
  - Plan: [docs/plans/editions-and-accounts.md](docs/plans/editions-and-accounts.md). Decisions: [ADR 5](docs/adr/0005-hosted-and-local-editions.md). Data belongs to a workspace; plans in config grant features and limits; billing is a driving adapter; OIDC authentication; synced folders are used from one device at a time, with no merging; no sync between local and cloud data.
  - Left: step 5's check by hand with two machines (the plan's "Verification after every step"), the plan's "Loose ends", and deleting the plan. Its follow-ups are already in this list.
- [ ] `M` **Security basics:**
  - security headers and a CSP on the server;
  - request size limits and rate limiting on the API;
  - `govulncheck` and `pnpm audit` in CI.
  - **SQL injection and other security checks** (merged in: decide which checks, such as `gosec` in golangci-lint).
  - Shapes: "No flash of the wrong theme or style" (section 4), whose inline script the CSP must allow, and any page that loads outside scripts.
- [ ] `M` **Automated accessibility checks:** axe in the Playwright e2e tests, across every style × theme.
  - Early, so that every later screen is checked as it's built.

## 2. Decisions that shape foundations, then the foundations they shape

- [ ] `?` **Other features along these same lines likely to be required by template users** that we brainstorm.
  - Shapes: the template-wide sweep (which features it must make easy), and the rest of this list, so it comes first in this section.
- [ ] `?` `decide` **Consider how distributed workloads would function for projects using this template**. The hosted edition serving multiple users and relying on computationally intensive background tasks, for instance, is a future we haven't considered, if it even can be planned for ahead of deciding on a specific problem space. Even if the problem domain is relatively simplistic, a single server serving the web application to a theoretical large number of concurrent users isn't scalable, nor is a single database instance. Need to determine how much of this problem is worth considering at the template-level, if any.
  - Related: background jobs or domain events (outbox pattern) - (this very likely will be relevant for my first intended use of the template). Probably part of the same decision.
  - Shapes: logging, metrics, and traces (tracing matters once there's more than one server), concurrent-edit protection, the server and storage design, notifications, backups, the audit log, and load testing.
- [ ] `L` `decide` **Teams: workspace members and roles.** Today each account has one personal workspace. Teams would let several accounts share a workspace, each with a role that sets what they may do. (A follow-up from the [editions plan](docs/plans/editions-and-accounts.md).)
  - Shapes: concurrent-edit protection (teams are what make two people edit the same record), workspace scoping, plans and limits (per workspace or per member), account flows, the audit log, and notifications.
- [ ] `M` `decide` **Crash & Error Reporting**: pre-integrated boundaries (like Global Exception Handlers or React Error Boundaries) that intercept crashes and prepare them for upstream analytics.
- [ ] `M` `decide` **Telemetry & Analytics**: a unified tracking hook or service layer to measure core user events, app launch performance, and page/view transitions.
  - Shapes (both of the above): logging, metrics, and traces, and the shared feedback patterns' error boundary.
- [ ] `M` **Logging, metrics, and traces:**
  - structured `slog` logs with a request ID on every request;
  - a `/metrics` endpoint;
  - optional OpenTelemetry tracing.
  - Depends on: the distributed workloads, crash reporting, and telemetry decisions. Used by: "Say why a browser sign-in failed" (section 4), and anything else that logs.
- [ ] `M` **Shared feedback patterns:** toasts, an error boundary, and a proper 404 page.
  - Depends on: the crash reporting decision (the error boundary). Used by: "Say when the server can't be reached" (section 4), form validation, and any later screen.
- [ ] `M` **Concurrent-edit protection:** a `version` field or ETag on updates, so concurrent edits don't silently overwrite each other.
  - A convention every update follows, in the API and storage. Depends on: the distributed workloads and teams decisions.
- [ ] `M` Template-wide sweep to ensure developers using this template will be able to add domain-specific features while touching as few places as possible (Core, API, Docs, Storage adapters, CLI, TUI, and Web I think covers it but I could be missing some). Also make sure `/feature` commands prompt for decisions about the web, TUI, and CLI interfaces, since these cannot all be identical. The intention is that shared code is maximized with reason so that new feature development requires as little work as is feasible.
  - Before the shared features, so that they're built the way it settles. Depends on: the brainstorm above (to know which features it must make easy), and the `platform` and `product` split, in the core and beyond.

## 3. Decide, then build

Each item here is decided just before it's built, in the order of what depends on what.

- [ ] `L` `decide` **Translations (i18n)**
  - Shapes: every piece of user-facing text in all five clients, so it comes first in this section: a "yes" is far cheaper before the features below add more text.
- [ ] `M` `decide` **Functional feature-flagging support** to toggle unreleased features or modify runtime behavior dynamically without redeploying. (Plans already turn features on per workspace, `docs/features.md`; flags would be a separate, run-time switch.)
  - Shapes: how every later feature is turned on and off.
- [ ] `M` `decide` **Audit log** (who changed what, and when)
  - Shapes: storage and the domain (every change is recorded). Depends on: the distributed workloads decision (domain events).
- [ ] `L` `decide` **Notifications feature**. This might not be relevant for all applications based on this template, but likely common enough to build in, and template users can remove if it's not applicable. This could be in-app notifications as well as boilerplate for email/mobile/etc. notifications.
  - Depends on: the distributed workloads decision (sending runs in the background). Shapes: the settings menu (notification preferences).
- [ ] `M` `decide` **Change application settings (e.g. theme, style, other domain-specific settings) from a settings menu in all clients**
  - Shapes: any later feature with a preference of its own.
- [ ] `?` `decide` **Multi-provider auth** for traditional email/password and standard OAuth providers. Mostly covered already: with `sign_in = "oidc"`, email and password and "Sign in with Google" and the like come from the sign-in provider, and the product never stores passwords (`docs/sign-in.md`, "Email and password"). Decide whether anything is left.
- [ ] `?` `decide` **Session management**: secure token refresh cycles, automatic logout on expiration, and multi-device session invalidation? Partly covered: sessions already expire and can be signed out (ADR 9). What's left would be "sign out everywhere" and listing sessions.
  - Shapes: "Make the session lifetime a server setting" (section 4).
- [ ] `S` `decide` **Sweep away expired sessions and unused sign-in codes.** A session is deleted only when it's looked up and found expired, so one nobody looks up again stays in storage for good, and so does a one-time sign-in code nobody redeemed. (A follow-up from the editions plan, steps 8a and 9.)
  - Depends on: the distributed workloads decision (a sweep runs in the background), and the session management decision.
- [ ] `M` `decide` **Sign in from the command line on a computer with no browser,** such as one reached over SSH, with the device code flow ([RFC 8628](https://www.rfc-editor.org/rfc/rfc8628)): the CLI shows a short code, and the person enters it in a browser on any other device. Today `starter login` needs a browser on the same computer. (A follow-up from the editions plan, step 9.)
- [ ] `M` `decide` **A paid licence key for the local edition,** behind the entitlements port, so a local copy can unlock a paid plan without an account. `docs/billing.md` already describes the idea. (A follow-up from the editions plan.)
  - Depends on: the teams decision (whether a licence is per person or per team).
- [ ] `M` `decide` **User account** flows for common tasks like password changes, account deletion, etc. **ONLY FOR HOSTED EDITIONS**. (Password changes belong to the sign-in provider; account deletion and its data are the product's.)
- [ ] `M`–`L` `decide` **Placeholder functionality for exporting workspace data to a file, importing workspace data from a file**--especially useful for converting from the local only version to the hosted version, or possibly for the data backups item below.
- [ ] `M` `decide` **Automatic periodic and manual data backups?**
  - Depends on: the export and import decision above, and the distributed workloads decision (periodic backups run in the background).
- [ ] `M` `decide` **Stub for application documentation/help section.** Could also lay the groundwork for easy tooltips the for template users, that potentially could themselves link to documentation pages?
- [ ] `M` `decide` **A sample/placeholder home page** for the hosted edition SPA that provides feature info, pricing details, etc., along with a login button, unless this is more appropriate as a separate hosted site).
- [ ] `S`–`M` `decide` **Report a bug or suggest a feature** feature
- [ ] `S`–`M` `decide` **Form validation helpers**
- [ ] `M` `decide` **Dockerized Mocking Infrastructure** such as a docker-compose stack to spin up auxiliary local services like mock authentication, testing databases, or local S3 object buckets seamlessly. (`task db:up` already starts a test Postgres, and `oidctest` fakes the sign-in provider in tests.)
- [ ] `?` `decide` **Load testing infrastructure**
  - Depends on: the distributed workloads decision.
- [ ] `M`–`L` `decide` **Synced data folders on Android,** through Android's Storage Access Framework, which lets an app use a folder another app (such as a cloud-sync app) provides. Today the Android app keeps its data in its own private folder. (A follow-up from the editions plan.)

## 4. Independent work

### Fixes found in the remote-mode check by hand
- [ ] `M` **Say when the server can't be reached:** the offline notice follows only the browser's own online signal (`navigator.onLine`, `useIsOnline`), so when the server stops, the web app and the desktop and Android apps in remote mode show the general "An error occurred" message instead (found in remote mode's check by hand, task 22, step A7). Tell a request that never reached the server apart from one the server refused, show a notice for it (agree the wording first), and retry once the server answers again. Respect `GET /api/v1/connection` as the offline notice does. Also: when the session lookup itself can't reach the server (for example an app built with a wrong `BUILT_IN_SERVER_URL`), the web app shows the items page as if sign-in were off, with only the general error; it should show the same notice instead.
  - Depends on: shared feedback patterns.
- [ ] `S` **Say why a browser sign-in failed:** the loopback page in `internal/adapters/browsersignin` shows "The server refused to exchange the sign-in code…" (`swapRefusedPage`) for every error from `swap`, including a request that never reached the server, and nothing logs the error. In remote mode's check by hand (task 22), creating an account on an Android phone failed twice with that page, and worked after the server was restarted, so the cause is unknown. Tell a connection failure apart from a refused code (`ErrCodeRefused`), show each its own page (agree the wording first), and log the underlying error so the next failure can be diagnosed (on Android it should reach `adb logcat`).
  - Depends on: logging, metrics, and traces.
- [ ] `M` **Bring the app forward after a browser sign-in:** when the desktop or Android app in remote mode signs in through the browser, the browser stays in front on "Sign-in complete", and the person has to switch back themselves (found in remote mode's check by hand, task 22).
  - Desktop: when the loopback port receives the code, call Wails' `Restore()` and `Focus()` on the window. Windows and many Linux desktops prevent focus stealing, so expect the app to flash in the taskbar or show a "ready" notice rather than take the screen; try it on both.
  - Android: since Android 10, an app in the background can't bring itself forward. Register a custom scheme (for example `starter://signed-in`) with an intent filter in `mobile/android/app/src/main/AndroidManifest.xml`, and give the "Sign-in complete" page a **Return to the app** button that opens it. The code still comes back through `127.0.0.1`, so the scheme carries nothing secret. Agree the wording first, cover the scheme in `task rename`, and try it on a phone, as Chrome may ask before opening the app.

### Settings and sessions
- [ ] `S` **Desktop app command-line flags:** `starter-desktop` reads no arguments, so `--data-dir` (which `bin/starter` takes) is silently ignored and the app opens its default folder. Give it the same `--data-dir`, `--storage`, and `--config` flags as the CLI, layered as in ADR 6, and have it refuse a flag it doesn't know instead of ignoring it (agree the message's wording first).
- [ ] `S` **Expand `~` in folder settings:** `data_dir` and the other paths are used as given, so `~/Dropbox/starter` from the settings file, `STARTER_DATA_DIR`, or a quoted `--data-dir` becomes a folder literally named `~`. Expand a leading `~` (and `~\` on Windows) to the user's home folder wherever a setting names a path, in `internal/config`, and say so in `docs/configuration.md`.
- [ ] `S` **Make the session lifetime a server setting.** A session lasts a fixed 30 days from sign-in (`domain.SessionLifetime` in `internal/core/domain/session.go`), which seems too long, and only a code change alters it. Make it a server setting (for example `session_lifetime`, and `STARTER_SESSION_LIFETIME`), layered as in ADR 6, so whoever runs the server chooses it and people using the app can't. It sets the cookie's `MaxAge` and the command line's and apps' bearer tokens alike. Refuse a value that isn't a positive length of time at start (agree the message's wording first), pick a shorter default, and document it in `docs/configuration.md` and `docs/sign-in.md`.
  - Depends on: the session management decision (section 3), which may change what the lifetime counts from.
  - Depends on: the session management decision.

### Front end
- [ ] `S` **No flash of the wrong theme or style:** a tiny inline script in `web/index.html` that applies both before React loads.
  - Depends on: security basics (the CSP).
- [ ] `S` **Installed-app polish** (follow-ups from SPEC.md §4.5):
  - precache the fonts and icons too (the service worker's default `globPatterns` cover only JS, CSS, and HTML, so offline the app falls back to system fonts until they're cached);
  - a dark-theme `theme-color` meta, so the browser bar matches the dark theme;
  - offline, the Items page shows "Loading…" under the offline notice, because TanStack Query pauses its requests. Show nothing, or a quieter placeholder.
- [ ] `M` Change the web app so that collapsing the sidebar (including the animation) is still possible even for wide displays. In those cases, rather than completely hiding the sidebar, it should be collapsed down to a narrow vertical strip with icons in place of the full text for each sidebar tab.

### Testing and tools
- [ ] `S` **Browser smoke runs against a deployment start no sign-in server.** With `E2E_BASE_URL` set, `web/playwright.config.ts` starts no server, but the `sign-in` project still runs `web/e2e/features/sign-in.feature` against its own base URL (`localhost`, the port after `PORT`), so it fails there. Skip that project when `E2E_BASE_URL` is set, or point it at a deployment with sign-in on.
- [ ] `S` **Seed data:** `task dev:seed` for realistic local data.
- [ ] `S` **Create a Claude command to review open Dependabot PRs and branches and provide guidance on which to merge, or if any should be deleted.**

### Desktop and mobile
- [ ] `M` `waits on hardware` **Desktop packaging:** installers for the desktop app (AppImage or .deb, .dmg, .msi), probably with the `wails3` CLI in `tools/go.mod`, plus an app icon from `web/public/icon.svg`. Try the macOS and Windows builds for real; so far only Linux has been run.
  - Plan: [docs/plans/desktop-packaging.md](docs/plans/desktop-packaging.md). Linux AppImage, .deb and .rpm; Windows NSIS setup .exe instead of .msi; macOS .dmg built and tried on a Mac, and built in CI for releases; unsigned for now.
  - Linux and Windows done 2026-11-04: packages built and tried by hand, and attached to the GitHub Release by `release.yml` (tried with a throwaway `v0.0.1` tag). Left: the macOS `.dmg` (step 4, on the Mac) and the rest of step 5.
- [ ] `S` **Pre-release tags for the desktop packages.** The Taskfile's `PACKAGE_VERSION` accepts only `vX.Y.Z`, so a tag like `v0.1.0-rc.1` builds installers numbered `0.0.0` (GoReleaser's CLI downloads get the right version). Each format writes a pre-release its own way: `.deb` and `.rpm` want `0.1.0~rc.1` (so it sorts before `0.1.0`), while Windows' version info takes only numbers. Then mark such a Release as a pre-release (GoReleaser's `release.prerelease: auto`), and update docs/ci-cd.md's "Making a release".
- [ ] `S` **Android app on a real phone:** so far only the emulator has run it. Check the keyboard pushes the page up instead of covering the field (not checked on the emulator: its keyboard was stuck in stylus mode), and the layout on a phone with a notch or cut-out. (Remote mode was checked on a phone on 2026-11-21; these two checks may still be open.)
- [ ] `S` **Android app icon:** the launcher icon is still Wails' default. Generate the `mipmap-*` icons (including round and adaptive ones) from `web/public/icon.svg`, alongside `pnpm --dir web icons`.
- [ ] `S` **Android system bars follow dark mode:** the status and gesture bars always use the light background (`mobile/android/.../res/values/colors.xml`). In the app's dark mode they should go dark with light icons. The web app would tell Android its theme, through the bridge or a Java hook.
- [ ] `M` **Android release builds and Play Store signing:** a `mobile:android:release` task (release APK or AAB, `production` build tag), signing with a keystore from environment variables (the generated `app/build.gradle` already reads `ANDROID_KEYSTORE_*`), and a reversed-domain app ID instead of `com.example.<slug>` (see `docs/mobile.md`, "The app ID"). Consider building the APK in CI (it needs the NDK).
- [ ] `L` `waits on hardware` **iOS app:** the Wails v3 mobile app on iPhone. It needs a Mac with full Xcode, and ideally a real iPhone. Its custom-scheme handler should pass request methods and bodies through, so it may not need the Android bridge (`docs/mobile.md`), but that's unverified. Add the generated Xcode project beside `mobile/android`, and cover the app ID in `task rename`.

## 5. Finishing

- [ ] `M` Project cleanup: go through the project and look for potential opportunities to pair down unnecessary files and leave a clean-as-possible slate for users of the template.
- [ ] `S` **Start the stored formats with no history.** The template's database schema should look as if it was designed all at once, so that a template user's own migrations clearly stand apart from it. How the schema grew while the template was built doesn't matter to them. Replace the SQL migrations in `internal/adapters/storage/sqlstore/migrations/` with two: one for the template's own tables, and one for the sample's table, which `--remove-sample` deletes. Start the JSON storage's record format again at version 1. Backwards compatibility isn't a concern.
  - After every item that adds a migration or changes a stored format, and before the first product starts, since after that each product's data depends on them.
- [ ] `M` **A guide to the template: what it includes, and how it's built.** Two ideas that overlap, so probably one guide with an architecture part (or two documents that link to each other). It describes everything above, so it's written once they're done:
  - Start a **"What This Template Includes" guide** which describes the features we've included so far, broken down into appropriate sections and subsections, with links to further information about specific features where available/appropriate. Keep this document updated as we implement new features for the template. This will be very valuable so template users know both _what_ is already included and therefore doesn't need to be done from scratch, as well as _how_ to work with those features, add to them, change them, etc. The guide should also discuss at a high level the technologies involved and a very high-level description of the application architecture.
  - Generate an architecture document, with diagrams, that explains the project directory structure and the purposes of each directory to give users of this template a foothold for starting new development. Also create a similar document with diagrams for each of the supported front-ends in place in their corresponding directory under `docs`, as well as the core Go library itself (just a stub for the template to be filled in later), and the REST API (also a stub).
- [ ] `M` **Generate a new TODO.md file that lives in the template permanently.** Last, as most items above add steps to it. Its purpose is to list out tasks for a new project that a developer should start with when using this template, including things like running the `init` script, defining the vision and the definition of MVP in an appropriate place for Claude to reference, and swapping out the mockups for new styles, among others. It should link to referenced files where appropriate. Once these initial tasks are complete, the TODO.md file can stick around as a place to keep new ideas for the application that have yet to be implemented. Add a note to CLAUDE.md that feature and bugfix ideas live here. Its setup tasks should point to the README's Prerequisites (tools, plus system packages such as the desktop app's GTK/WebKit libraries and the Linux and Windows packaging tools) rather than repeat them.
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
      - Run `task check` to confirm that everything works. It does not yet pass on Windows (see "Make `task check` pass on Windows" above).
    - Replace the placeholder `Categories=Utility;` in the Linux desktop file (the `.desktop` file in `build/desktop/linux/`) with the app's own (from the [freedesktop menu spec](https://specifications.freedesktop.org/menu-spec/latest/category-registry.html)). The template needs some category, because the AppImage tool refuses a desktop file without one.
    - Choose where the server is hosted: follow the README's "One-time Fly.io setup" (or replace `deploy.yml` for another host), then set the repository variable `DEPLOY_ENABLED` to `true`. Until then, deploys are skipped.
    - For a hosted deployment, create a Postgres database (Fly Postgres or any other), and give the server `STARTER_STORAGE=postgres` and the database's connection URL as the secret `STARTER_DATABASE_URL` (on Fly.io, with `fly secrets set`).
    - To make the desktop or Android app a client of your hosted server, build it with its address: `BUILT_IN_SERVER_URL=https://your-server.example.com task build:desktop` (and the same for the Android build). The address is the lowest layer of `server_url`, so a setting or `STARTER_SERVER_URL` still wins, and a program built with one has no local mode.
    - Define the product's plans in `internal/adapters/plans/plans.jsonc`: its limits, its features (each with a lasting key, a display `name`, and `whenOff`; see `docs/features.md`), and the plans hosted workspaces will be on, with `newWorkspacePlan` naming the one a new account's workspace starts on (`docs/features.md`, "The starting plan"). The template ships `local` (no limits) and the examples `free` and `pro`, and its only feature is the sample's `item-export`. The local edition's plan is `local` unless a build picks another with `-ldflags "-X <module>/internal/bootstrap.LocalPlan=<plan>"`; `STARTER_PLANS_FILE` replaces the file at run time.
    - Decide how customers will pay: by hand, Stripe, or a merchant of record such as Paddle. `docs/billing.md` walks through each, and says what to do about tax, failed payments and a paid local licence. The template takes no payments, and needs an account with whichever provider is chosen, its keys and its webhook signing secret.
    - Choose a sign-in provider for the hosted edition, such as Auth0, Clerk, Keycloak, or Zitadel (`docs/sign-in.md`, "Choosing a provider": email and password accounts come from the provider, and moving them to another provider later can cost users a password reset). Register the app with it as a server-side web application, register `<public_url>/auth/oidc/callback` as its redirect URL, and turn on its sign-up. Then give the server `STARTER_SIGN_IN=oidc`, `STARTER_PUBLIC_URL`, `STARTER_OIDC_ISSUER`, and `STARTER_OIDC_CLIENT_ID`, and the client secret as the secret `STARTER_OIDC_CLIENT_SECRET` (on Fly.io, with `fly secrets set`). The template ships with no provider, so a server with `sign_in = "oidc"` stops at start until these are set (`docs/sign-in.md`, "Setting up a provider").
