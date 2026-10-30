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
2. [ ] `desktop.DataDir`: falls back to `<user config dir>/starter`; an error from the lookup is returned.
   - Red: `TestDataDirFallsBackToUserConfigDir` (unset → `filepath.Join(<config dir>, "starter")`). The lookup-error case is a separate red after this one.
   - Green: falls back to `filepath.Join(configDir, "starter")`. The lookup's error is discarded (`configDir, _ :=`) until the next red demands it, so the item stays open.
   - Refactor: no change. A constant for `"starter"` adds nothing: the rename's whole-word slug rewrite (`scripts/internal/project/rewrite.go`) already catches the literal, as it does `STARTER_DATA_DIR`. The two lookup stubs are one-liners; reconsider a helper if the error-case test makes a third.
   - Item 11: confirm with the dry run that the folder name and SPEC.md §4.6's `~/.config/starter` are rewritten.
3. [ ] `desktop.Open(env)`: serves `GET /api/v1/items` (empty list) from jsonfile storage in the data folder, and the data is still there after `Close` and a second `Open`.
4. [ ] `desktop.Open(env)`: a second `Open` on a folder that's in use fails with `jsonfile.ErrLocked`.
5. [ ] Acceptance: add the `desktop` driver (temp data folder, `httptest` + `httpclient`) and `@desktop` tag routing; bind the three desktop steps (thin: close/reopen, open a second copy, assert `ErrLocked`). `task accept` goes green on all four drivers.
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
