# Plan: native front ends (phone layout, installable web app, desktop, mobile)

- TODO item: "create a matching PWA-based native Android and iOS app project, as well as a standalone desktop client" in [`TODO.md`](../../TODO.md).
- Temporary: delete this file in the same change that ticks that TODO item.
- Each step is its own `/feature` → `/red` → `/green` → `/refactor` cycle with its own `docs/wip/<slug>.md`. Stop for review between steps.

## Decisions
- **One UI.** Every shell runs the existing SPA (`web/`) unchanged.
- **Desktop: Wails v3** (v3.0.0 is stable). The SPA's normal `/api/*` requests go to `httpapi.NewHandler` *in-process*, through Wails' asset handler. No port is opened, and `api/openapi.yaml` stays the single contract.
  - Rejected: Wails bindings, because they would be a second API surface beside REST. Electron and Tauri were rejected because they need a second runtime or language next to Go.
- **Mobile: Wails v3 too, Android first, local-only first.** The phone runs the same Go core and REST handler in-process as the desktop app, so it works offline with its data on the device. A small bridge carries `/api` on Android (see the step 4 spike). Wails mobile is experimental, so the Capacitor plan (remote-only, with CORS) is kept as the fallback if it breaks or stalls. Decided after the spike on 2026-10-30.
  - Remote mode (cloud data) comes later, the same way as on the desktop (`bootstrap.NewRemote` behind the in-process handler), with no sync.
- **iOS:** not built or tested until there's a Mac with full Xcode, and ideally a real iPhone. Playwright WebKit on an iPhone viewport stands in for Safari. Commit the Android project only for now.
- **Order:** phone-width layout first, then the installable web app, then desktop, then mobile.
- **This machine:** KVM is available (`/dev/kvm`) for the Android emulator. The shell's default Node is v22, but the repo needs ≥ 24 (`nvm use`).

## Steps

### 1. Phone-width SPA layout — done
SPEC.md §4.4. `MobileNav` drawer in `web/src/routes/Layout.tsx`, `@phone` e2e projects (Pixel, iPhone/WebKit, phone-sized Firefox). Follow-up still open in TODO.md: run `task test:init` to confirm `--remove-sample` still works.

### 2. Installable web app — done
SPEC.md §4.5. `vite-plugin-pwa` (manifest, icons from `web/public/icon.svg` via `pnpm icons`, and a service worker that never caches `/api`). The server sends `.webmanifest` with the right type. `OfflineNotice` appears on every page. Manual testing: [`docs/installable-app.md`](../installable-app.md). Polish follow-ups are in TODO.md ("Installed-app polish").

