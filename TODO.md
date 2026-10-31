# Starter template backlog (temporary)

Ideas for things every project built from this template is likely to need.
Work through them one at a time (each usually starts with `/feature`), tick them off, and delete this file when it's empty.

## Do first
- [x] **"Make it mine" init script.** e.g. `task init -- --module github.com/you/app --name "Acme"`. It should:
  - rename the Go module, the "Starter" brand, the binaries and the Docker image;
  - optionally remove the sample `Item` domain.
- [x] **Create a "Rename project" script** based on the init script above. Allows for easy rebranding or changing from an initial "working title" later in development. Unlike the init script, it does not get deleted after running. Where logic can be shared between this and the init script and it makes sense to do so, avoid duplicating code.
- [x] In addition to the existing CLI, TUI, and web app, **create a matching PWA-based native Android and iOS app project, as well as a standalone desktop client** based on Electron or a similar technology to avoid duplicating the web UI. Help me think through the right frameworks/libraries for doing this that involves the least amount of duplicated code and extra work, and also help me figure out what software-based mobile emulators I can use on my Linux machine to test these out before loading onto physical devices.
  - Done (2026-10-30): a phone-width layout, an installable web app, a Wails v3 desktop app, and a Wails v3 Android app with its data on the phone. Decisions are in [ADR 4](docs/adr/0004-native-shells.md); emulator setup is in [docs/mobile.md](docs/mobile.md). iOS and the other follow-ups are the items below.
