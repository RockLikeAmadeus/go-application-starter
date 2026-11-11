# 5. Hosted and local editions share one core

- Status: accepted. Workspaces (step 1), plan limits (step 2), plan features (step 3), plan changes (step 4), plain text in a synced folder (step 5), SQLite (step 6), Postgres (step 7, ADR 8), and sign-in with the dev sign-in (step 8a, ADR 9) are built; the rest is planned. See [`docs/plans/editions-and-accounts.md`](../plans/editions-and-accounts.md).
- Date: 2026-10-31

## Context

A product built from this template may be sold as a hosted service: many users, data kept separate per account, and some features only on paid plans. The same product may also ship as free single-user tools on local data: the CLI and TUI, and local-only copies of the desktop and mobile apps. We want both from one core and one set of data structures, without a second code path for each.

## Decision

### Two editions, chosen at the composition root

- The **hosted edition** is multi-user, with Postgres on a server (ADR 8). It's reached by the web app, by the CLI and TUI after `starter login`, and by the desktop and mobile apps in remote mode (ADR 4).
- The **local edition** is single-user, with plain text or SQLite in a data folder the user chooses: the CLI, the TUI, and the desktop and mobile apps in local mode.
- **The server can run as either edition.** With sign-in on, it is the hosted edition. With sign-in off, it is a single-user local edition, for someone hosting it for themselves (a home server or NAS). It uses the same fixed local workspace as the desktop app, and the web app works the same either way.
- Each copy of a front-end uses one edition, never both.
- The core never knows which edition it's in. `internal/bootstrap` plugs in the adapters for one or the other, as it already does for storage.

### Data belongs to a workspace

- Every record has a `WorkspaceID`. In the hosted edition, each user gets a personal workspace, so teams can be added later as workspace members and roles, with no data migration. The local edition has exactly one implicit workspace.
- **Use cases ask a driven port for the current workspace:** `ports.WorkspaceResolver`. They pass the answer **explicitly** to repositories.
  - The local edition's resolver, `workspace.Local`, always answers the local workspace, so no local front end has to set anything.
  - The hosted edition's resolver, `workspace.FromContext`, reads what the session middleware put in the request's `context` (`workspace.WithID`). A request with no valid session has no workspace (ADR 9).
  - A missing or empty workspace is refused with `domain.ErrNoWorkspace` before storage is touched.
  - This refines the first plan, where use cases read the context directly. With a resolver port, local front ends need no changes, and a forgotten context value fails closed instead of leaking.
- The `storagetest` contract suite checks isolation: one workspace never sees another's records. Every storage backend must pass it.

### Features and limits come from plans

- A **plan** grants on/off **features** and numeric **limits**. Plans are defined in config, not code, so pricing can change without a release.
- For each limit, config says what happens when a workspace is over it: keep the data and refuse new records, or make that data read-only. For each feature, config says whether the UI shows it locked, with an upgrade hint, or hides it.
- Use cases ask an `EntitlementsProvider` port. The REST API exposes the current workspace's entitlements, so every UI can follow the same rules.
- The local edition's workspace is on a plan named at build time. A paid local licence could later switch it.
- Built in step 2: limits, in a plans file that is JSON with full-line comments (`internal/adapters/plans/plans.jsonc`, replaceable with `STARTER_PLANS_FILE`), and `GET /api/v1/entitlements`. A limit's error is a 403 naming the limit. Built in step 3: on/off features, in the same file's `features` section, each with a display name and `whenOff` (`show-locked` or `hide`). A use case refuses a feature that's off with a 403 naming it, and the web app, CLI help and TUI show it by the plan (`docs/features.md`).

### Billing is a driving adapter

- An admin-only `ChangePlan` use case is the core's only billing concept. Anything that takes money (Stripe or Paddle webhooks, an admin command) is a driving adapter that calls it.
- The template ships only the admin command, and a guide (`docs/billing.md`) for adding a provider.
- `ChangePlan` is its own driving port, `ports.BillingService`, and takes the workspace as an argument, because an admin or a provider acts on someone else's workspace. An admin reaches it only through `starter admin set-plan`, beside the hosted edition's storage. No API changes plans until sign-in and roles exist.
- Each workspace's plan is stored by the storage adapters, through `ports.WorkspacePlanRepository`.
- **The local edition never reads a stored plan.** Its plan stays fixed at the one named at build time, so a local user can't upgrade their own copy, and `starter admin set-plan` refuses there.
- A plan change checks no limits: moving to a smaller plan keeps the data, which the over-limit rules then govern.
- Built in step 4: all of the above, in every storage adapter.
- Built in step 8a: with sign-in on, the copy is the hosted edition. `bootstrap.NewBilling` then returns the real billing service on the configured storage, and each workspace's plan is its stored plan. A workspace with no stored plan is on the plans file's `newWorkspacePlan`. With sign-in off, `bootstrap.NewBilling` keeps refusing.

