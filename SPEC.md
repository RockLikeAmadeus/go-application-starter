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
| **Item** | A named thing the user wants to track. Placeholder entity for the walking skeleton. |
| **Theme** | The colour scheme of the web app: **dark** or **light**. |

## 4. Capabilities

### 4.1 Manage items _(walking skeleton)_

Users can create, list, rename and delete items.

**Rules**
- An item's name is required and is trimmed of surrounding whitespace.
- Names are at most 200 characters.
- Items are listed oldest first.

**Specification:** [`features/items.feature`](features/items.feature) · UI journeys: [`web/e2e/features/items.feature`](web/e2e/features/items.feature)

<!-- /feature adds new capability sections below this line -->

### 4.2 Choose a theme _(web only)_

Users can switch the web app between a dark and a light theme, so it suits their eyes and surroundings.

**Rules**
- A user's chosen theme is remembered in their browser and wins over everything else.
- With no chosen theme, the app follows the system's colour-scheme preference.
- With neither a choice nor a system preference, the app uses the dark theme.

**Examples**
- First visit, system prefers light → light theme.
- First visit, no system preference → dark theme.
- System prefers dark, user switches to light and comes back later → light theme.

**Specification:** [`web/e2e/features/theme.feature`](web/e2e/features/theme.feature). This is a presentation concern with no core behaviour, so there are no Godog scenarios.

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
