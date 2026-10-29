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

## 4. Capabilities

### 4.1 Manage items _(walking skeleton)_

Users can create, list, rename and delete items.

**Rules**
- An item's name is required and is trimmed of surrounding whitespace.
- Names are at most 200 characters.
- Items are listed oldest first.

**Specification:** [`features/items.feature`](features/items.feature) · UI journeys: [`web/e2e/features/items.feature`](web/e2e/features/items.feature)

<!-- /feature adds new capability sections below this line -->

## 5. Non-functional requirements

| Area | Requirement |
|---|---|
| Availability | _e.g. best effort; single instance while on plain-text storage_ |
| Data | _e.g. data survives restarts; backups via volume snapshots_ |
| Security | _e.g. no auth yet; see "Out of scope"_ |
| Accessibility | Web UI is keyboard navigable, and controls have accessible names (also used by tests) |

## 6. Out of scope (for now)

- Authentication and authorization
- Multi-instance deployments (blocked by the jsonfile adapter; see ADR 0003)
