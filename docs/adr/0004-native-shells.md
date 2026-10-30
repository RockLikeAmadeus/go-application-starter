# 4. Native shells run the same web app over the same REST API

- Status: accepted (desktop built; mobile planned: Wails, Android first)
- Date: 2026-10-30

## Context

Users want the app outside a browser tab: a desktop app that works without a server, and later phone apps. We already have a React SPA that talks to the core only through the REST API in `api/openapi.yaml` (ADR 2). We don't want a second UI, a second API surface, or a second runtime beside Go.

## Decision

Every native shell runs the existing SPA from `web/` unchanged, and the SPA keeps calling its own `/api/*`.

### Desktop: Wails v3, in-process

- `cmd/desktop` (build tag `desktop`) opens a Wails v3 window. Wails' asset handler is `httpapi.NewHandler`, so the SPA's files and its `/api/*` requests are answered **in-process**. No network port is opened, and the REST contract stays the only one.
- The wiring lives in `internal/adapters/desktop`, which doesn't import Wails. `desktop.Open` picks the data folder (`STARTER_DATA_DIR`, else `<user settings folder>/starter`), opens jsonfile storage through `bootstrap.NewEmbedded`, and returns the handler. Unit tests and a Godog driver use it without a window. The build tag keeps Wails' cgo and GTK dependency out of `go test`, lint and CI.
- **Local-only for now, remote mode kept open.** A desktop app will run either on local data (jsonfile in its data folder) or connected to a hosted service. Remote mode would be `desktop.Open` choosing `bootstrap.NewRemote` behind the same in-process handler, so the SPA doesn't change and no CORS is needed. Nothing outside `desktop.Open` may assume local storage.
- **No sync between local and cloud data.** A copy of the app works on one or the other, never both, and switching modes migrates nothing. So remote mode needs no conflict handling, offline queue or merging.
- **The server says whether a connection is needed.** `GET /api/v1/connection` returns `{"required": false}` when the data is local (`httpapi.Options.LocalData`). The SPA then hides its offline notice and stops TanStack Query from pausing requests when the network drops. The SPA does not sniff for Wails (`window._wails`): that says "Wails", not "local data", and would be wrong in remote mode.
- **No service worker.** The desktop handler answers `/sw.js` with 404. The app's files are built in, so precaching adds nothing, and "update on next open" would show the old version once after each upgrade.

Rejected:
- **Wails bindings:** a second API surface beside REST.
- **Electron, Tauri:** a second runtime or language next to Go.
- **Capacitor for mobile** (the first plan): remote-only, so no offline use on the phone. It remains the fallback.
- **A local HTTP server on a port:** other programs on the machine could reach the data, and it adds port clashes.

### Mobile: Wails v3 too, Android first (planned)

The phone runs the same core and REST handler in-process, with its data on the phone, so it works offline. A spike on the Android emulator (2026-10-30) proved this, with one catch: Wails' Android asset path sends every request to Go as a `GET` with no body, and drops every status except 200. So on Android the web app sends `/api` requests over the Wails bridge, through a small Go service that runs them through the same `http.Handler`. `/api` on the asset path answers 404, so a request that misses the bridge fails loudly. iOS waits for a Mac.

Wails mobile is experimental (outside its stability promise). The fallback is Capacitor, remote-only: mature, but it needs CORS and an API base URL, and can't work offline without reimplementing the core.

## Consequences

- One UI and one contract: a feature built for the browser works in every shell, and the SPA's tests cover it.
- Building the desktop app needs system packages (GTK 4 and WebKitGTK 6.0 on Linux; see the README). Plain `go build ./...` and `task check` don't.
- The desktop app locks its data folder like the server does, so a second copy, or the CLI against the same folder, fails fast with "data in use".
- Packaging (AppImage, .deb, .dmg, .msi) and remote mode are follow-ups in `TODO.md`.
