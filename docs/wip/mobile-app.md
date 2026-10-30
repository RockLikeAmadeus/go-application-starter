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
1. [x] **`desktop.Open` takes the settings-folder lookup** (`os.UserConfigDir` on the desktop, the app's storage on the phone). Red: with no `STARTER_DATA_DIR`, `Open` creates the data folder at `<lookup>/starter`.
   - Red ✔: `TestOpenCreatesTheDataFolderUnderTheGivenSettingsFolder` (`open_test.go`). `Open` now takes `userConfigDir` as its second parameter but still ignores it; every caller passes `os.UserConfigDir`.
   - Green ✔: `Open` passes `userConfigDir` to `DataDir`, and no longer imports `os`.
   - Refactor ✔: dropped the new test's `XDG_CONFIG_HOME` guard, which it no longer needed. `openWith` and `TestOpenKeepsItemsAfterClosingAndReopening` now pass a fake `configDirAt(t.TempDir())`, so no unit test sees the real settings folder. `DataDir`'s comment now says `os.UserConfigDir` is the desktop's lookup, not "production"'s. Left alone on purpose: the acceptance drivers still pass `os.UserConfigDir`, to mirror `cmd/desktop`'s wiring (`STARTER_DATA_DIR` always wins there). No `t.Parallel`: the repo doesn't use it anywhere.
2. [x] **`mobile.Serve(h, method, path, body) Reply`** runs one request through `h` and returns its status, content type and body. Red: a `POST` with a JSON body reaches the handler with its method, body and `Content-Type: application/json`, and its 201 reply comes back.
   - Red ✔: `TestServeSendsTheRequestAndReturnsTheReply` (`internal/adapters/mobile/serve_test.go`). New package `mobile`, with a stub `Serve` that returns an empty `Reply{Status, ContentType, Body}`. `Body` is a string, with JSON tags, so it crosses the Wails bridge as plain JSON.
   - Green ✔: `Serve` builds the request with `httptest.NewRequest` (setting `Content-Type: application/json`), records the reply with `httptest.NewRecorder`, and returns its status, `Content-Type` and body.
   - Refactor ✔: production code unchanged, since it's already short and clear. The test now keeps the `*http.Request` and its body instead of four separate strings, and fails with "the handler was not called" instead of a nil-pointer panic if `Serve` stops calling it. Left alone on purpose: `httptest.NewRequest` can panic, which is moved to the new task 3a because fixing it changes behaviour. `Content-Type` is set on every request, which is harmless.
3. [x] **`mobile.Serve`** returns an error reply's status and problem document unchanged (e.g. 422), and a query string reaches the handler.
   - Red ✘, passed at once: `TestServePassesTheQueryAndReturnsAnErrorReplyUnchanged` (`serve_test.go`). Task 2's `Serve` already does this (`httptest.NewRequest` parses the query, and the recorder returns any status, type and body). Kept as a regression test, as you decided, with no green or refactor step.
3a. [x] **`mobile.Serve`** answers a request it can't build (e.g. a malformed path) with a 400 reply instead of panicking. `httptest.NewRequest` panics, so this probably means switching to `http.NewRequest`. Red: a malformed path returns status 400.
   - Red ✔: `TestServeAnswersAMalformedPathWithBadRequest` (`serve_test.go`) uses `/api/v1/items/%zz`, an invalid escape. It recovers from the panic, so the failure is an assertion: "Serve panicked: invalid NewRequest arguments". It doesn't use a space, because `http.NewRequest` accepts one in the path and the test would then depend on how green builds the request.
   - Green ✔: `Serve` builds the request with `http.NewRequestWithContext(context.Background(), …)`, and returns `Reply{Status: 400}` when that fails. The handler isn't called.
   - Refactor ✔: removed the test's `recover` wrapper, which only turned the red panic into an assertion; a panic fails a Go test anyway. `Serve`'s doc comment now says a request it can't build gets a 400 without reaching `h`. Left alone on purpose: `httptest.NewRecorder` stays, because it's a plain in-memory writer that never panics. The empty `Host`/`RequestURI`/`RemoteAddr` is left for task 6 to check against the real `httpapi` handler. The 400 has no problem body, because no test asks for one.
4. [x] **Android asset handler:** `/index.html` serves what `/` serves.
   - Red ✔: `TestAssetHandlerServesTheWebAppForIndexHTML` (`internal/adapters/mobile/assets_test.go`) wraps `http.FileServerFS` over a one-file `MapFS`, which reproduces `httpapi`'s 301 for `/index.html`. Fails with "status 301, body \"\"". The new `mobile.AssetHandler(next http.Handler) http.Handler` (`assets.go`) is a stub that returns `next`, because returning nil would panic rather than fail an assertion. Task 5 adds to the same function.
   - Green ✔: `AssetHandler` rewrites `r.URL.Path` from `/index.html` to `/` before calling `next`.
   - Refactor ✔: `AssetHandler` now changes a clone of the request (`r.Clone`) instead of the caller's, as middleware should. Its doc comment explains the rule: Android asks for `/index.html`, the file server redirects, and Wails treats any status but 200 as "not found". Task 5 should add its `/api` rule to that comment. Left alone on purpose: `RawPath` is always empty for `/index.html`. `httpapi`'s `spaHandler` still changes the request in place, but that's outside this cycle.
5. [x] **Android asset handler:** `/api/*` answers 404, and other paths pass through.
   - Red ✔: `TestAssetHandlerAnswersAPIRequestsWithNotFound` (`assets_test.go`) fails with "status 200, body {\"items\":[]}". The pass-through half needs no red, because `AssetHandler` already hands every other path to `next`. Only `/index.html` is tested so far. A pass-through regression test (e.g. `/assets/app.js` reaches `next`) was suggested, as with task 3.
   - Regression ✔ (you asked for it): `TestAssetHandlerPassesOtherPathsThrough` checks that `/assets/app.js` reaches `next` unchanged and returns its reply. It passed at once.
   - Green ✔: `AssetHandler` answers any path starting with `/api/` with `http.NotFound`, before the `/index.html` rewrite and without calling `next`.
   - Refactor ✔: `AssetHandler`'s doc comment now lists both rules and why: the asset path makes every request a GET with no body and treats any status but 200 as "not found". Code unchanged. Left alone on purpose: a bare `/api` passes through, because nothing is served there. The two rules stay in one handler, because it's short and splitting it would add more than it saves.
6. [x] **`mobile.Open(getenv, storagePath, webAssets)`** returns `Handler` (the asset handler from items 4–5) and `API` (for the bridge), with `LocalData` set. Red: `GET /api/v1/connection` through `Serve` on `API` says no connection is required, and the same path on `Handler` is a 404.
   - Red ✔: `TestOpenAnswersTheAPIOverTheBridgeOnly` (`internal/adapters/mobile/open_test.go`) opens with an empty environment and `storagePath` = a temp folder, as the phone would. It fails with "Open returned API <nil> and Handler <nil>, want both". Stub in `open.go`: `App{Handler, API}`, `Open(getenv, storagePath string, webAssets fs.FS)` returning an empty `App`, and a no-op `Close`. `storagePath` is a string because `application.Android.StoragePath()` returns one.
   - Green ✔: `Open` calls `desktop.Open` with a lookup that returns `storagePath`. `API` is the desktop handler, `Handler` is `AssetHandler` around it, and `Close` closes the desktop app. This also confirms task 3a's open question: the real `httpapi` handler works with the empty `Host`/`RequestURI` that `http.NewRequest` leaves.
   - Refactor ✔: comments only, no code changes. `App`, its `Handler`/`API` fields and `Open` now say which handler is for the asset path and which for the bridge, where the data folder goes (`<storagePath>/starter` unless `STARTER_DATA_DIR`), what production passes in, and that `/sw.js` answers 404 as on the desktop. The package doc says `mobile` reuses `desktop`'s wiring because the phone keeps its data locally. Left alone on purpose: the `mobile` → `desktop` import stays, because a shared package would only pay off with a third user.
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
- 2026-10-30, task 2 refactor: added task 3a. `Serve` uses `httptest.NewRequest`, which panics on a malformed method or path and would crash the bridge call. Fixing that changes behaviour, so it needs its own red test rather than a refactor.
