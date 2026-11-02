# WIP: A settings file layered under the environment variables

- Backlog: `TODO.md`, "Do first", **Convert configuration from drawing from environment variables to a config file**.
- Spec: [SPEC.md §4.12](../../SPEC.md#412-keep-settings-in-a-settings-file), the glossary's **Settings file**, and §4.6's data folder rule.
- Scenarios: [features/settings-file.feature](../../features/settings-file.feature) (`@desktop`). Steps: `test/acceptance/settings_file_steps.go`. Run with `task accept -- --godog.tags=@wip`.
- Plan made in plan mode (2026-11-02): `~/.claude/plans/let-s-take-a-look-wild-grove.md`, outside the repo; everything needed is copied below.
- Temporary: delete this file in the change that removes the `@wip` tags and ticks the TODO item.

## Decisions
- **Add a layer, don't replace env vars** (chosen 2026-11-02). Servers are configured through env vars (Dockerfile, `docker-compose.yml`, `fly.*.toml`, CI, secrets); the standard pattern is layers. The gap was the local edition: desktop users can't set env vars, and the CLI, TUI and desktop app share a data folder but had no shared settings.
- **Precedence:** built-in defaults < settings file < env vars (`STARTER_*`, and `PORT`) < CLI flags. **Env over file is deliberate:** the user first preferred file over env; it's what Viper, koanf, git, the AWS CLI, npm and Docker do, and it keeps one-off overrides, deployment env vars and secrets, and test isolation working. The risk of a forgotten env var beating the file is answered by `starter info` showing each setting's source.
- **koanf** (`github.com/knadh/koanf/v2`, `providers/file`, `providers/confmap`, `parsers/toml/v2`; check current versions with `go get …@latest`), over Viper (large dependency tree, global by default, case-insensitive keys) and a hand-rolled parser. **TOML.**
- **Keys:** `storage`, `data_dir`, `port`, `server_url`, `log_format`, `plans_file`. Env var = `STARTER_` + upper case, except `PORT`.
- **Where the file is:** for the CLI, TUI and desktop app, `STARTER_CONFIG_FILE` (or the CLI's `--config`) if set, else `<settings folder>/starter/config.toml`. A missing default file is fine; a missing named file is an error. The server reads a file only when `STARTER_CONFIG_FILE` names one, so `task dev` isn't redirected by a developer's own file. Android: nothing new (`mobile.Open` → `desktop.Open` finds no file in private storage).
- **Strict:** an unknown key or bad TOML fails at start, naming the file and the key.
- **Data folder rule** (`config.LocalDataDir`): `--data-dir` > `STARTER_DATA_DIR` > the file's `data_dir` > `<settings folder>/starter`. The server's default stays `./data`.
- **Testability:** `internal/config` stops reading `os` itself. A `config.Env{Getenv, UserConfigDir}` describes the computer; `cmd/*` pass `os.Getenv` and `os.UserConfigDir`. Env values reach koanf through `confmap` from the known keys, not koanf's env provider, so tests inject them.
- **Sources:** `config.Load` records where each key came from (load each layer into its own koanf instance and note the keys it set). `starter info` prints `Settings file: <path or none>`, then each setting with its value and source: `built in`, `settings file`, `environment (STARTER_DATA_DIR)`, `option (--data-dir)`.
- **Flags** keep their explicit overrides in `internal/adapters/cli/root.go`; koanf's `posflag` only if it simplifies during refactor.
- **ADR 6** (`docs/adr/0006-layered-configuration.md`) records: layers not replacement, env over file and why, koanf over Viper, TOML, the server reading a file only when named.
- **Shown failing (2026-11-02):** "The settings file chooses the data folder" (the CLI names `<settings>/starter`), "The desktop app uses the settings file's data folder" (no `.lock` in `work-data`), "The command line says where a setting came from" (no `STARTER_DATA_DIR` in `info`), "A settings file with a setting the app doesn't know stops it" (the command line worked). "A setting chosen when starting overrides the settings file" already passes, since nothing reads the file yet; it guards the precedence once the file is read.

## Tasks
1. [x] Refactor, then red: `config.Load(env Env)` reads the environment through `env.Getenv` instead of `os` (`cmd/server` and the CLI pass `os.Getenv`). First red: `Load` reads `STARTER_STORAGE` from the injected env (new `internal/config/config_test.go`).
   - Red: `config.Env{Getenv}` and `Load(env Env)` added, unused, so the test compiled (`cmd/server` and the CLI's `openApp` pass `config.Env{Getenv: os.Getenv}`). `TestLoadReadsTheEnvironmentItIsGiven` failed with `Storage = "jsonfile", want "memory"`.
   - Green: `Load` reads every variable through `env.Getenv`; the `getenv` helper takes the lookup function; `internal/config` no longer imports `os`.
   - Refactor: doc comments on `Load` and `getenv`. `Env` gets `UserConfigDir` in task 6, when something needs it. The CLI still passes `os.Getenv` itself until task 8.
2. [x] `config`: values from a TOML settings file are used (koanf: file provider + TOML parser, unmarshalled into `Config`).
   - Red: `Load(env, settingsFile string)` got its new parameter (unused; callers pass `""`). `TestLoadUsesTheSettingsFile` (writes `storage` and `data_dir` to a temp `config.toml`) failed with `Storage, DataDir = "jsonfile", "./data", want "memory", "/srv/work-data"`.
   - Green: koanf v2.4.1 (`providers/file` v1.2.1, `parsers/toml/v2` v2.2.2; `go mod tidy` also bumped the indirect `pelletier/go-toml/v2` to v2.4.3 and added `fsnotify`) loads the file; each setting is `getenv(env, STARTER_X, getenv(k.String, key, default))`. Not unmarshalled into `Config`: the per-setting fallback is simpler and already gives each setting one place (it will also suit sources, task 10). koanf's `String` formats a TOML number as text, so `port = 8081` works.
   - Refactor: the package, `Config` field and `Load` doc comments name each setting's key in the file and the order (env, file, default).
3. [x] `config`: env vars override the settings file.
   - Red: `TestTheEnvironmentOverridesTheSettingsFile` passed straight away, as the plan log predicted: task 2's lookup checks the environment first. Kept as a guard (chosen 2026-11-02), pinning the deliberate env-over-file order against later rewrites. No green or refactor.
4. [x] `config`: an unknown key fails, naming the file and the key.
   - Red: `TestASettingTheAppDoesntKnowFailsNamingTheFileAndSetting` (`colour = "blue"`) failed with `Load error = <nil>, want one naming <file> and "colour"`.
   - Green: after loading, any key outside the six settings fails with `settings file <path>: unknown setting "colour"`. A TOML table's keys come out dotted (`a.b`), so they're refused too.
   - Refactor: reading and checking the file moved into `readSettingsFile(path)`, with the keys in `settingKeys`. Left: the six keys also appear in `Load`'s lookups; task 10 (sources) will want one table of settings anyway, so merging them waits for it. `@wip` progress unchanged (nothing passes a file to `Load` until task 8).
5. [x] `config`: bad TOML fails, naming the file.
   - Red: `TestASettingsFileThatIsntTOMLFailsNamingTheFile` (an unterminated string) passed straight away: task 2 already wraps koanf's read error as `read the settings file <path>: …`. Kept as a guard (chosen 2026-11-02) for SPEC §4.12's rule. No green or refactor.
6. [x] 6a. `config.LocalConfigFile(env)` is `<settings folder>/starter/config.toml`.
   - Red: `Env` gained `UserConfigDir` and `LocalConfigFile` a stub returning `""`, so the test compiled. `TestLocalConfigFileIsConfigTomlInTheAppsSettingsFolder` failed with `LocalConfigFile = "", want "<settings>/starter/config.toml"`.
   - Green: it joins `env.UserConfigDir()` with `starter/config.toml` (no existence check yet: 6c).
   - Refactor: doc comment; the app's folder name is `appFolder`, shared with `LocalDataDir`. Left: wrapping `UserConfigDir`'s error like `LocalDataDir` does ("find the data folder: …"): no test asks for it yet, and 6c touches the same lines.
   6b. [x] `STARTER_CONFIG_FILE` names another settings file, which `Load` then requires (a missing named file fails, naming it).
   - Red: `TestStarterConfigFileNamesAnotherSettingsFile` failed with `LocalConfigFile = "<settings>/starter/config.toml", want "/srv/starter/settings.toml"`.
   - Green: `LocalConfigFile` returns `STARTER_CONFIG_FILE` when set.
   - Refactor: the doc comment covers it, and says `Load` fails on a missing named file. That part needs no code: koanf's file provider can't read a missing file, and task 2's wrapping names it. Not unit-tested separately (it would pass straight away, like tasks 3 and 5); the CLI's `--config` (task 9) is the natural place to pin it.
   6c. [x] When the default settings file doesn't exist, there is none (`""`), so `Load` uses the environment and defaults only.
   - Red: `TestWithoutTheDefaultSettingsFileThereIsNone` failed with `LocalConfigFile = "<settings>/starter/config.toml", want none`.
   - Green: `LocalConfigFile` returns `""` when `os.Stat` says the default file doesn't exist. Any other error (say, no permission) keeps the path, so `Load` fails naming the file instead of silently skipping it.
   - Refactor: the doc comment says the default file is optional. Left: wrapping `UserConfigDir`'s error (see 6a); still no test asks for it, and the CLI and desktop app report it as it is.
7. [x] `config.LocalDataDir` uses the loaded `data_dir` (from the file or env) before the settings folder.
   - Red: `LocalDataDir(env Env, settingsFile string)` replaced `(getenv, userConfigDir)`, with `settingsFile` unused; every caller (`desktop.Open`, `cmd/starter`, the acceptance steps, the tests) passes `""` for now. It takes the file rather than `Load`'s `DataDir`, whose default is the server's `./data` and can't be told from a chosen value. `TestLocalDataDirUsesTheSettingsFilesDataDir` failed with `LocalDataDir = "/home/someone/.config/starter", want "/srv/work-data"`.
   - Green: after `STARTER_DATA_DIR`, it reads the file with `readSettingsFile` (so an unknown key or bad TOML fails here too) and uses its `data_dir`.
   - Refactor: gofmt on the two callers whose lines grew; the doc comment states the new order. Left: the file is read twice when a front end calls both `Load` and `LocalDataDir`, and `STARTER_DATA_DIR` before the file repeats `Load`'s lookup. Task 10 (sources) reshapes `Load`; merging them waits for it.
8. [x] CLI: `NewRootCmd` takes `config.Env` instead of `localDataDir`; commands load the local settings file. Update `cmd/starter`, the test helpers and `runCommandLine`. Scenarios "chooses the data folder" and "doesn't know stops it" should pass.
   - Red: first, with behaviour unchanged, `NewRootCmd(newApp, env config.Env, out)`; `openApp` passes `env` to `Load` and `LocalDataDir` (no file yet), so the CLI no longer reads `os` itself. The CLI tests became hermetic: `computer(settings, vars)` is a fake computer, and `runCLI` uses one with an empty settings folder. That replaced `noLocalDataDir`: every command now looks for its settings file in the settings folder, so "the settings folder is never looked up" can't hold any more. `TestWithoutDataDirCommandsUseTheLocalDataFolder` lost its `t.Setenv`. Then `TestCommandsUseTheLocalSettingsFile` (`data_dir` in `<settings>/starter/config.toml`) failed with `Data folder: <settings>/starter`.
   - Green: `openApp` finds the file with `config.LocalConfigFile(env)` and passes it to `Load` and `LocalDataDir`.
   - Refactor: `openApp`'s comment names the settings file. `@wip`: 3 of 5 pass ("chooses the data folder", "overrides", "doesn't know stops it"); left are the desktop app (task 13) and sources (task 12). Left: the acceptance step "it names the desktop app's data folder" still calls `LocalDataDir(env, "")`; task 13 should pass the settings file there too, once the desktop app reads it.
9. [x] CLI: `--config` names the settings file for one command.
   - Red: the `--config` flag was registered with nothing behind it, so the test failed on the missing behaviour rather than an unknown flag. `TestConfigNamesTheSettingsFileForOneCommand` (`data_dir` in a `work.toml` outside the settings folder) failed with `Data folder: <settings>/starter`.
   - Green: `openApp` reads the `--config` file when given, else `config.LocalConfigFile(env)`.
   - Refactor: that choice is `settingsFileFor(configFile, env)`; the flag's variable is `configFile`, after the option; the two settings-file tests share `writeSettings(t, path, content)`. Left: a missing `--config` file failing, naming it, isn't unit-tested at the CLI: `Load` already does it (koanf can't read the file; task 2's wrapping names it), so a test would pass straight away.
10. [x] `config`: `Load` reports each setting's source (built in, settings file, environment variable by name).
   - Red: `Config.Sources map[string]string` (by the setting's key in the file) added, unset. `TestLoadSaysWhereEachSettingCameFrom` (`storage` in the file, `STARTER_DATA_DIR` set, `log_format` and `port` neither) failed with every source `""`.
   - Green: `Load`'s `setting` lookup records `environment (<VAR>)`, `settings file` or `built in` as it picks each value. One map rather than separate koanf instances per layer (the plan's idea): the lookup already visits the layers in order.
   - Refactor: one table, `envVars` (key → environment variable), replaces `settingKeys` and the variable names in `Load`'s calls; `readSettingsFile` checks keys against it; the unused `getenv` helper is gone. Left: `LocalDataDir` still checks `STARTER_DATA_DIR` before the file itself; it returns a folder, not a `Config`, and the CLI (task 12) will mark `data_dir`'s source from what it chose.
11. [x] CLI: `starter info` prints `Settings file: <path>` or `none`.
   - Red: `TestInfoNamesTheSettingsFile`, a table (`--config <file>`, and no file), failed in both cases: `info` printed only the data line and the plan.
   - Green: `Load` keeps the file it read in `Config.SettingsFile`; `info` prints `Settings file: <path or none>` first.
   - Refactor: where the data is moved into `dataLocation(cfg)`, making room for task 12's list; `newInfoCmd`'s doc comment names the settings file. Left for task 15: `docs/data-locations.md`'s sample `starter info` output lacks the new line.
12. [ ] CLI: `starter info` lists each setting with its value and source; flags show as `option (--…)`. Scenario "where a setting came from" should pass.
13. [ ] Desktop: `desktop.Open` loads the local settings file (data folder, plans file). Scenario "desktop app uses the settings file's data folder" should pass.
14. [ ] Server: `cmd/server` loads `STARTER_CONFIG_FILE` when set, else no file.
15. [ ] Docs: new `docs/configuration.md` (every key, env var and option, precedence, file locations, a commented example), ADR 6, README's configuration paragraph (link), `docs/data-locations.md` ("Other files the app reads"), CLAUDE.md Gotchas (where the file is, strictness, the server reads one only when named). `.env.example` needs a hand check by the user (Claude can't read it here).
16. [ ] Finish: `task check`, `task test:init`, `task test:rename` (the path contains the slug), by hand (`starter info` with and without a file; the desktop app with `data_dir` set; `task dev` unaffected), remove `@wip`, tick the TODO item, delete this file.

## Changes to the plan
- 2026-11-02, task 6: split into 6a (the default path), 6b (`STARTER_CONFIG_FILE`) and 6c (a missing default file), one test each.
- 2026-11-02, task 5: passed straight away (task 2's error wrapping); kept as a guard.
- 2026-11-02, task 2: the simplest green (env, else file, else default per setting) already makes env override the file, so task 3's test will probably pass straight away; decide then whether to keep it as a guard. It did; kept as a guard.
- 2026-11-02, task 2: values aren't unmarshalled into `Config` (the plan said so); a lookup per setting is simpler.
