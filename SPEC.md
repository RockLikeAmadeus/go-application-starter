# Product Specification

> **Template note:** This is a stub. Replace the placeholders as the product takes shape. `/feature` adds a capability section here before writing Gherkin. Keep this document about **what** and **why**. **How** belongs in `docs/adr/`.

## 1. Purpose

_One paragraph: who the users are, what problem this solves, and what success looks like._

## 2. Users and front-ends

| Persona | Front-end | Editions (§3) | Notes |
|---|---|---|---|
| _e.g. End user_ | Web (React SPA) | Hosted, with sign-in (planned). Or local: a self-hosted server for one user, with no sign-in (works today) | Installable (§4.5) |
| _e.g. Power user / operator_ | CLI (`starter`) and TUI (`starter tui`) | Local by default: the core runs inside the CLI, with data in the desktop app's data folder (§4.11). Hosted, after `starter login --server …` (planned) | Today `--server` reaches only a server without sign-in |
| _e.g. Integrations_ | REST API (`api/openapi.yaml`) | Whichever edition the server runs | |
| _e.g. Single user on one computer_ | Desktop app (`starter-desktop`) | Local, with its data on the computer (§4.6). Hosted, in remote mode (planned) | The web app in its own window |
| _e.g. Single user on an Android phone_ | Mobile app (Android) | Local, with its data on the phone (§4.7). Hosted, in remote mode (planned) | The web app as a phone app |

Each copy of a front-end uses one edition, never both. Nothing is synced or moved between local and hosted data. The planned parts are being built step by step: see [`docs/plans/editions-and-accounts.md`](docs/plans/editions-and-accounts.md) and [ADR 5](docs/adr/0005-hosted-and-local-editions.md).

## 3. Glossary (ubiquitous language)

Use these terms consistently in code, Gherkin, and UI copy.

| Term | Meaning |
|---|---|
| **Theme** | The colour scheme of the web app: **dark** or **light**. Independent of the style. |
| **Style** | A named look for the web app: its layout, fonts, shape and density. Independent of the theme; every style works in dark and light. |
| **Desktop app** | The web app packaged as a program for the user's computer. In the local edition it keeps its own data and needs no server. Working on hosted data instead (remote mode) is planned. |
| **Command line** | The `starter` program's commands (the CLI), run one at a time in a terminal. |
| **Terminal app** | The full-screen app the command line opens with `starter tui` (the TUI), worked with the keyboard. |
| **Settings file** | An optional file of the app's settings (such as the data folder) on the user's computer, shared by the desktop app and the command line. |
| **Data folder** | The folder where the local edition keeps its data: on the user's computer, shared by the desktop app and the command line, or in the app's private storage on the phone. |
| **Mobile app** | The web app packaged as an app for the user's phone (Android for now). In the local edition it keeps its own data on the phone and needs no server. Working on hosted data instead (remote mode) is planned. |
| **Edition** | How a copy of the product runs. The **hosted edition** serves many users from a server, each signing in to an account (planned). The **local edition** is for one user, with no account: data on their own device, or on a server they host for themselves. Both run the same app, and each copy uses one edition only. |
| **Workspace** | The space a piece of data belongs to. In the hosted edition, each user has their own workspace. The local edition has exactly one. |
| **Plan** | What a workspace is allowed to do: which features it has and what its limits are (for example, free or paid). |
| **Feature** | A capability that a plan turns on or off. |
| **Limit** | A number a plan sets, such as how many items a workspace may have. |
| **Admin** | Someone who runs the product and can move a workspace to another plan, for example when its owner starts or stops paying. |
| **Ready** | A server is ready when it can reach its data, so it can serve requests. |
| **Conflict copy** | A second copy of a file that a file-sync service leaves in a synced folder when the same file changed in two places. |
| **Device** | A computer that uses a data folder, known by its name. |

<!-- sample:start -->
The sample capability (§4.1) adds one term:

| Term | Meaning |
|---|---|
| **Item** | A named thing the user wants to track. Placeholder entity for the walking skeleton. |
<!-- sample:end -->

## 4. Capabilities

<!-- sample:start -->
### 4.1 Manage items _(walking skeleton)_

Users can create, list, rename and delete items.

