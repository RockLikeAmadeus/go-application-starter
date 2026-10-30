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
2. [x] Web: add `vite-plugin-pwa` with the manifest (see Decisions), `public/icon.svg`, the `pnpm icons` script and the generated PNGs, plus a `theme-color` meta tag and an `apple-touch-icon` link in `index.html` (iOS ignores manifest icons). Demanded by "The app can be installed".
   - Red: the e2e scenario "The app can be installed" (the page links no manifest).
   - Green:
     - `vite-plugin-pwa@2.0.0` and `@vite-pwa/assets-generator@2.0.0` (dev dependencies).
     - `web/public/icon.svg`: a plum rounded square with a blush inner square, echoing the sidebar's logo mark.
     - `web/pwa-assets.config.ts` (`minimal2023Preset`) and a `pnpm icons` script. They generated, into `web/public/`: `pwa-64x64.png`, `pwa-192x192.png`, `pwa-512x512.png`, `maskable-icon-512x512.png`, `apple-touch-icon-180x180.png` and `favicon.ico`.
     - `VitePWA` in `web/vite.config.ts` with the manifest written out by hand, `theme_color`/`background_color` `#e6eaee` (frost, the light theme's background), and **`injectRegister: false`**. The plugin still emits `sw.js` but nothing registers it, so the service worker stays task 3's job.
     - `index.html`: `theme-color` meta, favicon (`.ico` + `.svg`) and `apple-touch-icon` links.
   - Results:
     - "The app can be installed" passes. The two offline scenarios still fail as before: no service worker controls the page.
     - Web unit tests (42), lint, typecheck and the format check pass.
   - Refactor:
     - `vite.config.ts`: the manifest's two colours are one named constant, `appBackground`, with a comment saying it is frost from `palette.css` and that `index.html`'s `theme-color` meta repeats it.
     - `public/icon.svg`: a comment saying it is the source for every icon (run `pnpm icons` after changing it) and that its colours come from `palette.css`.
     - Deliberately left:
       - The plugin's `pwaAssets` option, which would generate the icons at build time and derive the manifest list and `index.html` links. It would undo the decision to commit the PNGs, put `sharp` into every build (including the Alpine Docker stage), and change the emitted `index.html`. That's a decision to make on its own, not a tidy-up.
       - The colours can't be shared with `palette.css` at build time without new code. Comments keep the copies findable.
     - Possible follow-up (a behaviour change, so not done here): a dark-scheme `theme-color` meta.
3. [x] Web: the service worker: `autoUpdate`, precache the build, the navigation fallback with its deny list, no runtime caching. Demanded by the `Given I have opened the app before` step of both offline scenarios. Expect them to fail next on the missing offline notice.
   - Red: both offline scenarios time out in `Given I have opened the app before`, because no service worker controls the page.
   - Green: in `VitePWA`, `injectRegister: false` became `registerType: "autoUpdate"`. With the default `injectRegister: "auto"`, the plugin adds a `registerSW.js` script to `index.html`. `workbox.navigateFallbackDenylist` excludes `/api/`, `/healthz` and `/readyz`. The plugin already defaults `navigateFallback` to `index.html`, and there's no `runtimeCaching`. The built `sw.js` has `NavigationRoute(…"index.html", {denylist: […]})` and precaches 9 entries.
   - Results:
     - "The app can be installed" passes.
     - "The app opens without a connection" gets past the offline reload and shows the navigation. It now fails only on `the app says it is offline` (no `status` element yet), the right reason for tasks 4–6.
     - "The offline notice goes when the connection comes back" **passed vacuously**: its only `Then` was a negative, and no notice exists yet. At the user's request it now asserts `And the app says it is offline` before `When the connection comes back`, so it fails on the missing notice like the other offline scenario.
     - Web unit tests (42) pass.
   - Noticed, not done (no scenario demands it): the default `globPatterns` (`**/*.{js,css,html}`) leave out the `.woff2` fonts and the icons. Offline, the app falls back to system fonts until they're cached. If wanted, it's a behaviour change for its own step.
   - Refactor (`web/vite.config.ts`):
     - The plugin setup moved out of `defineConfig` into a `pwa` constant, with a comment on what it does (manifest, caches the built app but not the API, updates quietly) pointing at SPEC.md §4.5. `plugins` is back to one line.
     - The deny list now has a comment saying why: the API must always reach the server, and so must the health probes.
     - Checked after the change: the built `sw.js` has the same deny list, and the `@wip` scenarios are unchanged (1 passes; the 2 offline ones fail on the missing notice). `task lint`, `task test` and `task accept` pass.
     - Deliberately left: a separate `web/pwa.config.ts`. One constant in the same file is enough at this size.
4. [x] `OfflineNotice` shows "You're offline…" in a `status` region when the browser is offline at first render (stub `navigator.onLine`).
   - Red: `web/src/features/offline/OfflineNotice.test.tsx` › "says the app is offline when the browser starts without a connection" (`vi.spyOn(navigator, "onLine", "get")` → `false`) fails with `Unable to find an accessible element with the role "status"`. The compile stub `OfflineNotice` in `OfflineNotice.tsx` returns `null`. It's a new `features/offline/` folder: the notice is app-wide, not sample code.
   - Green: `OfflineNotice` renders `<p role="status">` that holds "You're offline. Data will load when the connection comes back." while `navigator.onLine` is false. It reads `navigator.onLine` once per render; it doesn't listen for changes yet (task 5). Web suite: 43 pass. No Godog `@wip` scenarios; the e2e offline scenarios still need `Layout` to render it (task 6).
   - **For refactor:**
     - The paragraph has no styling: no palette colour or size. Probably wait for task 6, where it gets a place in `Layout` and its look can be judged in context. Styling doesn't change what the test sees.
     - The message is an inline string inside JSX. Fine while it's the only one.
5. [x] `OfflineNotice` shows the message when an `offline` event fires, and clears it when an `online` event fires.
   - 5a Red (offline event): `OfflineNotice.test.tsx` › "says the app is offline as soon as the connection drops" fails. It renders online, flips the `navigator.onLine` stub to `false`, and dispatches `offline` on `window` inside `act`. The `status` region exists but is empty: nothing re-renders the component on the event. Task 4's **For refactor:** notes are still open; they fit task 6.
   - 5a Green: `OfflineNotice` reads `navigator.onLine` through `useSyncExternalStore`. A module-level `subscribe` listens only for `offline` on `window`, since the `online` event isn't demanded yet (5b). Web suite: 44 pass; lint clean.
   - 5b Red (online event): `OfflineNotice.test.tsx` › "clears the notice as soon as the connection comes back" fails. It renders offline, flips the stub to `true`, and dispatches `online` inside `act`. The `status` region still holds "You're offline. Data will load when the connection comes back.", because `subscribe` listens only for `offline`.
   - 5b Green: `subscribe` also listens for `online` on `window` and removes both listeners on cleanup. Web suite: 45 pass; lint clean.
   - Refactor:
     - `features/offline/useOnline.ts`: a `useOnline()` hook now holds the subscription and the `useSyncExternalStore` call, named like `useTheme`/`useStyle`. `OfflineNotice` is just markup.
     - `subscribe` loops over one `connectionEvents` list (`online`, `offline`) to add and remove its listeners, so the event names appear once.
     - Tests: a `connection(online)` helper stubs `navigator.onLine` and returns a `change(nowOnline)` function that flips the stub and dispatches the matching event inside `act`. The two event tests are now three lines each.
     - `task lint`, `task test` (45 web) and `task accept` pass.
     - Deliberately left:
       - No separate `useOnline.test.ts`: the three component tests already drive every path through the hook, and new tests would be new assertions, not a refactor.
       - Task 4's styling note waits for task 6, as planned.
6. [ ] `Layout` renders the offline notice. Then:
   - Red: `Layout.test.tsx` › "tells the user when the app is offline, on every page" (`navigator.onLine` stubbed `false`) fails with `Unable to find an accessible element with the role "status"`. It sits outside the sample block, since the notice isn't sample code. `afterEach` now also calls `vi.restoreAllMocks()` to undo the `navigator.onLine` spy.
   - Green: `Layout` renders `<OfflineNotice />` in `<main>`, right after `<MobileNav>` and before the page's `Outlet`, so it shows on every page. Web suite: 46 pass; lint clean.
   - Results:
     - All three `@wip` install scenarios pass on Chromium.
     - Full `task e2e`: 54 passed (the 51 from before plus these 3), so the service worker breaks nothing else.
   - Refactor (`OfflineNotice.tsx`, also closing task 4's styling note):
     - The `status` region is now an always-present `<div>` that holds a styled `<p>` only while offline. Empty, it takes no space (measured: 0px tall online, 34–38px offline), and it stays in the page so screen readers announce the message when it appears.
     - The `<p>` is a quiet banner made only of palette and style tokens: `rounded-control border border-line bg-sunken px-4 py-2 text-sm text-muted`, `mb-6` (`terminal:mb-4 terminal:text-xs`, matching the Menu button's spacing).
     - A doc comment explains the empty-region choice.
     - Checked offline in Crisp, Terminal and Soft × light and dark (desktop Chromium screenshots): readable everywhere, and it doesn't compete with the page heading.
     - `task lint`, `task test` (46 web), `task accept` and the full `task e2e` (54) pass.
     - Deliberately left: a shared `navigator.onLine` stub helper across `Layout.test.tsx` and `OfflineNotice.test.tsx`. It's one line in `Layout.test.tsx`.
   - Still to do in this task, after `/refactor`:
     - remove the `@wip` tags;
     - check "updates quietly" by hand: build, open, change something, rebuild, then reopen twice.
7. [ ] Housekeeping:
   - `task rename -- --name "Acme" --dry-run` shows the manifest name renamed;
   - CLAUDE.md: the service worker never caches `/api`, and icons come from `pnpm icons`;
   - delete this file.

## Changes to the plan
- Task 3: "The offline notice goes when the connection comes back" passed before any notice existed. Added `And the app says it is offline` to its setup, so it now proves the notice appears and then goes.
