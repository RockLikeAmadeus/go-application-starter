---
paths:
  - "cmd/desktop/**"
  - "internal/adapters/desktop/**"
---

# Desktop app (Wails v3, ADR 4)

- `cmd/desktop` (build tag `desktop`) is a Wails v3 window around `desktop.Open`'s handler. `internal/adapters/desktop` holds the wiring with no Wails import: data folder, storage, in-process handler, and remote mode.
- Build: `task build:desktop` (`bin/starter-desktop`; needs the desktop build deps) · `task build:desktop:windows` (`bin/starter-desktop.exe`, cross-compiled; needs no desktop build deps). Run: `task run:desktop`, which uses your real data folder.
- On Linux it needs GTK 4 and WebKitGTK 6.0: `sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev` in a real terminal. Without them, the build fails with "Package 'gtk4' … not found". Wails' older GTK 3 stack (`-tags gtk3`) is dropped in Wails v3.1, so don't rely on it. Plain `go build ./...` and `task check` skip `cmd/desktop`, so they don't need these packages.
- `desktop.Open` takes `jsonfile` or `sqlite` from the settings (never memory), and refuses `postgres`, before choosing a data folder: the desktop and mobile apps reach hosted data through a server (remote mode), never the database. While the app is open, the embedded CLI and TUI (and a second copy of the app) fail fast on jsonfile with "Your data folder is in use by another program on this computer ({path})." (see `.claude/rules/storage.md`).
- In the local edition, the desktop app keeps its data on the computer (`httpapi.Options.IsDataLocal`), so `GET /api/v1/connection` answers `{"isRequired": false}`. The SPA then hides the offline notice and keeps TanStack Query online (`useIsConnectionRequired`, `useLocalDataOnline`). New offline behaviour in the SPA must respect that answer. Remote mode leaves `IsDataLocal` unset, so the answer is `true` and the offline notice comes back. The desktop handler answers `/sw.js` with 404, so the service worker never runs there.
- **Remote mode** (ADR 4, `docs/sign-in.md`): with `server_url` set (the setting, `STARTER_SERVER_URL`, or `config.BuiltInServerURL`, from `BUILT_IN_SERVER_URL` in the Taskfile), `desktop.Open` opens `bootstrap.NewRemote` (program `DesktopOrMobileApp`) behind the same handler, with no data folder, storage, or lock, and the `storage` setting ignored. It sets `httpapi.Options.ClientSignIn` (`bootstrap.NewAppClientSignIn`) and `OpenBrowser` from the opener `Open` takes; `cmd/desktop` passes Wails' `Browser.OpenURL`. The apps keep their own token (keyring service `go-application-starter-app`, or `app-tokens.json`), apart from the command line's.
- The app ID is `com.example.starter`, set in `cmd/desktop`. Packaging is in `.claude/rules/desktop-packaging.md`.
- WebKitGTK drops registered `--tw-*` variables inside `@starting-style` (see `.claude/rules/web-styles.md`).
- Keep the Wails versions in step (see `.claude/rules/dependencies.md`).
