# WIP: workspace scoping

Step 1 of [docs/plans/editions-and-accounts.md](../plans/editions-and-accounts.md). Delete this file in the same change that removes the `@wip` tags.

- Spec: [SPEC.md §4.8](../../SPEC.md) "Keep each workspace's data separate"
- Scenarios: [features/workspaces.feature](../../features/workspaces.feature) (tagged `@workspaces`; run on the memory and jsonfile drivers only: see `drivers` in `test/acceptance/drivers.go`)
- Decisions: [ADR 5](../adr/0005-hosted-and-local-editions.md)

## Decisions
- **A driven port supplies the workspace** (agreed 2026-10-31, refining the plan's "use cases read it from the context"). Use cases call `ports.WorkspaceResolver.CurrentWorkspace(ctx) (domain.WorkspaceID, error)`. There are two adapters in a new `internal/adapters/workspace` package:
  - `Local`: always `domain.LocalWorkspace`. `bootstrap.NewEmbedded` wires it, so CLI, TUI, server, desktop and mobile change nothing.
  - `FromContext`: reads the ID that `workspace.WithID(ctx, id)` put in the context, and returns `domain.ErrNoWorkspace` when there isn't one. Step 8's auth middleware will set it. The acceptance tests use it now.
- **Fail closed:** without a workspace, a use case returns `domain.ErrNoWorkspace` and doesn't touch the repository.
- **Repositories take the workspace explicitly:** `FindByID(ctx, ws, id)`, `FindAll(ctx, ws)` and `Delete(ctx, ws, id)`. `Save(ctx, item)` uses `item.WorkspaceID`. Another workspace's item gives `domain.ErrNotFound`, exactly like a missing one. Item IDs stay globally unique (UUIDs). The contract suite gets isolation tests, so every backend is checked.
- **`domain.Item` gets a `WorkspaceID`.** `domain.WorkspaceID` is a string type, and `domain.LocalWorkspace` is `"local"`.
- **The API doesn't change** (agreed 2026-10-31). The workspace is implied by who's asking, so `api/openapi.yaml`, the SPA and the CLI output are untouched. Mapping `ErrNoWorkspace` to HTTP (probably 401) waits for step 8, when it first becomes reachable over HTTP.
- **jsonfile:** records gain `workspaceId`. A record without one (saved before this change) reads as the local workspace. The file format version stays 1, because the field is additive. Step 5 changes the layout anyway.
- **Acceptance harness:** the memory and jsonfile drivers build the core with `FromContext`. A `Before` hook puts the local workspace in every scenario's context, so features that don't name a workspace behave like the local edition. Workspace steps pass a context for the named workspace. The no-workspace steps pass a context without one. http, desktop and mobile keep `Local`, through their real wiring.
- **Sample vs template:** the workspace type, the resolver port and its adapters are template code (they stay after `--remove-sample`). The workspace scenarios, their steps and the Item isolation tests are sample code: add `features/workspaces.feature` and the new steps file to `SampleFiles` in `scripts/initproject/manifest.go`.

## Tasks
1. [ ] **domain:** `NewItem` takes a workspace and sets `Item.WorkspaceID` (`domain.WorkspaceID`, `domain.LocalWorkspace`, `domain.ErrNoWorkspace`).
2. [ ] **storagetest, Save keeps the workspace:** a saved item comes back with its `WorkspaceID` (`assertItemEqual` compares it). This changes the `ItemRepository` port signatures, so memory and jsonfile (record field) must follow.
3. [ ] **storagetest isolation, FindAll:** lists only the given workspace's items.
4. [ ] **storagetest isolation, FindByID:** another workspace's item gives `ErrNotFound`.
5. [ ] **storagetest isolation, Delete:** another workspace's item gives `ErrNotFound`, and the item is still there.
6. [ ] **jsonfile:** a record without `workspaceId` reads as `domain.LocalWorkspace` (FindAll and FindByID).
7. [ ] **adapters/workspace:** `Local` always returns `domain.LocalWorkspace`.
8. [ ] **adapters/workspace:** `FromContext` returns the ID set by `WithID`, and `domain.ErrNoWorkspace` when none is set.
9. [ ] **app, CreateItem** stamps the resolved workspace on the new item. Add a resolver fake to `testsupport` (fixed, or no workspace).
10. [ ] **app, ListItems, GetItem, RenameItem and DeleteItem** pass the resolved workspace to the repository.
11. [ ] **app, fail closed:** with no workspace, every use case returns `ErrNoWorkspace` and doesn't touch the repository (table test).
12. [ ] **bootstrap:** `NewEmbedded` wires `workspace.Local`. Update the acceptance harness's `newItems` to take a resolver (memory and jsonfile get `FromContext` plus the `Before` hook; http keeps `Local`).
13. [ ] **Acceptance steps** for `features/workspaces.feature`, in a new `test/acceptance/workspace_steps.go`. Thin: they build a context and call `ports.ItemService`.
14. [ ] **Template hygiene:** add the scenarios and steps file to `SampleFiles`, then run `task test:init`.
15. [ ] **Docs:**
    - CLAUDE.md Rules: the workspace comes from `ports.WorkspaceResolver`, repositories take it explicitly, and every repository contract includes isolation. Also add `internal/adapters/workspace` to the Architecture block.
    - ADR 5: the resolver refinement.
    - ADR 3 and the jsonfile package doc: the `workspaceId` field.
    - The editions plan: mark step 1 done.
16. [ ] **Finish:** `task check`, `task accept` (the `@wip` scenarios green), then remove the `@wip` tags and this file.

## Changes to the plan
_None yet._
