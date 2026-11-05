# 5. Hosted and local editions share one core

- Status: accepted. Workspaces (step 1) and plan limits (step 2) are built; the rest is planned. See [`docs/plans/editions-and-accounts.md`](../plans/editions-and-accounts.md).
- Date: 2026-10-31

## Context

A product built from this template may be sold as a hosted service: many users, data kept separate per account, and some features only on paid plans. The same product may also ship as free single-user tools on local data: the CLI and TUI, and local-only copies of the desktop and mobile apps. We want both from one core and one set of data structures, without a second code path for each.

## Decision

### Two editions, chosen at the composition root

- The **hosted edition** is multi-user, with Postgres on a server. It's reached by the web app, by the CLI and TUI after `starter login`, and by the desktop and mobile apps in remote mode (ADR 4).
- The **local edition** is single-user, with plain text or SQLite in a data folder the user chooses: the CLI, the TUI, and the desktop and mobile apps in local mode.
- **The server can run as either edition.** With sign-in on, it is the hosted edition. With sign-in off, it is a single-user local edition, for someone hosting it for themselves (a home server or NAS). It uses the same fixed local workspace as the desktop app, and the web app works the same either way.
- Each copy of a front-end uses one edition, never both.
- The core never knows which edition it's in. `internal/bootstrap` plugs in the adapters for one or the other, as it already does for storage.

### Data belongs to a workspace

- Every record has a `WorkspaceID`. In the hosted edition, each user gets a personal workspace, so teams can be added later as workspace members and roles, with no data migration. The local edition has exactly one implicit workspace.
- **Use cases ask a driven port for the current workspace:** `ports.WorkspaceResolver`. They pass the answer **explicitly** to repositories.
  - The local edition's resolver, `workspace.Local`, always answers the local workspace, so no local front end has to set anything.
  - The hosted edition's resolver, `workspace.FromContext`, reads what the auth middleware put in the request's `context` (`workspace.WithID`).
  - A missing or empty workspace is refused with `domain.ErrNoWorkspace` before storage is touched.
  - This refines the first plan, where use cases read the context directly. With a resolver port, local front ends need no changes, and a forgotten context value fails closed instead of leaking.
- The `storagetest` contract suite checks isolation: one workspace never sees another's records. Every storage backend must pass it.

### Features and limits come from plans

- A **plan** grants on/off **features** and numeric **limits**. Plans are defined in config, not code, so pricing can change without a release.
- For each limit, config says what happens when a workspace is over it: keep the data and refuse new records, or make that data read-only. For each feature, config says whether the UI shows it locked, with an upgrade hint, or hides it.
- Use cases ask an `EntitlementsProvider` port. The REST API exposes the current workspace's entitlements, so every UI can follow the same rules.
- The local edition's workspace is on a plan named at build time. A paid local licence could later switch it.
- Built in step 2: limits, in a plans file that is JSON with full-line comments (`internal/adapters/plans/plans.jsonc`, replaceable with `STARTER_PLANS_FILE`), and `GET /api/v1/entitlements`. A limit's error is a 403 naming the limit. On/off features come in step 3.

### Billing is a driving adapter

- An admin-only `ChangePlan` use case is the core's only billing concept. Anything that takes money (Stripe or Paddle webhooks, an admin command) is a driving adapter that calls it.
- The template ships only the admin command, and a guide (`docs/billing.md`) for adding a provider.

### Authentication

- Hosted edition only: an OIDC adapter behind a port, plus a dev-only fake login. We store no passwords.
- The CLI and TUI are local by default, and can log in to a hosted server.

### Synced data folders: one device at a time

- A local data folder may sit in a folder that a file-sync service (Dropbox, Syncthing, iCloud Drive, …) keeps in sync, as long as only one device uses it at a time.
- **The app doesn't merge.** The plain-text format keeps one file per record, so any clash is limited to that record. The app reports the sync service's conflict copies, and warns, best effort, when another device seems to be using the folder.
- SQLite can be corrupted by file-sync services, so the app warns when a SQLite folder looks synced.
- There is no sync between local and cloud data (ADR 4). For live use from several devices, use the hosted edition.

Rejected:
- **Merging edits from several devices inside a synced folder:** it needs watching files, tombstones and merge rules for every entity, and can't be tested against real sync services in CI. It's a lasting maintenance cost and still unreliable.
- **The app's own sync engine** (a change log replicated through a server, or a CRDT library): a product in its own right, and it would undo ADR 4's no-sync decision.
- **Build tags for gated features:** a second mechanism beside plans, and more build combinations to test.
- **A separate core or schema for the local edition:** it would mean double the code and drift between the two.

## Consequences

- Every repository query is scoped to a workspace. The contract suite enforces it, so it costs little to keep up.
- The local edition stays cheap: today's app, plus an implicit workspace and a fixed plan.
- The real upfront work (authentication, Postgres, billing) is what any hosted product needs anyway, and the template ships only reference adapters for it.
- Synced folders on Android need its document picker (Storage Access Framework), so they're a follow-up. The Android app keeps its data in its private storage.
