# WIP: warn when SQLite's data folder looks synced

- TODO item: "Warn when SQLite's data folder looks synced" in [`TODO.md`](../../TODO.md) (a follow-up to step 6 of the [editions plan](../plans/editions-and-accounts.md)).
- Spec: [SPEC.md §4.14](../../SPEC.md#414-keep-data-in-sqlite) (rules on synced folders, and the last three examples).
- Business scenarios: [`features/synced-folder.feature`](../../features/synced-folder.feature) (desktop driver, plus one on the server driver). Steps: `test/acceptance/synced_folder_steps.go`.
- Terminal journey: the `@wip` scenario in [`test/terminal/features/terminal-app.feature`](../../test/terminal/features/terminal-app.feature).
- Temporary: delete this file in the same change that removes the `@wip` tags.

## Decisions (agreed 2026-11-03)

- **Warn, never refuse**, and name the service: "Your data folder looks synced by Dropbox. A sync service can damage SQLite data: move the data folder out of it, or keep your data as plain text." The message is built in one place (`bootstrap`), and every front end shows it as given.
- **Silenceable:** a new setting, `ignore_synced_folder` (`STARTER_IGNORE_SYNCED_FOLDER`), a bool, default false. A bad value fails at start, like every setting. It shows in `starter info` through `Config.Settings`.
- **SQLite only.** Plain text in a synced folder is supported (one device at a time, ADR 5), so it gets no warning. Memory and remote (`--server`) mode have no local folder, so they get none either. The Android app stays on plain text, so it never warns.
- **Where it shows (all four, the user's choice):**
  - the desktop app's window, and the server's web app: a notice in the SPA, from a new `GET /api/v1/warnings` → `{"warnings": [{"message": "…"}]}` (an object per warning, so fields can be added later);
  - the CLI: `warning: …` on stderr for every embedded command, so stdout and scripts are untouched;
  - the TUI: a warning line on its screen, above whichever screen is open (template code wrapping the first screen, so it survives `--remove-sample`);
  - the server: a `slog` warning when it starts.
- **Detection is template code with no core dependency:** a new package `internal/adapters/syncedfolder`, `Detect(dir) (service string, ok bool)`. It resolves `dir` to an absolute path (the server's default is `./data`), then checks the folder and each of its parents for:
  - marker files a sync service leaves at the top of its folder: `.dropbox` (Dropbox), `.stfolder` (Syncthing), `.sync_*.db` or `._sync_*.db` (Nextcloud, also ownCloud);
  - folder names: `Dropbox`, `OneDrive` and `OneDrive - <org>`, `Google Drive` and `My Drive`, `iCloud Drive`, `iCloudDrive` (Windows) and `Mobile Documents` (macOS), and macOS's `Library/CloudStorage/<Provider>-…` (the provider from the prefix; `GoogleDrive` → Google Drive).
  It's a best guess: false positives are what the setting is for. Conflict-copy detection and the "in use on another device" lock are step 5 of the editions plan, not this.
- **Not core:** warnings aren't a business capability, so there's no driving port. `bootstrap.App.Warnings []string` carries them, as `App.Ready` carries readiness; `httpapi.Options.Warnings` serves them, as `Options.Ready` does.
- **No browser e2e journey:** `task e2e`'s server runs on memory storage. Godog proves the API's answer through the desktop and server drivers, and Vitest proves the SPA shows it and where.
- Acceptance setup: `syncedDataFolder` in the steps fakes one service of each detection kind (a marker file: Dropbox; a marker folder: Syncthing; a folder name: OneDrive). The full list is unit-tested in `syncedfolder`.

## Tasks

- [x] 1. `syncedfolder.Detect`: a folder inside one holding `.dropbox` looks synced by Dropbox; a folder with no signs anywhere above it doesn't.
  - Red: `TestDetect_AFolderInsideOneHoldingDotDropboxLooksSyncedByDropbox` failed: `Detect` = `("", false)`, want `("Dropbox", true)`.
  - Green: `Detect` walks up from `dir` to the root, `Dropbox` at the first `.dropbox`.
  - Refactor: nothing; a marker table waits until task 2 brings a second marker.
  - Cycle 2, no red: `TestDetect_AFolderWithNoSignsAnywhereAboveItDoesNotLookSynced` passed at once (the walk already ends with `("", false)`), kept as a guard; a mutation always answering Dropbox made it fail. Refactor: a `folderInside` test helper.
- [ ] 2. `Detect`: `.stfolder` above the folder → Syncthing.
- [ ] 3. `Detect`: `.sync_*.db` or `._sync_*.db` above the folder → Nextcloud.
- [ ] 4. `Detect`: folder names (table test): `Dropbox`, `OneDrive`, `OneDrive - Contoso`, `Google Drive`, `My Drive`, `iCloud Drive`, `iCloudDrive`, `Mobile Documents`.
- [ ] 5. `Detect`: macOS `Library/CloudStorage/<Provider>-…` → that provider (`Dropbox-Personal` → Dropbox, `GoogleDrive-me@example.com` → Google Drive, `OneDrive-Personal` → OneDrive).
- [ ] 6. `Detect`: a relative folder is checked from where the program runs (its parents count).
- [ ] 7. `config`: `ignore_synced_folder` from the settings file or `STARTER_IGNORE_SYNCED_FOLDER`, default false; a value that isn't a bool fails `Load`; it's listed by `Settings`. Add it to `envVars`, `Load`, `Config.Settings` and `docs/configuration.md`.
- [ ] 8. `bootstrap.NewEmbedded`: on SQLite in a synced folder, `App.Warnings` holds the message naming the service; none on jsonfile in the same folder, none with `IgnoreSyncedFolder`, none in an ordinary folder. `NewRemote` has none.
- [ ] 9. API: `GET /api/v1/warnings` in `api/openapi.yaml`, `task gen`; `httpapi.Options.Warnings` answered as given, an empty list when unset.
- [ ] 10. `desktop.Open` serves the core's warnings → the desktop app scenarios in `features/synced-folder.feature` pass (all but the command-line ones).
- [ ] 11. `cmd/server`: `handlerOptions` passes the warnings, and the server logs each with `log.Warn` at start → the `@server` scenario passes.
- [ ] 12. CLI: every embedded command prints each warning as `warning: …` on stderr (not stdout); nothing when there are none → the command-line scenarios pass.
- [ ] 13. TUI: a `warningLine` helper in `styles.go`, with its test in `styles_test.go`.
- [ ] 14. TUI: the warnings show above whichever screen opens first; `starter tui` passes the app's warnings → the terminal journey passes. Run `task test:init` (this wraps the sample's screen).
- [ ] 15. SPA: a warnings notice (`web/src/features/warnings/`) reading `GET /api/v1/warnings` (generated types): shows each message, nothing when there are none or the request fails; placed in `Layout` beside `OfflineNotice`. Vitest, by role.
- [ ] 16. Finish: `docs/data-locations.md` and ADR 7 ("is planned" → what happens), CLAUDE.md's SQLite Gotcha (one line on the warning and the setting), tick the TODO.md item, and note it under step 6 of the editions plan. Remove the `@wip` tags and delete this file. Run `task check` and `task test:init`.

## Changes to the plan

_None yet._
