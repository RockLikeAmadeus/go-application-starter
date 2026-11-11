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
- **CSRF:** with sign-in on, a request that changes something (not `GET`, `HEAD`, or `OPTIONS`) is refused (403, `cross_site`) unless it comes from the same origin (`Sec-Fetch-Site: same-origin` or `none`, or, without that header, an `Origin` that is missing or matches the host), cookie or not, so the sign-in form is covered too. Go's `http.CrossOriginProtection` wraps the whole handler. `SameSite=Lax` covers most cases already; this closes the rest. With sign-in off, nothing changes. The detail is the agreed "The server refused this request because it did not come from the application's own pages. Please try again from the application."
- **The web app learns its state from `GET /api/v1/session`** (in the OpenAPI file): `{"isSignInOn": …, "name": …}`, always 200. Signed out with sign-in on, it shows a sign-in screen linking to `/auth/sign-in`. Signed in, it shows a sign-out button. With sign-in off, neither. Their wording is agreed with the user in tasks 23 and 24.
- **Stored plans switch on with sign-in** (`docs/billing.md`, "Switching on stored plans"). The plans file names the starting plan for new workspaces in a new top-level field, `newWorkspacePlan`, which must be one of its plans. The built-in file and the test plans start on `free`. Locked features name their plans (hints on). `bootstrap.NewBilling` returns the real service when `sign_in` is on, so `starter admin set-plan` works beside a hosted server's data. The local edition keeps refusing.
- **Acceptance drivers:** `@sign-in` runs on memory, sqlite, postgres, and http. memory, sqlite, and postgres call `SignInService` directly, as the dev sign-in would. http signs in through `/auth/dev`, with a cookie jar per account. jsonfile, desktop, mobile, and server leave `@sign-in` out. In task 22, http also runs `@workspaces`, where workspace "Alpha" is the personal workspace of an account named "Alpha", and "someone with no workspace" hasn't signed in.
- **The browser journey** needs a server with sign-in on, so it gets its own Playwright project and server (built with `devsignin`, `STARTER_SIGN_IN=dev`), leaving the other journeys on a server without sign-in. It's written in task 25, not in the spec commit, as an undefined Playwright step breaks the whole browser suite.
- **Sample boundary:** everything here is template code. The sample part is the one scenario where an account adds an item, and its step. Run `task test:init` in tasks 21, 22, and 26.
- **SPEC.md §6 keeps "Authentication and authorization"** until 8b, when real sign-in exists. Authorization (roles) stays out of scope until teams.

## Tasks

Each is one `/red` → `/green` → `/refactor` cycle.

The core:
- [x] 1. `domain.NewIdentity(issuer, subject, name)` refuses a blank issuer or subject with a `*domain.ValidationError`. An `Account` has an ID, an identity, and a workspace.
  - Red: `TestNewIdentity_RefusesABlankIssuerOrSubject` (empty or whitespace issuer or subject) failed: nil error, want a `*ValidationError`.
  - Green: `NewIdentity` refuses a whitespace-only issuer or subject with the agreed messages ("The sign-in issuer cannot be empty.", "The sign-in subject cannot be empty."), and stores both untrimmed, as subjects are opaque.
  - Refactor: nothing worth changing. The code is two guard clauses.
- [x] 2. `domain.Session` knows when it's expired: `IsExpiredAt(now)` is true at or after its `ExpiresAt`.
  - Red: `TestSession_IsExpiredAtOrAfterItsExpiry` (before, at, and after) failed: `IsExpiredAt` = false at and after `ExpiresAt`, want true.
  - Green: `domain.Session` (only `ExpiresAt` for now; task 3 adds the rest) and `IsExpiredAt(now)` = `!now.Before(ExpiresAt)`.
  - Refactor: nothing worth changing. The method is one comparison; the session's other fields wait for task 3, which needs them.
