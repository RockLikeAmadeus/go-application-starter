# Sign in through a sign-in provider (OIDC)

Step 8b of [the editions plan](../plans/editions-and-accounts.md).

- Spec: [SPEC.md §4.18](../../SPEC.md#418-sign-in-to-the-hosted-edition), plus §5 (Security) and §6
- Scenarios: [`features/sign-in-provider.feature`](../../features/sign-in-provider.feature) (`@sign-in-provider`, 6 scenarios, each tagged `@wip`). A browser journey for "Create an account" is added in task 22, in [`web/e2e/features/sign-in.feature`](../../web/e2e/features/sign-in.feature). [`features/sign-in.feature`](../../features/sign-in.feature) must also pass on the new oidc driver.
- Decisions: [ADR 9](../adr/0009-sign-in-and-sessions.md) (to update), [`docs/sign-in.md`](../sign-in.md) (to update)

## Decisions

Agreed with the user (2026-11-11):
- **Library:** [`github.com/coreos/go-oidc/v3`](https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc) checks ID tokens, and [`golang.org/x/oauth2`](https://pkg.go.dev/golang.org/x/oauth2) swaps the code, with its PKCE helpers (`GenerateVerifier`, `S256ChallengeOption`, `VerifierOption`). Not zitadel/oidc: its own cookies and routes would fight ours.
- **One `public_url` setting** (`STARTER_PUBLIC_URL`), such as `https://app.example.com`. The redirect URL is `<public_url>/auth/oidc/callback`. Required with `sign_in = "oidc"`, optional otherwise. An `https://` public URL marks the session cookie `Secure` even over plain HTTP, which settles the TODO.md item "Mark the session cookie `Secure` behind an HTTPS proxy". Without one, today's rule (`r.TLS != nil`) stays.
- **"Create an account"** (added 2026-11-12): the sign-in screen has a "Create an account" link beside "Sign in" (the user chose that wording), to `GET /auth/sign-up`. With OIDC, that goes to `GET /auth/oidc/sign-up`, which starts like `/auth/oidc` but asks the provider for its sign-up page with [`prompt=create`](https://openid.net/specs/openid-connect-prompt-create-1_0.html). The parameter is sent only when the provider's discovery lists `create` in `prompt_values_supported`, as a provider that doesn't know it may refuse the request. Otherwise the person gets the sign-in page, which links to the provider's sign-up. The link always shows while signed out with sign-in on: the web app doesn't learn what the provider supports. With the dev sign-in, `/auth/sign-up` goes to `/auth/dev`, so the browser journey can prove the link with the dev sign-in.
- **The name follows the provider** (the 8a loose end "A changed name isn't saved"): `SignIn` saves the identity's name on an existing account when it differs. `AccountRepository.Save` already replaces by ID, with a contract test, so no storage change.

Decided in planning:
- **Flow:** authorization code with PKCE (S256), `state`, and `nonce`. Scopes `openid profile email`. The client secret is sent too (a confidential client).
- **Packages:** `internal/adapters/oidc` does the protocol, with no routes or cookies: `New(ctx, Settings)` discovers the provider, `Start()` gives the authorization URL and an `Attempt` (state, nonce, verifier, each fresh from `crypto/rand`), and `Finish(ctx, attempt, code)` returns a `domain.Identity`. `httpapi` defines a small interface for it (`SignInProvider`) and owns the routes and cookies, so its tests use a fake. `internal/adapters/oidc/oidctest` is the fake provider, an `httptest` server with its own RSA key, signing with go-jose (go-oidc's own dependency). Only tests import `oidctest`: add a depguard rule when it's created.
- **The identity:** issuer `iss`, subject `sub`, and the name from the `name` claim, else `preferred_username`, else `email`, else the subject. No email is stored.
- **Routes:** `GET /auth/sign-in` goes to `GET /auth/oidc`, which sets the attempt cookie and redirects to the provider. The provider returns to `GET /auth/oidc/callback`. `sign_in` is `off`, `dev`, or `oidc`, never two at once.
- **The attempt cookie** `oidc_sign_in`: `HttpOnly`, `SameSite=Lax` (the provider's redirect back is a top-level `GET`, which Lax still sends it with), `Path=/auth/oidc`, 10 minutes, `Secure` by the same rule as the session cookie. Cleared at the callback, whatever the outcome. It holds the attempt as is: another site can't read or set it.
- **A refused sign-in** answers a plain HTML page saying signing in didn't work, with a link to `/auth/sign-in` to try again, status 400 (wording to agree in its cycle). Causes: the provider's `error` parameter (such as `access_denied` when someone cancels), no attempt cookie, a `state` that doesn't match, or `Finish` failing. The log gets the cause; the page doesn't.
- **Discovery at start:** `bootstrap` calls `oidc.New`, so an unreachable or wrong issuer stops the server starting, with a clear message. `/readyz` doesn't check the provider: people already signed in keep working while it's down.
- **Sign-out is local only:** no RP-initiated logout at the provider. The guide says so.
- **Acceptance:** a new `oidc` driver, the hosted edition over HTTP with the OIDC adapter against `oidctest.Provider`, runs `@sign-in` and `@sign-in-provider`. The other drivers leave out `@sign-in-provider`. The driver maps a scenario's name to a person at the fake provider (subject and name), so the admin step must find the account by the driver's identity for a name, not `devIdentity`. No browser or terminal journey: the web app doesn't change (its link already goes to `/auth/sign-in`), and the oidc driver follows the redirects with a cookie jar, as a browser does.
- Token checks (signature, audience, expiry, issuer, nonce), the settings, and `Secure` are unit tests, not Gherkin.

## Tasks

Each is one `/red` → `/green` → `/refactor` cycle.

The OIDC adapter (`internal/adapters/oidc`):
- [ ] 1. `oidc.New` discovers the provider from its issuer, and fails for an unreachable issuer. Adds go-oidc and x/oauth2 to the main `go.mod`, and starts `oidctest.Provider` with discovery and keys.
- [ ] 2. `Start` returns the provider's authorization URL with the client ID, redirect URL, scopes, `state`, `nonce`, and an S256 PKCE challenge, and an `Attempt` holding them. Two starts share nothing.
- [ ] 3. `Finish` swaps the code, sending the PKCE verifier and client secret, checks the ID token, and returns the identity (issuer, subject, `name`). The fake gains its authorize and token endpoints, with a person to sign in.
- [ ] 4. The name falls back to `preferred_username`, then `email`, then the subject (table test).
- [ ] 5. `Finish` refuses an ID token signed with another key, meant for another client, expired, from another issuer, or with another nonce (table test; the fake's misbehaviour switches).
- [ ] 6. `Finish` refuses when the provider rejects the code: a wrong PKCE verifier, or a code used twice.
- [ ] 6b. `StartSignUp` adds `prompt=create` when the provider's discovery lists `create` in `prompt_values_supported`, and leaves it out otherwise (the fake gets a switch for it, and records the prompt it was sent).

The core:
- [ ] 7. `SignIn` for a known identity whose name changed saves the new name, so `CurrentAccount` shows it. (Ticks the 8a loose end.)

Settings (`internal/config`):
- [ ] 8. `sign_in = "oidc"` is accepted, with `public_url`, `oidc_issuer`, `oidc_client_id`, and `oidc_client_secret` (`STARTER_…`), left out of the settings file like `sign_in`.
- [ ] 9. With `oidc`, a missing `public_url`, `oidc_issuer`, `oidc_client_id`, or `oidc_client_secret` is refused, naming the setting (wording to agree).
- [ ] 10. A `public_url` that isn't an absolute `http://` or `https://` address is refused (wording to agree).

The server (`internal/adapters/httpapi`):
- [ ] 11. With a `SignInProvider`, `GET /auth/oidc` sets the attempt cookie and redirects to `Start`'s URL, and `GET /auth/sign-in` goes to `/auth/oidc`.
- [ ] 12. `GET /auth/oidc/callback` with the attempt's `state` calls `Finish`, signs in, sets the session cookie, clears the attempt cookie, and redirects to `/`.
- [ ] 13. A callback with the provider's `error` parameter answers the "signing in didn't work" page, 400, and clears the attempt cookie (wording to agree).
- [ ] 14. The same page for no attempt cookie, a `state` that doesn't match, and `Finish` failing (table test), with the cause logged.
- [ ] 14b. `GET /auth/sign-up` goes to the adapter's sign-up: `/auth/oidc/sign-up` with OIDC, which sets the attempt cookie and redirects to `StartSignUp`'s URL, or `/auth/dev` with the dev sign-in. `/auth/` is already in the service worker's `navigateFallbackDenylist`, so nothing changes there.
- [ ] 15. With an `https://` public URL, the session and attempt cookies are `Secure` over plain HTTP. (Ticks the TODO.md item.)

Wiring:
- [ ] 16. `bootstrap` with `sign_in = "oidc"` builds the OIDC adapter from the settings, against `oidctest.Provider`, and an unreachable issuer fails at start (wording to agree).
- [ ] 17. `cmd/server`'s `handlerOptions` passes the provider and the public URL to `httpapi.Options`.

Acceptance (`test/acceptance`; run `task test:init`, as the driver has a sample block):
- [ ] 18. The `oidc` driver, running `@sign-in` and `@sign-in-provider`, with the other drivers leaving out `@sign-in-provider`. The admin step finds the account by the driver's identity for a name. `sign-in.feature` passes on it.
- [ ] 19. Steps for a name changed at the provider, and for another person with the same name.
- [ ] 20. Steps for cancelling at the provider, coming back without starting, being told signing in didn't work, and not being signed in.
- [ ] 20b. Steps for creating an account at the provider, which page the provider showed first, and a provider that can't start at sign-up.

The web app:
- [ ] 21. `SignInScreen` shows a "Create an account" link to `/auth/sign-up` beside "Sign in" (Vitest, by role). Agree with the user whether "Please sign in to continue." should change, now that the screen also serves new people.
- [ ] 22. A browser journey in `web/e2e/features/sign-in.feature`: signed out, create an account as "Browser Grace" (the dev sign-in's form), see the items page, signed in as "Browser Grace".

- [ ] 23. **Docs and finish.**
  - `docs/sign-in.md`: setting up a provider (registering the app, the redirect URL, the settings, the client secret as a secret, and turning on its sign-up), a sequence diagram of the redirects, the journey of someone new (create an account → the provider's sign-up page → back, with the product's account made at the first sign-in), `prompt=create` and what happens with a provider that doesn't support it, the provider and email-and-password notes (including the cost of moving password accounts between providers), local-only sign-out, and how it's tested.
  - `docs/configuration.md` (the four settings), ADR 9 (status, the name rule, `public_url` and `Secure`), `.claude/rules/server.md` (the `/auth/oidc` routes), CLAUDE.md's editions paragraph, and README if it mentions sign-in.
  - Template users' steps in TODO.md (choose a provider, register the app, set the settings), and tick the `Secure` TODO item.
  - Ask the user to check `.env.example` for the new settings (Claude can't read it).
  - Remove the `@wip` tags, mark step 8b done in the plan, tick the "changed name" loose end, and delete this file.

## Changes to the plan
- 2026-11-12: added "Create an account" (SPEC.md §4.18's new rule and example, 2 scenarios, and tasks 6b, 14b, 20b, 21, and 22). The user found that nothing covered someone new, and wants a link that goes straight to the provider's sign-up page. The browser journey is written by task 22, not here: playwright-bdd stops the whole suite while a step is undefined, so `task e2e` would fail on the branch until then (8a did the same). Docs and finish became task 23.
