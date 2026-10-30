# Starter template backlog (temporary)

Ideas for things every project built from this template is likely to need.
Work through them one at a time (each usually starts with `/feature`), tick them off, and delete this file when it's empty.

## Do first
- [ ] **"Make it mine" init script.** e.g. `task init -- --module github.com/you/app --name "Acme"`. It should:
  - rename the Go module, the "Starter" brand, the binaries and the Docker image;
  - optionally remove the sample `Item` domain.
- [ ] **SQL storage adapter + migrations** (SQLite first). It must implement `ports.ItemRepository` and pass `storagetest.RunItemRepositoryContract`. It lifts the jsonfile adapter's single-instance limit (ADR 0003).
- [ ] **Authentication behind a port.** An `Authenticator` port with a simple session or OIDC adapter, wired through the REST API, the SPA, and the CLI's `--server` mode. Currently listed in SPEC.md §6 "Out of scope".
- [ ] **Phone-width SPA layout.** The sidebar is a fixed `w-60`, so the app breaks on phones. Add a collapsible sidebar or bottom navigation, and check it in all three styles and both themes.

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
- [ ] **Automated accessibility checks:** axe in the Playwright e2e tests, across every style × theme.

## Maybe (depends on the products built from it)
- [ ] Translations (i18n)
- [ ] Background jobs or domain events (outbox pattern)
- [ ] Audit log (who changed what, and when)
