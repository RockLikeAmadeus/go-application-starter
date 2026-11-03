# Plan: desktop packaging (installers and app icon)

- TODO item: "Desktop packaging" in [`TODO.md`](../../TODO.md).
- Temporary: delete this file in the same change that ticks that TODO item. Move any open follow-ups into `TODO.md` first.
- Each step is its own `/feature` → `/red` → `/green` → `/refactor` cycle with its own `docs/wip/<slug>.md`. Stop for review between steps.

## Decisions
- **Tooling: the `wails3` CLI already in `tools/go.mod`.** It has everything needed: `generate icons` (`.ico`, `.icns`), `generate .desktop`, `generate appimage`, `generate syso` (Windows icon and version info), `tool package` (nfpm: `.deb`, `.rpm`; and `.dmg`), and `generate build-assets` (templates for `Info.plist`, NSIS, nfpm). We call these from our own Taskfile, as we do for Android, because this repo isn't laid out as a Wails project (no `wails3 task`).
- **Formats (chosen 2026-10-30):**
  - Linux: AppImage, `.deb` and `.rpm`.
  - Windows: an NSIS setup `.exe`, not `.msi`: NSIS is Wails' usual Windows installer, and `makensis` runs on Linux (`apt install nsis`).
  - macOS: a `.app` bundle in a `.dmg`.
- **Unsigned for now.** Document the macOS Gatekeeper and Windows SmartScreen warnings, and how to get past them. Signing from environment variables (`wails3 sign`) is a TODO.md follow-up.
- **Where each one is built and tried:**
  - Linux: built and tried on this machine.
  - Windows: cross-compiled here (Wails v3 needs no CGO on Windows: `CGO_ENABLED=0 GOOS=windows`). The user has a Windows PC or VM to install it on and try it.
  - macOS: it needs CGO and the macOS SDK, so it can't be cross-compiled from Linux. It's built and tried **on the user's Mac** (the Xcode Command Line Tools are enough; full Xcode is only for iOS). Release builds come from a GitHub Actions `macos-latest` runner (step 5), so releases don't depend on that Mac.
- **One icon source:** `web/public/icon.svg`, as for the web app. Generated files are committed, like the PWA PNGs. Each format is generated in the step that first uses it.
- **Packaging files live in `build/desktop/`** (nfpm config, NSIS script, `Info.plist`, `.desktop` file, icons), generated once with `wails3 generate build-assets` and then trimmed to what we use. `docs/desktop.md` says how to regenerate them when Wails is upgraded.
- **Names and IDs:** package name `starter`, display name "Starter", bundle ID `com.example.starter`, built by the same rule as the Android app ID (`com.example.<slug>` without hyphens: see `docs/mobile.md`, "The app ID"), so reuse that code. `task rename` and `task init` must cover every new occurrence.
- **Version:** from the git tag in CI; `0.0.0-dev` locally. The CLI already gets its version through ldflags (`.goreleaser.yaml`), so do the same for the desktop app.
- **How it's tested:** packaging is mostly build tooling, so Godog has little to say. Each `/feature` should decide whether a scenario fits (for example, "the desktop app reports its version"). Otherwise, unit-test whatever Go we add (version, rename coverage), add a smoke check that the package builds and contains the expected files (`task test:package` or similar), and try it by hand.

## Steps

### 1. App icon — done
SPEC.md §4.6 (the icon rule). `desktop.Open` returns the web app's `pwa-512x512.png` as `App.Icon` (`iconOf` in `internal/adapters/desktop/open.go`, tested on the real `web/public`), and `cmd/desktop` passes it to Wails with the GTK app ID `com.example.starter`. `task rename` already covered the ID; `androidAppID` became `appID`, and `docs/rename.md` and `docs/template.md` say both apps use it. Checked by hand on Linux: the app runs, with the generic icon expected until step 2.
- `desktop.Open` returns the product's icon: the web app's `pwa-512x512.png`, which `pnpm --dir web icons` already makes from `icon.svg`. `cmd/desktop` hands it to Wails (`application.Options.Icon`), which shows it in the Windows title bar and taskbar and the macOS Dock.
- `cmd/desktop` sets the GTK application ID to `com.example.starter`, so step 2's `.desktop` file can match the window.
- **Linux can't show it yet:** Wails v3's GTK 4 `setIcon` does nothing, because GTK 4 dropped window icons. The desktop takes the icon from the installed `.desktop` file, so it first shows in step 2.

