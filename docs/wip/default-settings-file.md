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

- [ ] 1. `config` gives the default settings file's text for a data folder: `storage`, `data_dir` and `ignore_synced_folder` set, and it parses back with `Load` to the same values. No comments yet.
- [ ] 2. The text has `server_url` and `plans_file` commented out, and the comments, with wording agreed with the user first. It still parses, and leaves out `port` and `log_format`.
- [ ] 3. `config` creates the default settings file when it's missing: it writes the text from task 1 to `LocalConfigFile`'s default path (making the app's folder), never overwrites an existing file, does nothing when `STARTER_CONFIG_FILE` is set, and returns no error when it can't write.
- [ ] 4. Acceptance helper: `addSetting` sets a key in an existing `config.toml` instead of appending a second one. `settings-file.feature` still passes.
- [ ] 5. `desktop.Open` creates the default settings file before reading the settings. Define the desktop steps "my settings file sets … to …" and "my settings file doesn't set …"; turns the first two scenario outlines green.
- [ ] 6. `mobile.Open` creates no settings file. Define "the mobile app has created no settings file"; turns the `@mobile` scenario green.
- [ ] 7. The CLI and TUI create the default settings file when `--config` names none. Define "I delete my settings file"; turns "The command line creates the settings file when there is none" green.
- [ ] 8. Define "I replace my settings file with one choosing the data folder …" and "my settings file is as I left it"; turns "A settings file that's already there is kept as it is" green.
- [ ] 9. Define "I name a settings file of my own in the environment / with a command option" and "I have no settings file in my settings folder"; turns the last outline green.
- [ ] 10. Docs: `docs/configuration.md` ("Which file is read", the example file matching the generated text), `.claude/rules/configuration.md`, README if it describes first run.
- [ ] 11. Run `task test:init`, `task check` and `task accept`, remove `@wip`, and delete this file.

## Changes to the plan
