---
paths:
  - "web/**"
  - ".nvmrc"
---

# Web app (SPA)

- Node ≥ 24 is required (see `.nvmrc`). TypeScript is pinned to 6.0.x until typescript-eslint supports 7. Dependabot ignores TypeScript 7 until then; drop that rule in `.github/dependabot.yml` when unpinning.
- Unit tests: Vitest + Testing Library, querying by role or label. Stub `fetch` rather than mocking modules.
- Components every screen can use (e.g. `ErrorMessage`) go in `web/src/components`.
- Regenerate the app icons with `pnpm --dir web icons` after editing `web/public/icon.svg`; commit the PNGs and `build/desktop/windows/icon.ico`, which it also writes. The desktop app uses `pwa-512x512.png` too.
- The generated API client (`*.gen.ts`) comes from `api/openapi.yaml` via `task gen`. Never hand-edit it.

## Installable app (SPEC.md §4.5, `docs/installable-app.md`)

The SPA is an installable web app (`VitePWA` in `web/vite.config.ts`). Its service worker precaches only the built app and must never cache `/api`: data always comes from the server. Offline, page loads get the cached app. That includes any path the Go server serves outside `/api`, so add new server routes (like `/healthz` and `/readyz`) to `navigateFallbackDenylist`. A deploy reaches an open or installed app the next time it's opened. The service worker is off under `task dev`; test it with `task e2e` or a built server.

## Local data and offline behaviour

In the local edition, the desktop app keeps its data on the computer (`httpapi.Options.IsDataLocal`; remote mode, which is planned, will leave it unset), so `GET /api/v1/connection` answers `{"isRequired": false}`. The SPA then hides the offline notice and keeps TanStack Query online (`useIsConnectionRequired`, `useLocalDataOnline`). New offline behaviour in the SPA must respect that answer. The desktop handler answers `/sw.js` with 404, so the service worker never runs there.

## Android bridge (`web/src/api/client.ts`)

On Android, the SPA's `/api` requests go over the Wails bridge (`apiFetch` → `bridgeFetch` → `main.Bridge.Serve` → `mobile.Serve`). Details and constraints are in `.claude/rules/mobile.md`. In short: `client.ts` calls `Bridge` by that name, and `apiFetch` checks for `window.wails.invokeAsync` before importing `@wailsio/runtime`, because the import replaces `window.wails`. Keep `@wailsio/runtime` pinned exactly (see `.claude/rules/dependencies.md`).
