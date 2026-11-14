# 9. Sign-in and sessions: a session cookie, sessions in storage

- Status: accepted. The dev sign-in (step 8a of [`docs/plans/editions-and-accounts.md`](../plans/editions-and-accounts.md)), the OIDC adapter (step 8b), and signing in from the command line and the terminal app (step 9) are built. [`docs/sign-in.md`](../sign-in.md) explains them.
- Date: 2026-11-11, updated 2026-11-12 for the OIDC adapter, and 2026-11-13 for the command line

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
- The core never sees how someone proved who they are. So the dev sign-in and OIDC are both adapters behind the same port, and the core doesn't change between them.
- **An account is found by its identity's issuer and subject, never by its name,** because names change and two people can share one. One identity has at most one account. SQLite and Postgres enforce this with a unique index, and `SignIn` looks the account up again when two first sign-ins race, so both share one account.
- **The name follows the identity.** When a known identity signs in with a name that differs from the account's, `SignIn` saves the new name. So someone who changes their name at the provider sees it in the app from their next sign-in. The provider's word is the only source of the name, as the product has no screen to change it.
- The first sign-in creates the account and its personal workspace. Both IDs come from `ports.IDGenerator`. No plan is stored for the new workspace: it is on the plans file's `newWorkspacePlan` until an admin changes it.

### A session cookie, not tokens in the browser

- The server gives the browser a cookie named `session`, holding the session's token. The cookie is `HttpOnly`, so page scripts can't read it. It is `SameSite=Lax`, so the browser leaves it off most requests that other sites start. It is `Secure`, so it never travels unencrypted, when the request came over HTTPS or the `public_url` setting starts with `https://` (see "The OIDC adapter" below). It lasts as long as the session, so it outlives the browser being closed.
- The web app holds no tokens at all. It learns whether someone is signed in from `GET /api/v1/session`.
- This is the "backend for frontend" pattern, which the IETF's draft on [OAuth for browser-based apps](https://datatracker.ietf.org/doc/draft-ietf-oauth-browser-based-apps/) ranks the most secure for browsers. A script injected into the page (cross-site scripting) can't steal a token it can't read.
- The command line and the terminal app send a bearer token in a header instead, since their requests don't come from a browser, and the desktop and mobile apps in remote mode (step 10) will too. The middleware accepts both, as "Clients that aren't browsers" below explains.

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

### The OIDC adapter (step 8b)

