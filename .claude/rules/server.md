---
paths:
  - "internal/adapters/httpapi/**"
  - "internal/adapters/httpclient/**"
  - "cmd/server/**"
  - "api/**"
  - "web/assets*.go"
---

# REST server and SPA serving

- The server embeds the SPA only when built with `-tags embedweb` after `pnpm build` (`task build` does both). Otherwise `/` shows a hint page.
- The SPA's service worker serves the cached app for offline page loads on any path the Go server serves outside `/api`, so add new server routes (like `/healthz` and `/readyz`) to `navigateFallbackDenylist` in `web/vite.config.ts`. (More on the installable app in `.claude/rules/web.md` and `docs/installable-app.md`.)
- In the local edition, the desktop app keeps its data on the computer (`httpapi.Options.IsDataLocal`; remote mode, which is planned, will leave it unset), so `GET /api/v1/connection` answers `{"isRequired": false}`. The desktop handler answers `/sw.js` with 404, so the service worker never runs there.
- Sign-in (ADR 9, `docs/sign-in.md`) lives under `/auth/`, outside `api/openapi.yaml`, as it is redirects and HTML forms: `GET /auth/sign-in` redirects to the configured adapter (`/auth/dev`, only with `Options.IsDevSignInOn`), and `POST /auth/sign-out` ends the session. `/auth/` is in `navigateFallbackDenylist`. With `Options.SignIn` set, `sessionMiddleware` puts the `session` cookie's workspace in the context (none when signed out, so use cases fail with `ErrNoWorkspace`, a 401), and `refuseCrossSite` (`http.CrossOriginProtection`) refuses cross-origin requests that change something (403 `cross_site`). A new handler needs no sign-in check of its own. `GET /api/v1/session` tells the web app whether sign-in is on and who is signed in.
