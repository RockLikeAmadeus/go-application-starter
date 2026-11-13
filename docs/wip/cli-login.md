# CLI and TUI login (editions-and-accounts step 9)

- Spec: SPEC.md §4.19 "Sign in from the command line" (and §2's command-line row).
- Scenarios: [`features/command-line-sign-in.feature`](../../features/command-line-sign-in.feature) (http and oidc drivers), [`test/terminal/features/command-line-sign-in.feature`](../../test/terminal/features/command-line-sign-in.feature).
- Plan: step 9 of [`docs/plans/editions-and-accounts.md`](../plans/editions-and-accounts.md). Background: [ADR 9](../adr/0009-sign-in-and-sessions.md), whose "A session cookie" section already says the command line sends a bearer token.

## Decisions (agreed 2026-11-13)
- **Flow: the browser and a loopback port (RFC 8252).** `starter login` listens on `127.0.0.1:<random port>`, opens the browser at the server's `/auth/native` with the port, a `state`, and a PKCE S256 challenge. The server runs its normal sign-in (dev or OIDC, skipped if the browser's cookie is already signed in), then redirects to `http://127.0.0.1:<port>/callback?code=…&state=…`. The CLI checks the state and swaps the code plus its PKCE verifier at `POST /auth/native/token` for a **new session token** (its own session, apart from the browser's). The CLI never talks to the provider. Step 10 (desktop and mobile remote mode) reuses all of it. Not chosen: device code (RFC 8628, a follow-up for SSH) and pasting a token.
- **One-time codes are kept in storage** (`sign_in_codes`: the code's SHA-256 hash, the account, the PKCE challenge, the expiry), so any server copy can redeem them. A code lasts 5 minutes, works once (taken and deleted in one step), and only with the matching verifier. Memory, SQLite, and Postgres only, like sessions.
- **The bearer token is a session token,** looked up as the cookie's is: 30 days, not renewed. The middleware reads `Authorization: Bearer`, then the cookie. `http.CrossOriginProtection` lets requests with no `Sec-Fetch-Site` or `Origin` through, so the CLI passes the cross-origin check unchanged. `POST /auth/sign-out` ends a bearer session too.
- **The native routes live under `/auth/`,** beside the other sign-in routes and outside `api/openapi.yaml` (ADR 9 "Routes outside the API"). The token endpoint answers small JSON, called by hand-written client code. The sign-in routes learn to return to a local path after signing in, refusing anything but a path on this server (no open redirect).
- **The token store: the OS credential store, falling back to a file** (agreed). `zalando/go-keyring` (Keychain, Windows Credential Manager, Secret Service over D-Bus, pure Go) first; where it's unavailable, a JSON file in the settings folder with mode 0600, and `starter login` says so. Tokens are keyed by the server URL, normalised (scheme, host, port, no trailing slash), so the CLI can be signed in to several servers. Tests use go-keyring's `MockInit` and a temp folder.
- **A `credential_store` setting** (`system`, the default: keyring then file; or `file`) lets the terminal journeys, CI, and people who don't want keyring prompts keep the token in the file only, so tests never touch the developer's real keyring. Added under ADR 6's layers like any setting.
- **Login stores a token only** (agreed). It never changes which data the CLI uses: the server still comes from `--server`, `STARTER_SERVER_URL`, or `server_url`, and the settings file is never written (§4.12). `starter login` and `starter logout` without a server are refused. Login to a server whose `GET /api/v1/session` says sign-in is off is refused.
- **`starter logout`** posts sign-out with the bearer token, then deletes the stored token whatever the server answered (unreachable, or session already gone).
- **The browser opener** honours `BROWSER` (as `gh` and `git` do), else the OS default (`xdg-open`, `open`, `start`), and prints the address when opening fails. The terminal journeys set `BROWSER` to a helper that completes the dev sign-in.
- **TUI: signs in and out by itself** (agreed 2026-11-13), sharing the CLI's stored session. On a server with sign-in on and no session (at start, or on a 401 mid-use) it shows a signed-out screen with a sign-in key; the key runs the same browser sign-in, with a waiting screen showing the address, Esc cancelling. Signed in, it shows the name and a sign-out key. No keys on local data or with sign-in off. Which keys, and all wording, are agreed in the cycles. No server picker (agreed: the server stays a setting).
- **Signing in and out live outside `cli` and `tui`** (a client-side sign-in package, named per `naming.md`): check the server signs people in, run the browser sign-in, save the token; end the server session and forget the token. The CLI, the TUI, and later the desktop app call it. The browser opener lives there too, with its output discarded so it can't draw over the TUI.
- **`httpclient` asks a token source on each request,** not a fixed token, so the TUI picks up a new or forgotten session without reconnecting.
- The CLI prints advice to run `starter login` on a 401 (`ErrNoWorkspace` from `httpclient`).
- **Acceptance:** `@command-line-sign-in` runs on the http and oidc drivers only, started hosted. The "command line" there is the real browser-sign-in adapter and `httpclient` with a memory token store, with an opener that drives the harness's cookie-jar browser through the dev sign-in or `oidctest`.
- All user-facing text (CLI output, the loopback page, refusals, the TUI message) is agreed with the user in the cycle that writes it.

## Tasks
1. [x] Domain: a sign-in code's PKCE check: `domain` verifies an S256 verifier against its challenge, refusing a wrong or empty one.
   - Red: `TestSignInCode_VerifiedByTheVerifierBehindItsChallenge` failed on the RFC 7636 pair (stub returned false). Green: `SignInCode.VerifiedBy` in `domain/signincode.go`. Refactor: nothing worth changing.
2. [x] Domain: a sign-in code expires `SignInCodeLifetime` (5 minutes) after it's issued.
   - Red: `TestSignInCode_ExpiresItsLifetimeAfterItsIssued` failed at and after the lifetime (stub never expired). Green: `NewSignInCode` (account, challenge, now), `ExpiresAt`, `IsExpiredAt`, `SignInCodeLifetime` in `domain/signincode.go`. Refactor: nothing worth changing. Later tasks add the code hash to the struct.
3. [x] Ports and memory: `ports.SignInCodeRepository` (save, and take: find and delete in one step, missing after) with a `storagetest` contract suite, passing on memory.
   - Red: `RunSignInCodeRepositoryContract` on memory failed (stub's Take found nothing). Green: `ports.SignInCodeRepository` (Save, Take), `memory.SignInCodeRepository`, `SignInCode.CodeHash` and `HashSignInCode` in `domain`. Refactor: nothing worth changing. Task 4 sets `CodeHash` when issuing.
4. [x] App: `SignInService.IssueSignInCode(ctx, sessionToken, challenge)` stores the code's hash for the session's account and returns the code.
   - Red: `TestIssueSignInCode_KeepsTheCodesHashForTheSessionsAccountAndReturnsTheCode` failed (stub returned no code, nothing kept). Green: `IssueSignInCode` in `app/sign_in_service.go`, via `CurrentAccount`; `NewSignInService` takes a `SignInCodeRepository` (nil in callers that don't issue codes; bootstrap wires it in task 7). Not on `ports.SignInService` yet: the httpapi tasks add it. Refactor: nothing worth changing.
4b. [x] App: `IssueSignInCode` refuses an unknown or expired session (`ErrNoWorkspace`) and keeps no code.
   - No red: `TestIssueSignInCode_RefusesAnUnknownOrExpiredSessionAndKeepsNoCode` passed at once, as task 4's `CurrentAccount` call already refuses both; kept as a guard (it fails when that check is removed). No green. Refactor: nothing worth changing.
5. [x] App: `RedeemSignInCode(ctx, code, verifier)` returns a new session token for the code's account.
   - Red: `TestRedeemSignInCode_GivesANewSessionTokenForTheCodesAccount` failed (stub returned no token, no session). Green: `RedeemSignInCode` takes the code by hash and starts a session. Refactor: extracted `startSession`, shared with `SignIn`. Not on `ports.SignInService` yet (httpapi tasks add it); the verifier isn't checked until task 6.
6. [x] App: redeeming refuses a code used before, a wrong verifier, and an expired code (one test each, table if natural).
   - Red: `TestRedeemSignInCode_RefusesACodeUsedBeforeAWrongVerifierAndAnExpiredCode` (table) failed for the wrong verifier and the expired code (a session was started); the used-before case passed already, as `Take` deletes. Green: `RedeemSignInCode` checks expiry and `VerifiedBy`, refusing with `ErrNotFound` like a missing code. Refactor: nothing worth changing.
7. [ ] sqlstore: migration and repository for sign-in codes; the contract suite passes on SQLite and Postgres; bootstrap wires the repository with the hosted storage.
8. [ ] httpapi: the session middleware accepts `Authorization: Bearer <token>`, ahead of the cookie.
9. [ ] httpapi: `POST /auth/sign-out` ends a bearer session (no cookie involved), and passes the cross-origin check without `Origin`.
10. [ ] httpapi: the sign-in routes (dev and OIDC) return to a `return` path after signing in; anything but a local path is ignored.
11. [ ] httpapi: `GET /auth/native` checks the loopback port, `state`, and challenge, keeps them in a short-lived cookie, and sends the browser to sign in, returning to `/auth/native/finish` (straight there when already signed in). Bad parameters answer 400.
12. [ ] httpapi: `GET /auth/native/finish` issues a code for the browser's session and redirects to `http://127.0.0.1:<port>/callback?code=…&state=…`; no attempt cookie or no session is refused.
13. [ ] httpapi: `POST /auth/native/token` swaps a code and verifier for a session token (JSON); a refused swap answers 400 without saying why.
14. [ ] Browser sign-in adapter (named per `naming.md`, no Cobra or Wails import): listens on `127.0.0.1:0`, makes the state and verifier, builds the `/auth/native` URL, and hands it to an opener.
15. [ ] Browser sign-in adapter: the callback with the right state answers a page saying to return to the terminal, and the adapter swaps the code for the token; a wrong state is refused and the wait goes on; the context ends the wait.
16. [ ] Token store: the file store (JSON in the settings folder, 0600, keyed by the normalised server URL): save, read, delete.
17. [ ] Token store: the keyring store (go-keyring, `MockInit` in tests), same contract as the file store.
18. [ ] Token store: the system store uses the keyring and falls back to the file when the keyring is unavailable, reporting which it used.
19. [ ] Config: the `credential_store` setting (`system` or `file`), with its env variable, in `starter info` and the default settings file's comments.
20. [ ] httpclient: every service sends a bearer token from a token source, asked on each request (one client option); no token sends no header.
21. [ ] bootstrap: `NewRemote` gives `httpclient` a token source reading the server's token from the configured store.
22. [ ] Client sign-in: the opener honours `BROWSER`, else the OS default, discards the program's output, and reports a failure so the caller shows the address.
23. [ ] Client sign-in: signing in refuses a server whose sign-in is off (`GET /api/v1/session`).
24. [ ] Client sign-in: signing in runs the browser sign-in, saves the token, and returns the account's name and which store held it.
25. [ ] Client sign-in: signing out ends the server session and forgets the token, even when the server refuses or can't be reached.
26. [ ] CLI: `starter login` refuses without a server, and otherwise signs in and says who is signed in (and when the file fallback held it), or prints the address when the browser didn't open.
27. [ ] CLI: `starter logout` refuses without a server, and otherwise signs out.
28. [ ] CLI: `starter info` with a server says who is signed in there, or that no one is.
29. [ ] CLI: a request refused for want of sign-in tells the user to run `starter login`.
30. [ ] TUI: on a server with sign-in on and no session, it shows the signed-out screen with the sign-in key; on local data or with sign-in off, no sign-in or sign-out key.
31. [ ] TUI: signed in, it shows who is signed in and the sign-out key.
32. [ ] TUI: the sign-in key starts the browser sign-in and shows the waiting screen with the address; Esc cancels and returns to the signed-out screen.
33. [ ] TUI: a finished sign-in loads the workspace's screen on the new session, with no restart; a failed one shows the error on the signed-out screen.
34. [ ] TUI: the sign-out key signs out and shows the signed-out screen.
35. [ ] TUI: a request refused for want of sign-in mid-use shows the signed-out screen.
36. [ ] Acceptance: route `@command-line-sign-in` to the http and oidc drivers (hosted), and add the command-line client and its steps; the Godog scenarios pass.
37. [ ] Terminal journeys: the server built with `devsignin`, a `BROWSER` helper that completes the dev sign-in, and the steps for the command line and the terminal app; the journeys pass.
38. [ ] Docs and finish: `docs/sign-in.md` (the command line's and the terminal app's sign-in, with a sequence diagram), ADR 9 (bearer tokens, the native routes, codes, the token store), `docs/configuration.md` (`credential_store`), `.claude/rules/server.md` (the `/auth/native` routes), `.claude/rules/tui.md` (the signed-out screen), README and CLAUDE.md (CLI and TUI hosted is no longer planned), `.env.example` if it lists settings; device code over SSH as a follow-up in the plan; remove the `@wip` tags, run `task test:init` and `task check`, mark step 9 done in the plan, and delete this file.

## Changes to the plan
- 2026-11-13: the terminal app signs in and out by itself (the user's request), rather than sending people to `starter login`. Signing in and out moved out of the CLI into a shared client-side package, `httpclient` asks a token source per request, and tasks 30–35 and two terminal journeys were added. No server picker in the TUI (agreed).
- 2026-11-13: task 4 split: 4 issues a code, 4b refuses an unknown or expired session (one test each).