**Rules**
- An item's name is required and is trimmed of surrounding whitespace.
- Names are at most 200 characters.
- Items are listed oldest first.

**Specification:** [`features/items.feature`](features/items.feature) · UI journeys: [`web/e2e/features/items.feature`](web/e2e/features/items.feature) · terminal journeys: [`test/terminal/features/items.feature`](test/terminal/features/items.feature)
<!-- sample:end -->

<!-- /feature adds new capability sections below this line -->

### 4.2 Choose a theme _(web, desktop, and mobile only)_

Users can switch the web app between a dark and a light theme, so it suits their eyes and surroundings.

**Rules**
- A user's chosen theme is remembered in their browser and wins over everything else.
- With no chosen theme, the app follows the system's colour-scheme preference. Browsers report "no preference" as light, so first-time visitors whose system isn't set to dark see the light theme.

**Examples**
- First visit, system prefers dark → dark theme.
- First visit, system prefers light (or has no setting) → light theme.
- System prefers dark, user switches to light and comes back later → light theme.

**Specification:** [`web/e2e/features/theme.feature`](web/e2e/features/theme.feature). This is a presentation concern with no core behaviour, so there are no Godog scenarios.

### 4.3 Choose a style _(web, desktop, and mobile only)_

Users can switch the web app between styles, so it suits how they like to work: compact and dense, or roomy and soft.

**Rules**
- The app offers a fixed list of styles. Out of the box it offers **Crisp**, **Terminal** and **Soft**.
- A first visit uses the first style in the list.
- A user's chosen style is remembered in their browser and wins on later visits. A remembered style the app no longer offers falls back to the first style.
- Style and theme are independent: changing one never changes the other.
- Every style uses the product's single colour palette.

**Examples**
- First visit → Crisp.
- User switches to Terminal and comes back later → Terminal.
- In the dark theme, user switches to Soft → Soft style, still dark.

**Specification:** [`web/e2e/features/style.feature`](web/e2e/features/style.feature). Like the theme, this is presentation only, so there are no Godog scenarios.

### 4.4 Use on a phone _(web only)_

Users can use the web app comfortably on a phone, so they aren't tied to a desktop browser. This also lays the groundwork for the installable and mobile apps.

**Rules**
- On a phone-sized screen, the app's navigation and Appearance controls are tucked away behind a menu, so the page content gets the full width.
- The menu starts closed. Opening it shows the navigation and the Appearance controls.
- The menu closes when the user goes somewhere from it, presses Escape, or taps outside it.
- At phone width, pages never need sideways scrolling, in every style and both themes.
- On a wide screen, the navigation is always visible and there is no menu.

**Examples**
- Phone, first visit → the content fills the screen and a menu button is shown; the navigation is hidden.
- Phone, user opens the menu and switches to Soft → Soft style; the menu stays open until dismissed.
- Phone, Terminal style, dark theme → no sideways scrolling.
- Laptop → the sidebar is shown and there is no menu button.

**Specification:** [`web/e2e/features/phone.feature`](web/e2e/features/phone.feature). Presentation only, so there are no Godog scenarios. Scenarios tagged `@phone` run on an emulated Android phone (Chromium), an emulated iPhone (WebKit) and a phone-sized Firefox window. The rest run on desktop Chromium.

### 4.5 Install the web app _(web only)_

Users can install the web app on their phone or computer, so it opens from the home screen or app list in its own window, like a native app. It needs no app store.

**Rules**
- The app can be installed under the product's name and icon. It opens in its own window, at the home page.
- After one visit, the app opens even without a connection.
- Without a connection, the app says it is offline. When the connection comes back, the notice goes away.
- Data always comes from the server. The app never shows a stored copy that may be out of date.
- A new version installs quietly in the background and takes over the next time the app is opened.

**Examples**
- User installs the app on their phone → it opens from the home screen, full screen, at the home page.
- User opens the installed app on a plane → the app opens and says it is offline. It shows no data.
- User lands, and the connection comes back → the notice goes away and data loads.

**Specification:** [`web/e2e/features/install.feature`](web/e2e/features/install.feature). Presentation and caching only, so there are no Godog scenarios. These scenarios run on desktop Chromium, where Playwright supports service workers.

### 4.6 Use the desktop app

