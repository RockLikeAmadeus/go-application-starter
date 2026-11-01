# Plan: hosted and local editions (workspaces, plans and entitlements, local-only tools)

- TODO item: "Hosted and local editions" in [`TODO.md`](../../TODO.md).
- Temporary: delete this file in the same change that ticks that TODO item. Move any open follow-ups into `TODO.md` first.
- Each step except step 0 is its own `/feature` → `/red` → `/green` → `/refactor` cycle, with its own `docs/wip/<slug>.md`. Stop for review between steps.
- The lasting decisions are in [ADR 5](../adr/0005-hosted-and-local-editions.md). This file is the working plan.

## Goal
All clients can operate as commercial products, with data kept separate per account and features gated by paid plans. CLI, TUI and free local-only desktop and mobile apps stay single-user tools on local data (SQLite or plain text, in a folder the user chooses, which can be a cloud-synced one). Both run on the same core and data structures.

## Decisions (agreed 2026-10-31)
- **Scope: template seams.** The template ships the ports, a simple reference adapter for each, and a sample gated feature. Each product picks its real auth, billing and hosting.
- **Two editions, one core.** The **hosted edition** is multi-user, with Postgres on a server. The **local edition** has one implicit workspace and local storage. The core never knows which edition it's in; `internal/bootstrap` picks the adapters.
- **The owner is a workspace.** Every record has a `WorkspaceID`. Each user gets a personal workspace now, and teams later only add members and roles, with no data migration. The current workspace travels in `context`, set by the driving adapter: the auth middleware in the hosted edition, a fixed local workspace in the local edition. Use cases read it and pass it **explicitly** to repositories, so the `storagetest` contract suite can check isolation (workspace A never sees B's records).
- **Plans and entitlements.** Plans are defined in config, not code. Each plan lists on/off **features** and **limits**.
  - For each limit, config says what happens when a workspace is over it (for example after a downgrade): `block-new` keeps the data and refuses new records; `read-only` makes that data read-only.
  - For each feature, config says whether the UI shows it locked, with an upgrade hint, or hides it.
  - Use cases ask an `Entitlements` port. The API exposes the current workspace's entitlements, so the web app, CLI and TUI can show or hide features.
- **Local feature set: a named plan** from the same config (for example `local`), chosen at build time. A paid local licence later just switches the plan. Local builds show no upgrade hints.
- **Billing: a port plus a manual adapter.** An admin-only `ChangePlan(workspace, plan)` use case changes a plan. Payment providers are driving adapters that call it, as the REST handlers call the core. The reference adapter is an admin command, `starter admin set-plan`. A detailed guide for template users goes in `docs/billing.md`, covering:
  - the concepts and the plans config;
  - running paid accounts by hand;
  - a Stripe walkthrough: checkout, webhooks, which events to handle, duplicate deliveries, and the Stripe CLI;
  - merchant-of-record options (Paddle, Lemon Squeezy);
  - downgrades, failed payments and grace periods;
  - a future paid local licence (an offline signed key).
- **Authentication:** an OIDC adapter, which works with any provider (Auth0, Clerk, Keycloak, Google, …), plus a dev-only fake login for tests and `task dev`. We store no passwords. Hosted edition only.
- **CLI and TUI:** local by default. `starter login --server …` lets a hosted user work on their account through `httpclient`.
- **No sync between local and cloud data** (ADR 4).
- **Synced folders: Level 1 only.** A local data folder may live in a Dropbox, Syncthing, iCloud Drive or Google Drive folder, used from **one device at a time**. The app:
  - uses a plain-text layout of one file per record;
  - detects the sync service's conflict copies and tells the user;
  - shows a best-effort "in use on another device" warning (a lock file holding the device and a heartbeat).

  There's no merging and no live use from several devices. Live use from several devices is what the hosted edition is for. (Level 2, merging inside a synced folder, and Level 3, the app's own sync engine, were rejected as too costly to build and maintain.)
- **Local formats: plain text or SQLite, the user's choice.** SQLite can be corrupted by file-sync services, so the app warns when its folder looks synced.
- **Android:** data stays in the app's private storage. Synced folders on Android need its document picker (Storage Access Framework: content URIs, not file paths), so that's a follow-up.

## Steps

### 0. Record the plan — done (docs only)
This file, [ADR 5](../adr/0005-hosted-and-local-editions.md), and a merge of the two duplicate TODO entries into one that links here. The SQL-adapter, authentication and desktop/mobile remote-mode TODO items now point at their steps below. SPEC.md §3 now defines edition, workspace, plan, feature and limit, and §6 links this plan. §2 (personas) waits until the editions exist.

### 1. Workspace scoping — done
SPEC.md §4.8 · [`features/workspaces.feature`](../../features/workspaces.feature) · [ADR 5](../adr/0005-hosted-and-local-editions.md) "Data belongs to a workspace".
- **Domain:** every item has a `WorkspaceID`. `domain.LocalWorkspace` (`"local"`) is the local edition's one workspace, and `domain.ErrNoWorkspace` is returned when a request has none.
- **Resolver port:** use cases ask `ports.WorkspaceResolver` for the current workspace and pass it explicitly to every repository call. `internal/adapters/workspace` has two resolvers:
  - `Local`, which `bootstrap` wires for every current front end;
  - `FromContext`, which reads what `WithID` set and refuses a missing or empty workspace.
