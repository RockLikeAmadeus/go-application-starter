# WIP: Install the web app

- Spec: [SPEC.md §4.5](../../SPEC.md#45-install-the-web-app-web-only)
- Scenarios: [`web/e2e/features/install.feature`](../../web/e2e/features/install.feature) · steps: [`web/e2e/steps/install.steps.ts`](../../web/e2e/steps/install.steps.ts) (reuses "I see the navigation" from `phone.steps.ts` and "I open the app for the first time" from `theme.steps.ts`)
- Parent plan: step 2 of the native front ends work (phone layout → **installable web app** → Wails desktop → Capacitor mobile).

Run: `task e2e -- --project chromium --grep @wip` (`task e2e` rebuilds the SPA first, which the service worker needs).

## Decisions
- **Updates:** quietly. `vite-plugin-pwa` with `registerType: "autoUpdate"`: the new service worker takes over without a prompt, so there's no update UI to build. No scenario covers it (it needs two builds); check it by hand once (task 6).
- **Offline:** the app shell opens offline and says it is offline; it never shows data from a stored copy.
  - The service worker precaches only the build output (`globPatterns`) and has **no `runtimeCaching`**, so `/api/*` always goes to the network.
  - `navigateFallback: "index.html"` serves client-side routes offline. `navigateFallbackDenylist` excludes `/api/`, `/healthz` and `/readyz`.
  - There is no scenario for "never an old copy": offline, TanStack Query *pauses* requests (`networkMode: "online"`), so no request ever reaches the service worker and such a scenario passes whatever the cache does. The config above is what enforces the rule.
  - While paused, the Items page still shows "Loading…" under the offline notice. Acceptable for now. TanStack refetches by itself when the connection returns.
- **Offline notice:** an app-wide component (not sample code) rendered by `Layout`, driven by `navigator.onLine` and the `online`/`offline` events (`useSyncExternalStore`). It is a `role="status"` live region that is always in the page and holds the message only while offline, so screen readers announce the change.
- **Manifest:** `name` and `short_name` are the brand ("Starter"), `display: "standalone"`, `start_url: "/"`, and `theme_color`/`background_color` come from the palette (`web/src/styles/palette.css`). `scripts/rename` already rewrites the whole-word brand in every tracked text file, so `vite.config.ts` needs no new rename code. Confirm with `task rename -- --name "Acme" --dry-run` (task 7).
- **Icons:** one source, `web/public/icon.svg` (the logo mark: the rounded square in the sidebar accent). `@vite-pwa/assets-generator` makes the PNGs (64, 192, 512, maskable 512, apple-touch 180) through a `pnpm icons` script. The PNGs are committed, so builds don't depend on the generator. A new project replaces the SVG and runs `pnpm icons`.
- **Content type:** Go's `mime` reads the freedesktop `globs2` table when it exists (and then skips `/etc/mime.types`). Neither that table, Go's built-in table nor the distroless image knows `.webmanifest`, so the server sends `text/plain`. Measured: `mime.TypeByExtension(".webmanifest")` is `""` on this machine. Fix it in `httpapi`, demanded by a Go test, so it doesn't depend on the host. `.js`, `.png` and `.svg` are in Go's built-in table.
- **Where the scenarios run:** untagged, so on desktop Chromium only. Playwright's service-worker support is Chromium-only, and installing doesn't depend on screen size.
- Config-only changes (the plugin, the manifest, the service worker, the icons, `index.html` links) can't be seen by a unit test. The failing e2e scenarios demand them, as the CSS did in the phone feature.

## Tasks
1. [x] Go: `NewHandler` serves `/manifest.webmanifest` from the web assets with `Content-Type: application/manifest+json`. First test file for the package: `internal/adapters/httpapi/server_test.go` (black-box, `fstest.MapFS` as `WebAssets`). Likely green: register the type for `httpapi` (`mime.AddExtensionType`) or set it in `spaHandler`.
   - Red: `server_test.go` › `TestServesWebAppManifestAsManifestJSON` fails with `Content-Type = "text/plain; charset=utf-8", want "application/manifest+json"`. The file is served (200); only its type is wrong, because the file server sniffed the JSON as plain text.
   - Green: in `spaHandler`, a request whose path ends in `.webmanifest` gets `Content-Type: application/manifest+json` before the file server runs (the file server keeps a type that is already set). `task test:go` passes. No Godog `@wip` scenarios exist for this feature; the e2e "can be installed" scenario still fails on the missing manifest link (task 2).
   - Refactor: added a comment in `spaHandler` saying why the type is set by hand: no MIME table Go reads, nor the distroless image, knows `.webmanifest`. Deliberately left:
     - Checking the type after the `index.html` fallback, or registering it with `mime.AddExtensionType`. Either would change what a request for a *missing* `.webmanifest` gets (today: `index.html` labelled as a manifest), which is a behaviour change, not a refactor. Nothing requests such a file, so it's not worth its own red/green cycle unless that changes.
     - An extension → type map: one entry doesn't justify it.
2. [ ] Web: add `vite-plugin-pwa` with the manifest (see Decisions), `public/icon.svg`, the `pnpm icons` script and the generated PNGs, plus a `theme-color` meta tag and an `apple-touch-icon` link in `index.html` (iOS ignores manifest icons). Demanded by "The app can be installed".
3. [ ] Web: the service worker: `autoUpdate`, precache the build, the navigation fallback with its deny list, no runtime caching. Demanded by the `Given I have opened the app before` step of both offline scenarios. Expect them to fail next on the missing offline notice.
4. [ ] `OfflineNotice` shows "You're offline…" in a `status` region when the browser is offline at first render (stub `navigator.onLine`).
5. [ ] `OfflineNotice` shows the message when an `offline` event fires, and clears it when an `online` event fires.
6. [ ] `Layout` renders the offline notice. Then:
   - run the `@wip` scenarios, and the full `task e2e` (the service worker must not break the other 51);
   - remove the `@wip` tags;
   - check "updates quietly" by hand: build, open, change something, rebuild, then reopen twice.
7. [ ] Housekeeping:
   - `task rename -- --name "Acme" --dry-run` shows the manifest name renamed;
   - CLAUDE.md: the service worker never caches `/api`, and icons come from `pnpm icons`;
   - delete this file.

## Changes to the plan
