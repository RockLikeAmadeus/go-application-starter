# 4. Native shells run the same web app over the same REST API

- Status: accepted (desktop and Android built; iOS planned)
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

### Mobile: Wails v3 too, Android first

The phone runs the same core and REST handler in-process, with its data on the phone, so it works offline. A spike on the Android emulator (2026-10-30) proved this, with one catch: Wails' Android asset path sends every request to Go as a `GET` with no body, and drops every status except 200.

- `cmd/mobile` (build tag `android`) wraps `internal/adapters/mobile`, which, like `desktop`, doesn't import Wails. `mobile.Open` calls `desktop.Open` with the app's private storage as the settings folder, so the data folder is `<storage>/starter` and `STARTER_DATA_DIR` still wins.
- **`/api` goes over the Wails bridge.** On Android the web app's `fetch` hands each API request to one bound Go method, `main.Bridge.Serve(method, path, body)`. It runs the request through the same `http.Handler` (`mobile.Serve`) and returns the status, content type and body, which the web app turns back into a `Response`. This is a transport for the REST contract, not a second API: the bridge has one method, and every route, status and error comes from `httpapi`. In browsers and on the desktop the web app keeps using plain `fetch`, and loads no Wails code.
- **`/api` on the asset path answers 404**, so a request that misses the bridge fails loudly. In the spike, a `POST` that fell through was answered as a `GET` with 200, and the web app reported a save that never happened.
- The acceptance tests run every feature through `mobile.Serve`, as a fifth driver, so the bridge is covered without a phone. The Android project is Wails' generated Gradle project, committed under `mobile/android`, and the Taskfile builds it with the NDK and Gradle directly. Setup, debugging and upgrade steps are in `docs/mobile.md`.
- iOS waits for a Mac. Its custom-scheme handler should pass methods and bodies through, so it may not need the bridge, but that's unverified.

Wails mobile is experimental (outside its stability promise), and the bridge depends on how its Android code behaves today. The fallback is Capacitor, remote-only: mature, but it needs CORS and an API base URL, and can't work offline without reimplementing the core.

## Consequences

- One UI and one contract: a feature built for the browser works in every shell, and the SPA's tests cover it.
- Building the desktop app needs system packages (GTK 4 and WebKitGTK 6.0 on Linux; see the README), and the Android app needs a JDK, the Android SDK and the NDK. Plain `go build ./...` and `task check` need neither.
- The desktop app locks its data folder like the server does, so a second copy, or the CLI against the same folder, fails fast with "data in use".
- Wails upgrades need care on Android: the Go module and the `@wailsio/runtime` npm package must move together, and the generated Gradle project may need regenerating.
- Packaging (AppImage, .deb, .dmg, .msi), remote mode, iOS, and Play Store signing are follow-ups in `TODO.md`.