Users can run the app as a program on their own computer, so they can use it without a server, an account or a connection. This section describes the desktop app in the local edition. Working on hosted data instead (remote mode) is planned: see [`docs/plans/editions-and-accounts.md`](docs/plans/editions-and-accounts.md).

**Rules**
- The desktop app shows the same app as the web app, with the same styles and themes, in its own window.
- It keeps the user's data in a data folder on their computer. It needs no server and no connection, and it doesn't open a network port.
- By default the data folder is the user's settings folder for the app (on Linux, `~/.config/starter`). The user can choose another one in the settings file (§4.12), or when they start the app.
- Data is still there after the app is closed and opened again.
- With plain-text storage (the default), only one copy of the app can use a data folder at a time. A second copy says the data is already in use, and doesn't start. SQLite storage (§4.14) has no such limit.
- In local-only mode, it never says it is offline, because it doesn't need a connection.
- It shows the product's icon, the same one as the installed web app (§4.5), in the taskbar or Dock. On Linux the icon comes from the installed app, so a copy that wasn't installed from a package may show a generic icon.
- On Linux, users install it from a package for their system (`.deb` for Debian, Ubuntu and Pop!_OS; `.rpm` for Fedora and openSUSE), or run it as an AppImage with nothing to install. Once installed, it appears in the app list under the product's name and icon, and uninstalling it leaves the user's data folder alone.
- On Windows, users install it with a setup program (`starter-desktop-<version>-setup.exe`). It installs for the current user only, so it needs no administrator, and puts the app in the Start menu under the product's name and icon. Windows lists it under installed apps, where it can be uninstalled. Uninstalling leaves the user's data folder alone. The setup program isn't signed yet, so Windows SmartScreen warns about it before it runs.

**Examples**
- User adds "Widget", closes the app and opens it again → "Widget" is still there.
- User opens the app twice → the second copy says the data is in use.
- User opens the app on Windows or a Mac → the taskbar or Dock shows the product's icon.
- User installs the `.deb` on Pop!_OS → "Starter" appears in the app list with its icon, and opens from there.
- User runs the setup program on Windows → "Starter" appears in the Start menu with its icon, and opens from there.
- User uninstalls it on Windows and installs it again → their items are still there.
- User opens the app with the network off → the app works and shows no offline notice.

**Specification:** [`features/desktop.feature`](features/desktop.feature). The desktop app also runs every other feature's scenarios. Opening no network port, showing no offline notice, and the icon are checked by unit tests and by hand. The Linux packages' and the Windows setup program's contents are checked by `test/packaging` (`task test:package`), and installing them by hand.

### 4.7 Use the mobile app

Users can install the app on their Android phone and use it without a server, an account or a connection. This section describes the mobile app in the local edition. Working on hosted data instead (remote mode) is planned: see [`docs/plans/editions-and-accounts.md`](docs/plans/editions-and-accounts.md).

**Rules**
- The mobile app shows the same app as the web app on a phone (§4.4), with the same styles and themes.
- It keeps the user's data in its data folder, in the app's private storage on the phone. It needs no server and no connection.
- Data is still there after the app is closed and opened again.
- In local-only mode, it never says it is offline, because it doesn't need a connection.
- Removing the app from the phone removes its data.
- Android only for now; iPhone comes later.

**Examples**
- User adds "Widget", closes the app and opens it again → "Widget" is still there.
- User turns on airplane mode and adds "Widget" → it is saved, and the app shows no offline notice.

**Specification:** [`features/mobile.feature`](features/mobile.feature). The mobile app also runs every other feature's scenarios, except the desktop app's. Showing no offline notice is checked by unit tests and by hand.

### 4.8 Keep each workspace's data separate

All data belongs to a workspace, so that when one server serves many users, nobody sees or changes anyone else's data. This is the groundwork for the hosted edition (ADR 5).

**Rules**
- Every piece of data belongs to exactly one workspace: the one it was created in.
- In a workspace, only that workspace's data can be seen or changed. Data in another workspace behaves as if it doesn't exist.
- Without a workspace, no data can be seen or changed.
- The local edition has exactly one workspace and uses it automatically. Its users never see or choose a workspace.

<!-- sample:start -->
**Examples**
- Workspace Alpha has "Widget", and Beta has "Gadget" → Alpha lists only "Widget", and Beta only "Gadget".
- Someone in Beta tries to rename or delete Alpha's "Widget" → they're told it wasn't found, and Alpha still has "Widget".

