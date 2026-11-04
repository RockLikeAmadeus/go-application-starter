# Where the data lives

This page lists where each part of the app keeps its data: the user's records, and the few smaller things kept beside them. Each depends on how the app is run and on which operating system.

Names on this page are the project's own: the slug `starter`, the `STARTER_` environment variables and the app ID `com.example.starter`. `task init` and `task rename` keep them up to date (see [docs/template.md](template.md) and [docs/rename.md](rename.md)).

## The rules behind every case

- **One data folder per copy.** The core keeps its data through a storage adapter, which `internal/bootstrap` picks. With `jsonfile` (the default), the data is plain JSON in a **data folder**: one file per kind of record (the sample's is `items.json`), plus a `.lock` file (see [ADR 3](adr/0003-plain-text-storage.md)). With `sqlite`, it's one SQLite database in the data folder, `data.db`, with SQLite's `data.db-wal` and `data.db-shm` beside it (see [ADR 7](adr/0007-sqlite-storage.md)). With `memory`, nothing is written: the data is lost when the process ends.
- **Choosing another folder.** `data_dir` in the settings file chooses the folder for every front end that reads that file ([docs/configuration.md](configuration.md)). `STARTER_DATA_DIR` overrides the file, and the CLI's `--data-dir` overrides both for one command.
- **One process per folder, with `jsonfile`.** While a copy has a data folder open, it holds a lock on `.lock`. A second process that opens the same folder fails fast with "data directory is locked". The lock uses `flock` on Linux and macOS, and `LockFileEx` on Windows.
- **Several processes per folder, with `sqlite`.** There's no lock: the desktop app, the CLI and TUI, and several servers on one disk can all use `data.db` at once, and each sees what the others saved. They must be on the same computer (or disk): SQLite can't be shared over a network drive.
- **Temporary files.** With `jsonfile`, saving writes a `<file>.*.tmp` beside each JSON file and renames it over the file. The server's readiness check writes and removes `.ready-*.tmp`. Only a crash at the wrong moment leaves either one behind.
- **Newer data is refused.** With `sqlite`, opening the data brings its layout up to date with the app. Data saved by a newer version of the app isn't opened ("the data was saved by a newer version of the app") and is left unchanged.
- **Nothing syncs.** Each copy has its own data. Copying a data folder (or putting it in a cloud-synced folder) works, but don't let two devices use it at the same time ([ADR 5](adr/0005-hosted-and-local-editions.md)). Don't put a `sqlite` data folder in a cloud-synced one at all: a sync service can damage the database. When it looks synced (by its path, or the files a sync service leaves in it or a folder above it), every front end warns, naming the service, and keeps working: the desktop app's and the server's web app show a notice, the CLI prints `warning: …` on stderr, the TUI shows a line above its screen, and the server logs it at start. It's a best guess: `ignore_synced_folder = true` ([docs/configuration.md](configuration.md)) turns it off.
- **Workspaces share the files.** Each record carries its `workspaceId` (in SQLite, a `workspace_id` column). A hosted server keeps every workspace's records in the same files, and the local edition's one workspace is just the only ID in them.

## At a glance

| How it runs | Where the data is | Chosen by |
|---|---|---|
| Server (`bin/server`, `task run:server`) | `./data`, relative to the folder it was started in | `STARTER_DATA_DIR`, `STARTER_STORAGE`, or the settings file `STARTER_CONFIG_FILE` names |
| `task dev` | `./data` in the repository (gitignored) | as for the server |
| Docker (`task docker:up`) | `/data` in the container, on the named volume `app-data` | `Dockerfile`, `docker-compose.yml` |
| Fly.io (staging, production) | `/data`, on the Fly volume `app_data` | `fly.staging.toml`, `fly.production.toml` |
| CLI and TUI, on their own | The desktop app's data folder (below), wherever they're run from | `--data-dir`, `STARTER_DATA_DIR`, the settings file, `--storage` |
| CLI and TUI with `--server` | On that server: nothing is stored locally | `--server`, `STARTER_SERVER_URL` |
| Desktop app, Linux | `~/.config/starter` (`$XDG_CONFIG_HOME/starter`) | the settings file, `STARTER_DATA_DIR` |
| Desktop app, Windows | `%AppData%\starter` (`C:\Users\<you>\AppData\Roaming\starter`) | the settings file, `STARTER_DATA_DIR` |
| Desktop app, macOS (not tried yet) | `~/Library/Application Support/starter` | the settings file, `STARTER_DATA_DIR` |
| Android app | `/data/data/com.example.starter/files/starter`, the app's private storage | fixed |
| Web app in a browser | On the server it was loaded from | — |
| Browser e2e tests (`task e2e`) | In memory, never on disk | `STARTER_STORAGE=memory` in `web/playwright.config.ts` |
| Acceptance tests (`task accept`) | In memory, or a throwaway temp folder per scenario | `test/acceptance` |

## The server

`cmd/server` reads its settings from the environment, over the settings file `STARTER_CONFIG_FILE` names, if any ([docs/configuration.md](configuration.md)):

- `STARTER_STORAGE` (`storage` in the file) is `jsonfile` (the default), `sqlite` or `memory`.
- `STARTER_DATA_DIR` (`data_dir`) is the data folder, `./data` by default. That path is relative to the process's working folder, not to the binary, so start the server from the same place each time, or set an absolute path.

**In Docker**, the image sets `STARTER_STORAGE=jsonfile` and `STARTER_DATA_DIR=/data`, and declares `/data` a volume. `docker-compose.yml` mounts the named volume `app-data` there, so the data survives `docker compose down` and rebuilds. Docker prefixes the volume's name with the Compose project's (`docker volume ls`), and `docker compose down -v` deletes it.

**On Fly.io**, both apps mount the Fly volume `app_data` at `/data`. A Fly volume belongs to one machine, so run one machine per app, on `jsonfile` (because of the lock) or `sqlite` (which can't be shared between machines) alike.

**Hosted edition (planned):** one server holds every workspace's data, in the same files, each record tagged with its workspace. The Postgres adapter (planned) will move this into a database server. The data folder then gives way to a connection string.

## The CLI and the TUI

On their own (embedded), the CLI and TUI run the core in-process and open the data folder themselves. They use **the desktop app's data folder** (see the next section), so the three share one set of data, and the folder you run a command from makes no difference (SPEC.md §4.11). The folder is, in order:

1. `--data-dir`, for one command;
2. `STARTER_DATA_DIR`, as for the desktop app;
3. `data_dir` in the settings file (`config.toml` beside the default folder, or the one `STARTER_CONFIG_FILE` or `--config` names), as for the desktop app;
4. `starter` in the user's settings folder: `~/.config/starter` on Linux, `%AppData%\starter` on Windows, `~/Library/Application Support/starter` on macOS.

`starter info` says which folder that is, where each setting came from, and the workspace's plan:

```console
$ starter info
Settings file: none
  storage = "jsonfile"  # built in
  data_dir = "/home/you/.config/starter"  # built in
  port = 8080  # built in
  server_url = ""  # built in
  log_format = "text"  # built in
  plans_file = ""  # built in
Data folder: /home/you/.config/starter
Plan: local
```

With `--storage memory`, nothing is saved: `starter info` says the data is kept in memory, and no folder is looked up.

On `jsonfile`, because of the lock, the CLI and TUI can't open the folder while the desktop app (or a server) has it open: they say "data directory is locked" and change nothing. Close the app first, or point the CLI at a running server: `starter --server http://localhost:8080 …`. On `sqlite` (`storage = "sqlite"` in the settings file, which the desktop app reads too), there's no lock, so they can use the folder while the app is open. With `--server` (or `STARTER_SERVER_URL`), the CLI and TUI store nothing locally, and `starter info` names the server instead.

During development, `task run:cli` and `task run:tui` use the same folder as `task run:desktop`: your real one. `task dev`'s server keeps the repository's `./data`.

## The desktop app

The desktop app (`cmd/desktop`) uses `sqlite` when the settings say `storage = "sqlite"`, and `jsonfile` otherwise (never `memory`). Its data folder, which the CLI and TUI share, is `STARTER_DATA_DIR` when set, else `data_dir` in the settings file (`config.toml` in the folder below, or the one `STARTER_CONFIG_FILE` names), and otherwise `starter` inside the user's settings folder (Go's `os.UserConfigDir`; see `internal/config/localdata.go`):

| OS | Data folder |
|---|---|
| Linux, including the AppImage | `$XDG_CONFIG_HOME/starter`, which is `~/.config/starter` when `XDG_CONFIG_HOME` isn't set |
| Windows | `%AppData%\starter` |
| macOS | `~/Library/Application Support/starter` |

macOS is what Go's `os.UserConfigDir` returns there; the macOS app hasn't been built or tried yet.

**Installing and uninstalling** never touch the data folder:
- On Linux, the `.deb` and `.rpm` put the program in `/usr/bin/starter-desktop`, and the AppImage is one file wherever you keep it. Removing the package, or deleting the AppImage, leaves `~/.config/starter` alone.
- On Windows, the setup program installs the program for the current user into `%LocalAppData%\Programs\Starter`. Uninstalling it leaves `%AppData%\starter` alone.

**The web view's own data.** The window is a system web view, and it keeps its own data beside the app's: the chosen style and theme (local storage), plus caches. None of it is the user's records.

| OS | Web view data | Removed on uninstall |
|---|---|---|
| Linux (WebKitGTK) | `~/.local/share/com.example.starter/` (local storage) and `~/.cache/com.example.starter/`, named after the app ID | No |
| Windows (WebView2) | `%AppData%\starter-desktop.exe\` (named after the program) | Yes, by the uninstaller |
| macOS (WKWebView) | Not checked yet | — |

## The Android app

The Android app (`cmd/mobile`) uses the desktop app's wiring, with the app's private storage in place of the settings folder: `/data/data/com.example.starter/files/starter`. Other apps can't read it, and neither can the user without root. A debug build can be inspected with `adb shell run-as com.example.starter ls files/starter` (see [docs/mobile.md](mobile.md)).

- Updating the app (or `adb install -r`) keeps the data. Uninstalling it deletes the data, along with the WebView's own storage (the chosen style and theme).
- The data belongs to the app ID. Changing `com.example.starter` after people have installed the app leaves their data behind in the old app.

## The web app in a browser

The SPA keeps no records itself: it reads and writes them through the server's API, so they live wherever that server keeps them (see above). The service worker caches only the built app, never `/api` ([docs/installable-app.md](installable-app.md)).

The browser keeps two settings per site in local storage: `theme` and `style`. They belong to that browser and that site address, so another browser, a private window, or the same app on another address starts with the defaults. Clearing the site's data resets them, and leaves the records alone.

## Other files the app reads

- **The settings file.** Optional: `config.toml` in the app's folder in the user's settings folder, beside the default data folder (`~/.config/starter/config.toml` on Linux), for the CLI, TUI and desktop app. `STARTER_CONFIG_FILE` names another, and the server reads one only then. Nothing writes to it. See [docs/configuration.md](configuration.md).
- **The plans file.** The plans are built into the binary (`internal/adapters/plans/plans.jsonc`). `STARTER_PLANS_FILE` names a file to read instead. Nothing writes to it.
- **Logs** go to standard output (`STARTER_LOG_FORMAT`), never to a file. In Docker or on Fly.io, the platform keeps them.
