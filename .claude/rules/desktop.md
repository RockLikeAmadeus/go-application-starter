---
paths:
  - "cmd/desktop/**"
  - "internal/adapters/desktop/**"
---

# Desktop app (Wails v3, ADR 4)

- `cmd/desktop` (build tag `desktop`) is a Wails v3 window around `desktop.Open`'s handler. `internal/adapters/desktop` holds the wiring with no Wails import: data folder, storage, in-process handler.
- Build: `task build:desktop` (`bin/starter-desktop`; needs the desktop build deps) · `task build:desktop:windows` (`bin/starter-desktop.exe`, cross-compiled; needs no desktop build deps). Run: `task run:desktop`, which uses your real data folder.
- On Linux it needs GTK 4 and WebKitGTK 6.0: `sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev` in a real terminal. Without them, the build fails with "Package 'gtk4' … not found". Wails' older GTK 3 stack (`-tags gtk3`) is dropped in Wails v3.1, so don't rely on it. Plain `go build ./...` and `task check` skip `cmd/desktop`, so they don't need these packages.
- `desktop.Open` takes `jsonfile` or `sqlite` from the settings (never memory). While the app is open, the embedded CLI and TUI (and a second copy of the app) fail fast with "data directory is locked" on jsonfile (see `.claude/rules/storage.md`).
- In the local edition, the desktop app keeps its data on the computer (`httpapi.Options.LocalData`; remote mode, which is planned, will leave it unset), so `GET /api/v1/connection` answers `{"required": false}`. The SPA then hides the offline notice and keeps TanStack Query online (`useConnectionRequired`, `useLocalDataOnline`). New offline behaviour in the SPA must respect that answer. The desktop handler answers `/sw.js` with 404, so the service worker never runs there.
- The app ID is `com.example.starter`, set in `cmd/desktop`. Packaging is in `.claude/rules/desktop-packaging.md`.
- WebKitGTK drops registered `--tw-*` variables inside `@starting-style` (see `.claude/rules/web-styles.md`).
- Keep the Wails versions in step (see `.claude/rules/dependencies.md`).
