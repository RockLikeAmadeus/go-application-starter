# WIP: mobile app

- Spec: [SPEC.md §4.7 Use the mobile app](../../SPEC.md#47-use-the-mobile-app), plus the §2 row and the glossary ("Mobile app", "Data folder").
- Scenarios: [`features/mobile.feature`](../../features/mobile.feature) (`@mobile @wip`), plus every other feature except `desktop.feature`, run through the new mobile driver.
- Parent plan: [`docs/plans/native-front-ends.md`](../plans/native-front-ends.md), step 4. Its "Spike: Wails v3 on Android" section has the evidence behind every decision below.
- Delete this file in the same change that removes the `@wip` tag.

## Decisions
- **Wails v3 on Android, local-only.** The phone runs `desktop.Open`'s wiring: the core, jsonfile storage and `httpapi.NewHandler` with `LocalData`, serving the built web app in-process. iOS, remote mode, release signing and an app icon are out of scope (TODO.md follow-ups at the end).
- **Why a bridge:** Wails' Android asset path sends every request to Go as `GET` with an empty body (`bridge.serveAsset(path, "GET", "{}")`), and treats any status but 200 as "not found". So the web app's `/api` requests go over the Wails bridge instead: a Go service runs each request through the same `http.Handler` and returns the status, content type and body.
- **Testable wiring lives in a new `internal/adapters/mobile`, with no Wails import**, like `desktop`. `cmd/mobile/main.go` (`//go:build android`) only wraps it: it registers the Android main, adds a one-line Wails service around `mobile.Serve`, and opens one window.
- **Two handlers.** `mobile.Open` returns:
  - `Handler`, for Wails' asset path: `/index.html` serves `/` (Android asks for it and Go's file server would answer with a 301), and `/api/*` answers **404**, so a request that misses the bridge fails loudly. The spike showed what happens otherwise: a `POST` turned into a `GET` that answered 200, and the web app showed success while nothing was saved.
  - `API`, the plain handler, for the bridge.
- **Data folder:** `desktop.Open` hard-codes `os.UserConfigDir`, so it gains the lookup as a parameter. The phone passes the app's private storage (`application.Android.StoragePath()`); `STARTER_DATA_DIR` still wins, as on the desktop. The folder is `<storage>/starter`.
- **Web app:** `web/src/api/client.ts` checks for Android's bridge (`window.wails.invokeAsync`) **once, when the module loads**, because importing the Wails runtime replaces `window.wails`. When it's there, requests go through `bridgeFetch`, which calls `mobile.Serve` via the runtime's `Call.ByName` and rebuilds a normal `Response` (with no body for 204). Make `bridgeFetch` take the call function as an argument, so Vitest can pass a fake instead of mocking modules. The `@wailsio/runtime` npm package is pinned to the Go module's version (3.0.0; check it exists on npm) and imported dynamically, so the browser bundle doesn't grow.
- **Acceptance:** a `mobile` driver opens `mobile.Open` on a temporary data folder and drives it with `httpclient` over an `http.RoundTripper` that calls `mobile.Serve`. So every feature's scenarios cross the same bridge as the phone. Tag routing:
  - `memory`, `jsonfile`, `http`: `~@desktop && ~@mobile`
  - `desktop`: `~@mobile`
  - `mobile`: `~@desktop`
- **Steps:** "I close the mobile app and open it again" reuses the harness's `reopen` hook. The scenario is inside sample markers, like the desktop one.
- **Android project:** Wails' generated Gradle project, committed under `mobile/android/`, with `applicationId` `com.example.starter` and the app name "Starter". `wails3` goes in `tools/go.mod` so the project can be regenerated on Wails upgrades. Our Taskfile builds it directly (NDK cross-compile, then `gradlew`), not with `wails3 task`, because our layout isn't a Wails project.
- **This machine:** the Android SDK is in `~/Android/Sdk` (platform-tools, android-35, build-tools 34 and 35, NDK 26.3.11579264, emulator, an x86_64 system image) and the AVD is `spike` (Pixel 7). JDK 21 is installed.
- **Wails mobile is experimental.** If it breaks or stalls, the fallback is the Capacitor plan in the parent plan.

## Tasks
1. [ ] **`desktop.Open` takes the settings-folder lookup** (`os.UserConfigDir` on the desktop, the app's storage on the phone). Red: with no `STARTER_DATA_DIR`, `Open` creates the data folder at `<lookup>/starter`.
2. [ ] **`mobile.Serve(h, method, path, body) Reply`** runs one request through `h` and returns its status, content type and body. Red: a `POST` with a JSON body reaches the handler with its method, body and `Content-Type: application/json`, and its 201 reply comes back.
3. [ ] **`mobile.Serve`** returns an error reply's status and problem document unchanged (e.g. 422), and a query string reaches the handler.
4. [ ] **Android asset handler:** `/index.html` serves what `/` serves.
5. [ ] **Android asset handler:** `/api/*` answers 404, and other paths pass through.
6. [ ] **`mobile.Open(getenv, storagePath, webAssets)`** returns `Handler` (the asset handler from items 4–5) and `API` (for the bridge), with `LocalData` set. Red: `GET /api/v1/connection` through `Serve` on `API` says no connection is required, and the same path on `Handler` is a 404.
7. [ ] **Mobile acceptance driver and tag routing** (test code). Turns `features/mobile.feature` green and runs every other non-desktop feature over the bridge.
8. [ ] **Web: `bridgeFetch(call)`** sends the method, path with query, and body, and returns a `Response` with the reply's status, content type and body.
9. [ ] **Web: `bridgeFetch`** returns a body-less `Response` for 204.
10. [ ] **Web: `client.ts` uses the bridge** when `window.wails.invokeAsync` exists when the module loads, and plain `fetch` otherwise (both cases tested, re-importing the module).
11. [ ] **`@wailsio/runtime`**, dynamically imported, passed to `bridgeFetch` as `Call.ByName("main.Proxy.Do", …)` or whatever name the service gets. Wiring: check the browser bundle doesn't grow.
12. [ ] `cmd/mobile/main.go` (`//go:build android`) and the Wails service wrapper. Wiring, no unit test. Builds with the NDK.
13. [ ] `mobile/android/`: the generated Gradle project, with `applicationId` and app name set. `wails3` in `tools/go.mod`.
14. [ ] Taskfile `mobile:android:build` and `mobile:android:run`.
15. [ ] Docs:
    - `docs/mobile.md`: SDK setup with the exact `sdkmanager` packages, the licences, the emulator, `chrome://inspect`, plus the spike's lasting findings (the bridge, the 404 guard, the `window.wails` trap).
    - The README Prerequisites row.
    - CLAUDE.md: architecture, Commands rows, Gotchas.
    - ADR 4's mobile section, from "planned" to "built".
16. [ ] Rename/init: the dry run covers the `applicationId` and the app name; `task test:rename` and `task test:init` (the fenced scenario; anything in `mobile` that uses `Item`).
17. [ ] Manual check on the emulator: create, rename, delete, the empty-name error, persistence after killing the app, airplane mode with no notice and saving working, and the phone-width drawer.
18. [ ] Remove `@wip`. Add follow-ups to TODO.md (iOS, mobile remote mode, Play Store release builds and signing, app icon). Mark step 4 done in the parent plan, delete this file, and `task check` green.

## Changes to the plan
