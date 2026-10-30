# Starter template backlog (temporary)

Ideas for things every project built from this template is likely to need.
Work through them one at a time (each usually starts with `/feature`), tick them off, and delete this file when it's empty.

## Do first
- [x] **"Make it mine" init script.** e.g. `task init -- --module github.com/you/app --name "Acme"`. It should:
  - rename the Go module, the "Starter" brand, the binaries and the Docker image;
  - optionally remove the sample `Item` domain.
- [x] **Create a "Rename project" script** based on the init script above. Allows for easy rebranding or changing from an initial "working title" later in development. Unlike the init script, it does not get deleted after running. Where logic can be shared between this and the init script and it makes sense to do so, avoid duplicating code.
- [ ] In addition to the existing CLI, TUI, and web app, **create a matching PWA-based native Android and iOS app project, as well as a standalone desktop client** based on Electron or a similar technology to avoid duplicating the web UI. Help me think through the right frameworks/libraries for doing this that involves the least amount of duplicated code and extra work, and also help me figure out what software-based mobile emulators I can use on my Linux machine to test these out before loading onto physical devices.
  - Plan, decisions and progress: [`docs/plans/native-front-ends.md`](docs/plans/native-front-ends.md) (steps 1–2 done; next: desktop app).
- [ ] Figure out how we can structure this so that the web and mobile applications can be potentially commercial products with data separated into individual user accounts, with feature configuration possible so that it's a simple matter to offer the product for free with certain features locked behind paid accounts. At the same time, I would like to have the TUI, CLI, and a local-only free version of the desktop and mobile apps that operate as *local-only* tools for a *single user*, with data stored in a local sql database or plain text. These products would use the same backend data structures, excepting for having separate user account data. This is a major change to the original concept so will need significant upfront planning and I expect you to ask me lots of clarifying questions, if it is indeed even realistic (it may be too much of an initial lift and maintenance headache to be worth doing).
  - Decided: **no sync between local and cloud data.** A copy of the app works on either local data or cloud data, never both, and switching between them migrates nothing. (The desktop app is designed to allow both modes; see [`docs/wip/desktop-app.md`](docs/wip/desktop-app.md), "Keep remote mode open".)
- [ ] **SQL storage adapter + migrations** (SQLite first). It must implement `ports.ItemRepository` and pass `storagetest.RunItemRepositoryContract`. It lifts the jsonfile adapter's single-instance limit (ADR 0003).
- [ ] **Authentication behind a port.** An `Authenticator` port with a simple session or OIDC adapter, wired through the REST API, the SPA, and the CLI's `--server` mode. Currently listed in SPEC.md §6 "Out of scope".
- [x] **Phone-width SPA layout.** The sidebar is a fixed `w-60`, so the app breaks on phones. Add a collapsible sidebar or bottom navigation, and check it in all three styles and both themes.
  - [ ] Run `task test:init` to confirm `--remove-sample` still works now that the sidebar's content is shared with the phone menu (`SidebarContent` in `web/src/routes/Layout.tsx`).

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
- [ ] Generate a new TODO.md file that lives in the template permanently. It's purpose is to list out tasks for a new project that a developer should start with when using this template, including things like running the `init` script, defining the vision in an appropriate place for Claude to reference, and swapping out the mockups for new styles, among others. It should link to referenced files where appropriate. Once these initial tasks are complete, the TODO.md file can stick around as a place to keep new ideas for the application that have yet to be implemented. Add a note to CLAUDE.md that feature and bugfix ideas live here. Its setup tasks should point to the README's Prerequisites (tools, plus system packages such as the desktop app's GTK/WebKit libraries) rather than repeat them.
- [ ] Generate an architecture document, with diagrams, that explains the project directory structure and the purposes of each directory to give users of this template a foothold for starting new development. Also create a similar document with diagrams for each of the supported front-ends in place in their corresponding directory under `docs`, as well as the Go library itself (just a stub for the template to be filled in later), and the REST API (also a stub).
