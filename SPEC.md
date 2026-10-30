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

## 3. Glossary (ubiquitous language)

Use these terms consistently in code, Gherkin, and UI copy.

| Term | Meaning |
|---|---|
| **Theme** | The colour scheme of the web app: **dark** or **light**. Independent of the style. |
| **Style** | A named look for the web app: its layout, fonts, shape and density. Independent of the theme; every style works in dark and light. |

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

**Specification:** [`web/e2e/features/phone.feature`](web/e2e/features/phone.feature). Presentation only, so there are no Godog scenarios. Scenarios tagged `@phone` run on an emulated Android phone (Chromium) and an emulated iPhone (WebKit). The rest run on desktop Chromium.

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