### 2. Linux packages — done
SPEC.md §4.6 (the Linux install rule). `task package:desktop:linux` builds `bin/starter-desktop_<version>_amd64.deb`, `…-<version>-1.x86_64.rpm` and `…-<version>-x86_64.AppImage` (one format alone: `:deb`, `:rpm`, `:appimage`) from `build/desktop/linux/` (`nfpm.yaml`, `app.desktop`, `refresh.sh`). They install `/usr/bin/starter-desktop`, the desktop file and icons under the app ID, with GTK 4/WebKitGTK 6.0 dependencies (Debian and rpm names), 0755/0644 modes, and scripts that refresh the app list and icon cache. Versions: a release tag loses its `v`, anything else is `0.0.0`. Checked by `test/packaging` (`task test:package`, 11 tests, about 2.5 minutes), by `task rename`'s new "no file is named after the old names" step, and by hand on Pop!_OS (install, app list and dock icon, remove keeps the data, AppImage). The placeholder `Categories=Utility;`, `maintainer` and missing `description` are steps for template users in TODO.md.
- Follow-ups for step 5: build release packages only in CI (a runner with only WebKitGTK 6.0, so the AppImage doesn't bundle 4.1); decide whether CI runs `task test:package`; `linuxdeploy` comes unpinned from GitHub's "continuous" release (Wails' choice).
- Other follow-up: arm64 packages (a cross toolchain or an arm64 runner). (`test/packaging` is linted: `.golangci.yml` sets the `packaging` build tag, for `task lint` and CI alike.)

### 3. Windows installer — done
SPEC.md §4.6 (the Windows install rule). `task package:desktop:windows` cross-compiles `bin/starter-desktop.exe` (`task build:desktop:windows`: `CGO_ENABLED=0`, a GUI program, with the icon, manifest and version info from a temporary `.syso`) and wraps it in `bin/starter-desktop-<version>-setup.exe` with `makensis`, from `build/desktop/windows/` (`installer.nsi` trimmed from Wails' template, `wails_tools.nsh` as generated, `app.manifest`, `info.json`, `icon.ico` made by `pnpm --dir web icons`). It installs for the current user (`%LOCALAPPDATA%\Programs\Starter`, no admin prompt), adds a Start menu shortcut only, carries the WebView2 bootstrapper, and its uninstaller removes WebView2's `%AppData%\starter-desktop.exe` but keeps the items in `%AppData%\starter`. `info.json` is the one source of the product name, publisher and copyright for both programs (placeholders are a template-user step in TODO.md). Checked by `test/packaging/nsis_test.go` (`debug/pe`, 7-Zip, and `makensis -V4`'s log in `bin/windows/makensis.log`), and by hand on Windows on 2026-11-02: SmartScreen warning, install, Start menu icon, items kept across restarts, Properties and Task Manager names, uninstall, reinstall with the items back.
- Generate `icon.ico` from `icon.svg` (`wails3 generate icons`), from the same command as the other icons.
- Cross-compile `bin/starter-desktop.exe` on Linux, with the icon and version info embedded (`wails3 generate syso`).
- `build/desktop/windows/`: the NSIS script (Start menu shortcut, uninstaller, and the WebView2 bootstrapper if Wails' template includes it).
- `task package:desktop:windows` → `bin/starter-desktop-<version>-setup.exe`. `makensis` is a new system dependency: README Prerequisites, `task setup`, Gotchas.
- Try it by hand on the Windows PC or VM: install, launch from the Start menu, create and rename items, restart and check the data persists (under `%AppData%`), uninstall. Note the SmartScreen warning.

### 4. macOS app and `.dmg` — not started
- Generate `icon.icns` from `icon.svg` (`wails3 generate icons`), from the same command as the other icons.
- `build/desktop/darwin/Info.plist` (bundle ID, name, version, icon).
- `task package:desktop:macos` → `bin/Starter.app` in `bin/starter-desktop-<version>.dmg`. It only runs on macOS; on Linux it says so and exits.
- A universal binary (`wails3 tool lipo`, arm64 and amd64), so one `.dmg` runs on Apple silicon and Intel Macs.
- Set up the Mac: the Xcode Command Line Tools (`xcode-select --install`), Go, Node ≥ 24, pnpm and Task. Check `task setup` and `task run:desktop` work there too, and record anything macOS-specific in README Prerequisites.
- Try it by hand on the Mac: open the `.dmg`, drag the app to Applications, get past Gatekeeper (right-click → Open, or System Settings → Privacy & Security), launch it, create and rename items, quit and relaunch to check the data persists (under `~/Library/Application Support`), and check the Dock icon and the app menu's name.

### 5. Release wiring and housekeeping — in progress (Linux and Windows part)
- **Split 2026-11-02** (see the log): the Linux and Windows part is done now, while the Mac is being set up; the `.dmg` job and its docs come after step 4. Done in small pieces, with each CI idea explained as it comes, at the user's request (new to CI/CD).
- Done so far: `release.yml` builds the Linux packages and the Windows setup program on a `v*` tag or **Run workflow**, kept as artifacts (jobs `desktop-linux`, `desktop-windows`, on `ubuntu-24.04`); both tried by hand on GitHub 2026-11-02. A `publish` job adds both to the tag's Release once GoReleaser and the builds succeed; docs/ci-cd.md has "Making a release".
- Attach all the installers to the GitHub Release on a `v*` tag (`release.yml`), beside GoReleaser's CLI archives. Linux and Windows build on `ubuntu-latest`; the `.dmg` builds on `macos-latest` with the same `task package:desktop:macos` used on the Mac.
- `docs/desktop.md`: building each package, regenerating the Wails assets, the unsigned-app warnings, and what's been tried on which OS.
- CLAUDE.md: Commands rows, the `build/desktop/` architecture line, Gotchas. SPEC.md §4.6: how the desktop app is installed.
- Run `task rename -- --name "Acme Tasks" --dry-run`, `task test:rename` and `task test:init` to check every new name and ID is covered.
- TODO.md: tick the item, add follow-ups (signing, notarization, auto-update), and delete this file.

## Verification after every step
`task check` (lint, unit, acceptance, generated-code drift and e2e), plus that step's by-hand check.

## Changes to the plan
- 2026-10-30: the user has a Mac after all. macOS is now built and tried by hand there (step 4) instead of only in CI; the CI `macos-latest` build moves to release wiring (step 5). Dropped the "trying macOS for real" follow-up.
- 2026-10-30, step 1: Wails v3 can't set a window icon on Linux (GTK 4), so step 1 is slimmed to handing the icon to Wails (seen on Windows and macOS) and setting the GTK app ID; Linux shows the icon from step 2's `.desktop` file. `.ico` and `.icns` move to steps 3 and 4, where they're first used. The icon is the web app's `pwa-512x512.png`, read from the built web app, so there's no second copy. No Godog scenario: it's presentation, like §4.5.
- Step 1 done. Along the way, the phone-width menu drawer was found not to slide in under WebKitGTK 2.52 (the desktop app's engine), which drops a `@starting-style` declaration that reads an unset registered custom property, as Tailwind's `-translate-x-full` does. Fixed separately in `web/src/routes/Layout.tsx` (a plain `translate` value), checked with a WebKitGTK script and by hand. No test covers it: neither jsdom nor Playwright's newer WebKit shows the bug.
- Step 2 done. Changes along the way: one task per format, so tests can build the `.deb` alone; the desktop file has a placeholder category after all, because `appimagetool` requires one; source files keep generic names, because `task rename` doesn't rename files; and checks the plan didn't list (file modes, a real maintainer, the `.rpm`'s own dependency names).
- Step 3 done. Changes along the way: per-user install only and a Start menu shortcut only (chosen 2026-10-31); the setup program's publisher and product name come from `info.json` too, so a project fills them in once; 7-Zip can't show the compressed install script, so the install folder, shortcuts and uninstaller are read from `makensis -V4`'s log, and the admin level from the manifest resource. The Linux package build was switched off in the package check while the step was in progress, and is back on.
- 2026-11-02: step 5's Linux and Windows release wiring and docs pulled ahead of step 4, at the user's request, because step 4 waits on setting up the Mac. Step 5 finishes after step 4 with the `macos-latest` `.dmg` job and the macOS docs.
- 2026-11-02: added `docs/ci-cd.md` (plain-language guide to the workflows, Dependabot, the deploy switch and the manual desktop build), at the user's request. Step 5's release work updates it as the desktop jobs grow.
- 2026-11-02, step 5: the Windows job's first run failed: `go tool wails3`, built with cgo on Linux, needs GTK. Both Windows tasks now set `CGO_ENABLED=0`, so the Windows build needs no desktop build deps, as CLAUDE.md already claimed. Desktop jobs pin `ubuntu-24.04` rather than `ubuntu-latest`.
- 2026-11-02, step 5: pre-release tags (`v0.1.0-rc.1`) still give the installers version `0.0.0`; moved to TODO.md as a follow-up, at the user's choice. The Release is tried with a plain throwaway tag instead, deleted afterwards.