### 3. Wails v3 desktop app — done
SPEC.md §4.6, [`features/desktop.feature`](../../features/desktop.feature), [ADR 4](../adr/0004-native-shells.md).
- `internal/adapters/desktop` (no Wails import): `DataDir` (`STARTER_DATA_DIR`, else `<user settings folder>/starter`) and `Open(getenv, webAssets)`, which wires jsonfile storage behind `httpapi.NewHandler` with `LocalData` set and `/sw.js` answered with 404.
- `cmd/desktop` (`//go:build desktop`): one Wails v3 window around that handler, so the SPA and its API run in-process and no port is opened. `task run:desktop` and `task build:desktop` (`bin/starter-desktop`).
- `GET /api/v1/connection` tells the SPA whether a connection is needed. In local mode the SPA hides the offline notice and keeps TanStack Query online (`useConnectionRequired`, `useLocalDataOnline`).
- A Godog `desktop` driver runs every feature through `desktop.Open`. `@desktop` scenarios cover reopening and the data-folder lock.
- Linux build deps are GTK 4 and WebKitGTK 6.0 (Wails v3's default; GTK 3 goes away in v3.1). They're documented in the README and CLAUDE.md.
- Checked by hand on Linux: create, rename and delete; the data persists after a restart; no listening socket; no offline notice, and saving works with the network off.
- Follow-ups in TODO.md: desktop remote mode, and packaging (including trying macOS and Windows).

### 4. Wails mobile app, Android, local-only (`/feature mobile app`)
The app runs on an Android phone with its data on the phone, and works offline. See the spike below for what was proven and why each piece exists.

- **Spec:** a SPEC.md §4.7 "Use the mobile app", like §4.6. It covers local data on the phone, offline use, and one copy per phone. Add an "Android phone" row to §2.
- **Testable wiring, without Wails** (TDD, new package `internal/adapters/mobile`, or shared with `desktop` if that reads better):
  - **Data folder:** the app's private storage (`application.Android.StoragePath()`). `desktop.Open` hard-codes `os.UserConfigDir`, so the folder lookup must be passed in. Reuse `DataDir`'s `STARTER_DATA_DIR` rule.
  - **Android handler:** serve `/` when asked for `/index.html` (Go's file server would redirect). **Answer `/api/*` on the asset path with 404**, so a request that misses the bridge fails loudly instead of turning a `POST` into a `GET` that "succeeds".
  - **Bridge:** `Serve(h, method, path, body) Reply`, which runs one request through the handler and returns the status, content type and body. The Wails service in `cmd/mobile` is a one-line wrapper around it.
- **Web app** (Vitest, stubbing the bridge rather than mocking modules):
  - `web/src/api/client.ts` sends requests over the bridge when Android's Wails bridge is present. That's checked **once, at load**, because importing the Wails runtime replaces `window.wails`. Status, body and content type come back as a normal `Response`. In the browser and on the desktop nothing changes.
  - Add the `@wailsio/runtime` npm package at the version matching the Go module (v3.0.0), loaded only on Android (dynamic import), so the browser bundle doesn't grow. Bumping Wails means bumping both.
- **Entrypoint:** `cmd/mobile/main.go` (`//go:build android`), wiring only, like `cmd/desktop`: it registers the Android main, opens the core, adds the bridge service and one window.
- **Android project:** Wails' generated Gradle project, committed under `mobile/android/`. Set `applicationId` to `com.example.starter` and the app name to "Starter". The `wails3` CLI goes in `tools/go.mod`, to regenerate the project when Wails is upgraded.
- **Taskfile:** `mobile:android:build` (web build, NDK cross-compile of `libwails.so`, `gradlew assembleDebug`) and `mobile:android:run` (start the emulator if needed, install, launch). Don't run `wails3 task`: our layout isn't a Wails project.
- **Docs:**
  - `docs/mobile.md`: the SDK install with the command-line tools (the exact `sdkmanager` packages from the spike, plus accepting the licences), the emulator, `adb`, and WebView debugging with `chrome://inspect`. Move the spike's lasting findings (why the bridge exists, the traps) here too, because this plan gets deleted.
  - README Prerequisites: an optional "Android build deps" row (Android SDK, NDK, JDK 21: `openjdk-21-jdk-headless`).
  - CLAUDE.md: architecture lines, Commands rows, Gotchas (the bridge and why, the 404 guard, the runtime version pinned to the Go module, and that Wails mobile is experimental).
- **ADR 4:** update its mobile section from "Capacitor, planned" to what was built.
- **Verify by hand on the emulator,** as for the desktop:
  - create, rename and delete items; an empty name shows the error;
  - kill and relaunch the app, and check the data persists;
  - in airplane mode, check there's no offline notice and saving works;
  - check the phone-width drawer.

  A scripted smoke check over Chrome DevTools, like the spike's, is optional. Android isn't built in CI for now (it needs the NDK).
- **Out of scope:** iOS (needs a Mac), remote mode, Play Store signing and release builds, and an app icon from `web/public/icon.svg`. All of these go to TODO.md as follow-ups.

#### Spike: Wails v3 on Android (2026-10-30)
Throwaway code in the scratchpad, not in the repo. It used a copy of the repo, `cmd/mobile` (`//go:build android`), Wails' generated Gradle project, and an Android 35 x86_64 emulator on KVM.

- **What worked:**
  - `desktop.Open(getenv, web.Assets())` compiled for Android with no changes: the core, jsonfile storage, the REST handler and the embedded web app. The data folder is `application.Android.StoragePath()/starter`, the app's private storage.
  - Through the real web app: list, create, rename and delete all worked. An empty name shows "name must not be empty" (our 422). Items survived the app being killed and relaunched.
  - In airplane mode, `navigator.onLine` is false, there's no offline notice, and saving works instantly. `GET /api/v1/connection` and `useLocalDataOnline` did their job unchanged.
  - Size: the release `libwails.so` for arm64 is 11 MB (about 5 MB compressed). The debug APK for the emulator was 17 MB.
- **What needed a shim:** Wails' Android asset path sends every request to Go as `GET` with an empty body (`bridge.serveAsset(path, "GET", "{}")`), and treats any status other than 200 as "not found". So:
  - **`/api` goes over the Wails bridge** instead. A Go service (`Proxy.Do(method, path, body)`, about 30 lines) runs the request through the same `http.Handler` with `httptest`. A `fetch` hook in `web/src/api/client.ts` (about 20 lines) calls it when Android's bridge (`window.wails.invokeAsync`) is present. Every status, body and content type comes back intact.
  - **`/index.html`:** Android requests it for `/`, and Go's file server answers with a 301, so the handler must serve `/` for it.
  - **The Wails runtime JS** can't be served at `/wails/runtime.js`, because Android treats every `/wails/runtime…` path as a runtime call. The spike served it at another path. A real version would bundle the `@wailsio/runtime` npm package.
- **Traps found:**
  - Importing the Wails runtime replaces `window.wails`, so the bridge check must run once, before the import.
  - Before that fix, a `POST` fell back to the asset path, was answered as a `GET` with status 200, and the web app showed success while nothing was saved. **A real version must make `/api` on the asset path fail loudly**, for example with a 404, which Android turns into a failed fetch.
  - The service worker registration fails harmlessly (`/sw.js` is a 404).
- **Not tested:** iOS (needs a Mac with full Xcode). Its custom-scheme handler should receive methods and bodies, so it may need no shim, but that's unverified. A real phone wasn't tested either, only the emulator.
- **Upkeep risks:** Wails mobile is experimental (outside its stability promise). The shim depends on Android behaviour in Wails that could change. In 3.0.0 the Android bridge is exposed to every frame (a fix is planned for 3.0.1; we embed no third-party content). The generated Gradle and Java project would live in the repo. Android builds need the NDK and a JDK as well as the SDK.
- **Compared with Capacitor:** no CORS, no API base URL and no second JS project. Local and remote modes work the same as on the desktop. Offline use on the phone comes for free, whereas Capacitor would need the core reimplemented, or gomobile, to get it.

#### Fallback: Capacitor, remote-only
The original plan, kept in case Wails mobile breaks or stalls. It needs server CORS and an API base URL in the web app, and it can't work offline without reimplementing the core.
- **Server CORS** (TDD in `httpapi`):
  - `Options.AllowedOrigins`, set from `STARTER_CORS_ORIGINS` (a comma-separated list; empty turns it off).
  - The middleware answers preflight requests, and sets `Access-Control-Allow-Origin` only for listed origins.
  - Capacitor's origins are `https://localhost` (Android) and `capacitor://localhost` (iOS).
  - Tests in `internal/adapters/httpapi` and `internal/config`.
- **SPA base URL:** in `web/src/api/client.ts`, `baseUrl = import.meta.env.VITE_API_BASE_URL || location.origin`. Unit-test it with a stubbed `fetch`.
- **Project:** `mobile/`, with its own `package.json` (`@capacitor/core`, `cli`, `android`, `ios`).
  - `capacitor.config.ts`: `webDir: "../web/dist"`, appId `com.example.starter`, appName `Starter`.
  - Dev setup: `cleartext` plus `http://10.0.2.2:8080`, for the emulator.
  - The service worker inside the Capacitor WebView: check how it behaves, and probably skip registering it there.
- **Taskfile:**
  - `mobile:sync`: build the web app with `VITE_API_BASE_URL`, then run `cap sync`.
  - `mobile:android`: `cap run android`.
- **`docs/mobile.md`**, the emulator setup guide:
  - Android Studio, the SDK, and a Pixel AVD on KVM;
  - `adb`, and Chrome's `chrome://inspect` for WebView debugging;
  - iOS options: a real iPhone later, a macOS CI runner, or Appetize/BrowserStack.
- **Verify:**
  - `task dev:server` with `STARTER_CORS_ORIGINS=https://localhost`;
  - `task mobile:android` on the Pixel AVD;
  - create, rename and delete items from the emulator, and check the drawer layout on the device.

### 5. Housekeeping (folded into steps 3–4 as they land)
- **Rename:** `scripts/rename` and `scripts/internal/project` must cover the Android `applicationId` (`com.example.starter`) and app name (the Wails window title and `bin/starter-desktop` are done: step 3).
  - The whole-word brand rewrite already covers anything that literally says "Starter" (it did for the manifest).
  - Check the appId (`com.example.starter`) with `task rename -- --name "Acme Tasks" --dry-run`, then run `task test:rename` and `task test:init`.
  - Don't put quoted block-marker strings in prose.
- **SPEC.md:** §2 rows for desktop and mobile.
- **CLAUDE.md:**
  - Architecture lines for `cmd/mobile`, `mobile/android` and the mobile adapter (desktop: done in step 3);
  - Commands rows;
  - Gotchas.
- **TODO.md:** tick the item and delete this file.

## Verification after every step
`task check` (lint, unit, acceptance, generated-code drift and e2e).

## Changes to the plan
- Step 1: added a phone-sized Firefox e2e project. Playwright can't emulate a mobile device in Firefox.
- Step 2: updates install quietly, with no prompt; offline, the app opens and says so. There's no "never an old copy" scenario, because TanStack Query pauses its requests while offline. The service worker's configuration enforces that rule.
- Step 3: the desktop app gets Godog scenarios after all (`features/desktop.feature`, `@desktop`) and its own acceptance driver, which runs every feature through `desktop.Open`'s handler. The adapter's seam is `desktop.Open(env)` instead of `NewHandler(app, log)`. Remote mode and packaging are deferred to TODO.md. The SPA must hide the offline notice in the desktop shell.
- Step 3 done. Changes along the way:
  - `desktop.Open(getenv, webAssets)` replaced `NewHandler(app, log)`.
  - Hiding the offline notice wasn't enough, because TanStack Query also pauses requests offline. So the server reports `GET /api/v1/connection`, and the SPA stays online when no connection is needed.
  - The service worker is turned off by a 404 on `/sw.js`, with no web change.
  - Linux deps are GTK 4/WebKitGTK 6.0, not GTK 3/WebKitGTK 4.1.
  - The desktop tests that use `Item` live in a sample file, so `--remove-sample` still works.
- Step 4: after a spike on 2026-10-30 (findings under step 4), the user chose **Wails mobile, Android first, local-only** over Capacitor, remote-only. That gives offline use on the phone for a small bridge. Capacitor is kept as the fallback. The Decisions section was updated to match.