- [ ] **Desktop and mobile remote mode:** let the desktop and Android apps work on cloud data instead of local data: `desktop.Open` (and so `mobile.Open`) chooses `bootstrap.NewRemote` behind the same in-process handler, and leaves `LocalData` unset so the offline notice comes back. Choose the mode with an env var first (on the phone, a setting). Needs the accounts and authentication items (forwarding the user's session). No sync: see [ADR 4](docs/adr/0004-native-shells.md). This is step 10 of the [editions plan](docs/plans/editions-and-accounts.md).
- [ ] **Desktop packaging:** installers for the desktop app (AppImage or .deb, .dmg, .msi), probably with the `wails3` CLI in `tools/go.mod`, plus an app icon from `web/public/icon.svg`. Try the macOS and Windows builds for real; so far only Linux has been run.
  - Plan: [docs/plans/desktop-packaging.md](docs/plans/desktop-packaging.md). Linux AppImage, .deb and .rpm; Windows NSIS setup .exe instead of .msi; macOS .dmg built and tried on a Mac, and built in CI for releases; unsigned for now.
- [ ] **iOS app:** the Wails v3 mobile app on iPhone. It needs a Mac with full Xcode, and ideally a real iPhone. Its custom-scheme handler should pass request methods and bodies through, so it may not need the Android bridge (`docs/mobile.md`), but that's unverified. Add the generated Xcode project beside `mobile/android`, and cover the app ID in `task rename`.
- [ ] **Android release builds and Play Store signing:** a `mobile:android:release` task (release APK or AAB, `production` build tag), signing with a keystore from environment variables (the generated `app/build.gradle` already reads `ANDROID_KEYSTORE_*`), and a reversed-domain app ID instead of `com.example.<slug>` (see `docs/mobile.md`, "The app ID"). Consider building the APK in CI (it needs the NDK).
- [ ] **Android app icon:** the launcher icon is still Wails' default. Generate the `mipmap-*` icons (including round and adaptive ones) from `web/public/icon.svg`, alongside `pnpm --dir web icons`.
- [ ] **Android system bars follow dark mode:** the status and gesture bars always use the light background (`mobile/android/.../res/values/colors.xml`). In the app's dark mode they should go dark with light icons. The web app would tell Android its theme, through the bridge or a Java hook.
- [ ] **Android app on a real phone:** so far only the emulator has run it. Check the keyboard pushes the page up instead of covering the field (not checked on the emulator: its keyboard was stuck in stylus mode), and the layout on a phone with a notch or cut-out.
- [ ] Change the web app so that collapsing the sidebar (including the animation) is still possible even for wide displays. In those cases, rather than completely hiding the sidebar, it should be collapsed down to a narrow vertical strip with icons in place of the full text for each sidebar tab.
- [ ] **SQL storage adapter + migrations** (SQLite first). It must implement `ports.ItemRepository` and pass `storagetest.RunItemRepositoryContract`. It lifts the jsonfile adapter's single-instance limit (ADR 0003). This is step 6 of the [editions plan](docs/plans/editions-and-accounts.md); Postgres is step 7.
- [ ] **Authentication behind a port.** An `Authenticator` port with a simple session or OIDC adapter, wired through the REST API, the SPA, and the CLI's `--server` mode. Currently listed in SPEC.md §6 "Out of scope". This is step 8 of the [editions plan](docs/plans/editions-and-accounts.md): OIDC plus a dev-only fake login, with CLI and TUI login in step 9.
- [x] **Phone-width SPA layout.** The sidebar is a fixed `w-60`, so the app breaks on phones. Add a collapsible sidebar or bottom navigation, and check it in all three styles and both themes.
  - [x] Run `task test:init` to confirm `--remove-sample` still works now that the sidebar's content is shared with the phone menu (`SidebarContent` in `web/src/routes/Layout.tsx`). Passed on 2026-10-30.

## Operability
- [ ] **Logging, metrics and traces:**
  - structured `slog` logs with a request ID on every request;
  - a `/metrics` endpoint;
  - optional OpenTelemetry tracing.
- [ ] **Security basics:**
  - security headers and a CSP on the server;
  - request size limits and rate limiting on the API;
  - `govulncheck` and `pnpm audit` in CI.
- [ ] **Release automation:** GoReleaser (or similar), a changelog convention, and Renovate or Dependabot.

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
- [ ] Background jobs or domain events (outbox pattern)
- [ ] Audit log (who changed what, and when)

## Other
- [ ] Generate a new TODO.md file that lives in the template permanently. It's purpose is to list out tasks for a new project that a developer should start with when using this template, including things like running the `init` script, defining the vision and the definition of MVP in an appropriate place for Claude to reference, and swapping out the mockups for new styles, among others. It should link to referenced files where appropriate. Once these initial tasks are complete, the TODO.md file can stick around as a place to keep new ideas for the application that have yet to be implemented. Add a note to CLAUDE.md that feature and bugfix ideas live here. Its setup tasks should point to the README's Prerequisites (tools, plus system packages such as the desktop app's GTK/WebKit libraries and the Linux and Windows packaging tools) rather than repeat them.
  - Steps the new file must give **the developer using the template** (the template deliberately leaves these unset; they are not tasks for us):
    - Describe the desktop app in its Linux desktop file (the `.desktop` file in `build/desktop/linux/`): add a `Comment`, and replace the placeholder `Categories=Utility;` with the app's own (from the [freedesktop menu spec](https://specifications.freedesktop.org/menu-spec/latest/category-registry.html)). The template needs some category, because the AppImage tool refuses a desktop file without one.
    - Describe the Linux packages in `build/desktop/linux/nfpm.yaml`: add a `description` (until then, `apt show` and software centres say "no description given"), and replace the placeholder `maintainer` (`…@example.com`) with a real name and address.
    - Describe the Windows program in `build/desktop/windows/info.json`: replace the placeholder `FileDescription` (Task Manager shows it as the process's name; the template uses the product name), `CompanyName` and `LegalCopyright` with the project's own. Explorer's Properties → Details shows all three.
- [ ] **Init asks for the app's description and publisher.** The desktop packages' descriptions, publisher, maintainer and copyright (the template-user steps above) are filled in by hand for now. `task init` could take them as flags or prompts and write them into `build/desktop/`, so a new project starts with none of the template's placeholders.
- [ ] Notate in CLAUDE.md that all bugfixes and new features should be functional for all clients: web, desktop, mobile, cli, and tui.
- [ ] Generate an architecture document, with diagrams, that explains the project directory structure and the purposes of each directory to give users of this template a foothold for starting new development. Also create a similar document with diagrams for each of the supported front-ends in place in their corresponding directory under `docs`, as well as the Go library itself (just a stub for the template to be filled in later), and the REST API (also a stub).
- [ ] **Hosted and local editions:** let web, desktop and mobile be commercial products, with data kept separate per account and features gated by paid plans, while the CLI, TUI and local-only desktop and mobile apps stay single-user tools on local data (plain text or SQLite, in a folder the user chooses, which can be a cloud-synced one), on the same core and data structures.
  - Plan: [docs/plans/editions-and-accounts.md](docs/plans/editions-and-accounts.md). Decisions: [ADR 5](docs/adr/0005-hosted-and-local-editions.md). Data belongs to a workspace; plans in config grant features and limits; billing is a driving adapter; OIDC authentication; synced folders are used from one device at a time, with no merging; no sync between local and cloud data.
- [ ] **Fix `task accept -- -run 'TestFeatures/<driver>'`.** CLAUDE.md documents it for running one acceptance driver, but on 2026-10-30 `-run 'TestFeatures/memory$'` still ran all five drivers. A guess, not yet checked: `pflag.Parse()` in `test/acceptance/acceptance_test.go`'s `TestMain` interferes with `go test`'s `-run`. Fix it, or correct the CLAUDE.md tip.
- [ ] Project cleanup: go through the project and look for potential opportunities to pair down unnecessary files and leave a clean-as-possible slate for users of the template.
