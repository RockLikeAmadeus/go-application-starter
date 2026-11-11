# 9. Sign-in and sessions: a session cookie, sessions in storage

- Status: accepted. The dev sign-in is built (step 8a of [`docs/plans/editions-and-accounts.md`](../plans/editions-and-accounts.md)). The OIDC adapter is planned (step 8b).
- Date: 2026-11-11

## Context

The hosted edition (ADR 5) serves many people from one server, and each of them must see only their own workspace. So the server has to know who sent each request. It needs three things for that:

- a way for someone to prove who they are, which ADR 5 leaves to a sign-in provider through OpenID Connect (OIDC), so the product stores no passwords;
- a way for the browser to show, on every later request, that it has already signed in;
- a way to end that, when someone signs out or a session gets old.

The template also needs to test all of this, in CI and on a developer's computer, without a real sign-in provider.

## Decision

### The seam: an identity, and a driving port

- A **sign-in adapter** works out who someone is, as a `domain.Identity`: the issuer (who vouches for them), the subject (the issuer's ID for them, which never changes), and a display name.
- The adapter then calls the driving port `ports.SignInService`. `SignIn(ctx, identity)` returns a session token. `CurrentAccount(ctx, token)` returns the session's account. `SignOut(ctx, token)` ends the session.
- The core never sees how someone proved who they are. So the dev sign-in now, and OIDC in step 8b, are both adapters behind the same port, and the core doesn't change between them.
- **An account is found by its identity's issuer and subject, never by its name,** because names change and two people can share one. One identity has at most one account. SQLite and Postgres enforce this with a unique index, and `SignIn` looks the account up again when two first sign-ins race, so both share one account.
- The first sign-in creates the account and its personal workspace. Both IDs come from `ports.IDGenerator`. No plan is stored for the new workspace: it is on the plans file's `newWorkspacePlan` until an admin changes it.

### A session cookie, not tokens in the browser

- The server gives the browser a cookie named `session`, holding the session's token. The cookie is `HttpOnly`, so page scripts can't read it. It is `SameSite=Lax`, so the browser leaves it off most requests that other sites start. It is `Secure` over HTTPS, so it never travels unencrypted. It lasts as long as the session, so it outlives the browser being closed.
- The web app holds no tokens at all. It learns whether someone is signed in from `GET /api/v1/session`.
- This is the "backend for frontend" pattern, which the IETF's draft on [OAuth for browser-based apps](https://datatracker.ietf.org/doc/draft-ietf-oauth-browser-based-apps/) ranks the most secure for browsers. A script injected into the page (cross-site scripting) can't steal a token it can't read.
- The command line (step 9) and the desktop and mobile apps in remote mode (step 10) will send a bearer token in a header instead, since their requests don't come from a browser. The middleware will then accept both.

### Sessions are kept in the app's storage

- Sessions and accounts are stored through `ports.SessionRepository` and `ports.AccountRepository`, beside the rest of the data. So every server copy on one database knows every session, and signing out ends a session at once, everywhere.
- **Storage keeps only the token's SHA-256 hash.** The token is 32 random bytes from `ports.TokenGenerator` (`crypto/rand`). So a leaked copy of the database holds no session anyone can use.
- A session lasts 30 days from sign-in (`domain.SessionLifetime`), timed by `ports.Clock`, and isn't renewed. An expired session counts as missing, and is deleted when it's found.
- Sign-in works on memory, SQLite, and Postgres storage only. The server refuses sign-in with plain-text storage at start, because plain text is local edition only (ADR 5), so jsonfile has no account or session repository.

### Signed out is "no workspace"

- The session middleware puts a valid session's workspace in the request's context (`workspace.WithID`). Any other request runs with no workspace.
- The use cases already refuse a request with no workspace (`domain.ErrNoWorkspace`), before they touch storage. So being signed out fails closed, with no check of its own in each handler. A route someone forgets to protect still can't see any data.
- `problem.go` maps `ErrNoWorkspace` to a 401 problem, and `httpclient` maps it back.

### The dev sign-in is only in dev builds

- The dev sign-in signs in whoever types a name, with the issuer `dev` and the name as the subject. It stands in for a provider in `task dev` and in the tests.
- Its wiring is behind the Go build tag `devsignin`. `task dev`, `task run:server`, `task run:cli`, `task run:tui`, and the browser journey's server pass the tag. The acceptance tests' http driver builds the handler in-process, with the dev sign-in turned on in its options. `task build` doesn't pass the tag, and a build without it refuses `sign_in = "dev"` at start. So a public server can't turn it on by mistake.
- Building with the tag doesn't turn sign-in on. The `sign_in` setting still decides, and it is `off` by default.

### Routes outside the API

- Signing in is redirects and HTML forms, not JSON, so its routes live under `/auth/`, outside `api/openapi.yaml`.
- `GET /auth/sign-in` goes to the configured sign-in adapter. For the dev sign-in, that is `GET /auth/dev`, a plain form that posts to `POST /auth/dev`. `POST /auth/sign-out` ends the session and clears the cookie.
- With sign-in on, a request that changes something (not `GET`, `HEAD`, or `OPTIONS`) is refused with a 403 `cross_site` problem unless it comes from the server's own origin. Go's [`http.CrossOriginProtection`](https://pkg.go.dev/net/http#CrossOriginProtection) does the check, from the browser's `Sec-Fetch-Site` header or its `Origin`. This protects against cross-site request forgery (CSRF), where another site makes the browser send a request with its cookie. `SameSite=Lax` already stops most of these, and the check closes the rest, such as another origin on the same site.

Rejected:
- **Tokens in the browser** (the web app holds the OIDC tokens, in memory or local storage): any script injected into the page could read and send them.
- **Signed sessions with no storage** (a JSON Web Token, or a signed cookie, that the server checks without a lookup): signing out couldn't end a session before it expires, short of a list of ended sessions, which is storage again.
- **Sessions in server memory:** they'd be lost on every restart, and one server copy wouldn't know another's sessions.
- **A dev sign-in turned on only by a setting:** a release build with a wrong setting would let anyone sign in as anyone.

## Consequences

- Every new front end that reaches the hosted server from a browser gets sign-in for free, from the cookie. Clients that aren't browsers wait for the bearer token in step 9.
- Each request with a cookie costs one session lookup and one account lookup in storage.
- Expired sessions that nobody looks up stay in storage until a sweep removes them. The sweep is a follow-up in the editions plan.
- Behind a proxy that handles HTTPS, the server sees plain HTTP, so it doesn't mark the cookie `Secure` yet. This is a follow-up in `TODO.md`.
- Roles, such as an admin who can change plans from the web app, wait for teams.