**Specification:** [`features/workspaces.feature`](features/workspaces.feature). These scenarios run on the memory, jsonfile, sqlite and postgres drivers only: until sign-in exists (`docs/plans/editions-and-accounts.md`, step 8), nothing over HTTP can pick a workspace, and the desktop and mobile apps are the local edition.
<!-- sample:end -->

### 4.9 Plans, features and limits

Each workspace is on a **plan**, which turns **features** on or off and sets its **limits**: for example, whether it can export its items, and how many items it may have. Paid plans can then offer more than free ones (ADR 5). Plans are defined in configuration, not code, so they can change without a new release.

**Rules: features**
- A plan turns each feature on or off. A feature the plan doesn't turn on is off.
- A feature that's off can't be used: trying is refused, saying the feature isn't on the workspace's plan. Front ends never decide this themselves; they only show what the plan says.
- Each feature has a name that people read (for example "Item export"), set with the plan, so its wording can change without a release.
- Each feature says how the front ends show it while it's off:
  - **locked:** it's shown, but marked as disabled on the workspace's plan, with a hint naming the plans that have it (for example, "Available on the pro plan");
  - **hidden:** it isn't shown at all, and a refusal names no plan.
- The local edition gives no upgrade hints: a locked feature there just says it isn't available.
- Moving to a plan that has a feature turns it on straight away, and moving to one that doesn't turns it off.

**Rules: limits**
- A workspace's plan sets its limits. A plan without a given limit has no limit on that thing.
- A workspace can always see its plan, which features it has, how each one that's off is shown, and its limits.
- At a limit, nothing more can be created. Whatever was already there stays.
- A workspace can end up **over** a limit, for example after moving to a smaller plan. Each limit says what happens then:
  - **block new:** existing data can still be changed and deleted; only new data is refused;
  - **read only:** existing data can still be seen and deleted, but not changed, until the workspace is back within the limit.
- The local edition is on a plan chosen when the app is built (by default, `local`, with no limits).

<!-- sample:start -->
**Examples**
- The free plan allows 3 items. A workspace on it with 3 items tries to create a fourth → it's told "You have reached the limit of 3 items on the free plan.", and still has 3 items.
- The free plan allows 3 items and makes them read only when over. A workspace with 4 items moves to it → it can't rename an item, but can delete one; with 3 items left, it can rename again.
- The pro plan has no item limit → a workspace on it can create a fourth item.
- **Item export** is the sample feature: it gives the workspace's items as a CSV file (spreadsheet text), one row per item, oldest first. The local and pro plans have it; the free plan shows it locked.
  - A workspace on the pro plan exports its items → it gets "First" and "Second", in that order.
  - A workspace on the free plan tries to export → it's told "Item export is not included in your current plan. To use this feature, please upgrade to the pro plan." The web app shows the export button locked, with the note "Available on the pro plan". `starter --help` lists the command with "(not included in your current plan; available on the pro plan)", and `starter item export` refuses with the same message.
  - With item export hidden when off, a workspace on the free plan sees no export button, and `starter --help` doesn't list the command.

**Specification:** [`features/plans.feature`](features/plans.feature) (limits) and [`features/plan-features.feature`](features/plan-features.feature) (features). These scenarios choose each workspace's plan, so they run only on the memory, jsonfile, sqlite, postgres and http drivers. The desktop and mobile apps are on the local plan ([`features/desktop.feature`](features/desktop.feature), [`features/mobile.feature`](features/mobile.feature)).
<!-- sample:end -->

### 4.10 Tell when the server is ready

The server says whether it is **ready**: whether it can reach its data and so serve requests. That way a deployment only sends users to a server that can serve them, and a broken one is noticed. This is for whoever runs the server, not its users. It doesn't depend on which kinds of data the product has.

**Rules**
- A server that can reach its data is ready.
- A server that can't reach its data isn't ready, and it says that the data is unavailable.
- Readiness is checked every time it's asked about, so a server whose data comes back is ready again without a restart.
- A server that keeps its data in memory is always ready.

**Examples**
- The server has just started on its data folder → ready.
- The data folder is removed while the server runs → not ready: the data is unavailable.
- The data folder is put back → ready again.

