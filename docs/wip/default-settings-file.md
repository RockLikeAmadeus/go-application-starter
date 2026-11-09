# Create a default settings file

- Spec: [SPEC.md §4.12](../../SPEC.md#412-keep-settings-in-a-settings-file) (the rules from "When the desktop app, the CLI or the TUI starts…" on, and the last example)
- Scenarios: [`features/default-settings-file.feature`](../../features/default-settings-file.feature) (12 `@desktop` and 1 `@mobile`, each tagged `@wip`)
- Related: [`features/settings-file.feature`](../../features/settings-file.feature), [ADR 6](../adr/0006-layered-configuration.md), [`docs/configuration.md`](../configuration.md), `.claude/rules/configuration.md`

## Decisions

Agreed with the user (2026-11-09):
- **Who creates it:** the desktop app, the CLI and the TUI, at start, when the default settings file (`config.LocalConfigFile`'s `config.toml` in the app's folder in the user's settings folder) is missing. Never the server (it never calls `LocalConfigFile`) and never the Android app.
- **Never when another file is named:** with `STARTER_CONFIG_FILE` or `--config`, nothing is created (a missing named file stays an error, as today).
- **Never overwrites:** an existing file is left exactly as it is. Create it with `O_EXCL`, so two programs starting at once can't clobber each other.
- **Contents:** `storage = "jsonfile"`, `data_dir = "<this computer's full default data folder>"` and `ignore_synced_folder = false`, set. `server_url` and `plans_file` (no default) appear commented out, with a note. `port` and `log_format` are left out: only the server reads them, and it never reads this file.
- **The comments in the file are user-facing text.** Agree their exact wording with the user (`.claude/rules/user-facing-text.md`) in the task that writes them, before writing them.
- **If it can't be written** (read-only folder, no settings folder), carry on silently on the built-in defaults, as today, and try again next run. No warning.

Found while planning:
- **Android goes through `desktop.Open`** (`internal/adapters/mobile/open.go`), so creating the file in `desktop.Open` alone would create one in the phone's storage. `mobile.Open` must open without creating it. Name the option or the split when it's needed (task 6), following `.claude/rules/naming.md`.
- **The acceptance helper `localComputer.addSetting` appends** to `config.toml`. Once the desktop app creates the file at the driver's start, appending `data_dir` gives a duplicate key and `settings-file.feature` breaks. Task 4 makes it set a key in place, before task 5 wires the desktop app.
- **Where the code lives:** the file's text comes from a pure function in `internal/config` (unit-tested with no files); creating it sits beside `config.LocalConfigFile`, which already uses `os.Stat`. The CLI and TUI share `cli`'s `settingsFileFor`, so one wiring covers both; the TUI needs no scenario of its own.
- **`starter info` will say "settings file"** for storage, data_dir and ignore_synced_folder after the first run, not "built in". That's expected. Check `internal/adapters/cli/info_test.go` (it expects `# built in` lines) and the terminal journeys still pass, and fix a test only if it asserted on a real settings folder by accident.
- "The created settings file changes nothing" already passes: it's a regression guard, and must still pass at the end.

## Tasks

- [x] 1. `config` gives the default settings file's text for a data folder: `storage`, `data_dir` and `ignore_synced_folder` set, and it parses back with `Load` to the same values. No comments yet.
  - Red: `TestTheDefaultSettingsFileSetsTheStorageDataFolderAndSyncedFolderSettings` (`defaultsettings_test.go`) failed: all three settings came back "built in" from the stub's empty text.
  - Green: `config.DefaultSettingsFileText(dataDir)` returns the three lines, quoting `data_dir` with `%q` (so a Windows path's backslashes are escaped).
  - Refactor: `storage` comes from `StorageJSONFile`; left `Load`'s fallbacks unshared with the text, as two literals don't justify a table yet.
- [x] 2. The text has `server_url` and `plans_file` commented out, and the comments, with wording agreed with the user first. It still parses, and leaves out `port` and `log_format`.
  - Red: `TestTheDefaultSettingsFileDescribesEachSettingAndLeavesOutTheRest` (`defaultsettings_test.go`, pinning the wording agreed 2026-11-09) failed: the text had only the three settings, with no comments.
  - Green: `DefaultSettingsFileText` returns the agreed text, the header and each setting's note, with `server_url` and `plans_file` commented out.
  - Refactor: nothing worth changing; the test pins the whole text on purpose, and task 1's test still covers parsing and a Windows path's escaping.
- [x] 3. `config` creates the default settings file when it's missing: it writes the text from task 1, for the default data folder, to `LocalConfigFile`'s default path, making the app's folder.
  - Red: `TestTheDefaultSettingsFileIsCreatedWhenMissing` (`localdata_test.go`) failed: no `starter/config.toml` after the stub `CreateDefaultSettingsFile`.
  - Green: `CreateDefaultSettingsFile(env)` makes the app's folder and writes `DefaultSettingsFileText` for it to `config.toml` (plain `os.WriteFile` for now; 3b makes it `O_EXCL`).
  - Refactor: `LocalConfigFile` and `CreateDefaultSettingsFile` share `defaultSettingsFile(env)`; kept the `error` return, which 3d decides on.
- [x] 3b. It never overwrites an existing settings file.
  - Red: `TestTheDefaultSettingsFileIsNeverOverwritten` (`localdata_test.go`) failed: an existing `config.toml` was replaced by the default text.
  - Green: `CreateDefaultSettingsFile` opens the file with `O_CREATE|O_EXCL`, so an existing one is left alone (and its `ErrExist` returned, for 3d to decide).
  - Refactor: moved the exclusive open, write and close into `writeNewFile(file, text)`; the doc comment says it never overwrites.
- [x] 3c. It does nothing when `STARTER_CONFIG_FILE` is set.
  - Red: `TestNoDefaultSettingsFileIsCreatedWhenStarterConfigFileNamesOne` (`localdata_test.go`) failed: `starter/config.toml` was created though `STARTER_CONFIG_FILE` named another file.
  - Green: `CreateDefaultSettingsFile` returns nil before touching any file when `STARTER_CONFIG_FILE` is set.
  - Refactor: the doc comment says so; left the two `STARTER_CONFIG_FILE` reads unshared, as a helper for one `Getenv` adds nothing.
- [x] 3d. It returns no error when it can't write (a read-only or blocked app's folder).
  - Red: `TestTheDefaultSettingsFileReturnsNoErrorWhenItCantWrite` (now `…CarriesOnWhenItCantWrite`) (`localdata_test.go`) failed: making the app's folder, blocked by a file, returned `not a directory`.
  - Green: `CreateDefaultSettingsFile` returns nil when it can't make the app's folder or write the file (an existing file's `ErrExist` included).
  - Refactor: the doc comment says it carries on silently; kept the `error` return, which 3e decides on.
- [x] 3e. It returns no error when there's no settings folder (`UserConfigDir` fails).
  - Red: `TestTheDefaultSettingsFileReturnsNoErrorWithoutASettingsFolder` (now `…CarriesOnWithoutASettingsFolder`, `localdata_test.go`) failed: it returned `UserConfigDir`'s error.
  - Green: `CreateDefaultSettingsFile` returns nil when it can't find the settings folder.
  - Refactor: it could no longer fail, so it returns nothing; the 3d and 3e tests are renamed `…CarriesOn…` and assert, instead, that the file in the app's folder's place is left as it was (3d) and nothing is created in the working directory (3e); the doc comment covers a missing settings folder.
- [x] 4. Acceptance helper: `addSetting` sets a key in an existing `config.toml` instead of appending a second one. `settings-file.feature` still passes.
  - Red: `TestAddSettingSetsAKeyTheSettingsFileAlreadyHas` (now `TestSetSettingReplacesAKey…`, `test/acceptance/settings_file_steps_test.go`) failed: `Load` refused the file, as `data_dir` was defined twice.
  - Green: `addSetting` reads the file and replaces the key's line when it has one, else appends it.
  - Refactor: renamed `addSetting` to `setSetting`, since it now sets a key in place; left the regexp built per call, as a test helper's few calls don't need it shared.
- [x] 5. `desktop.Open` creates the default settings file before reading the settings. Define the desktop steps "my settings file sets … to …" and "my settings file doesn't set …"; turns the first two scenario outlines green.
  - Red: `TestOpenCreatesTheDefaultSettingsFile` (`internal/adapters/desktop/open_test.go`) failed: no `config.toml` in the settings folder.
  - Green: `desktop.Open` calls `config.CreateDefaultSettingsFile` first; steps in `test/acceptance/default_settings_file_steps.go`. Both outlines and "changes nothing" pass.
  - Refactor: `Open`'s doc comment now says it creates the file; nothing else worth changing.
- [x] 6. `mobile.Open` creates no settings file. Define "the mobile app has created no settings file"; turns the `@mobile` scenario green.
  - Red: `TestOpenCreatesNoSettingsFile` (mobile) failed: Open left `starter/config.toml` in the storage folder.
  - Green: `desktop.Open` = `config.CreateDefaultSettingsFile` + new `desktop.OpenWithoutDefaultSettingsFile`, which `mobile.Open` calls; step defined via new `harness.phoneStorage`. `@mobile` scenario passes.
  - Refactor: `mobile.Open`'s doc comment now says it creates no settings file. Left `desktop.Open` building `config.Env` twice: both functions take the same arguments, and it's one line.
- [x] 7. The CLI and TUI create the default settings file when `--config` names none. Define "I delete my settings file"; turns "The command line creates the settings file when there is none" green.
  - Red: `TestCommandsCreateTheDefaultSettingsFileWhenThereIsNone` (`internal/adapters/cli/info_test.go`) failed: no `starter/config.toml` after `info`.
  - Green: `cli`'s `settingsFileFor` calls `config.CreateDefaultSettingsFile` when `--config` names no file; `TestInfoNamesTheSettingsFile`'s "none" case now runs on a read-only settings folder; step "I delete my settings file" defined. The scenario passes.
  - Refactor: `settingsFileFor`'s doc comment now says it creates the file; nothing else worth changing (the TUI goes through the same function).
- [x] 8. Define "I replace my settings file with one choosing the data folder …" and "my settings file is as I left it"; turns "A settings file that's already there is kept as it is" green.
  - No red: with the two steps defined (`default_settings_file_steps.go`), the scenario passed straight away, as 3b already never overwrites; kept as a guard (it fails when the exclusive open is swapped for `O_TRUNC`). No green.
  - Refactor: nothing worth changing; the new steps follow the file's others.
- [x] 9. Define "I name a settings file of my own in the environment / with a command option" and "I have no settings file in my settings folder"; turns the last outline green.
  - No red: with the three steps defined (`default_settings_file_steps.go`; `--config` through a new `world.commandLineOptions`), both examples passed straight away, as 3c and task 7 already create nothing when a file is named; kept as guards (each fails when its step stops naming the file). No green.
  - Refactor: nothing worth changing; the new steps follow the file's others.
- [ ] 10. Docs: `docs/configuration.md` ("Which file is read", the example file matching the generated text), `.claude/rules/configuration.md` (including a rule for adding a setting: it goes into `config.DefaultSettingsFileText` with a note above it, commented out when it has no useful default; the note's wording is agreed with the user first; and the example file in `docs/configuration.md` changes to match), README if it describes first run.
- [ ] 10b. A guard test: every setting `config` knows about appears in `DefaultSettingsFileText`, set or commented out, so adding a setting without updating the text fails `task test`.
- [ ] 11. Run `task test:init`, `task check` and `task accept`, remove `@wip`, and delete this file.

## Changes to the plan

- 2026-11-09: split task 3 into 3, 3b, 3c and 3d, one behaviour and one test each.
- 2026-11-09: split 3d: 3e covers no settings folder, a separate failure needing its own test.
- 2026-11-09: at the user's request, task 10 also adds a rule to `.claude/rules/configuration.md` for keeping the default settings file in step with new settings, and task 10b adds a guard test that enforces it.
