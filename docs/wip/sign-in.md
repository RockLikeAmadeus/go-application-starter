# Sign in to the hosted edition (with the dev sign-in)

Part 8a of step 8 of [the editions plan](../plans/editions-and-accounts.md). Part 8b adds the OIDC adapter behind the same seam.

- Spec: [SPEC.md §4.18](../../SPEC.md#418-sign-in-to-the-hosted-edition)
- Scenarios: [`features/sign-in.feature`](../../features/sign-in.feature) (`@sign-in`, 6 scenarios, each tagged `@wip`; the last is sample). Task 22 adds the http driver to [`features/workspaces.feature`](../../features/workspaces.feature), and task 25 adds the browser journey `web/e2e/features/sign-in.feature`.
- Decisions: [ADR 5](../adr/0005-hosted-and-local-editions.md) (to update), a new ADR 9 for sign-in and sessions (task 26)

## Decisions

Agreed with the user (2026-11-10):
- **Step 8 is split in two.** 8a (this feature) builds all of sign-in and proves it with the dev sign-in. 8b adds the OIDC adapter, tested against a small fake OIDC provider inside the tests.
- **Session cookie, not tokens in the browser** (the backend-for-frontend pattern, ranked most secure by the IETF's browser-apps draft). The server signs people in and gives the browser an `HttpOnly`, `SameSite=Lax` cookie (`Secure` over HTTPS) that page scripts can't read. The web app holds no tokens. The CLI in step 9 will send a bearer token instead, so the middleware will accept both later.
- **Sessions and accounts are kept in the app's storage**, so every server copy on one database knows them and signing out ends a session at once.
- **Sign-in works on memory, SQLite, and Postgres storage only.** Memory is for the tests and `task dev`. The server refuses sign-in with jsonfile at start, so jsonfile gets no account or session repository.

Decided in planning:
- **The seam:** a sign-in adapter (the dev sign-in now, OIDC in 8b) works out who someone is, as a `domain.Identity` (issuer, subject, name), and calls the driving port `ports.SignInService`. The core never sees how they proved it. Methods: `SignIn(ctx, identity)` returns a session token; `CurrentAccount(ctx, token)` returns the session's account (its name and workspace); `SignOut(ctx, token)`.
- **Accounts:** `ports.AccountRepository` finds an account by its identity (issuer and subject, never the name) and saves new ones. It's not workspace-scoped, as an account is what picks the workspace. A new account's ID and its personal workspace's ID come from `ports.IDGenerator`. No stored plan is written: `plans.StoredPlanOf` answers the starting plan for a workspace with none.
- **Sessions:** the token is 32 random bytes from a new driven port, `ports.TokenGenerator` (a `crypto/rand` adapter, and a `testsupport` fake). Storage keeps only the token's SHA-256 hash, so a leaked database holds no working sessions. A session lasts 30 days from sign-in (`ports.Clock`), with no renewal for now. An expired session is treated as missing and deleted when found. A sweep for expired sessions nobody looks up is a plan follow-up.
- **Signed out is "no workspace":** with no valid session, the middleware sets no workspace, so the use cases fail closed with `domain.ErrNoWorkspace`, as they already do. `problem.go` maps it to a 401 problem, and `httpclient` maps it back. The step "told to sign in" asserts `ErrNoWorkspace`.
- **The setting:** `sign_in` (`STARTER_SIGN_IN`), `off` by default, or `dev`. 8b adds `oidc`. Only the server and the CLI and TUI's admin commands read it. With it off, nothing changes: the `Local` resolver and the fixed local plan.
- **Dev sign-in only in dev builds:** its wiring sits behind the build tag `devsignin`. `task dev`, the acceptance http driver (which builds the handler in-process), and the browser journey's server use it. `task build` doesn't, and there `sign_in = "dev"` is refused at start.
- **Routes outside the API:** sign-in is redirects and an HTML form, not JSON, so it lives under `/auth/` beside the API, not in `api/openapi.yaml`. `GET /auth/sign-in` goes to the configured sign-in adapter (for dev, `GET /auth/dev`, a plain form asking for a name, posting to `POST /auth/dev`). `POST /auth/sign-out` ends the session and clears the cookie. Both go in `navigateFallbackDenylist` (`.claude/rules/server.md`). The dev form is developer-only text, so it isn't agreed as user-facing text.
- **CSRF:** with a session cookie, a request that changes something (not `GET` or `HEAD`) is refused unless it comes from the same site (`Sec-Fetch-Site: same-origin`, or an `Origin` matching the host). `SameSite=Lax` covers most cases already; this closes the rest.
- **The web app learns its state from `GET /api/v1/session`** (in the OpenAPI file): `{"isSignInOn": …, "name": …}`, always 200. Signed out with sign-in on, it shows a sign-in screen linking to `/auth/sign-in`. Signed in, it shows a sign-out button. With sign-in off, neither. Their wording is agreed with the user in tasks 23 and 24.
- **Stored plans switch on with sign-in** (`docs/billing.md`, "Switching on stored plans"). The plans file names the starting plan for new workspaces in a new top-level field, `newWorkspacePlan`, which must be one of its plans. The built-in file and the test plans start on `free`. Locked features name their plans (hints on). `bootstrap.NewBilling` returns the real service when `sign_in` is on, so `starter admin set-plan` works beside a hosted server's data. The local edition keeps refusing.
- **Acceptance drivers:** `@sign-in` runs on memory, sqlite, postgres, and http. memory, sqlite, and postgres call `SignInService` directly, as the dev sign-in would. http signs in through `/auth/dev`, with a cookie jar per account. jsonfile, desktop, mobile, and server leave `@sign-in` out. In task 22, http also runs `@workspaces`, where workspace "Alpha" is the personal workspace of an account named "Alpha", and "someone with no workspace" hasn't signed in.
- **The browser journey** needs a server with sign-in on, so it gets its own Playwright project and server (built with `devsignin`, `STARTER_SIGN_IN=dev`), leaving the other journeys on a server without sign-in. It's written in task 25, not in the spec commit, as an undefined Playwright step breaks the whole browser suite.
- **Sample boundary:** everything here is template code. The sample part is the one scenario where an account adds an item, and its step. Run `task test:init` in tasks 21, 22, and 26.
- **SPEC.md §6 keeps "Authentication and authorization"** until 8b, when real sign-in exists. Authorization (roles) stays out of scope until teams.

## Tasks

Each is one `/red` → `/green` → `/refactor` cycle.

The core:
- [ ] 1. `domain.NewIdentity(issuer, subject, name)` refuses a blank issuer or subject with a `*domain.ValidationError`. An `Account` has an ID, an identity, and a workspace.
- [ ] 2. `domain.Session` knows when it's expired: `IsExpiredAt(now)` is true at or after its `ExpiresAt`.
- [ ] 3. `ports.SignInService`, `AccountRepository`, `SessionRepository`, and `TokenGenerator`, with `testsupport` fakes where needed. `app.SignInService.SignIn` with a new identity creates an account and a personal workspace, and returns a token whose session (stored by hash only) expires 30 days later.
- [ ] 4. `SignIn` with a known identity reuses its account and workspace, even when the name has changed.
- [ ] 5. `CurrentAccount(ctx, token)` returns the session's account, and `ErrNoWorkspace` for an unknown or expired token (deleting an expired session).
- [ ] 6. `SignOut(ctx, token)` deletes the session, so `CurrentAccount` then fails. Signing out of an unknown session isn't an error.

Storage:
- [ ] 7. `memory.NewAccountRepository` passes a new `storagetest` account contract (find by issuer and subject, not by name; a missing one is `ErrNotFound`).
- [ ] 8. `memory.NewSessionRepository` passes a new `storagetest` session contract (find by hash, delete).
- [ ] 9. Migration `00004_accounts_and_sessions.sql` and `sqlstore.NewAccountRepository` pass the account contract on SQLite and Postgres.
- [ ] 10. `sqlstore.NewSessionRepository` passes the session contract on SQLite and Postgres.

Plans:
- [ ] 11. The plans file's `newWorkspacePlan` names the starting plan for new workspaces. A missing field, or a plan the file doesn't have, fails at start. The built-in file and the test plans start on `free`.

The server:
- [ ] 12. `httpapi` session middleware: a valid session cookie puts its workspace in the request's context (`workspace.WithID`). No cookie, or an unknown one, leaves none.
- [ ] 13. `problem.go` maps `ErrNoWorkspace` to a 401 problem, and `httpclient` maps it back.
- [ ] 14. The dev sign-in routes: `GET /auth/sign-in` and `GET /auth/dev` show the form, `POST /auth/dev` signs in as the given name (issuer `dev`) and sets the cookie, and `POST /auth/sign-out` ends the session and clears it. Add them to `navigateFallbackDenylist`.
- [ ] 15. With a session cookie, a request that changes something from another site is refused (403).
- [ ] 16. `GET /api/v1/session` in `api/openapi.yaml` (`task gen`), served from `CurrentAccount`, and `httpclient` reads it.

Settings and wiring:
- [ ] 17. The `sign_in` setting (`STARTER_SIGN_IN`): `off` by default, or `dev`; any other value is refused. `starter info` shows it.
- [ ] 18. `bootstrap` refuses sign-in with jsonfile storage at start, naming memory, SQLite, and Postgres. (Agree the message.)
- [ ] 19. The dev sign-in is wired only in builds with the `devsignin` tag; without it, `sign_in = "dev"` is refused at start. `task dev` and `task run:*` builds that need it pass the tag; `task build` doesn't. (Agree the message.)
- [ ] 20. With sign-in on, `bootstrap` wires the hosted server: the `FromContext` resolver, the session middleware, the account and session repositories from its storage, stored plans (`plans.StoredPlanOf` with `newWorkspacePlan`, hints on), and the real `bootstrap.NewBilling`.

Acceptance:
- [ ] 21. The `@sign-in` steps on memory, sqlite, postgres, and http, and the driver tags that leave it out of jsonfile, desktop, mobile, and server. The sign-in scenarios pass. Sample: `task test:init`.
- [ ] 22. `@workspaces` runs on the http driver too, each workspace a signed-in account's. Update the feature's comment and SPEC.md §4.8. Sample: `task test:init`.

The web app:
- [ ] 23. Signed out with sign-in on, the web app shows a sign-in screen linking to `/auth/sign-in`, and no data. (Agree the wording.)
- [ ] 24. Signed in, the web app shows a sign-out button, which ends the session; with sign-in off, there's neither. (Agree the wording.)
- [ ] 25. Browser journey `web/e2e/features/sign-in.feature` (`@wip`): sign in with the dev sign-in, see the items page, sign out, and be asked to sign in again. Its own Playwright project and server with sign-in on.

- [ ] 26. **Docs and finish:** a new ADR 9 (sign-in and sessions: the session cookie and why, sessions in storage, signed out as no workspace, the dev sign-in build tag). Update ADR 5, a new `docs/sign-in.md` guide (how sign-in works, the dev sign-in, sessions; 8b adds providers and the email-and-password note), `docs/configuration.md` (`sign_in`), `docs/billing.md` (stored plans are now switched on), `docs/workspaces.md`, README, SPEC.md §2 and §5, and `.claude/rules/` (`workspaces-and-plans.md`: `ErrNoWorkspace` is mapped and stored plans are on; `testing.md`: the `@sign-in` tag; `server.md`: the `/auth/` routes). Add a template-user step under the TODO.md starter-list item if one arises. Remove the `@wip` tags, run `task check`, `task e2e`, and `task test:init`, delete this file, and mark 8a done in the plan.

## Changes to the plan
