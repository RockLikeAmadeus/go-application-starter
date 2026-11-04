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
- In the local edition, the desktop app keeps its data on the computer (`httpapi.Options.LocalData`; remote mode, which is planned, will leave it unset), so `GET /api/v1/connection` answers `{"required": false}`. The desktop handler answers `/sw.js` with 404, so the service worker never runs there.