- **Storage:** memory and jsonfile return only the given workspace's items, and treat another workspace's item as missing. Contract-suite cases check that for `FindAll`, `FindByID` and `Delete`. jsonfile records carry `workspaceId`. There's no migration (the template has no users).
- **Fail closed:** with no workspace, every use case returns `ErrNoWorkspace` before touching storage (a table test).
- **Acceptance:** the workspace scenarios run on the memory and jsonfile drivers, which use `FromContext`. A `Before` hook puts every other scenario in the local workspace. The http, desktop and mobile drivers stay on `Local` until sign-in (step 8).
- Users see no change. Left for step 8: mapping `ErrNoWorkspace` to HTTP, and running the isolation scenarios on the http driver.

### 2. Plans and entitlements — not started
- The plans config (format and location to decide in the `/feature`), the `Entitlements` port and its config adapter, and the local plan chosen at build time.
- `GET /api/v1/entitlements` (`api/openapi.yaml`, then `task gen`). `httpclient` maps it back.
- Sample: a limit on items on the free plan, with the over-limit behaviour set per limit. Inside the sample blocks, so `--remove-sample` strips it. Run `task test:init`.

### 3. Sample paid feature: CSV export of items — not started
- On/off gating end to end: the use case checks the entitlement. In the web app, CLI and TUI it's shown locked or hidden, as config says per feature.
- Inside the sample blocks. Run `task test:init`.

### 4. Changing a workspace's plan — not started
- The admin-only `ChangePlan` use case and `starter admin set-plan`.
- `docs/billing.md`, the template users' guide described above.

### 5. Sync-safe plain text — not started
- jsonfile v2: one file per record. No migration from the v1 `items.json` (the template has no users).
- Detect conflict copies. Warn about a heartbeat lock file from another device.
- Update ADR 3.

### 6. SQLite adapter and migrations — not started
- This is the existing "SQL storage adapter" TODO item. It passes every contract suite, including isolation.
- The user chooses plain text or SQLite. SQLite warns when its folder looks synced.

### 7. Postgres adapter — not started
- For the hosted edition. It shares the schema with SQLite where possible, and is tested against a Postgres container in CI.

### 8. Authentication — not started
- This is the existing "Authentication behind a port" TODO item.
- The OIDC adapter and the dev-only fake login.
- Auth middleware that sets the workspace, and a personal workspace created at first login.
- Sign-in is optional on the server. With it off, the server is a self-hosted local edition: it uses the `Local` resolver, as it does today.
- SPA login and logout. The isolation scenarios run on the http driver.
- Remove "Authentication and authorization" from SPEC.md §6.

### 9. CLI and TUI login — not started
- `starter login --server …`, with the token stored locally. `httpclient` sends it.

### 10. Desktop and mobile remote mode — not started
- This is the existing TODO item, now unblocked by step 8.

## Follow-ups for TODO.md (when this plan finishes)
- Synced folders on Android (Storage Access Framework).
- Teams: workspace members and roles.
- A paid local licence key, behind the entitlements port.

## Verification after every step
- `task check`, plus `task accept` with the step's `@wip` scenarios passing on every relevant driver.
- From step 1: every storage adapter passes the isolation contract.
- Step 5: a by-hand check with the data folder in a Syncthing or Dropbox folder on two machines, opened one at a time, plus a forced conflict copy that shows the warning.
- From step 8: `task e2e` for the login journey.

## Plan changes
- 2026-10-31: plan created from the design discussion.
- 2026-10-31: step 1 refines how the workspace reaches the core: use cases ask a driven port, `ports.WorkspaceResolver`, instead of reading the context directly. bootstrap plugs in a fixed local resolver for the local edition, and the hosted edition's resolver reads what the auth middleware put in the context. Missing workspaces are refused. The API doesn't expose the workspace yet.
- 2026-11-01: the server can also run as a self-hosted local edition, with sign-in off (step 8). Recorded in ADR 5. SPEC.md, CLAUDE.md and README.md now describe which clients run in which edition, marking the parts still to build as planned.
- 2026-11-01: no backwards compatibility while this is still the template, which has no users. Data saved before a change isn't migrated: step 1 drops its "old records read as local" task, and step 5 no longer migrates the v1 `items.json`. Products built from the template handle their own compatibility.
- 2026-11-01: after step 1, more of the workspace work now survives `--remove-sample`, so template users don't redo it for each entity:
  - a generic port, `ports.WorkspaceRepository[T]`, which `ItemRepository` now embeds;
  - a generic isolation suite, `storagetest.RunWorkspaceIsolation`, which the Item contract now calls instead of its own isolation cases;
  - a guide with a checklist, `docs/workspaces.md`, linked from CLAUDE.md's workspace rule, the port and `/feature`'s planning step;
  - a comment in `bootstrap.NewEmbedded`, outside the sample block, naming `workspace.Local{}` beside the clock and IDs.
