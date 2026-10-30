# Plan: native front ends (phone layout, installable web app, desktop, mobile)

- TODO item: "create a matching PWA-based native Android and iOS app project, as well as a standalone desktop client" in [`TODO.md`](../../TODO.md).
- Temporary: delete this file in the same change that ticks that TODO item.
- Each step is its own `/feature` → `/red` → `/green` → `/refactor` cycle with its own `docs/wip/<slug>.md`. Stop for review between steps.

## Decisions
- **One UI.** Every shell runs the existing SPA (`web/`) unchanged.
- **Desktop: Wails v3** (v3.0.0 is stable). The SPA's normal `/api/*` requests go to `httpapi.NewHandler` *in-process*, through Wails' asset handler. No port is opened, and `api/openapi.yaml` stays the single contract.
  - Rejected: Wails bindings, because they would be a second API surface beside REST. Electron and Tauri were rejected because they need a second runtime or language next to Go.
- **Mobile: Capacitor** wrapping `web/dist`, **remote-only** for now: the app talks to a server. Running the core on the device (gomobile) waits for the local-only/accounts TODO item.
- **iOS:** not tested until a real iPhone is available. Playwright WebKit on an iPhone viewport stands in for Safari. Commit `mobile/android` only; add `mobile/ios` (`cap add ios`) once there's a Mac.
- **Order:** phone-width layout first, then the installable web app, then desktop, then mobile.
- **This machine:** KVM is available (`/dev/kvm`) for the Android emulator. The shell's default Node is v22, but the repo needs ≥ 24 (`nvm use`).

## Steps

### 1. Phone-width SPA layout — done
SPEC.md §4.4. `MobileNav` drawer in `web/src/routes/Layout.tsx`, `@phone` e2e projects (Pixel, iPhone/WebKit, phone-sized Firefox). Follow-up still open in TODO.md: run `task test:init` to confirm `--remove-sample` still works.

### 2. Installable web app — done
SPEC.md §4.5. `vite-plugin-pwa` (manifest, icons from `web/public/icon.svg` via `pnpm icons`, and a service worker that never caches `/api`). The server sends `.webmanifest` with the right type. `OfflineNotice` appears on every page. Manual testing: [`docs/installable-app.md`](../installable-app.md). Polish follow-ups are in TODO.md ("Installed-app polish").

### 3. Wails v3 desktop app — in progress ([`docs/wip/desktop-app.md`](../wip/desktop-app.md))
- `docs/adr/0004-native-shells.md` covers both shells: the in-process REST transport, the rejected options above, and remote-only mobile for now.
- **A testable adapter with no Wails dependency:** `internal/adapters/desktop/`.
  - `DataDir(env)` returns `STARTER_DATA_DIR` if set, otherwise `<os.UserConfigDir()>/starter`.
  - `NewHandler(app *bootstrap.App, log)` reuses `httpapi.NewHandler(httpapi.Options{Items: …, WebAssets: web.Assets()})`.
  - Black-box tests: GET `/api/items` through the handler, and `DataDir`'s precedence.
- **A thin entrypoint:** `cmd/desktop/main.go`, behind `//go:build desktop`.
  - Flow: `config.Load` → override `DataDir` → `bootstrap.NewEmbedded` → `application.New(application.Options{Assets: application.AssetOptions{Handler: handler}})`, plus one window.
  - The build tag keeps Wails' cgo/GTK dependency out of `go test ./internal/...` and `golangci-lint ./...`.
- Optional remote mode: `STARTER_SERVER_URL` → `bootstrap.NewRemote`, still served through the same in-process handler.
- **Taskfile:**
  - `run:desktop`: `go run -tags desktop,embedweb ./cmd/desktop`, depends on `web:build`.
  - `build:desktop`: builds `bin/starter-desktop`.
  - Packaging (AppImage, .deb, .dmg, .msi via the `wails3` CLI in `tools/go.mod`) is a follow-up, unless the CLI makes it trivial.
- **Linux build dependencies:** `libgtk-3-dev libwebkit2gtk-4.1-dev`. Document them in the README and in CLAUDE.md's Gotchas, and install them in CI if CI builds the desktop app.
- **Gotcha to document:** the desktop app locks its data directory, as the server does.
- **Acceptance:** the core is unchanged, so Godog is too. Verify by hand:
  - `task run:desktop`, then create, rename and delete an item;
  - restart, and check the data persists;
  - check no TCP port is opened (`ss -ltnp`).
- The service worker shouldn't run inside Wails (the app is served in-process and is always "installed"). Check whether `registerSW.js` misbehaves in the webview; if so, skip registration there.

### 4. Capacitor mobile app, remote mode (`/feature mobile app`)
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
- **Rename:** `scripts/rename` and `scripts/internal/project` must cover the Capacitor appId/appName, the Wails window title and `bin/starter-desktop`.
  - The whole-word brand rewrite already covers anything that literally says "Starter" (it did for the manifest).
  - Check the appId (`com.example.starter`) with `task rename -- --name "Acme Tasks" --dry-run`, then run `task test:rename` and `task test:init`.
  - Don't put quoted block-marker strings in prose.
- **SPEC.md:** §2 rows for desktop and mobile.
- **CLAUDE.md:**
  - Architecture lines for `cmd/desktop`, `internal/adapters/desktop` and `mobile/`;
  - Commands rows;
  - Gotchas.
- **TODO.md:** tick the item and delete this file.

## Verification after every step
`task check` (lint, unit, acceptance, generated-code drift and e2e).

## Changes to the plan
- Step 1: added a phone-sized Firefox e2e project. Playwright can't emulate a mobile device in Firefox.
- Step 2: updates install quietly, with no prompt; offline, the app opens and says so. There's no "never an old copy" scenario, because TanStack Query pauses its requests while offline. The service worker's configuration enforces that rule.
- Step 3: the desktop app gets Godog scenarios after all (`features/desktop.feature`, `@desktop`) and its own acceptance driver, which runs every feature through `desktop.Open`'s handler. The adapter's seam is `desktop.Open(env)` instead of `NewHandler(app, log)`. Remote mode and packaging are deferred to TODO.md. The SPA must hide the offline notice in the desktop shell.