- [x] 3. `ports.SignInService`, `AccountRepository`, `SessionRepository`, and `TokenGenerator`, with `testsupport` fakes where needed. `app.SignInService.SignIn` with a new identity creates an account and a personal workspace, and returns a token whose session (stored by hash only) expires 30 days later.
  - Red: `TestSignIn_WithANewIdentityCreatesAnAccountWorkspaceAndThirtyDaySession` failed: empty token, no account, and no session stored.
  - Green: `SignIn` saves an account (ID, then workspace ID, from `IDGenerator`) and a session keyed by the token's hex SHA-256, expiring 30 days from now. Ports hold only the methods a test uses so far.
  - Refactor: the lifetime and hashing moved into `domain` (`SessionLifetime`, `NewSession`, `HashSessionToken`); save errors are wrapped. The test keeps its own SHA-256 so it pins the hash independently.
- [x] 4. `SignIn` with a known identity reuses its account and workspace, even when the name has changed.
  - Red: `TestSignIn_WithAKnownIdentityReusesItsAccountEvenWhenTheNameHasChanged` failed: a second account `id-1` was saved, and the session pointed to it, not `account-1`.
  - Green: `SignIn` finds the account by identity (new `AccountRepository.FindByIdentity`, issuer and subject only) and creates one only on `ErrNotFound`; other find errors are returned, wrapped.
  - Red (review fix): `TestSignIn_WhenTheAccountCantBeLookedUpIsRefusedAndSavesNothing`, checked against the code without the find-error branch, failed: nil error, want the lookup's error. Green: the branch already in place.
  - Refactor: the find-or-create moved into `accountFor`, so `SignIn` reads as account, then session. Whether a changed name is stored is left open: no task asks for it.
