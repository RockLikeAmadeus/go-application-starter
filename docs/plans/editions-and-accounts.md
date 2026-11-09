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
  - Use cases ask an `EntitlementsProvider` port. The API exposes the current workspace's entitlements, so the web app, CLI and TUI can show or hide features.
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
  - shows a best-effort "in use on another device" warning (a file per device, holding the device and a heartbeat).

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

### 2. Plans and entitlements — done
SPEC.md §4.9 · [`features/plans.feature`](../../features/plans.feature) · [ADR 5](../adr/0005-hosted-and-local-editions.md) "Features and limits come from plans".
- **Domain:** `Entitlements` (a plan's name and its limits) with `CheckCreate` and `CheckChange`, `Limit` (built with `NewLimit`), and `LimitError`. Over a `block-new` limit nothing new can be added; over a `read-only` one nothing can be changed either. Deleting always works.
- **Ports and use case:** the driven `ports.EntitlementsProvider` and the driving `ports.EntitlementsService` (`CurrentEntitlements`, failing closed without a workspace).
- **Plans file:** `internal/adapters/plans` reads `plans.jsonc` (JSON with full-line comments, built in, with `local`, `free` and `pro`), or `STARTER_PLANS_FILE`'s file. The local plan is `bootstrap.LocalPlan`, set at build time. A bad file or an unknown local plan fails at start.
- **API:** `GET /api/v1/entitlements`, and a limit error as a 403 problem with `limit`, `max` and `isReadOnly`; `httpclient` maps both back. The server, desktop and Android apps serve it, and remote mode reads it.
- **Sample:** `ItemService` checks the items limit when creating and renaming, and refuses when the plan or the count can't be looked up. `free` has the sample items limit (3, block-new), inside a sample block. `task test:init` passes.
- **Acceptance:** the harness has its own plans; `@plans` scenarios run on memory, jsonfile and http. Desktop and mobile check they're on the local plan.
- Found along the way, and fixed: template code that depended on sample code (now a rule in CLAUDE.md); `/readyz` and generic UI helpers in the sample are TODO.md items.

### 3. Sample paid feature: CSV export of items — done
SPEC.md §4.9 · [`features/plan-features.feature`](../../features/plan-features.feature) · [`docs/features.md`](../features.md) · [ADR 5](../adr/0005-hosted-and-local-editions.md) "Features and limits come from plans".
- Done 2026-11-05: on/off features end to end, all template code. Each feature in the plans file's `features` section has a display `name` and a `whenOff` (`show-locked` or `hide`). `Entitlements.CheckFeature` fails closed and returns a `FeatureError`, whose upgrade hint names the plans that have the feature (there are no hints in the local edition). The API serves the features in `/entitlements`, with a 403 `feature` problem that `httpclient` maps back. The CLI help notes locked commands without opening storage, the TUI has `gatedKeyHelp`, and the web app has `useFeature` and `FeatureGate`, with no pop-in.
- Sample: `item-export` (on for `local` and `pro`, locked for `free`): `ExportItems`, `GET /api/v1/items/export`, `starter item export`, and the Items page's Export button, all in sample files or blocks. The scenarios run on every driver, plus the browser and terminal journeys. `task test:init` passes.

### 4. Changing a workspace's plan — done
SPEC.md §4.15 · [`features/plan-changes.feature`](../../features/plan-changes.feature) · [`docs/billing.md`](../billing.md) · [ADR 5](../adr/0005-hosted-and-local-editions.md) "Billing is a driving adapter".
- Done 2026-11-06: the admin-only `ChangePlan` use case (`ports.BillingService`), reached by `starter admin set-plan`, with no API. Each workspace's plan is stored through `ports.WorkspacePlanRepository` in memory, jsonfile and sqlite, and `plans.StoredPlanOf` reads it. `ports.PlanCatalog` refuses an unknown plan.
- The local edition's plan stays fixed, so `bootstrap.NewBilling` refuses every copy until step 8 wires the stored plans. Only the acceptance harness changes plans for now. All template code; the sample use is the item scenarios that move plans. `task test:init` passes.
- `docs/billing.md`, the template users' guide described above.

### 5. Sync-safe plain text — done
SPEC.md §4.16 · [`features/synced-plain-text.feature`](../../features/synced-plain-text.feature) · [ADR 3](../adr/0003-plain-text-storage.md) · [ADR 5](../adr/0005-hosted-and-local-editions.md) "Synced data folders".
- Done 2026-11-08: jsonfile v2 keeps each record in `<table>/<key>.json` (`{"version": 2, "key": …, "record": …}`), with no migration from the v1 `items.json`. A one-record lookup reads only that record's file.
- A file whose key doesn't match its name is a conflict copy, whatever the sync service called it: reading skips it, and `jsonfile.Store.ConflictCopies` lists them for a warning in `App.Warnings`, whether or not the folder looks synced.
- The heartbeat lock file became one mark per device, `.devices/<device>.json` (`internal/adapters/syncedfolder`), written only when the folder looks synced and `ignore_synced_folder` is off, renewed every minute and removed on closing. Another device's mark under five minutes old is a warning naming it.
- All template code; the sample has two thin scenarios. ADR 3, ADR 5, `docs/data-locations.md`, `docs/configuration.md` and `.claude/rules/storage.md` are updated. `task test:init` passes.

### 6. SQLite adapter and migrations — done
SPEC.md §4.14 · [`features/sqlite-storage.feature`](../../features/sqlite-storage.feature) · [ADR 7](../adr/0007-sqlite-storage.md) · `internal/adapters/storage/sqlite`.
- Done 2026-11-03: `storage = "sqlite"` keeps the data in `<data dir>/data.db` (modernc.org/sqlite, goose migrations, WAL, no lock), on the server, the CLI and TUI, and the desktop app. A new `sqlite` acceptance driver runs every business scenario; the desktop and server drivers cover sharing with the CLI, reopening and readiness.
- This is the existing "SQL storage adapter" TODO item. It passes every contract suite, including isolation.
- The user chooses plain text or SQLite (`storage = "sqlite"`). SQLite has no lock: the desktop app, the CLI and TUI, and several servers on one disk can share it.
- The warning when SQLite's folder looks synced is deferred to a follow-up, now in TODO.md.
  - Done 2026-11-04 as that follow-up: every front end warns, naming the service, when SQLite's data folder looks synced by its path or marker files (`internal/adapters/syncedfolder`); `ignore_synced_folder` turns it off. [`features/synced-folder.feature`](../../features/synced-folder.feature).

### 7. Postgres adapter — not started
- For the hosted edition. It shares the schema with SQLite where possible, and is tested against a Postgres container in CI.

### 8. Authentication — not started
- This is the existing "Authentication behind a port" TODO item.
- The OIDC adapter and the dev-only fake login.
- Auth middleware that sets the workspace, and a personal workspace created at first login.
- Stored plans in the hosted edition: `bootstrap.NewBilling` wires `app.NewBillingService` beside the hosted storage (the `Local` server keeps refusing), `plans.StoredPlanOf` feeds its `plans.Provider`, and the default plan for a new personal workspace is chosen here (see `docs/billing.md`, "Switching on stored plans").
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
- 2026-11-01: step 2 covers limits only. On/off features, with show-locked or hide, move to step 3, where CSV export is their first user. The plans file is JSON with full-line `//` comments, so it can carry sample markers.
- 2026-11-01: step 2 finished. Along the way:
  - `NewRemote` reads the server's plan too, so the CLI and TUI with `--server` see it.
  - The item endpoints keep their `default` problem response, which covers the limit's 403, rather than listing it.
  - `free` and `pro` ship in the plans file as example data, for sign-in (step 8) to put hosted workspaces on.
  - Template code had come to depend on sample code (`bootstrap`'s plans, httpclient's `toError`). Both are fixed, CLAUDE.md now says to keep machinery in template code with one thin sample use, and to run `task test:init` whenever a cycle touches a sample boundary. The other cases found are TODO.md items (`/readyz`, generic UI helpers).
- 2026-11-03: step 6 (SQLite) started before step 5, at the user's request; neither depends on the other. Its synced-folder warning moves to a follow-up. Decided: modernc.org/sqlite (pure Go), goose migrations, no lock (WAL), so SQLite lifts ADR 3's one-program limit.
- 2026-11-03: step 6 finished, recorded in ADR 7. The synced-folder warning went into TODO.md now, not at the end of the plan, so it isn't lost. The template keeps a no-op baseline migration, so SQLite still opens with the sample removed.
- 2026-11-04: step 6's follow-up, the synced-folder warning, finished (TODO.md). It detects by path and marker files only; step 5's conflict-copy detection stays in step 5.
- 2026-11-05: step 3 finished. Along the way:
  - Features got a required display `name`, so template users choose the wording; the CLI help lists locked commands as disabled rather than hiding them, and reads only the entitlements (`bootstrap.NewEntitlementsService`), never opening storage.
  - The web app avoids pop-in: `Layout` prefetches the entitlements for the session, gated pages wait for them in their loading state, and `FeatureGate` has a same-size disabled placeholder.
  - The desktop download was checked by hand on WebKitGTK. `docs/features.md` is the checklist for gating a new feature.
- 2026-11-05: step 4 started. Decided: the local edition's plan stays fixed (`starter admin set-plan` refuses there, so until step 8 only the tests change plans); ChangePlan is reached only by the admin command, with no API until sign-in; each workspace's plan is stored by the storage adapters (`ports.WorkspacePlanRepository`).
- 2026-11-06: step 4 finished. Its stored-plan wiring moved to step 8 here, since the task list that held it is gone. `task test:init` found a template step depending on the sample's `in` helper (from task 12c, which didn't run it), so `in` moved to `steps.go`.
- 2026-11-06: step 5 started. The "lock file" became one mark per device (`.devices/<device>.json`), since two devices rewriting one file would make a conflict copy in exactly the case we warn about. A device writes its mark only when the folder looks synced, and both warnings are checked when the data is opened.
- 2026-11-08: step 5 finished. Its task list split most tasks into one test each, added direct one-record reads with benchmarks (a question about the cost of opening many files), and moved the plans' records to one file per workspace. The by-hand check with two machines (Verification) is still to do.
