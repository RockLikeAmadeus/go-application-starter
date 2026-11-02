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
3. [ ] `config`: env vars override the settings file.
4. [ ] `config`: an unknown key fails, naming the file and the key.
5. [ ] `config`: bad TOML fails, naming the file.
6. [ ] `config.LocalConfigFile(env)`: `STARTER_CONFIG_FILE`, else `<settings folder>/starter/config.toml`; a missing default file means no file, a missing named file is an error.
7. [ ] `config.LocalDataDir` uses the loaded `data_dir` (from the file or env) before the settings folder.
8. [ ] CLI: `NewRootCmd` takes `config.Env` instead of `localDataDir`; commands load the local settings file. Update `cmd/starter`, the test helpers and `runCommandLine`. Scenarios "chooses the data folder" and "doesn't know stops it" should pass.
9. [ ] CLI: `--config` names the settings file for one command.
10. [ ] `config`: `Load` reports each setting's source (built in, settings file, environment variable by name).
11. [ ] CLI: `starter info` prints `Settings file: <path>` or `none`.
12. [ ] CLI: `starter info` lists each setting with its value and source; flags show as `option (--…)`. Scenario "where a setting came from" should pass.
13. [ ] Desktop: `desktop.Open` loads the local settings file (data folder, plans file). Scenario "desktop app uses the settings file's data folder" should pass.
14. [ ] Server: `cmd/server` loads `STARTER_CONFIG_FILE` when set, else no file.
15. [ ] Docs: new `docs/configuration.md` (every key, env var and option, precedence, file locations, a commented example), ADR 6, README's configuration paragraph (link), `docs/data-locations.md` ("Other files the app reads"), CLAUDE.md Gotchas (where the file is, strictness, the server reads one only when named). `.env.example` needs a hand check by the user (Claude can't read it here).
16. [ ] Finish: `task check`, `task test:init`, `task test:rename` (the path contains the slug), by hand (`starter info` with and without a file; the desktop app with `data_dir` set; `task dev` unaffected), remove `@wip`, tick the TODO item, delete this file.

## Changes to the plan
- 2026-11-02, task 2: the simplest green (env, else file, else default per setting) already makes env override the file, so task 3's test will probably pass straight away; decide then whether to keep it as a guard.
- 2026-11-02, task 2: values aren't unmarshalled into `Config` (the plan said so); a lookup per setting is simpler.