### Authentication

- Hosted edition only: an OIDC adapter behind a port, plus a dev-only fake login. We store no passwords.
- Built in step 8a: the port (`ports.SignInService`), accounts, sessions in a cookie and in storage, and the dev sign-in, in dev builds only. ADR 9 records how. The OIDC adapter is step 8b.
- The CLI and TUI are local by default, and can log in to a hosted server.

### Synced data folders: one device at a time

- A local data folder may sit in a folder that a file-sync service (Dropbox, Syncthing, iCloud Drive, …) keeps in sync, as long as only one device uses it at a time.
- **The app doesn't merge.** The plain-text format keeps one file per record, so any clash is limited to that record (ADR 3). The app reports the sync service's conflict copies, and warns, best effort, when another device seems to be using the folder.
- **Plain text is never used in the hosted edition,** so a plain-text data folder only ever holds one workspace. The files need no folder per workspace, and nothing in them needs to perform well across many workspaces. `jsonfile` still takes the workspace on every call and passes the isolation suite, because the ports are shared.
- **A conflict copy is a file that doesn't match its own key** (ADR 3). Every copy of the app warns about them on plain text, naming the files, whether or not the folder looks synced: a copy is itself the evidence, so `ignore_synced_folder` doesn't turn it off.
- **Each device marks the folder in a file of its own,** `<data dir>/.devices/<device>.json`, holding the device and the time. One shared lock file would itself get a conflict copy when two devices rewrite it, in exactly the case we warn about. The device is the computer's name (`os.Hostname()`), so the desktop app and the CLI on one computer are one device, and two computers with the same name look like one. With no name, the app leaves no mark.
- **A mark is written only when the folder looks synced:** `syncedfolder.Detect` finds a service and `ignore_synced_folder` is off. So the default folder, servers, and the Android app make no background writes, and a sync setup we can't detect gets no warning. Every plain-text copy still reads other devices' marks.
- **Timing:** the mark is renewed every minute and removed on closing (`App.Close`), and a mark five minutes old or more doesn't count, to allow for slow sync services. So a device that crashed or lost power stops counting by itself. Times are compared on this device's clock, so a skewed clock can miss or invent a warning.
- **Both warnings are checked when the data is opened,** and go in `App.Warnings`, so every front end shows them with no new plumbing. A desktop app left open sees a new conflict copy the next time it opens.
- **Where it lives:** the conflict copy scan is `jsonfile.Store.ConflictCopies`, since only it knows the layout. Marks are in `internal/adapters/syncedfolder` (`KeepMark`, `OtherDevices`), beside `Detect`, and `internal/bootstrap` composes them. Neither adapter imports the other. The local `.lock` still stops two programs on one computer.
- SQLite can be corrupted by file-sync services, so the app warns when a SQLite folder looks synced. Marks are plain text only.
- There is no sync between local and cloud data (ADR 4). For live use from several devices, use the hosted edition.

Rejected:
- **Merging edits from several devices inside a synced folder:** it needs watching files, tombstones, and merge rules for every entity, and can't be tested against real sync services in CI. It's a lasting maintenance cost and still unreliable.
- **The app's own sync engine** (a change log replicated through a server, or a CRDT library): a product in its own right, and it would undo ADR 4's no-sync decision.
- **Build tags for gated features:** a second mechanism beside plans, and more build combinations to test.
- **A separate core or schema for the local edition:** it would mean double the code and drift between the two.

## Consequences

- Every repository query is scoped to a workspace. The contract suite enforces it, so it costs little to keep up.
- The local edition stays cheap: today's app, plus an implicit workspace and a fixed plan.
- The real upfront work (authentication, Postgres, billing) is what any hosted product needs anyway, and the template ships only reference adapters for it.
- Synced folders on Android need its document picker (Storage Access Framework), so they're a follow-up. The Android app keeps its data in its private storage.
