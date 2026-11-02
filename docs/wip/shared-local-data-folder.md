# WIP: The CLI and TUI share the desktop app's data folder

- Spec: [SPEC.md §4.11](../../SPEC.md#411-share-local-data-between-the-desktop-app-and-the-command-line), plus the glossary's **Data folder** and the CLI row of §2.
- Scenarios: [features/local-data.feature](../../features/local-data.feature) (`@desktop`, so only the desktop driver runs them). Steps: `test/acceptance/local_data_steps.go`. Run them with `task accept -- --godog.tags=@wip`.
- Docs to update: [docs/data-locations.md](../data-locations.md) (not committed yet: it, its README link and the Windows-lock TODO item were written just before this feature and go in with its first commit).
- Temporary: delete this file in the change that removes the `@wip` tags.

## Decisions
- **Why:** the CLI and TUI defaulted to `./data` relative to the folder they're run in, so two runs from different folders silently used two sets of data. The desktop app already uses the user's settings folder, and ADR 5 treats the CLI, TUI and desktop app as one local edition. (Chosen 2026-11-02.)
- **The server keeps `./data`.** Docker and Fly.io set `STARTER_DATA_DIR=/data`, a server run as a system service often has no proper settings folder, and `task dev` keeps using the repository's `./data`. A self-hosted server can set `STARTER_DATA_DIR` to share the folder.
- **Precedence for the CLI and TUI:** `--data-dir` > `STARTER_DATA_DIR` > `<settings folder>/starter`. The last two are exactly the desktop app's rule (`desktop.DataDir` today).
- **The rule moves to `config.LocalDataDir(getenv, userConfigDir)`**, so the desktop app and the CLI share one rule, and the CLI doesn't import the desktop adapter. Its three existing tests move with it.
- **The CLI gets its default through a new `NewRootCmd` dependency**, a `func() (string, error)` that `cmd/starter` builds from `config.LocalDataDir(os.Getenv, os.UserConfigDir)`. Tests and the acceptance steps pass their own, so nothing reads the real settings folder. With `--server`, it's never called.
- **`starter info` (template command, chosen 2026-11-02):** prints the data folder (or the server, with `--server`) and the workspace's plan. It opens the app, so it fails with "data directory is locked" while the desktop app is open. It exists because, with the sample removed, `starter tui` is the only command that opens the data, and a scenario can't drive it; it also answers "where is my data?" for users.
- **Dev tasks use the real folder (chosen 2026-11-02):** `task run:cli` and `task run:tui` share `~/.config/starter` with `task run:desktop`. No Taskfile change.
- **The desktop driver now opens the app on a fake settings folder with no `STARTER_DATA_DIR`** (`localComputer` in `test/acceptance/drivers.go`), so the default rule is what's tested. The command line runs beside it in-process (`runCommandLine`), with `bootstrap.New` as `cmd/starter` does.
- **The TUI** opens the app through the same `openApp` as every other command, so the CLI tests cover its data folder. Check `starter tui` by hand at the end.
- **Shown failing (2026-11-02):** the three template scenarios fail with `unknown command "info" for "starter"` (the lock scenario: `expected the data to be in use, got unknown command "info"`). The sample scenario's `I list the items from the command line` and `the command line shows "Widget"` are undefined: defining them today would run `item list` on `./data` in `test/acceptance`.

## Tasks
1. [ ] Refactor (no red): move `desktop.DataDir` to `config.LocalDataDir`, with its tests (`internal/config/localdata_test.go`). `desktop.Open` and the acceptance step `itNamesTheDesktopAppsDataFolder` use it.
2. [ ] CLI: `starter info` prints the data folder it opened (`--data-dir` given). Test in a new `internal/adapters/cli/info_test.go` (template file; the `run` helper in `item_test.go` is sample, so `info_test.go` needs its own).
3. [ ] CLI: `starter info` fails with `jsonfile.ErrLocked` while another program has the data folder open (open a `jsonfile.Store` on it first).
4. [ ] CLI: `starter info` prints the workspace's plan (`local`), through `app.Entitlements`.
5. [ ] CLI: without `--data-dir`, commands use the folder from the new `NewRootCmd` dependency (the `AppFactory` receives it as `cfg.DataDir`). Update `cmd/starter`, the sample's `run` helper and `runCommandLine` (pass `config.LocalDataDir(w.local.getenv, w.local.userConfigDir)`). The first two template scenarios should pass after this.
6. [ ] CLI: when the settings folder can't be found, a command that needs the data fails, saying so ("find the data folder: …").
7. [ ] CLI: with `--server`, the settings folder isn't looked up (a failing lookup doesn't matter), and `starter info` names the server instead of a data folder.
8. [ ] CLI: with `--storage memory`, `starter info` says the data is kept in memory and not saved, instead of naming a folder.
9. [ ] Sample: define `I list the items from the command line` and `the command line shows "…"` in `test/acceptance/item_steps.go`, through `runCommandLine`. No production code.
10. [ ] Docs: `docs/data-locations.md` (the CLI and TUI rows and section, `starter info`), CLAUDE.md (the Gotcha about the jsonfile lock: the CLI and the desktop app now share a folder by default), README (CLI usage, `starter info`). Check `.env.example`'s `STARTER_DATA_DIR` comment.
11. [ ] Finish: run `starter tui` and `starter info` by hand, `task check`, `task test:init` (new template code beside sample code), remove the `@wip` tags, delete this file.

## Changes to the plan