- **The flow is the authorization code flow, with PKCE (S256), a `state`, and a `nonce`.** The server is a confidential client, so it sends its client secret when it swaps the code. It asks for the scopes `openid`, `profile`, and `email`.
- **The libraries:** [`github.com/coreos/go-oidc/v3`](https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc) checks ID tokens, and [`golang.org/x/oauth2`](https://pkg.go.dev/golang.org/x/oauth2) swaps the code, with its PKCE helpers.
- **The protocol and the routes are apart.** `internal/adapters/oidc` speaks the protocol and has no routes or cookies. `Start` and `StartSignUp` return the provider's URL, the state, and an opaque attempt string (the nonce and the PKCE verifier), and `Finish` takes the attempt and the code and returns a `domain.Identity`. `httpapi` defines its own small `SignInProvider` interface, in standard types only, and never imports `oidc`. So `httpapi`'s tests use a fake, and `oidc`'s tests use `oidctest`, a fake provider that only tests may import.
- **The identity** is the ID token's issuer (`iss`) and subject (`sub`). The name is the `name` claim, else `preferred_username`, else `email`, else the subject. No email address is stored.
- **The routes:** `GET /auth/sign-in` goes to `GET /auth/oidc`, which starts an attempt and redirects to the provider. The provider returns to `GET /auth/oidc/callback`. `GET /auth/sign-up` goes to `GET /auth/oidc/sign-up`, which asks for the provider's sign-up page with `prompt=create`, only when the provider's discovery lists it. `sign_in` is `off`, `dev`, or `oidc`, never two at once.
- **The attempt lives in a cookie,** `oidc_sign_in`, as `<state>.<attempt>`. It is `HttpOnly`, `SameSite=Lax` (the provider's redirect back is a top-level `GET`, which `Lax` still sends it with), on `Path=/auth/oidc`, for 10 minutes, and `Secure` by the session cookie's rule. The callback clears it whatever the outcome. So the server stores nothing for an attempt someone abandons.
- **A refused sign-in** answers a plain "Sign-in failed" page, with status 400 and a link to try again. Its causes are the provider's `error` parameter, no attempt cookie, a `state` that doesn't match, or `Finish` failing. The log gets the cause, and the page doesn't.
- **One `public_url` setting** gives the server's own address, such as `https://app.example.com`. The redirect URL is `<public_url>/auth/oidc/callback`. It is required with `sign_in = "oidc"`. An `https://` public URL marks the cookies `Secure` even when the server sees plain HTTP, as it does behind a proxy that handles HTTPS. Without one, only a TLS connection marks them `Secure`.
- **Discovery happens at start.** An unreachable or wrong issuer stops the server from starting, with a clear message. `/readyz` doesn't check the provider, so people already signed in keep working while it's down.
- **Sign-out is local only.** The server ends its own session, and doesn't sign the person out at the provider (RP-initiated logout).

### Clients that aren't browsers (step 9)

- **The command line signs in through the browser, with a loopback port** ([RFC 8252](https://datatracker.ietf.org/doc/html/rfc8252), OAuth for native apps). `starter login` listens on `127.0.0.1` at a free port, and opens the browser at `GET /auth/native` with the port, a `state`, and a PKCE S256 challenge. The browser signs in as usual, or skips that when it already has a session. `GET /auth/native/finish` then issues a one-time code and redirects to `http://127.0.0.1:<port>/callback` with the code and the state. The command line checks the state, and swaps the code and its verifier at `POST /auth/native/token` for a new session token. The command line never talks to the provider, so it works the same with the dev sign-in and with any OIDC provider.
- **The attempt lives in a cookie,** `native_sign_in`, on `Path=/auth/native`, for 10 minutes, as the OIDC attempt does. `/auth/native/finish` checks it again, and builds the loopback address from the port itself, so it can redirect only to `127.0.0.1`.
- **The sign-in routes take a `return` path,** so the browser comes back to `/auth/native/finish` after signing in. Anything but a path on this server is ignored, so the sign-in can't be used as an open redirect. The OIDC attempt cookie carries the path through the trip to the provider.
- **One-time codes are kept in storage** (`ports.SignInCodeRepository`, the `sign_in_codes` table), so any server copy can redeem them. Storage keeps the code's SHA-256 hash, its account, its PKCE challenge, and its expiry. A code lasts 5 minutes, works once (it is taken and deleted in one step), and only with the verifier behind its challenge. A refused swap answers a bare 400 that doesn't say why.
- **The bearer token is a session token.** It is a new session, apart from the browser's, looked up as the cookie's is, for 30 days and not renewed. The middleware reads `Authorization: Bearer` first, then the cookie. `POST /auth/sign-out` ends a bearer session too. `http.CrossOriginProtection` lets requests with no `Sec-Fetch-Site` or `Origin` through, so the command line passes the cross-origin check unchanged.
- **The native routes live under `/auth/`,** beside the other sign-in routes and outside `api/openapi.yaml`, as they are redirects, and a token endpoint called by hand-written client code.
- **The token is kept in the operating system's credential store,** through [`zalando/go-keyring`](https://github.com/zalando/go-keyring), falling back to `tokens.json` in the settings folder (mode `0600`) where there is none. The `credential_store` setting (`system` or `file`) lets CI, tests, and people who don't want the credential store's prompts use the file only. Tokens are keyed by the normalised server address, so the command line can be signed in to several servers.
- **Signing in stores a token only.** It never chooses the server or the data, which still come from `--server`, `STARTER_SERVER_URL`, or `server_url`. Nothing writes the settings file (SPEC.md §4.12).
- **The terminal app signs in and out by itself,** with a signed-out screen and the `l` and `o` keys, and shares the command line's stored session.
- **The client side is one package, `clientsignin`,** which the command line and the terminal app call, and the desktop and mobile apps will call in step 10. It imports no other adapter. `internal/bootstrap` passes in `httpclient`, `browsersignin`, and `tokenstore` through small interfaces. `httpclient` asks a token source on each request, so a new or forgotten session takes effect without reconnecting.

Rejected:
- **Tokens in the browser** (the web app holds the OIDC tokens, in memory or local storage): any script injected into the page could read and send them.
- **Signed sessions with no storage** (a JSON Web Token, or a signed cookie, that the server checks without a lookup): signing out couldn't end a session before it expires, short of a list of ended sessions, which is storage again.
- **Sessions in server memory:** they'd be lost on every restart, and one server copy wouldn't know another's sessions.
- **A dev sign-in turned on only by a setting:** a release build with a wrong setting would let anyone sign in as anyone.
- **[`zitadel/oidc`](https://github.com/zitadel/oidc) as the OIDC library:** it brings its own cookies and routes, which would fight the server's.
- **Trusting `X-Forwarded-Proto` to mark the cookie `Secure` behind a proxy:** the header is safe only from a known proxy, which needs a setting for the proxy's addresses. The `public_url` setting is needed anyway, and already says whether people connect over HTTPS.
- **The attempt kept in storage:** an abandoned attempt would stay behind, and the cookie already keeps it where another site can't read or set it.
- **The device code flow** ([RFC 8628](https://datatracker.ietf.org/doc/html/rfc8628)) for the command line: the person types a short code into a browser on any device. It suits computers with no browser, such as one reached over SSH, so it is a follow-up rather than the main flow.
- **Pasting a token** from the web app into the command line: clumsy, and it teaches people to copy secrets around.

## Consequences

- Every new front end that reaches the hosted server from a browser gets sign-in for free, from the cookie. A client that isn't a browser signs in through `clientsignin`, as the command line and the terminal app do.
- Each request with a cookie or a bearer token costs one session lookup and one account lookup in storage.
- The command line can sign in only where a browser runs on the same computer, as the loopback address is reachable from that computer only. Over SSH, that needs a browser on the remote computer until the device code flow is added.
- A one-time code that nobody redeems stays in storage after it expires, until a sweep removes it, as an expired session does. The sweep for both is a follow-up in the editions plan.
- Expired sessions that nobody looks up stay in storage until a sweep removes them. The sweep is a follow-up in the editions plan.
- Behind a proxy that handles HTTPS, the cookies are `Secure` only when `public_url` starts with `https://`. A server people reach over plain HTTP must not set an `https://` public URL, or the browser never sends the cookie back.
- The product never stores a password. Email and password accounts come from a provider that keeps its own accounts, so moving between providers can cost users a password reset.
- A provider that is down stops new sign-ins, but not sessions that already exist.
- Roles, such as an admin who can change plans from the web app, wait for teams.
