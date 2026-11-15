# 4. Native shells run the same web app over the same REST API

- Status: accepted (desktop and Android built; iOS planned)
- Date: 2026-10-30

## Context

Users want the app outside a browser tab: a desktop app that works without a server, and later phone apps. We already have a React SPA that talks to the core only through the REST API in `api/openapi.yaml` (ADR 2). We don't want a second UI, a second API surface, or a second runtime beside Go.

## Decision

Every native shell runs the existing SPA from `web/` unchanged, and the SPA keeps calling its own `/api/*`.

### Desktop: Wails v3, in-process

- `cmd/desktop` (build tag `desktop`) opens a Wails v3 window. Wails' asset handler is `httpapi.NewHandler`, so the SPA's files and its `/api/*` requests are answered **in-process**. No network port is opened, and the REST contract stays the only one.
- The wiring lives in `internal/adapters/desktop`, which doesn't import Wails. `desktop.Open` picks the data folder (`STARTER_DATA_DIR`, else `<user settings folder>/starter`), opens jsonfile storage through `bootstrap.NewEmbedded`, and returns the handler. Unit tests and a Godog driver use it without a window. The build tag keeps Wails' cgo and GTK dependency out of `go test`, lint, and CI.
- **Local or remote mode, chosen when the app opens.** A copy of the app runs either on local data (in its data folder) or on a hosted server, in **remote mode** (built in step 10 of `docs/plans/editions-and-accounts.md`, SPEC §4.20). With a server set (the `server_url` setting, `STARTER_SERVER_URL`, or an address built into the program, `config.BuiltInServerURL`), `desktop.Open` puts `bootstrap.NewRemote` behind the same in-process handler, so the SPA doesn't change and no CORS is needed: `httpclient` makes plain REST calls to the server. It then opens no data folder or storage, so nothing is locked and the `storage` setting is ignored. A change of setting takes effect at the next opening. Nothing outside `desktop.Open` may assume local storage.
- **Remote mode signs in through the system browser, like the command line** (ADR 9, `docs/sign-in.md`). The requests to the server come from the app's Go code, not from the window, so the window's cookie can't sign them in. Instead, the in-process handler signs in through `httpapi.Options.ClientSignIn`, an interface over `clientsignin` wired in bootstrap: the session answer says `signInMethod: browser`, the web app's sign-in screen calls `POST /api/v1/browser-sign-in` and shows a waiting screen with Cancel (`DELETE /api/v1/browser-sign-in`), and the Sign out button calls `DELETE /api/v1/session`, which forgets the token. The apps sign out through the API because their webview can't follow the redirect that the hosted web app's sign-out form gets (ADR 9). Nothing the window loads may answer with a redirect. The opener is passed in: `cmd/desktop` and `cmd/mobile` pass Wails' `Browser.OpenURL`, which also works on Android. The apps keep their own session token, apart from the command line's.
- **No sync between local and cloud data.** A copy of the app works on one or the other, never both, and switching modes migrates nothing. So remote mode needs no conflict handling, offline queue, or merging.
- **The server says whether a connection is needed.** `GET /api/v1/connection` returns `{"isRequired": false}` when the data is local (`httpapi.Options.IsDataLocal`). The SPA then hides its offline notice and stops TanStack Query from pausing requests when the network drops. The SPA does not sniff for Wails (`window._wails`): that says "Wails", not "local data", and would be wrong in remote mode.
- **No service worker.** The desktop handler answers `/sw.js` with 404. The app's files are built in, so precaching adds nothing, and "update on next open" would show the old version once after each upgrade.

Rejected:
- **Wails bindings:** a second API surface beside REST.
- **Electron, Tauri:** a second runtime or language next to Go.
- **Capacitor for mobile** (the first plan): remote-only, so no offline use on the phone. It remains the fallback.
- **A local HTTP server on a port:** other programs on the machine could reach the data, and it adds port clashes.

### Mobile: Wails v3 too, Android first

The phone runs the same core and REST handler in-process, with its data on the phone, so it works offline. A spike on the Android emulator (2026-10-30) proved this, with one catch: Wails' Android asset path sends every request to Go as a `GET` with no body, and drops every status except 200.

- `cmd/mobile` (build tag `android`) wraps `internal/adapters/mobile`, which, like `desktop`, doesn't import Wails. `mobile.Open` calls `desktop.Open` with the app's private storage as the settings folder, so the data folder is `<storage>/starter` and `STARTER_DATA_DIR` still wins.
- **Remote mode on Android comes from a built-in address.** Nobody can edit the app's settings on a phone, so a build sets `config.BuiltInServerURL` (`BUILT_IN_SERVER_URL` in the Taskfile). `mobile.Open` keeps the session token in a file in the app's private storage (`credential_store = file`, whatever the settings say), as Android has no Secret Service, and only the app can read that storage.
- **`/api` goes over the Wails bridge.** On Android the web app's `fetch` hands each API request to one bound Go method, `main.Bridge.Serve(method, path, body)`. It runs the request through the same `http.Handler` (`mobile.Serve`) and returns the status, content type, and body, which the web app turns back into a `Response`. This is a transport for the REST contract, not a second API: the bridge has one method, and every route, status and error comes from `httpapi`. In browsers and on the desktop the web app keeps using plain `fetch`, and loads no Wails code.
- **`/api` on the asset path answers 404**, so a request that misses the bridge fails loudly. In the spike, a `POST` that fell through was answered as a `GET` with 200, and the web app reported a save that never happened.
- The acceptance tests run every feature through `mobile.Serve`, as a fifth driver, so the bridge is covered without a phone. The Android project is Wails' generated Gradle project, committed under `mobile/android`, and the Taskfile builds it with the NDK and Gradle directly. Setup, debugging, and upgrade steps are in `docs/mobile.md`.
- iOS waits for a Mac. Its custom-scheme handler should pass methods and bodies through, so it may not need the bridge, but that's unverified.

Wails mobile is experimental (outside its stability promise), and the bridge depends on how its Android code behaves today. The fallback is Capacitor, remote-only: mature, but it needs CORS and an API base URL, and can't work offline without reimplementing the core.

## Consequences

- One UI and one contract: a feature built for the browser works in every shell, and the SPA's tests cover it.
- Building the desktop app needs system packages (GTK 4 and WebKitGTK 6.0 on Linux; see the README), and the Android app needs a JDK, the Android SDK and the NDK. Plain `go build ./...` and `task check` need neither.
- The desktop app locks its data folder like the server does, so a second copy, or the CLI against the same folder, fails fast, saying the data folder is in use.
- Wails upgrades need care on Android: the Go module and the `@wailsio/runtime` npm package must move together, and the generated Gradle project may need regenerating.
- Packaging (AppImage, .deb, .dmg, .msi), iOS, and Play Store signing are follow-ups in `TODO.md`.
- In remote mode the app needs a connection, and `GET /api/v1/connection` says so, so the offline notice comes back. Signing in through the phone's browser on Android, and coming back to the app, is still to be tried on a phone.
