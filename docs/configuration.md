# Configuration

Every front end takes the same settings, from up to four places. When a setting is in more than one, the later place wins:

1. **Built in:** the defaults below.
2. **The settings file:** an optional TOML file.
3. **The environment:** a `STARTER_…` variable (or `PORT`).
4. **A command-line option:** the CLI's `--data-dir`, `--storage` and `--server`, for one command.

So the settings file holds what you want every time, and an environment variable or an option overrides it for one run. Why environment variables beat the file, and why there is a file at all, is in [ADR 6](adr/0006-layered-configuration.md).

## The settings

| Key in the file | Environment variable | CLI option | Default | What it does |
|---|---|---|---|---|
| `storage` | `STARTER_STORAGE` | `--storage` | `jsonfile` | Where the core keeps its data: `jsonfile` (plain JSON in a data folder) or `memory` (lost when the process ends). The desktop app always uses `jsonfile`. |
| `data_dir` | `STARTER_DATA_DIR` | `--data-dir` | see below | The data folder for `jsonfile`. |
| `port` | `PORT` | | `8080` | The port the server listens on. `PORT` has no prefix: it's the name cloud platforms set. |
| `server_url` | `STARTER_SERVER_URL` | `--server` | none | Makes the CLI and TUI use a server at this URL instead of their own data. |
| `log_format` | `STARTER_LOG_FORMAT` | | `text` | The server's log format: `text` or `json`. |
| `plans_file` | `STARTER_PLANS_FILE` | | none | A plans file to use instead of the one built in (`internal/adapters/plans/plans.jsonc`). |

The default data folder depends on the front end. The server uses `./data`, relative to the folder it was started in. The CLI, TUI and desktop app share `starter` in the user's settings folder (`~/.config/starter` on Linux). [docs/data-locations.md](data-locations.md) has every case, on every OS.

## The settings file

It's [TOML](https://toml.io): one `key = value` per line, strings in quotes, `#` for comments.

```toml
# ~/.config/starter/config.toml

# Keep the data in a synced folder (one device at a time: see docs/data-locations.md).
data_dir = "/home/you/Sync/starter"

# Try out a plan's limits.
# plans_file = "/home/you/plans.jsonc"
```

**Which file is read:**
- **The CLI, TUI and desktop app** read `config.toml` in the app's folder in the user's settings folder, beside the default data folder:
  - Linux: `~/.config/starter/config.toml` (`$XDG_CONFIG_HOME/starter/config.toml`);
  - Windows: `%AppData%\starter\config.toml`;
  - macOS: `~/Library/Application Support/starter/config.toml`.

  `STARTER_CONFIG_FILE` names another file, and the CLI's `--config` does for one command.
- **The server** reads a file only when `STARTER_CONFIG_FILE` names one. It never looks in the user's settings folder, so a developer's own settings file doesn't change `task dev`'s server.
- **The Android app** reads none: its settings folder is the app's private storage, where nobody can put one.

**Mistakes stop the app.** The app checks the file when it starts, and refuses to start on settings it would ignore:
- a setting it doesn't know (a typo such as `datadir`) names the file and the setting;
- a file that isn't valid TOML names the file;
- a file that `STARTER_CONFIG_FILE` or `--config` names, but that doesn't exist, is a mistake too. The default file may be missing: then there is no settings file.

## Where did a setting come from?

`starter info` names the settings file it read, then lists every setting the way the file would write it, with where it came from:

```console
$ STARTER_LOG_FORMAT=json starter info
Settings file: /home/you/.config/starter/config.toml
  storage = "jsonfile"  # built in
  data_dir = "/home/you/Sync/starter"  # settings file
  port = 8080  # built in
  server_url = ""  # built in
  log_format = "json"  # environment (STARTER_LOG_FORMAT)
  plans_file = ""  # built in
Data folder: /home/you/Sync/starter
Plan: local
```

The sources are `built in`, `settings file`, `environment (STARTER_…)` and `option (--…)`. If a setting in your file seems to be ignored, this shows the environment variable or option that overrides it.

## In code

`internal/config` does all of this, and reads nothing from the process itself: `config.Env` gives it the environment and the user's settings folder (`os.Getenv` and `os.UserConfigDir` in `cmd/*`, fakes in tests).
- `config.LocalConfigFile(env)` finds the local front ends' settings file (`""` for none).
- `config.Load(env, settingsFile)` reads the file and the environment, records each setting's source in `Config.Sources`, and validates the result.
- `config.LocalDataDir(env, settingsFile)` picks the local front ends' data folder.

To add a setting: add a field to `config.Config`, its key and environment variable to `envVars` in `internal/config/config.go`, its lookup in `Load` and its line in `Config.Settings`, then add it to the table above.