- [x] 5. `CurrentAccount(ctx, token)` returns the session's account.
  - Red: `TestCurrentAccount_WithASignedInTokenReturnsTheSessionsAccount` failed: got the zero account, want `account-2` (Grace's).
  - Green: `CurrentAccount` finds the session by the token's hash (new `SessionRepository.FindByHash`) and returns its account (new `AccountRepository.Find`, by ID); both fakes implement them.
  - Refactor: the find errors are wrapped, as `SignIn`'s are. Mapping a missing session to `ErrNoWorkspace` is left to 5b, which demands it.
- [x] 5b. `CurrentAccount` returns `ErrNoWorkspace` for an unknown token.
  - Red: `TestCurrentAccount_WithAnUnknownTokenHasNoWorkspace` failed: got the wrapped `ErrNotFound` ("find session: …"), want `ErrNoWorkspace`.
  - Green: `CurrentAccount` returns `ErrNoWorkspace` when `FindByHash` gives `ErrNotFound`; other errors are still wrapped.
  - Refactor: the doc comment says an unknown token is `ErrNoWorkspace`. Nothing else worth changing.
- [x] 5c. `CurrentAccount` returns `ErrNoWorkspace` for an expired token, and deletes its session.
  - Red: `TestCurrentAccount_WithAnExpiredTokenHasNoWorkspaceAndDeletesTheSession` (clock at `ExpiresAt`) failed: nil error, want `ErrNoWorkspace`.
  - Green: `CurrentAccount` deletes an expired session (new `SessionRepository.Delete`, in the fake too) and returns `ErrNoWorkspace`; the delete's error is ignored.
  - Refactor: comments only: why the delete's error is ignored, and expiry in the `CurrentAccount` and port doc comments.
- [x] 6. `SignOut(ctx, token)` deletes the session, so `CurrentAccount` then fails.
  - Red: `TestSignOut_DeletesTheSessionSoItsTokenIsSignedOut` failed: the session was still stored, and `CurrentAccount` returned no error, want `ErrNoWorkspace`.
  - Green: `SignOut` (new on `ports.SignInService`) deletes the session by the token's hash.
  - Refactor: the delete's error is wrapped, as the other calls' are. The repeated Ada setup in three tests is left, as each test reads on its own.
- [x] 6b. Signing out of an unknown session isn't an error.
  - Red: `TestSignOut_OfAnUnknownSessionIsNotAnError` failed: `delete session: The requested record could not be found.` (`SessionRepository.Delete` now documents `ErrNotFound` for a missing session, as `ItemRepository` does, and `testsupport.StoredSessions` returns it.)
  - Green: `SignOut` treats `ErrNotFound` from the delete as already signed out.
  - Refactor: nothing worth changing; the guard clause and the test are already minimal.

Storage:
- [x] 7. `memory.NewAccountRepository` passes a new `storagetest` account contract: `FindByIdentity` with no saved account is `ErrNotFound`.
  - Red: `TestAccountRepositoryContract/FindByIdentity_with_no_saved_account_returns_ErrNotFound` failed: `err = <nil>, want ErrNotFound`.
  - Green: the new `memory.AccountRepository`'s `FindByIdentity` always returns `ErrNotFound`; `Find` and `Save` are stubs until 7d–7f.
  - Refactor: nothing worth changing; the stubs are left for 7b–7f to drive.
- [x] 7b. Account contract: `FindByIdentity` returns the saved account with the identity's issuer and subject, even under another name.
  - Red: `TestAccountRepositoryContract/FindByIdentity_returns_the_account_with_the_issuer_and_subject,_whatever_the_name` failed: `FindByIdentity: The requested record could not be found.`
  - Green: `memory.AccountRepository` keeps accounts by ID (`Save`), and `FindByIdentity` returns the one with the issuer and subject.
  - Refactor: nothing worth changing; an `Identity` matching method in `domain` was left, as the SQL repository will match in its query and no second caller needs it.
- [x] 7c. Account contract: `FindByIdentity` doesn't match an account with the same name but another issuer or subject.
  - Red: none. `TestAccountRepositoryContract/FindByIdentity_doesn't_match_an_account_with_the_same_name_but_another_issuer_or_subject` passed at once, as 7b's green already matches on issuer and subject; kept as a guard (it fails on a repository that matches by name).
  - Green: skipped, as no production code was demanded.
  - Refactor: nothing worth changing; the two contract cases share a saved account, but one setup helper for two cases isn't clearer yet.
- [x] 7d. Account contract: `Find` returns the saved account by its ID.
  - Red: `TestAccountRepositoryContract/Find_returns_the_saved_account_by_its_ID` failed: `got {ID: …}, want {ID:account-2 …}` (the stub returned an empty account).
  - Green: `memory.AccountRepository.Find` returns the account kept under the ID (an unknown ID is left to 7e).
  - Refactor: the contract's repeated Ada account is now one `ada` fixture shared by its cases.
- [x] 7e. Account contract: `Find` of an unknown ID is `ErrNotFound`.
  - Red: `TestAccountRepositoryContract/Find_of_an_unknown_ID_returns_ErrNotFound` failed: `err = <nil>, want ErrNotFound`.
  - Green: `memory.AccountRepository.Find` returns `ErrNotFound` when no account is kept under the ID.
  - Refactor: nothing worth changing; the two-line lookup matches the other memory repositories.
- [x] 7f. Account contract: `Save` replaces the account with the same ID.
  - Red: none. `TestAccountRepositoryContract/Save_replaces_the_account_with_the_same_ID` passed at once, as the memory map is keyed by ID; kept as a guard (it fails on a `Save` that keeps the first account).
  - Green: skipped, as no production code was demanded.
  - Refactor: nothing worth changing; two contract cases now save a list of accounts in a loop, but a helper for two isn't clearer yet.
- [x] 8. `memory.NewSessionRepository` passes a new `storagetest` session contract: `FindByHash` with no saved session is `ErrNotFound`.
  - Red: `TestSessionRepositoryContract/FindByHash_with_no_saved_session_returns_ErrNotFound` failed: `err = <nil>, want ErrNotFound`.
  - Green: the new `memory.SessionRepository`'s `FindByHash` always returns `ErrNotFound`; `Save` and `Delete` are stubs until 8b–8e.
  - Refactor: nothing worth changing; the stubs are left for 8b–8e to drive.
- [x] 8b. Session contract: `FindByHash` returns the saved session with the hash.
  - Red: `TestSessionRepositoryContract/FindByHash_returns_the_saved_session_with_the_hash` failed: `FindByHash: The requested record could not be found.`
  - Green: `memory.SessionRepository` keeps sessions in a map by token hash (`Save`), and `FindByHash` returns the one kept under the hash.
  - Refactor: the session comparison is now `assertSessionEqual` (expiry compared with `Equal`, like `assertItemEqual`), ready for 8c.
- [x] 8c. Session contract: `Save` replaces the session with the same token hash.
  - Red: none. `TestSessionRepositoryContract/Save_replaces_the_session_with_the_same_token_hash` passed at once, as the memory map is keyed by hash; kept as a guard (it fails on a `Save` that keeps the first session).
  - Green: skipped, as no production code was demanded.
  - Refactor: nothing worth changing; the save loop now appears twice in the contract, as in the account contract, but a helper for two isn't clearer yet.
- [x] 8d. Session contract: `Delete` removes the session, so `FindByHash` then returns `ErrNotFound`.
  - Red: `TestSessionRepositoryContract/Delete_removes_the_session,_so_FindByHash_returns_ErrNotFound` failed: `FindByHash after Delete: err = <nil>, want ErrNotFound`.
  - Green: `memory.SessionRepository.Delete` removes the session kept under the hash.
  - Refactor: nothing worth changing; `Delete` matches `memory.ItemRepository.Delete`, less its missing-record check, which 8e drives.
- [x] 8e. Session contract: `Delete` of a missing session is `ErrNotFound` (set in task 6b).
  - Red: `TestSessionRepositoryContract/Delete_of_a_missing_session_returns_ErrNotFound` failed: `err = <nil>, want ErrNotFound`.
  - Green: `memory.SessionRepository.Delete` returns `ErrNotFound` when no session is kept under the hash.
  - Refactor: nothing worth changing; `Delete` now matches `memory.ItemRepository.Delete`.
- [x] 9. Migration `00004_accounts_and_sessions.sql` and `sqlstore.NewAccountRepository` pass the account contract on SQLite.
  - Red: `TestAccountRepositoryContractOnSQLite` failed in all six cases, as the stub returns zero values and no errors (e.g. `err = <nil>, want ErrNotFound`).
  - Green: `00004_accounts_and_sessions.sql` creates `accounts` (only; task 10 adds `sessions` to the same unreleased file), and `sqlstore.AccountRepository` selects by issuer and subject or by ID and upserts on ID.
  - Refactor: `accountColumns` and a `findOne` helper share the select and scan; an index on `(issuer, subject)` for sign-in's lookup, left non-unique as no test demands refusing a second account with one identity.
- [x] 9b. `sqlstore.NewAccountRepository` passes the account contract on Postgres.
  - Red: none. `TestAccountRepositoryContractOnPostgres` passed at once, as task 9's SQL and migration are shared; kept as a guard (it fails on a broken `Save`).
  - Green: skipped, as no production code was demanded.
  - Refactor: nothing worth changing; the test matches `TestWorkspacePlanRepositoryContractOnPostgres`.
- [x] 10. The migration adds `sessions`, and `sqlstore.NewSessionRepository` passes the session contract on SQLite.
  - Red: `TestSessionRepositoryContractOnSQLite` failed in all five cases, as the stub returns zero values and no errors (e.g. `err = <nil>, want ErrNotFound`).
  - Green: `00004_accounts_and_sessions.sql` adds `sessions` (keyed by token hash, expiry as UTC text), and `sqlstore.SessionRepository` selects by hash, upserts on hash, and deletes with `ErrNotFound` when no row goes.
  - Refactor: `timeFormat`, `formatTime`, and a new `parseTime` moved from the sample `item_repository.go` to the template `times.go`, as the session repository needs them after `task init`. `task test:init` also found the session contract (since 8b) using the sample contract's `baseTime`, now in the template `storagetest/times.go`.
- [x] 10b. `sqlstore.NewSessionRepository` passes the session contract on Postgres.
  - Red: none. `TestSessionRepositoryContractOnPostgres` passed at once, as task 10's SQL and migration are shared; kept as a guard (it fails on a broken `Save`).
  - Green: skipped, as no production code was demanded.
  - Refactor: nothing worth changing; the test matches `TestAccountRepositoryContractOnPostgres`.

Plans:
- [x] 11. `plans.Parse` reads the plans file's `newWorkspacePlan`, the starting plan for new workspaces.
  - Red: `TestParse_ReadsTheNewWorkspacePlan` failed with `unknown field "newWorkspacePlan"`. To compile, `Parse` now returns a `plans.File` (`Catalog`, `NewWorkspacePlan`); `Default` and `Load` still return its `Catalog`.
  - Green: the plans document reads `newWorkspacePlan` into `File.NewWorkspacePlan`.
  - Refactor: `Parse`'s doc comment shows and explains the field. Left: `Default` and `Load` keep returning a `Catalog` until a test needs the field from them (11d, task 20).
- [x] 11b. A plans file without `newWorkspacePlan` is refused, so it fails at start. The test plans start on `free`.
  - Red: `TestParse_RefusesABadPlansFile/no_newWorkspacePlan` failed with "Parse succeeded, want an error". The valid fixtures in the plans, bootstrap, cli, and terminal tests now name a plan they have (`free`, or `local`).
  - Green: `Parse` refuses a file without `newWorkspacePlan` (checked after the plans, so other refusals keep their messages); `plans.jsonc` starts new workspaces on `free`, so the acceptance test plans do too.
  - Refactor: `Parse`'s doc comment says the field is required. Left: the `docs/features.md` example is a fragment, so it needs no field; task 26 documents it.
- [x] 11c. A `newWorkspacePlan` that isn't one of the file's plans is refused.
  - Red: `TestParse_RefusesABadPlansFile/a_newWorkspacePlan_that_isn't_a_plan` failed with "Parse succeeded, want an error".
  - Green: `Parse` refuses a `newWorkspacePlan` missing from the catalog, naming it.
  - Refactor: `Parse`'s doc comment says the plan must be one of the file's. Nothing else worth changing.
- [x] 11d. The built-in file starts new workspaces on `free`.
  - Red: none. `TestDefault_StartsNewWorkspacesOnFree` passed straight away, as 11b already set `plans.jsonc` to `free`; kept as a guard (it fails when the file says `local`). To compile, `Default` now returns the `plans.File`.
  - Green: skipped, as no production code was demanded.
  - Refactor: nothing worth changing. Left: `Load` keeps returning a `Catalog` until task 20 needs the field from it.

The server:
- [x] 12. `httpapi` session middleware: a valid session cookie puts its workspace in the request's context (`workspace.WithID`).
  - Red: `TestASessionCookiePutsItsAccountsWorkspaceInTheRequest` failed with workspace "none", want "ws-a". To compile: `Options.SignIn` and `httpapi.SessionCookie` (`session`).
  - Green: `NewHandler` wraps the mux in `sessionMiddleware` when `Options.SignIn` is set; a cookie whose `CurrentAccount` answers puts its workspace in the context.
  - Refactor: the middleware and `SessionCookie` moved to `httpapi/session.go`, with a comment on failing closed. Left: any `CurrentAccount` error (even a storage failure) runs the request signed out, which fails closed; 12b pins the signed-out cases.
- [x] 12b. No session cookie leaves no workspace in the request's context.
  - Red: none. `TestNoSessionCookieLeavesNoWorkspaceInTheRequest` passed straight away, as task 12's middleware only sets a workspace from a cookie; kept as a guard (it fails when the middleware sets one without a cookie).
  - Green: skipped, as no production code was demanded.
  - Refactor: nothing worth changing; the test reuses `requestWorkspace`.
- [x] 12c. An unknown session cookie leaves no workspace in the request's context.
  - Red: none. `TestAnUnknownSessionCookieLeavesNoWorkspaceInTheRequest` passed straight away, as task 12's middleware sets a workspace only when `CurrentAccount` answers; kept as a guard (a middleware that sets one despite the error fails it with workspace "leak").
  - Green: skipped, as no production code was demanded.
  - Refactor: nothing worth changing; the test reuses `requestWorkspace`, and folding the two signed-out tests into one table would hide which case failed for little gain.
- [x] 13. `problem.go` maps `ErrNoWorkspace` to a 401 problem (type `no_workspace`, detail `err.Error()`). Its message is now the agreed "Please sign in to continue."
  - Red: `TestNoWorkspaceIsAnUnauthorizedProblemAskingToSignIn` failed with status 500, want 401.
  - Green: `problemFor` maps `ErrNoWorkspace` to 401 `no_workspace` (`ProblemNoWorkspace`) with `err.Error()`, and `ErrNoWorkspace`'s message is "Please sign in to continue."
  - Refactor: nothing worth changing; the new case follows `ErrNotFound`'s, and the test follows the limit test's shape. `workspaces-and-plans.md` still says it isn't mapped: task 26 updates it, after 13b.
- [x] 13b. `httpclient` maps the `no_workspace` problem back to `ErrNoWorkspace`.
  - Red: `TestNoWorkspaceComesBackAsErrNoWorkspace` failed: the error read as "The server returned an error (401): …", not `ErrNoWorkspace`.
  - Green: `toError` maps `no_workspace` to `domain.ErrNoWorkspace`.
  - Refactor: nothing worth changing; the case and test follow their neighbours. `workspaces-and-plans.md` still says it isn't mapped: task 26 updates it.
- [x] 14. `POST /auth/dev` signs in as the given name (issuer `dev`, the name as the subject), sets the session cookie (`HttpOnly`, `SameSite=Lax`, path `/`), and goes to `/`. On with `httpapi.Options.IsDevSignInOn`.
  - Red: `TestTheDevSignInSignsInAsTheGivenNameAndSetsTheSessionCookie` failed with status 200 (the SPA's hint page), want 303 to `/`. To compile: `Options.IsDevSignInOn`.
  - Green: with `IsDevSignInOn`, `POST /auth/dev` signs in as the form's `name` (issuer `dev`), sets the cookie, and redirects (303) to `/`; a refused sign-in is a problem (`problemFor`), pinned by 14g.
  - Refactor: the handler moved to `httpapi/dev_sign_in.go` (`devSignIn`), the issuer is `httpapi.DevIssuer`, and `setSessionCookie` (in `session.go`) sets the cookie, ready for 14b, 14c, and 14f.
- [x] 14b. The session cookie lasts as long as the session (`domain.SessionLifetime`), so it outlives the browser.
  - Red: `TestTheSessionCookieLastsAsLongAsTheSession` failed with Max-Age 0, want 2592000.
  - Green: `setSessionCookie` sets `MaxAge` to `domain.SessionLifetime` in seconds.
  - Refactor: the dev sign-in tests share `postDevSignIn`, ready for 14c and 14g; `setSessionCookie` uses `MaxAge` (relative), so it needs no clock.
- [x] 14c. Over HTTPS, the session cookie is `Secure`.
  - Red: `TestTheSessionCookieIsSecureOnlyOverHTTPS` failed: posting to `https://…` gave Secure = false, want true (plain HTTP pinned to false, so `Secure: true` alone can't pass).
  - Green: `setSessionCookie` takes the request and sets `Secure` when `r.TLS != nil`.
  - Refactor: the test posts through `postDevSignInTo` (a path or full URL), which `postDevSignIn` now calls. A TLS-terminating proxy (`X-Forwarded-Proto`) isn't covered: not in the task.
- [x] 14d. `GET /auth/dev` shows the form asking for a name, posting to `POST /auth/dev`.
  - Red: `TestTheDevSignInFormAsksForANameAndPostsToTheDevSignIn` failed: `GET /auth/dev` served the SPA's hint page, with no form.
  - Green: with `IsDevSignInOn`, `GET /auth/dev` serves a plain HTML form (a `name` field and a submit button) posting to `/auth/dev`.
  - Refactor: the page is the constant `devSignInPage`, and `devSignInForm` is a plain handler function. Adding `/auth/` to `navigateFallbackDenylist` is left to 14e.
- [x] 14e. `GET /auth/sign-in` goes to the configured sign-in adapter (`/auth/dev`). Add `/auth/` to `navigateFallbackDenylist`.
  - Red: `TestSignInGoesToTheDevSignInWhenItIsOn` failed: `GET /auth/sign-in` answered 200 (the SPA's hint page), want 302 to `/auth/dev`.
  - Green: with `IsDevSignInOn`, `GET /auth/sign-in` redirects (302) to `/auth/dev`; `/auth/` is in `navigateFallbackDenylist`.
  - Refactor: the redirect is `signInGoesTo(adapterPath)` in `sign_in.go`, the route for any sign-in adapter, ready for OIDC in 8b.
- [x] 14f. `POST /auth/sign-out` ends the session, clears the cookie, and goes to `/`.
  - Red: `TestSigningOutEndsTheSessionClearsTheCookieAndGoesToTheWebApp` failed with status 200 (the SPA's hint page), want 303 to `/`.
  - Green: with `SignIn` set (any adapter), `POST /auth/sign-out` signs out the cookie's token, clears the cookie (Max-Age < 0), and redirects (303) to `/`; a failed sign-out is a problem (`problemFor`), pinned by 14h.
  - Refactor: clearing the cookie is `clearSessionCookie` in `session.go`, beside `setSessionCookie`; `signOut` sits in `sign_in.go`, as it serves every adapter.
- [x] 14h. A sign-out that fails (storage can't be reached) is a problem, and keeps the cookie.
  - Red: none. `TestASignOutThatFailsIsAProblemAndKeepsTheCookie` passed at once, as 14f's green already returns `problemFor`'s 500 before clearing the cookie; kept as a guard (clearing the cookie first turns it red).
  - Refactor: the sign-out tests share `postSignOut`; `signInRecorder` fails sign-outs with `signOutErr`.
- [x] 14g. A refused dev sign-in (a blank name) is a 422 `validation` problem, and sets no cookie.
  - Red: `TestABlankDevSignInNameIsAValidationProblemAndSetsNoCookie` failed with status 303 (signed in with a blank subject), want 422.
  - Green: `devSignIn` builds the identity with `domain.NewIdentity`, and answers its error with `problemFor` before signing in.
  - Refactor: nothing worth changing; the two early returns in `devSignIn` read in the order the checks happen, and merging them would hide it.
- [x] 15. With sign-in on, a request that changes something from another site (`Sec-Fetch-Site: cross-site`) is refused (403, `cross_site`).
  - Red: `TestASignOutFromAnotherSiteIsRefused` failed with status 303 (signed out), want 403. To compile: `ProblemCrossSite`.
  - Green: with `SignIn` set, `NewHandler` wraps the handler in `http.CrossOriginProtection`, whose deny handler writes the 403 `cross_site` problem with the agreed detail.
  - Refactor: the wrapping moved to `refuseCrossSite` in `httpapi/cross_site.go`, with a comment on what it closes beyond `SameSite=Lax`.
- [x] 15b. A request that changes something with an `Origin` from another host (no `Sec-Fetch-Site`) is refused.
  - Red: none. `TestASignOutWithAnOriginFromAnotherHostIsRefused` passed straight away, since `http.CrossOriginProtection` already checks `Origin`; kept as a guard (with the server's own origin it fails, 303). Green skipped.
  - Refactor: the cross-site tests share `signOutWithHeader` and `wantCrossSiteProblem`, ready for 15c.
- [x] 15c. A same-origin request that changes something, with the session cookie, still goes through.
  - Red: none. `TestASameOriginSignOutStillGoesThrough` passed straight away, as `http.CrossOriginProtection` lets `Sec-Fetch-Site: same-origin` through; kept as a guard (with `cross-site` it fails, 403, as task 15 shows). Green skipped.
  - Refactor: nothing worth changing; the test reuses `signOutWithHeader`, and its assertions match the sign-out test's.
- [x] 16. `GET /api/v1/session` in `api/openapi.yaml` (`task gen`), served from `CurrentAccount`: signed in, sign-in is on and it names the account.
  - Red: `TestTheSessionEndpointNamesTheSignedInAccount` failed, 404 `no such endpoint`.
  - Green: `Session` schema and `getSession` in `api/openapi.yaml` (`task gen`); the session middleware also puts the account in the context, and `GetSession` answers sign-in on (hard-coded) with its name.
  - Refactor: `withAccount` and `accountFrom` keep the context's account in one place; the hard-coded `IsSignInOn` and the empty name when signed out are left for 16b and 16c.
- [ ] 16b. Signed out with sign-in on, the session endpoint says sign-in is on, with no name.
- [ ] 16c. With sign-in off, the session endpoint says sign-in is off.
- [ ] 16d. `httpclient` reads the session endpoint.

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

- 2026-11-10, task 5: split into 5 (a valid token's account), 5b (an unknown token), and 5c (an expired token, deleted), as each needs its own test.
- 2026-11-10, task 6: split into 6 (signing out deletes the session) and 6b (signing out of an unknown session isn't an error), as each needs its own test.
- 2026-11-10, task 7: split into 7 (no saved account is `ErrNotFound`), 7b (found by issuer and subject, whatever the name), 7c (not matched by name), 7d (`Find` by ID), 7e (`Find` of an unknown ID), and 7f (`Save` replaces), as each needs its own test.
- 2026-11-10, task 8: split into 8 (no saved session is `ErrNotFound`), 8b (found by hash), 8c (`Save` replaces), 8d (`Delete` removes), and 8e (`Delete` of a missing session is `ErrNotFound`), as each needs its own test.
- 2026-11-10, task 9: split into 9 (the account contract on SQLite) and 9b (on Postgres), as each needs its own test.
- 2026-11-10, task 10: split into 10 (the session contract on SQLite) and 10b (on Postgres), as each needs its own test.
- 2026-11-10, task 11: split into 11 (`Parse` reads `newWorkspacePlan`), 11b (a missing field is refused), 11c (an unknown plan is refused), and 11d (the built-in file starts on `free`), as each needs its own test.
- 2026-11-10, task 12: split into 12 (a valid session cookie puts its workspace in the context) and 12b (no cookie, or an unknown one, leaves none), as each needs its own test.
- 2026-11-10, task 12b: split into 12b (no cookie leaves no workspace) and 12c (an unknown cookie leaves none), as each needs its own test.
- 2026-11-11, task 13: split into 13 (the server's 401 problem) and 13b (`httpclient` maps it back), as each needs its own test. The user agreed `ErrNoWorkspace`'s message, "Please sign in to continue.", which is also the problem's detail.
- 2026-11-11, task 14: split into 14 (`POST /auth/dev` signs in and sets the cookie), 14b (the cookie lasts as long as the session), 14c (`Secure` over HTTPS), 14d (`GET /auth/dev` shows the form), 14e (`GET /auth/sign-in` goes to it, and the denylist), 14f (signing out), and 14g (a refused dev sign-in, added at green, where the code needed an error path), as each needs its own test.
- 2026-11-11, task 14f: added 14h (a failed sign-out is a problem and keeps the cookie), at green, where the code needed an error path, as task 14 did for 14g.
- 2026-11-11, task 14g: a blank name is a 422 `validation` problem, not 400 (agreed with the user): `devSignIn` builds the identity with `domain.NewIdentity`, and `problemFor` maps its `ValidationError` (field `subject`, the existing message), as for any other invalid input.
- 2026-11-11, task 15: with sign-in on, the cross-origin check covers every request that changes something, cookie or not (agreed with the user), using `http.CrossOriginProtection`; the Decisions entry is updated. Split into 15 (`Sec-Fetch-Site` from another site), 15b (`Origin` from another host), and 15c (same-origin still goes through), as each needs its own test. The user agreed the problem's detail.
- 2026-11-11, task 16: split into 16 (signed in, it names the account), 16b (signed out), 16c (sign-in off), and 16d (`httpclient` reads it), as each needs its own test.
