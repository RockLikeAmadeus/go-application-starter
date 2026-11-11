---
paths:
  - "build/desktop/**"
  - "test/packaging/**"
  - "cmd/desktop/**"
  - ".github/workflows/release.yml"
  - "VERSION"
---

# Desktop packaging

Commands: `task package:desktop:linux` (`.deb`, `.rpm` and AppImage in `bin/`; one format: `…:deb`, `…:rpm`, `…:appimage`) · `task package:desktop:windows` (`bin/starter-desktop-<version>-setup.exe`; needs `makensis`) · check them: `task test:package` (slow; not part of `task check`).

`build/desktop/linux` holds the Linux packaging: `nfpm.yaml` (.deb, .rpm), `app.desktop`, install script. `build/desktop/windows` holds the Windows packaging: `icon.ico`, `app.manifest`, `info.json` (version info), `installer.nsi` + `wails_tools.nsh` (NSIS).

## Linux packages

The desktop file and icons are installed under the app ID (`com.example.starter`, set in `cmd/desktop`), so the desktop can match the window to them, but their sources keep generic names, because `task rename` rewrites names inside files and moves only `cmd/<slug>`. Package versions come from `VERSION`: a release tag (`v1.2.3`) loses its `v`, anything else becomes `0.0.0`. The AppImage bundles every WebKitGTK on the build machine, so build it where only WebKitGTK 6.0 is installed (not `libwebkit2gtk-4.1-dev`). `task test:package` runs the package check (`test/packaging`, build tag `packaging`); it takes a few minutes, mostly the AppImage.

## Windows setup program

Built on Linux. Both tasks set `CGO_ENABLED=0`: the program needs no C compiler, and it also keeps `go tool wails3` from linking GTK, so neither needs the desktop build deps (check with `PKG_CONFIG_LIBDIR=/nonexistent`, which hides them). Only `go build` gets `GOOS=windows`: set for the whole task, `go tool wails3` would build a Windows `wails3.exe`. Its icon, manifest, and version info go in through `cmd/desktop/rsrc_windows_amd64.syso`, which the task writes and deletes (gitignored); the WebView2 bootstrapper is written beside `installer.nsi` for `makensis` the same way. `info.json` is a template (`@VERSION@`) and the one source of the product name, publisher, and copyright for both the program and the setup program. It installs for the current user (`WAILS_INSTALL_SCOPE=user`: `%LOCALAPPDATA%\Programs`, no admin prompt) and keeps the items in `%AppData%\starter`; its uninstaller deletes only WebView2's `%AppData%\starter-desktop.exe`. `wails_tools.nsh` is Wails' generated file: replace it, don't edit it, on a Wails upgrade. The package check reads what the setup program does from `bin/windows/makensis.log` (`makensis -V4`), because the script inside it is compressed. It's unsigned, so Windows SmartScreen warns before it runs ("More info" → "Run anyway").