**Specification:** [`features/readiness.feature`](features/readiness.feature). These scenarios run only on the server driver, which wires the server as `cmd/server` does, on a data folder the tests can remove.

### 4.11 Share local data between the desktop app and the command line

On one computer, the desktop app, the CLI and the TUI are one local edition, so they work on the same data. A user who adds something in one finds it in the others, whichever folder they ran the command in. The command line can also say where that data is.

**Rules**
- By default the CLI and the TUI use the desktop app's data folder (§4.6): the user's settings folder for the app (on Linux, `~/.config/starter`). The folder they're run from makes no difference.
- The user can choose another data folder, for one command or for all of them, as for the desktop app.
- With plain-text storage (the default), only one program can use a data folder at a time. While the desktop app is open, the command line says the data is already in use, and changes nothing. With SQLite storage (§4.14), both can use it at once.
- `starter info` says which data folder the command line uses, and the workspace's plan.
- Connected to a server (`--server`), the command line uses that server's data, and keeps none on the computer.
- The server isn't part of this: it keeps its data where it's configured to (by default `./data`, where it was started).

**Examples**
<!-- sample:start -->
- User adds "Widget" in the desktop app, closes it, and lists items with `starter item list` → "Widget" is there.
<!-- sample:end -->
- User runs `starter info` from any folder on Linux → it names `~/.config/starter`.
- User runs `starter info` while the desktop app is open → it says the data is already in use.
- User runs `starter --data-dir ~/work-data info` → it names `~/work-data`.

**Specification:** [`features/local-data.feature`](features/local-data.feature). These scenarios run only on the desktop driver, which opens the desktop app on a temporary settings folder and runs the command line beside it. Where each front end keeps its data, on each OS, is in [docs/data-locations.md](docs/data-locations.md).

### 4.12 Keep settings in a settings file

Users of the local edition can keep their settings in a settings file, so they don't have to set them each time they start the app. The desktop app, the CLI and the TUI read the same file. Whoever runs a server can give it one too.

**Rules**
- The settings file is optional. Without one, the app uses its built-in defaults.
- By default it's `config.toml` in the user's settings folder for the app (on Linux, `~/.config/starter/config.toml`). `STARTER_CONFIG_FILE`, or `--config` for one command, names another one.
- It can hold any of the app's settings: the data folder, the storage, the server to connect to, the log format, the plans file and the port.
- A setting chosen when the app is started overrides the settings file: an environment variable over the file, and a command-line option over both.
- `starter info` says which settings file was read, and where each setting came from: built in, the settings file, an environment variable or a command-line option.
- A settings file the app can't understand (a mistake in it, or a setting it doesn't know) stops the app at start, saying which file and which setting. The app never starts on settings it ignored.
- A named settings file that doesn't exist is a mistake too. The default one may be missing.
- A server reads a settings file only when `STARTER_CONFIG_FILE` names one, never the user's settings folder.
- When the desktop app, the CLI or the TUI starts and the default settings file is missing, it creates one, so the user has a file to edit. The file sets the settings these programs use (the storage, the data folder and whether to ignore a synced folder) to the values they would have without it. The data folder is written as this computer's full path. The settings that have no default (the server to connect to and the plans file) are in it as comments. Creating it changes nothing about how the app runs.
- The app never changes a settings file that's already there. It creates none when `STARTER_CONFIG_FILE` or `--config` names one, and the server and the mobile app never create one.
- If the settings file can't be created (for example, the folder can't be written to), the app starts on its built-in defaults as before, and tries again next time.

**Examples**
- The settings file says the data folder is `~/work-data` → the desktop app and `starter info` both use `~/work-data`.
- The settings file says `~/work-data`, and the app is started with `STARTER_DATA_DIR=/tmp/scratch` → it uses `/tmp/scratch`, and `starter info` says that came from `STARTER_DATA_DIR`.
- The settings file has `colour = "blue"` → the app doesn't start, and says `colour` isn't a setting it knows.
- A user on Linux with no settings file opens the desktop app → `~/.config/starter/config.toml` now exists, with `storage = "jsonfile"`, `data_dir = "/home/<user>/.config/starter"` and `ignore_synced_folder = false`, and the app runs as it did before.

