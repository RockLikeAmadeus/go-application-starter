# WIP: desktop app

- Spec: [SPEC.md §4.6 Use the desktop app](../../SPEC.md#46-use-the-desktop-app)
- Scenarios: [`features/desktop.feature`](../../features/desktop.feature) (`@desktop @wip`), plus every scenario in [`features/items.feature`](../../features/items.feature) through the new desktop driver.
- Parent plan: [`docs/plans/native-front-ends.md`](../plans/native-front-ends.md), step 3.
- Delete this file in the same change that removes the `@wip` tag.

## Decisions
- **Wails v3**, with the SPA's `/api/*` requests answered in-process by `httpapi.NewHandler` through Wails' asset handler. No port, no bindings, one REST contract. See the parent plan for the rejected options. Record it in `docs/adr/0004-native-shells.md`.
- **All testable wiring lives in `internal/adapters/desktop/`, with no Wails import.** `cmd/desktop/main.go` (`//go:build desktop`) only creates the Wails app and window around `desktop.Open`'s handler. The build tag keeps Wails' cgo/GTK dependency out of `go test ./internal/...`, `task lint` and CI.
- **`desktop.Open(env)` is the seam the acceptance tests drive.** It picks the data folder, opens jsonfile storage through `bootstrap.NewEmbedded`, and returns the app (to close) and the HTTP handler. The desktop driver serves that handler on `httptest` and talks to it with `httpclient`, so the scenarios exercise the same wiring as the real window. (The Wails asset handler is the only part left to the manual check.)
- **The desktop driver uses the real clock**, because `Open` wires `bootstrap.NewEmbedded`. The items steps' `clock.Advance` has no effect there; real nanosecond timestamps keep "oldest first" ordered.
- **Driver routing by tag:** `@desktop` scenarios run only on the desktop driver. The desktop driver runs every feature. Combine each driver's tag filter with the user's `--godog.tags` (`&&`).
- **"Data in use"** is `jsonfile.ErrLocked`, which already exists. The step asserts `errors.Is(err, jsonfile.ErrLocked)`.
- **Data folder:** `STARTER_DATA_DIR` if set, otherwise `<os.UserConfigDir()>/starter`. Pass the user-config-dir lookup in, so the fallback is unit-testable without touching `$HOME`.
- **Out of scope for this step:** remote mode (`STARTER_SERVER_URL` in the desktop app) and packaging (AppImage, .deb, .dmg, .msi). Both go to TODO.md as follow-ups when the feature finishes.
- **Offline notice:** the SPA's `OfflineNotice` uses `navigator.onLine`, so a desktop app with the network off would wrongly say it is offline. The SPA needs to know it is in the desktop shell. How it detects that (e.g. a Wails-injected global, or the `wails:` origin) is settled by a quick check at item 6, before its `/red`. The same check decides whether the service worker must be skipped (item 7).

## Tasks
1. [x] `desktop.DataDir`: `STARTER_DATA_DIR` wins when set.
   - Red: `TestDataDirUsesStarterDataDirWhenSet` in `internal/adapters/desktop/datadir_test.go`. Signature `DataDir(getenv func(string) string, userConfigDir func() (string, error)) (string, error)`; a stub returns `""`.
   - Green: returns `getenv("STARTER_DATA_DIR")` unconditionally; the fallback is item 2.
   - Refactor: no change. The `"STARTER_DATA_DIR"` literal stays duplicated: `config.Load` spells every env key as a literal, so one exported constant would break that idiom for a single key. Desktop can't simply reuse `config.Load` either, because its `./data` default hides whether the variable was set. Revisit if a third place reads it. `userConfigDir` stays unused until item 2.
2. [x] `desktop.DataDir`: falls back to `<user config dir>/starter`; an error from the lookup is returned.
   - Red: `TestDataDirFallsBackToUserConfigDir` (unset → `filepath.Join(<config dir>, "starter")`). The lookup-error case is a separate red after this one.
   - Green: falls back to `filepath.Join(configDir, "starter")`. The lookup's error is discarded (`configDir, _ :=`) until the next red demands it, so the item stays open.
   - Refactor: no change. A constant for `"starter"` adds nothing: the rename's whole-word slug rewrite (`scripts/internal/project/rewrite.go`) already catches the literal, as it does `STARTER_DATA_DIR`. The two lookup stubs are one-liners; reconsider a helper if the error-case test makes a third.
   - Red: `TestDataDirReturnsTheUserConfigDirError` (lookup fails → `errors.Is(err, lookupErr)`; wrapping with context is allowed).
   - Green: returns the lookup's error unwrapped.
   - Refactor: the lookup error is wrapped as `find the data folder: %w`, so a startup failure says what it was doing. The tests use two helpers, `envOf(map)` for `os.Getenv` and `configDirAt(dir)` for `os.UserConfigDir`. Left: a table test. The error case asserts differently, so a table would need branching; three short tests read better.
   - Item 11: confirm with the dry run that the folder name and SPEC.md §4.6's `~/.config/starter` are rewritten.
3. [x] `desktop.Open(env)`: serves `GET /api/v1/items` (empty list) from jsonfile storage in the data folder, and the data is still there after `Close` and a second `Open`.
   - Red: `TestOpenServesAnEmptyItemListFromTheDataFolder` in `open_test.go` (200, `{"items":[]}`, data folder created). Stub in `open.go`: `Open(getenv) (*App, error)`, `App{Handler http.Handler}`, `(*App).Close() error`. The stub's handler is `http.NotFoundHandler()`, not nil, so the red is a 404 and not a panic. Persistence after reopening is the next red.
   - Green: `Open` = `DataDir(getenv, os.UserConfigDir)` → `bootstrap.NewEmbedded(config.Config{Storage: jsonfile, DataDir: dir})` → `httpapi.NewHandler(Options{Items})`; `Close` closes the bootstrap app. `Items` is inside sample markers, as in `cmd/server`.
   - Not yet wired (no test demands them): `WebAssets: web.Assets()` (needed by item 8; the window shows the hint page without it) and a `Logger` (httpapi falls back to `slog.Default()`).
   - Refactor: the wired `bootstrap.App` (local and field) is now `core`, so it no longer reads as the desktop `App` itself. `Open`'s doc names `STARTER_DATA_DIR` and the settings-folder fallback. Left: the test's JSON decoding inline; it's the only HTTP test in the package so far, so a helper waits for the reopen test.
   - Red (didn't go red): `TestOpenKeepsItemsAfterClosingAndReopening` passed on first run, because the previous green already wired real jsonfile storage and a real `Close`. Kept as a guard. It fails as it should when `Close` is made a no-op ("data directory is locked by another process" on the second `Open`); that mutation was reverted.
   - Refactor: `open_test.go` now uses two helpers, `createItem(t, h, name)` (expects 201) and `itemNames(t, h)` (expects 200, returns the names). Each test now reads as open → act → check names. The empty-list test keeps its 200 check inside `itemNames`. Left: an `openAt(t, env)` helper. The two tests open differently (one cleans up, one closes mid-test), so it would save little.
4. [-] ~~`desktop.Open(env)`: a second `Open` on a folder that's in use fails with `jsonfile.ErrLocked`.~~ Skipped: the "second copy" scenario in item 5 covers it end to end.
   - Expect this to pass on first run too: `bootstrap.NewEmbedded` returns jsonfile's error unwrapped. Write it as a guard (verify it with the same no-op mutation, inverted), or skip to item 5, whose "second copy" scenario covers it end to end.
5. [x] Acceptance: add the `desktop` driver (temp data folder, `httptest` + `httpclient`) and `@desktop` tag routing; bind the three desktop steps (thin: close/reopen, open a second copy, assert `ErrLocked`). `task accept` goes green on all four drivers.
   - Red: the new test case is the `desktop` driver in `test/acceptance/drivers.go` (`startDesktop`: temp `STARTER_DATA_DIR` → `desktop.Open` → `httptest` + `httpclient`). `driver.tags` routes scenarios: memory, jsonfile and http get `~@desktop`, and desktop runs everything. `joinTags` in `acceptance_test.go` combines that with `--godog.tags` using `&&`. Result: memory, jsonfile and http pass 8/8, with no desktop scenarios. The desktop driver passes all 8 items scenarios, and its 2 `@desktop` scenarios fail on the 3 undefined steps. With `--godog.tags=@wip`, the other drivers report "No scenarios" (not a failure).
   - Green needs only test code: bind the 3 steps. The harness needs a way to close and reopen the desktop app (swap `h.items` to a fresh server and client on the same folder), and a way to open a second copy. No production change is expected.
   - Green (test code only, as expected): `test/acceptance/desktop_steps.go` binds the 3 steps, registered from `steps.go`. The harness gained two hooks, set only by the desktop driver: `reopen` (closes the server and app, then opens both again on the same folder, swapping `h.items`) and `openCopy` (`desktop.Open` on the same folder; closes the copy if it opened). `task accept`: memory, jsonfile and http 8/8, desktop 10/10; `@wip` only: desktop 2/2.
   - Refactor: `startDesktop`'s closures became a `desktopApp` type (`getenv`, `app`, `srv`) with `open(h)` and an idempotent `close()`, so `reopen` is just close-then-open. The `httpclient.New` + `h.items` wiring, previously duplicated in `startHTTP` and `startDesktop`, is now `(*harness).driveOverHTTP(url)`. It keeps the sample markers in one place, which helps item 11's `--remove-sample` check. Left: nil checks on `w.reopen` and `w.openCopy` in the steps. The tag routing already keeps `@desktop` scenarios off other drivers, and a nil call would fail loudly anyway.
   - Item 11: `features/desktop.feature`'s reopen scenario uses the sample's item steps (`an item named`, `the items are`). Check that `task test:init` with `--remove-sample` either drops or rewrites that scenario, or that sample markers can go in Gherkin comments.
6. [ ] Web: in the desktop shell, `OfflineNotice` never shows, even with `navigator.onLine` false. (Settle the shell detection first; see Decisions.)
7. [ ] Web: the service worker isn't registered in the desktop shell, if the check at item 6 shows it misbehaves there.
8. [ ] `cmd/desktop/main.go` (`//go:build desktop`): Wails v3 app + one window titled "Starter", asset handler = `desktop.Open`'s handler, close the app on quit. Add `github.com/wailsapp/wails/v3` to `go.mod`.
9. [ ] Taskfile: `run:desktop` (`go run -tags desktop,embedweb ./cmd/desktop`, depends on `web:build`) and `build:desktop` (`bin/starter-desktop`). Linux build deps (`libgtk-3-dev libwebkit2gtk-4.1-dev`) in `task setup`'s notes, the README and CLAUDE.md Gotchas.
10. [ ] `docs/adr/0004-native-shells.md`.
11. [ ] Rename/init: the window title and `bin/starter-desktop` follow the rename. Check with `task rename -- --name "Acme Tasks" --dry-run`, then `task test:rename` and `task test:init`.
12. [ ] CLAUDE.md: architecture lines (`cmd/desktop`, `internal/adapters/desktop`), Commands rows, Gotchas (build deps; the desktop app locks its data folder like the server).
13. [ ] Manual check: `task run:desktop`; create, rename and delete an item; restart and see it persist; `ss -ltnp` shows no port; network off shows no offline notice.
14. [ ] Remove `@wip`, add follow-ups (remote mode, packaging) to TODO.md, mark step 3 done in the parent plan, delete this file. `task check` green.

## Changes to the plan
- Item 3: the reopen test passed on first run (behaviour came with the first green's real wiring). Kept as a guard after checking it fails when `Close` doesn't release storage. Item 4 is expected to behave the same.
- Item 4 skipped (the user went straight to item 5, as suggested): its rule ("data in use") is covered end to end by item 5's scenario, and the unit test would only have repeated `bootstrap`'s error passthrough.
