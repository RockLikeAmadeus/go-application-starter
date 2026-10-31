# Product Specification

> **Template note:** This is a stub. Replace the placeholders as the product takes shape. `/feature` adds a capability section here before writing Gherkin. Keep this document about **what** and **why**. **How** belongs in `docs/adr/`.

## 1. Purpose

_One paragraph: who the users are, what problem this solves, and what success looks like._

## 2. Users and front-ends

| Persona | Front-end | Notes |
|---|---|---|
| _e.g. End user_ | Web (React SPA) | |
| _e.g. Power user / operator_ | CLI (`starter`) and TUI (`starter tui`) | Embedded or `--server` mode |
| _e.g. Integrations_ | REST API (`api/openapi.yaml`) | |
| _e.g. Single user on one computer_ | Desktop app (`starter-desktop`) | The web app in its own window, with its data on the computer (§4.6) |
| _e.g. Single user on an Android phone_ | Mobile app (Android) | The web app as a phone app, with its data on the phone (§4.7) |

## 3. Glossary (ubiquitous language)

Use these terms consistently in code, Gherkin, and UI copy.

| Term | Meaning |
|---|---|
| **Theme** | The colour scheme of the web app: **dark** or **light**. Independent of the style. |
| **Style** | A named look for the web app: its layout, fonts, shape and density. Independent of the theme; every style works in dark and light. |
| **Desktop app** | The web app packaged as a program for the user's computer. It keeps its own data and needs no server. |
| **Data folder** | The folder where the desktop app or the mobile app keeps its data: on the user's computer, or in the app's private storage on the phone. |
| **Mobile app** | The web app packaged as an app for the user's phone (Android for now). It keeps its own data on the phone and needs no server. |

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

**Specification:** [`features/items.feature`](features/items.feature) · UI journeys: [`web/e2e/features/items.feature`](web/e2e/features/items.feature)
<!-- sample:end -->

<!-- /feature adds new capability sections below this line -->

### 4.2 Choose a theme _(web only)_

Users can switch the web app between a dark and a light theme, so it suits their eyes and surroundings.

**Rules**
- A user's chosen theme is remembered in their browser and wins over everything else.
- With no chosen theme, the app follows the system's colour-scheme preference. Browsers report "no preference" as light, so first-time visitors whose system isn't set to dark see the light theme.

**Examples**
- First visit, system prefers dark → dark theme.
- First visit, system prefers light (or has no setting) → light theme.
- System prefers dark, user switches to light and comes back later → light theme.

**Specification:** [`web/e2e/features/theme.feature`](web/e2e/features/theme.feature). This is a presentation concern with no core behaviour, so there are no Godog scenarios.

### 4.3 Choose a style _(web only)_

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

Users can run the app as a program on their own computer, so they can use it without a server, an account or a connection.

**Rules**
- The desktop app shows the same app as the web app, with the same styles and themes, in its own window.
- It keeps the user's data in a data folder on their computer. It needs no server and no connection, and it doesn't open a network port.
- By default the data folder is the user's settings folder for the app (on Linux, `~/.config/starter`). The user can choose another one when they start the app.
- Data is still there after the app is closed and opened again.
- Only one copy of the app can use a data folder at a time. A second copy says the data is already in use, and doesn't start.
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

Users can install the app on their Android phone and use it without a server, an account or a connection.

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

## 5. Non-functional requirements

| Area | Requirement |
|---|---|
| Availability | _e.g. best effort; single instance while on plain-text storage_ |
| Data | _e.g. data survives restarts; backups via volume snapshots_ |
| Security | _e.g. no auth yet; see "Out of scope"_ |
| Accessibility | Web UI is keyboard navigable, and controls have accessible names (also used by tests) |
| TUI navigation | Every list or selectable view supports Vim keys alongside the arrow keys: `j` down, `k` up |

## 6. Out of scope (for now)

- Authentication and authorization
- Multi-instance deployments (blocked by the jsonfile adapter; see ADR 0003)
