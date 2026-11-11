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
| `storage` | `STARTER_STORAGE` | `--storage` | `jsonfile` | Where the core keeps its data: `jsonfile` (plain JSON in a data folder, one program at a time), `sqlite` (one SQLite database, `data.db`, in the data folder, which several programs can use at once; [ADR 7](adr/0007-sqlite-storage.md)), `postgres` (a Postgres database, for hosted servers, at `database_url`; [ADR 8](adr/0008-postgres-storage.md)), or `memory` (lost when the process ends). The desktop app takes `jsonfile` or `sqlite`, refuses `postgres`, and uses `jsonfile` for anything else. The Android app uses `jsonfile`, and refuses `postgres` too. |
| `data_dir` | `STARTER_DATA_DIR` | `--data-dir` | see below | The data folder for `jsonfile` and `sqlite`. |
| `database_url` | `STARTER_DATABASE_URL` | | none | The database that `postgres` storage connects to, for example `postgres://user:password@host:5432/starter`. It usually holds a password, so set it in the environment, as a secret, rather than in a settings file. |
| `port` | `PORT` | | `8080` | The port the server listens on. `PORT` has no prefix: it's the name cloud platforms set. |
| `server_url` | `STARTER_SERVER_URL` | `--server` | none | Makes the CLI and TUI use a server at this URL instead of their own data. |
| `log_format` | `STARTER_LOG_FORMAT` | | `text` | The server's log format: `text` or `json`. |
| `plans_file` | `STARTER_PLANS_FILE` | | none | A plans file to use instead of the one built in (`internal/adapters/plans/plans.jsonc`). |
| `ignore_synced_folder` | `STARTER_IGNORE_SYNCED_FOLDER` | | `false` | `true` is for a folder that looks synced but isn't. It turns off the warning that SQLite's data folder looks synced (SPEC.md §4.14), and stops plain text marking the folder as in use on this device (§4.16). It doesn't turn off the conflict copy warning, as a conflict copy is itself the evidence. A value that isn't `true` or `false` stops the app at start. |

The default data folder depends on the front end. The server uses `./data`, relative to the folder it was started in. The CLI, TUI and desktop app share `starter` in the user's settings folder (`~/.config/starter` on Linux). [docs/data-locations.md](data-locations.md) has every case, on every OS.

## The settings file

It's [TOML](https://toml.io): one `key = value` per line, strings in quotes, `#` for comments.

When the desktop app, the CLI or the TUI starts and its default settings file (below) is missing, it creates one, with the values it would use without the file. On Linux, for a user called `you`, it looks like this:

```toml
# The settings for this application. The application created this file with the
# values it uses when no settings file exists.

# How the application keeps its data: "jsonfile" (plain-text storage, which
# one program on this computer can use at a time), "sqlite" (a database,
# which several programs can use at once), or "memory" (kept only until the
# program ends).
storage = "jsonfile"

# The folder in which the application keeps its data. The desktop
# application, the command line, and the terminal application share it.
data_dir = "/home/you/.config/starter"

# Set this to true if the data folder appears to be synced by a file sync
# service but is excluded from sync. The application then does not warn that
# SQLite data in the folder may be damaged, and plain-text storage does not
# mark the folder as in use on this computer.
ignore_synced_folder = false

# To make the command line and the terminal application use a server instead
# of the data on this computer, please remove the "#" below and give the
# server's address.
# server_url = "https://example.com"

# To use a plans file of your own instead of the one built into the
# application, please remove the "#" below and give the file's full path.
# plans_file = "/path/to/plans.jsonc"
```

`port`, `log_format`, and `database_url` are left out: they are for running a server, which never reads this file. Edit the file to change a setting: for example, `data_dir = "/home/you/Sync/starter"` keeps the data in a synced folder (one device at a time: see [docs/data-locations.md](data-locations.md)), and `storage = "sqlite"` lets the CLI work while the desktop app is open (but not in a synced folder, where a sync service can damage the database).

**When the file is created:**
- only when the default file is missing. A file that's already there is never changed, even if it's empty or sets nothing, and two programs starting at once can't both write it (it's opened with `O_EXCL`);
- never when `STARTER_CONFIG_FILE` or `--config` names another file;
- never by the server or the Android app;
- if it can't be written (a read-only folder, or no settings folder), the program carries on silently on the built-in settings, and tries again the next time it starts. Delete the file to get a fresh one.

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
- a file that `STARTER_CONFIG_FILE` or `--config` names, but that doesn't exist, is a mistake too. The default file may be missing: then the program creates it (above), or, if it can't, runs with no settings file.

## Where did a setting come from?

`starter info` names the settings file it read, then lists every setting the way the file would write it, with where it came from:

```console
$ STARTER_LOG_FORMAT=json starter info
Settings file: /home/you/.config/starter/config.toml
  storage = "jsonfile"  # settings file
  data_dir = "/home/you/Sync/starter"  # settings file
  database_url = ""  # built in
  port = 8080  # built in
  server_url = ""  # built in
  log_format = "json"  # environment (STARTER_LOG_FORMAT)
  plans_file = ""  # built in
  ignore_synced_folder = false  # settings file
Data folder: /home/you/Sync/starter
Plan: local
```

The sources are `built in`, `settings file`, `environment (STARTER_…)` and `option (--…)`. If a setting in your file seems to be ignored, this shows the environment variable or option that overrides it.

## In code

`internal/config` does all of this, and reads nothing from the process itself: `config.Env` gives it the environment and the user's settings folder (`os.Getenv` and `os.UserConfigDir` in `cmd/*`, fakes in tests).
- `config.CreateDefaultSettingsFile(env)` creates the default settings file when it's missing, as above. `desktop.Open` and the CLI's `settingsFileFor` (used by the TUI too) call it; `desktop.OpenWithoutDefaultSettingsFile`, which `mobile.Open` calls, doesn't.
- `config.DefaultSettingsFileText(dataDir)` is that file's text.
- `config.LocalConfigFile(env)` finds the local front ends' settings file (`""` for none).
- `config.Load(env, settingsFile)` reads the file and the environment, records each setting's source in `Config.Sources`, and validates the result.
- `config.LocalDataDir(env, settingsFile)` picks the local front ends' data folder.

To add a setting: add a field to `config.Config`, its key and environment variable to `envVars` in `internal/config/config.go`, its lookup in `Load` and its line in `Config.Settings`, then add it to the table above. If the desktop app, the CLI or the TUI reads it, add it to `config.DefaultSettingsFileText` too, with a note above it: set to its default, or commented out when it has no useful default. The note is user-facing text, so agree its wording first. Then change the example file above to match.