**Specification:** [`features/settings-file.feature`](features/settings-file.feature) and [`features/default-settings-file.feature`](features/default-settings-file.feature). These scenarios run on the desktop driver, on a temporary settings folder, apart from one on the mobile driver showing that the mobile app creates no settings file. Every setting, its environment variable and its option are listed in [docs/configuration.md](docs/configuration.md).

### 4.13 Use the command line and the terminal app

Users who live in a terminal can do their work with the command line, one command at a time, or in the terminal app, full screen and by keyboard. Both are the installed `starter` program, and work on the same data as the other front ends (§4.11).

**Rules**
- Each command prints its result, or says what went wrong. A command that fails ends with an error status, so scripts can tell.
- A command the program doesn't know fails, saying so.
- The terminal app fills the terminal, says how to quit it, and leaves the terminal as it found it when the user quits.
- In every list, the terminal app moves with the arrow keys and with Vim's `j` and `k` (§5).
- Connected to a server (`--server`), both work on the server's data instead of the computer's.

**Examples**
- User runs `starter info` → it names the data folder.
- User runs `starter --server http://localhost:8080 info` → it names the server.
- User runs `starter frobnicate` → it says the command is unknown, and ends with an error status.
- User runs `starter tui`, then presses `q` → the terminal shows what it showed before.
<!-- sample:start -->
- User runs `starter item add "Widget"`, then `starter item list` → "Widget" is listed.
- User opens the terminal app, presses `a`, types "Widget" and presses Enter → "Widget" is listed.
<!-- sample:end -->

**Specification:** the business rules are in [`features/`](features/), run against the driving ports. These journeys only prove the wiring: they run the built programs in a terminal, as a user would. [`test/terminal/features/command-line.feature`](test/terminal/features/command-line.feature), [`test/terminal/features/terminal-app.feature`](test/terminal/features/terminal-app.feature).

### 4.14 Keep data in SQLite

Data can be kept in a SQLite database instead of as plain text, so that several programs can use it at the same time: the desktop app and the command line on one computer, or several copies of a server on one disk. Plain text stays the default. This is step 6 of [`docs/plans/editions-and-accounts.md`](docs/plans/editions-and-accounts.md); the hosted edition's database, Postgres, is §4.17.

**Rules**
- The storage setting (§4.12) chooses how data is kept: `jsonfile` (plain text, the default), `sqlite` or `memory`. The server, the command line, the terminal app and the desktop app all honour it. The mobile app keeps plain text for now.
- SQLite keeps all the data in one database file in the data folder.
- Several programs can use the same SQLite data at once, and each sees what the others saved. (Plain-text data is still for one program at a time, §4.6 and §4.11.)
- Opening the data brings its layout up to date with the app. Data saved by a newer version of the app, with a layout this one doesn't know, is refused rather than changed.
- A server on SQLite is ready while it can reach its database, and not ready once the database is gone (§4.10).
- Each workspace's data stays separate, as with every kind of storage (§4.8).
- Don't keep SQLite data in a folder that a file-sync service (Dropbox, iCloud Drive, Google Drive, OneDrive, Syncthing, Nextcloud, …) syncs: it can damage the database. When the data folder looks synced, the app warns, naming the service, and keeps working:
  - the desktop app (and the server's web app) shows the warning in its window;
  - the command line prints it with every command, apart from the command's own output;
  - the terminal app shows it on its screen;
  - the server logs it when it starts.
- Telling whether a folder is synced is a best guess, from the folder's path and the files a sync service leaves in it. A user whose folder looks synced but isn't (it's excluded from sync, say) can turn the warning off with the `ignore_synced_folder` setting.
- Plain-text data doesn't get this warning: it's safe in a synced folder while one device at a time uses it (§4.16, ADR 5).

**Examples**
- The settings file says `storage = "sqlite"` → the desktop app keeps its data in SQLite, and `starter info` works while the desktop app is open.
<!-- sample:start -->
- User adds "Widget" in the desktop app and, with it still open, runs `starter item list` → "Widget" is listed.
<!-- sample:end -->
- A server on SQLite whose data folder is removed → not ready: the data is unavailable.
- A user opens data saved by a newer version of the app → it isn't opened, and they're told it's from a newer version.
- The desktop app keeps SQLite data in `~/Dropbox/starter` → its window warns: "Your data folder appears to be synced by Dropbox. A file sync service can damage SQLite data: please move your data folder out of the synced location before continuing. Alternatively, you can switch to plain-text storage, but your existing data will not be transferred automatically." `starter item list` still lists the items, with the same warning beside them.
- The same user sets `ignore_synced_folder = true` → no warning.
- Plain-text data in `~/Dropbox/starter` → no warning.

**Specification:** [`features/sqlite-storage.feature`](features/sqlite-storage.feature) and [`features/synced-folder.feature`](features/synced-folder.feature). Every other feature's business scenarios also run on SQLite storage. Refusing a newer layout, and which paths and files mean a synced folder, are checked by unit tests.

### 4.15 Change a workspace's plan

An admin can move a workspace to another plan, for example when its owner starts or stops paying. This is the product's only billing step: anything that takes payments (a payment provider's notifications, or an admin by hand) does it by moving the workspace to the plan that was paid for (ADR 5). How to connect a payment provider is in [`docs/billing.md`](docs/billing.md). This is step 4 of [`docs/plans/editions-and-accounts.md`](docs/plans/editions-and-accounts.md).

**Rules**
- Only an admin can change a workspace's plan, with `starter admin set-plan WORKSPACE PLAN`, run beside the hosted edition's data. Users can't change their own plan, and nothing in the web app or the API changes it. (An admin API waits for sign-in.)
- The new plan must be one of the plans in the plans file. Any other plan is refused, and the workspace stays on its plan.
- The change takes effect straight away: the workspace's features and limits are the new plan's (§4.9). Its data is kept, even when it's now over a limit.
- Moving one workspace leaves every other workspace on its plan.
- A workspace's plan is kept with its data, so it stays the same after the app restarts.
- The local edition's plan can't be changed: it's set when the app is built (§4.9). `starter admin set-plan` refuses there, saying so. Until the hosted edition has sign-in, every copy is the local edition, so only the tests change plans for now.

**Examples**
- An admin moves workspace Alpha from the free plan to the pro plan → Alpha's plan is pro, and it can export its items.
- An admin tries to move Alpha to the "platinum" plan, which the plans file doesn't have → it's refused, and Alpha stays on the free plan.
- An admin moves Alpha to the pro plan → Beta stays on the free plan.
- A user runs `starter admin set-plan local pro` on their own computer → it fails with "The plan for the local edition is set when the application is built and cannot be changed."

**Specification:** [`features/plan-changes.feature`](features/plan-changes.feature). The scenarios that move a workspace run on the memory, jsonfile, sqlite, postgres and http drivers; the one that switches between workspaces runs on memory, jsonfile, sqlite and postgres only; the local edition's refusal runs on the desktop driver. Keeping the plan after a restart is checked by each storage adapter's contract tests. The command's output and exit status are checked by a terminal journey ([`test/terminal/features/command-line.feature`](test/terminal/features/command-line.feature)).

### 4.16 Keep plain-text data in a synced folder

A user can keep plain-text data in a folder that a file-sync service (Dropbox, iCloud Drive, Google Drive, OneDrive, Syncthing, Nextcloud, …) syncs, and use it from one device at a time. The app makes a clash small and visible, and leaves the sync service and the user to settle it. For live use from several devices, there's the hosted edition (ADR 5).

**Rules**
- Plain-text data keeps each record in its own file, so a clash can touch only that record.
- The app never merges. When a sync service leaves a conflict copy in the data folder, the app warns, naming the file, and keeps working. A conflict copy isn't read as data, so nothing appears twice.
- The user settles a conflict by hand: they compare the copy with the original, then either delete the copy or use it to replace the original. Deleting the original instead would lose the record, as a copy isn't read as data. The warning goes once no conflict copy is left.
- A conflict copy is itself the evidence, so the warning comes whether or not the folder looks synced, and `ignore_synced_folder` doesn't turn it off.
- While a copy of the app has plain-text data open in a folder that looks synced, it marks the folder as in use on its device. It renews the mark every minute and removes it on closing. With `ignore_synced_folder` set, or in a folder that doesn't look synced, it leaves no mark.
- When another device's mark is less than five minutes old, the app warns, naming the device, and keeps working. An older mark doesn't count, so a device that crashed or lost power stops counting by itself.
- This is a best guess. A sync service can be slow to bring the mark over, and two computers with the same name look like one device.
- Both warnings appear wherever the §4.14 warning does, and are checked when the data is opened: by each command on the command line, and when the desktop app or the server starts.
- The mobile app keeps its data in private storage, which no sync service reaches, so none of this applies there.

**Examples**
- A user's Dropbox leaves `items/6f1c (conflicted copy).json` in the data folder → the desktop app warns: "Your data folder contains 1 conflicted copy, items/6f1c (conflicted copy).json, which appears to have been created by a file sync service. A change may be missing from your data. Please compare it with the original file, then either delete it or use it to replace the original before continuing."
- The same user deletes the file and opens the app again → no warning.
- The desktop app is open on a user's laptop, and they open the same folder on their desktop computer → it warns: "Your data folder appears to be in use on laptop. Running the application from two devices at once while plain-text storage is enabled can result in lost changes: please close the application on your other device before continuing." It keeps working.
- The laptop lost power an hour ago, with the app open → no warning.

**Specification:** [`features/synced-plain-text.feature`](features/synced-plain-text.feature), on the desktop driver. The layout of the files, what counts as a conflict copy for each sync service, and renewing the mark are checked by unit tests.

### 4.17 Keep data in Postgres

A server can keep its data in a Postgres database, so that several copies of it, on different machines, can share the data. This is the hosted edition's storage, step 7 of [`docs/plans/editions-and-accounts.md`](docs/plans/editions-and-accounts.md). Sign-in comes in step 8. Until then, a server on Postgres is a self-hosted local edition with one workspace, as on any other storage.

**Rules**
- The storage setting (§4.12) can be `postgres`. A new setting, `database_url`, says which database to use. Postgres without a `database_url` is refused when the program starts.
- The server, the command line and the terminal app honour it, so an operator can run commands such as `starter admin set-plan` on a hosted server's data. The desktop and mobile apps refuse it, because they keep local data. They reach hosted data through a server instead (remote mode, planned).
- Several servers, on one machine or many, can use the same database at once, and each sees what the others saved.
- Postgres and SQLite keep the same data in the same layout, defined once for both. Opening the database brings its layout up to date with the app. A database saved by a newer version of the app is refused rather than changed, as with SQLite (§4.14).
- A server on Postgres is ready while it can reach its database, and is not ready while it can't (§4.10). It recovers by itself once the database is back.
- Each workspace's data stays separate, as with every kind of storage (§4.8).

**Examples**
- A server with `storage = "postgres"` and a `database_url` → it is ready, and keeps its items in that database.
<!-- sample:start -->
- Two servers share one database, and a user adds "Widget" through one of them → the other lists "Widget" too.
<!-- sample:end -->
- The database stops → the server says it isn't ready, because its data is unavailable.
- `storage = "postgres"` with no `database_url` → the server doesn't start, and says the setting is missing.
- The desktop app's settings file says `storage = "postgres"` → it doesn't open the data, and says that it keeps local data only.

**Specification:** [`features/postgres-storage.feature`](features/postgres-storage.feature), on the server driver. Every other feature's business scenarios also run on Postgres storage. Refusing a newer layout, a missing `database_url` and the desktop app refusing Postgres are checked by unit tests. These tests need a Postgres database (`task db:up`). Without one they are skipped locally, and CI always runs them.

## 5. Non-functional requirements

| Area | Requirement |
|---|---|
| Availability | _e.g. best effort; single instance while on plain-text storage, several on one disk with SQLite, several on many machines with Postgres_ |
| Data | _e.g. data survives restarts; backups via volume snapshots_ |
| Security | _e.g. no sign-in yet; see "Out of scope". Planned: sign-in for the hosted edition. A self-hosted server for one user (local edition) keeps running without sign-in_ |
| Accessibility | Web UI is keyboard navigable, and controls have accessible names (also used by tests) |
| TUI navigation | Every list or selectable view supports Vim keys alongside the arrow keys: `j` down, `k` up |

## 6. Out of scope (for now)

- Authentication and authorization (planned: hosted edition, [`docs/plans/editions-and-accounts.md`](docs/plans/editions-and-accounts.md))
