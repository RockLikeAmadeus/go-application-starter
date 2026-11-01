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
- **jsonfile:** records gain `workspaceId`. The file format version stays 1. Data saved before this change isn't kept: the template has no users, so there's no backwards compatibility (2026-11-01). Step 5 changes the layout anyway.
- **Acceptance harness:** the memory and jsonfile drivers build the core with `FromContext`. A `Before` hook puts the local workspace in every scenario's context, so features that don't name a workspace behave like the local edition. Workspace steps pass a context for the named workspace. The no-workspace steps pass a context without one. http, desktop and mobile keep `Local`, through their real wiring.
- **Sample vs template:** the workspace type, the resolver port and its adapters are template code (they stay after `--remove-sample`). The workspace scenarios, their steps and the Item isolation tests are sample code: add `features/workspaces.feature` and the new steps file to `SampleFiles` in `scripts/initproject/manifest.go`.

## Tasks
1. [x] **domain:** `NewItem` takes a workspace and sets `Item.WorkspaceID`.
   - **Red (2026-11-01):** `TestNewItem_SetsWorkspace` in `internal/core/domain/item_test.go` fails: `WorkspaceID = "", want "ws-1"`.
   - **Green (2026-11-01):** `NewItem` sets `WorkspaceID` from its first parameter. `task test:go` passes.
   - **Refactor (2026-11-01):** `NewItem`'s struct literal is now one field per line. Deliberately left:
     - `WorkspaceID`'s doc comment, until the resolver exists (task 7) to say who supplies it.
     - `app.CreateItem`'s `""`: task 9 replaces it with the resolved workspace.
     - The test's `ws` constant: it reads fine as it is.
   - Compile scaffolding: `domain.WorkspaceID` in a new template file, `internal/core/domain/workspace.go` (not sample: `item.go` is in `SampleFiles`). `Item.WorkspaceID` field. `NewItem`'s first parameter is ignored for now.
   - `app.CreateItem` passes `""` until task 9 passes the resolved workspace.
   - `domain.LocalWorkspace` and `domain.ErrNoWorkspace` arrive with the tasks that first need them (7 and 8).
2. [x] **storagetest, Save keeps the workspace:** a saved item comes back with its `WorkspaceID` (`assertItemEqual` compares it).
   - **Red (2026-11-01):** contract case "Save keeps the item's workspace" fails on jsonfile (`WorkspaceID:` empty, want `ws-a`): its on-disk record has no workspace field. memory already passes (it stores the domain struct).
   - The port's signatures don't change here: `Save` already carries the workspace on the item. `FindByID`, `FindAll` and `Delete` gain the workspace parameter in tasks 3–5, where isolation demands it.
   - **Green (2026-11-01):** jsonfile's `itemRecord` has `WorkspaceID string` (`json:"workspaceId"`), and `toRecord` and `toDomain` copy it. `task test:go` passes.
   - **Refactor (2026-11-01):** `toRecord` and `toDomain` are now one field per line. Deliberately left:
     - The empty `"workspaceId": ""` written until task 9, which stops it at the source.
     - No `omitempty`, which would change the file format with no test asking for it.
     - The contract case's `want.WorkspaceID = "ws-a"`: tasks 3–5 will need items in named workspaces too, so a workspace parameter on the `item` helper can come then, when there's real duplication.
3. [x] **storagetest isolation, FindAll:** lists only the given workspace's items.
   - **Red (2026-11-01):** contract case "FindAll lists only the given workspace's items" fails on memory and jsonfile: `FindAll(ws-a)` returns both items.
   - Compile scaffolding: the port is now `FindAll(ctx, workspace domain.WorkspaceID)`, and both adapters ignore the parameter. `app.ListItems` passes `""` until task 10. The older contract cases pass `""`, the workspace their items have.
   - **Green (2026-11-01):** memory's and jsonfile's `FindAll` keep only items whose `WorkspaceID` matches before sorting. `task test:go` passes. Running apps are unaffected: until tasks 9 and 10, the app saves and lists with the same `""` workspace.
   - **Refactor (2026-11-01):**
     - `storageutil.SortedItems` became `WorkspaceItems(m, workspace)`, which filters and sorts, so both adapters' `FindAll` call one helper. The file is renamed `storageutil/items.go`; the folder is already in `SampleFiles`.
     - In the contract, items default to a named `workspace` (`ws-a`), and `itemIn(ws, …)` puts one in `otherWorkspace` (`ws-b`). This replaces the hand-set `WorkspaceID`s and the `FindAll(ctx, "")` calls. "Save keeps the item's workspace" now saves into `otherWorkspace`, so it still checks a workspace other than the default.
     - Deliberately left: jsonfile still builds a map of every record before `WorkspaceItems` filters it. Skipping other workspaces' records while decoding would save little, and step 5 (one file per record) rewrites this read path anyway.
4. [x] **storagetest isolation, FindByID:** another workspace's item gives `ErrNotFound`.
   - **Red (2026-11-01):** contract case "FindByID of another workspace's item returns ErrNotFound" fails on memory and jsonfile with `err = <nil>`: both return the other workspace's item.
   - Compile scaffolding: the port is now `FindByID(ctx, workspace, id)`, and both adapters ignore the workspace. `app.GetItem` and `app.RenameItem` pass `""` until task 10. The older contract cases pass the workspace their item was saved in. jsonfile's `TestItemsSurviveReopen` passes `""`, its item's workspace.
   - **Green (2026-11-01):** memory's and jsonfile's `FindByID` return `ErrNotFound` when the stored item's workspace doesn't match. `task test:go` passes.
   - **Refactor (2026-11-01):** the port's `FindByID` doc comment now says it returns `ErrNotFound` when the workspace has no item with the ID, including when the item is in another workspace. Deliberately left: the duplicated "missing, or in another workspace" check in the two adapters. `Delete` needs it too (task 5), and jsonfile compares its record's string rather than a domain item, so decide on a shared helper after task 5's green, with three uses to look at.
5. [x] **storagetest isolation, Delete:** another workspace's item gives `ErrNotFound`, and the item is still there.
   - **Red (2026-11-01):** contract case "Delete of another workspace's item returns ErrNotFound and keeps it" fails on memory and jsonfile. They return `err = <nil>` and delete the item, so `FindByID` in its own workspace then says "not found".
   - Compile scaffolding: the port is now `Delete(ctx, workspace, id)`, and both adapters ignore the workspace. `app.DeleteItem` passes `""` until task 10. The older Delete cases pass `workspace`.
   - **Green (2026-11-01):** memory's and jsonfile's `Delete` return `ErrNotFound`, and delete nothing, when the stored item's workspace doesn't match. `task test:go` passes.
   - **Refactor (2026-11-01):**
     - The port's `Delete` doc comment now matches `FindByID`'s: `ErrNotFound`, and nothing deleted, when the workspace has no item with the ID, including when it's in another workspace.
     - jsonfile's `itemRecord.WorkspaceID` is typed `domain.WorkspaceID`. The JSON is unchanged (still a string), and the casts in `toRecord`, `toDomain` and the checks are gone.
     - Each adapter has one `lookup(workspace, id)` holding the "missing, or in another workspace" rule. It's a method on memory's repository, where the caller holds the lock, and a function over the records map in jsonfile. `FindByID` and `Delete` use it in both.
     - Deliberately left: no shared `storageutil` helper, because jsonfile looks up records and memory looks up domain items, so one helper would need generics for two short functions. Task 6 (records without a workspace) will touch jsonfile's `lookup` and `toDomain`, so the local-workspace default may want to live in one of them.
6. [-] ~~**jsonfile:** a record without `workspaceId` reads as `domain.LocalWorkspace`.~~ **Dropped (2026-11-01):** no backwards compatibility while this is the template (see the log).
7. [x] **adapters/workspace:** `Local` always returns `domain.LocalWorkspace`.
   - **Red (2026-11-01):** `TestLocal_IsAlwaysTheLocalWorkspace` in `internal/adapters/workspace/workspace_test.go` fails: `CurrentWorkspace = "", want "local"`.
   - Compile scaffolding: `domain.LocalWorkspace = "local"` in `domain/workspace.go`. The `ports.WorkspaceResolver` interface, `CurrentWorkspace(ctx) (domain.WorkspaceID, error)`. A new template package `internal/adapters/workspace`, whose `Local` returns `"", nil` for now.
   - **Green (2026-11-01):** `workspace.Local.CurrentWorkspace` returns `domain.LocalWorkspace`. `task test:go` passes.
   - **Refactor (2026-11-01):** doc comments only. `Local` says the local edition has a single workspace, so every request works in `domain.LocalWorkspace`. `WorkspaceID` says use cases get the current workspace from a `ports.WorkspaceResolver`. Deliberately left: naming bootstrap in `Local`'s comment, which only becomes true in task 12.
8. [x] **adapters/workspace:** `FromContext` returns the ID set by `WithID`, and `domain.ErrNoWorkspace` when none is set.
   - Split into two red/green cycles, one test each: (a) the ID set by `WithID`; (b) `ErrNoWorkspace` when none is set.
   - **(a) Red (2026-11-01):** `TestFromContext_IsTheWorkspaceSetByWithID` fails: `CurrentWorkspace = "", want "ws-a"`. Compile scaffolding: `workspace.FromContext` returns `"", nil`, and `workspace.WithID` returns the context unchanged.
   - **(a) Green (2026-11-01):** `WithID` stores the ID under a private `idKey{}` in the context, and `FromContext` reads it back. A missing value gives `""` with no error, because nothing tests that yet: (b) does. `task test:go` passes.
   - **(a) Refactor (2026-11-01):** `idKey` now sits after `WithID`, with a comment that it's unexported so only `WithID` can set it. Deliberately left: `CurrentWorkspace`'s ignored ok, which (b) replaces.
   - **(b) Red (2026-11-01):** `TestFromContext_WithoutAWorkspaceIsRefused` fails: `err = <nil>, want ErrNoWorkspace`. Compile scaffolding: the `domain.ErrNoWorkspace` sentinel in `domain/errors.go`, which is template code, not sample.
   - **(b) Green (2026-11-01):** `FromContext` returns `domain.ErrNoWorkspace` when the context holds no workspace ID. `task test:go` passes.
   - **For refactor (b):**
     - Little to tidy. `FromContext`'s doc comment could say it refuses a request with no workspace (fail closed).
   - **(c) Empty workspace ID refused** (agreed 2026-11-01). `WithID(ctx, "")` must be refused like a missing workspace. Otherwise a bug that leaves a user's workspace ID empty would put every such request in one shared `""` workspace.
     - **Red (2026-11-01):** `TestFromContext_WithAnEmptyWorkspaceIsRefused` fails: `err = <nil>, want ErrNoWorkspace`. No scaffolding needed.
     - **Green (2026-11-01):** `FromContext` refuses when the ID is missing **or** empty (`!ok || id == ""`). `task test:go` passes.
     - **Refactor (2026-11-01):**
       - `FromContext` checks only `id == ""`, with a comment that a missing ID reads as `""`, so missing and empty are refused alike.
       - Its doc comment says it fails closed with `ErrNoWorkspace` for a missing or empty workspace.
       - Deliberately left: the three `FromContext` tests stay separate. One checks the returned ID and two check the error, and as a table they'd need a want-ID-or-error branch that reads worse than three short tests.
9. [x] **app, CreateItem** stamps the resolved workspace on the new item. Add a resolver fake to `testsupport` (fixed, or no workspace).
   - **Red (2026-11-01):** `TestCreateItem_PutsTheItemInTheCurrentWorkspace` in `internal/core/app/item_service_test.go` fails: `WorkspaceID = "", want "ws-a"`. `newService` now builds the service with `testsupport.FixedWorkspace("ws-a")`.
   - Compile scaffolding:
     - `testsupport.FixedWorkspace`, a resolver that always resolves to itself. The no-workspace fake waits for task 11.
     - `NewItemService(repo, workspaces, clock, ids)`, with the resolver stored but unused.
     - Callers: `bootstrap.NewEmbedded` and the acceptance harness's `newItems` pass `workspace.Local{}`. The tui and httpclient tests pass `testsupport.FixedWorkspace(domain.LocalWorkspace)`.
   - **Heads-up for green:** once `CreateItem` stamps the resolved workspace, the reads still pass `""` (task 10). Existing tests will then fail: `TestCreateItem_PersistsAndReturnsItem` (`GetItem`), the rename and delete app tests, and every acceptance scenario. Green's "fix what you broke" rule then means passing the resolved workspace on the reads too, which is task 10's work. If so, task 10 becomes "covered by existing tests" rather than its own red.
   - **Green (2026-11-01, agreed to include task 10):**
     - `CreateItem` resolves the workspace and passes it to `domain.NewItem`.
     - `GetItem`, `ListItems`, `RenameItem` and `DeleteItem` pass the same resolved workspace to the repository.
     - Every use case returns a resolver error as it is.
     - `task test:go` and `task accept` (not `@wip`) pass.
   - **Refactor (2026-11-01):** `NewItemService`'s signature is now one parameter per line (it had grown to about 130 characters). Deliberately left:
     - The repeated "resolve the workspace or return the error" block in the five use cases. A helper would still leave an `if err != nil` in each, so it saves one line per use case and hides what each one does. Look again only if task 11's fail-closed handling makes the block longer.
     - Wrapping resolver errors. `ErrNoWorkspace` reads clearly as it is, and nothing needs the extra context yet.
10. [x] **app, ListItems, GetItem, RenameItem and DeleteItem** pass the resolved workspace to the repository.
    - **Done in task 9's green (2026-11-01, agreed).** Once `CreateItem` stamped the resolved workspace, the existing tests demanded that the reads use it too. Those are `TestCreateItem_PersistsAndReturnsItem` (`GetItem`), `TestCreateItem_ValidationErrorIsNotPersisted` (`ListItems`), the rename and delete app tests, and every acceptance scenario. No separate red: a new test would have passed at once.
11. [x] **app, fail closed:** with no workspace, every use case returns `ErrNoWorkspace` and doesn't touch the repository (table test).
    - **Red attempted (2026-11-01): passes at once.** Task 9's green already made every use case resolve the workspace first and return the resolver's error. So `TestUseCases_WithoutAWorkspaceAreRefusedBeforeStorage` (all five use cases, with `testsupport.NoWorkspace` and an `untouchedRepo` that counts calls) is green without new code.
    - Shown to bite: with `ListItems` briefly changed to ignore the resolver error, the test failed (`err = <nil>, want ErrNoWorkspace`, `repository called 1 times, want 0`), and the change was reverted.
    - Kept as a guard of the fail-closed rule. New: `testsupport.NoWorkspace`, a resolver that always returns `ErrNoWorkspace`. Ticked with no green (agreed 2026-11-01): the behaviour already existed.
    - **Refactor (2026-11-01):**
      - A `newServiceWith(repo, workspaces)` helper builds the service with the fixed clock and sequential IDs. `newService` and the fail-closed test both use it, instead of repeating that setup.
      - The table's cases are all multi-line and take `ctx`.
      - `untouchedRepo`'s methods are laid out alike.
      - Deliberately left: `testsupport.NoWorkspace` as it is, small and doing one thing. The test runs one case per use case rather than looping over the `ports.ItemService` methods by reflection, which would be harder to read.
12. [x] **bootstrap:** `NewEmbedded` wires `workspace.Local`. Update the acceptance harness's `newItems` to take a resolver (memory and jsonfile get `FromContext` plus the `Before` hook; http keeps `Local`).
    - **No red/green (2026-11-01):**
      - **Bootstrap:** `NewEmbedded` already passed `workspace.Local{}`, the compile scaffolding from task 9. The desktop and mobile acceptance drivers cover it, since they open the app through `bootstrap`, so a missing resolver would fail every scenario on them.
      - **Harness:** test code, not production code. `newItems(repo, workspaces)`: memory and jsonfile get `workspace.FromContext{}`, and http gets `workspace.Local{}`, with a comment that it's the local edition like `cmd/server`.
    - **The hook is needed:** with `FromContext` but no hook, every non-`@wip` scenario failed on memory and jsonfile with "no workspace", and the other drivers passed. That's the fail-closed rule seen end to end.
    - **Hook added:** the `Before` hook in `steps.go` returns `workspace.WithID(ctx, domain.LocalWorkspace)`. All non-`@wip` scenarios pass on all five drivers. The `@wip` ones fail only on undefined steps (task 13). `task lint` and `task test` pass.
13. [x] **Acceptance steps** for `features/workspaces.feature`, in a new `test/acceptance/workspace_steps.go`. Thin: they build a context and call `ports.ItemService`.
    - **Done (2026-11-01):** `workspace_steps.go` registers six steps (from `steps.go`, inside the sample block).
      - `in(name)` builds a context for a named workspace.
      - The "has an item" and "has the items" steps reuse `anItemNamed` and `theItemsAre` with that context.
      - "Tries to view, rename or delete" finds the item in its owner's workspace, then calls from the requester's.
      - The no-workspace steps use a fresh `context.Background()` rather than the scenario's local-workspace context.
    - All six `@wip` scenarios pass on memory and jsonfile, and the other drivers skip them. **The full `task accept` is green.**
    - **Shown to bite:** with `storageutil.WorkspaceItems` briefly made to ignore the workspace, "Each workspace lists only its own items" failed (`items = ["Widget" "Gadget"], want ["Widget"]`), and the change was reverted.
14. [x] **Template hygiene:** add the scenarios and steps file to `SampleFiles`, then run `task test:init`.
    - Note (2026-11-01): `task test:init` fails until the `@wip` scenarios pass. The generated project's `go test ./test/...` runs them, and their steps are undefined until task 13. Seen when checking the new template-only block in CLAUDE.md: `init` stripped it, then failed only on these scenarios. So run `test:init` after task 13.
    - **Done (2026-11-01):**
      - `features/workspaces.feature` and `test/acceptance/workspace_steps.go` are in `SampleFiles`, beside the item ones. Their registration in `steps.go` is already inside the sample block.
      - The template code stays after `--remove-sample`: `domain.WorkspaceID`, `LocalWorkspace` and `ErrNoWorkspace`, `ports.WorkspaceResolver`, `adapters/workspace`, the `testsupport` resolver fakes, and the `Before` hook.
      - `task test:init` passes (exit 0).
    - The init tests don't check CLAUDE.md's template-only block, so a throwaway test (deleted afterwards) ran init's `stripBlocks` on the real CLAUDE.md. The backwards-compatibility rule and both markers are removed.
15. [x] **Docs:**
    - CLAUDE.md Rules: the workspace comes from `ports.WorkspaceResolver`, repositories take it explicitly, and every repository contract includes isolation. Also add `internal/adapters/workspace` to the Architecture block.
    - ADR 5: the resolver refinement.
    - ADR 3 and the jsonfile package doc: the `workspaceId` field.
    - The editions plan: mark step 1 done.
    - **Done (2026-11-01):**
      - **CLAUDE.md:** a workspace rule (the resolver port, the explicit workspace in repositories, isolation in every contract suite, `ErrNoWorkspace` before storage, which fakes to use). `ErrNoWorkspace` is added to the domain errors, with a note that it isn't mapped to HTTP until sign-in. The architecture block has a line for `internal/adapters/workspace`.
      - **ADR 5:** the resolver approach, with why it replaced reading the context directly. Its status says step 1 is built.
      - **ADR 3 and the jsonfile package doc:** records carry `workspaceId`, and lookups are per workspace.
      - **The editions plan:** step 1 is marked done, with a summary and what's left for step 8.
16. [ ] **Finish:** `task check`, `task accept` (the `@wip` scenarios green), then remove the `@wip` tags and this file.

## Changes to the plan
- 2026-11-01: task 2 no longer changes the repository signatures. Saving the workspace needs only the jsonfile record field. The workspace parameter on `FindByID`, `FindAll` and `Delete` moves to tasks 3–5, which demand it.
- 2026-11-01: task 6 moved to after task 12. Until tasks 9–12, the app saves and lists items with an empty workspace. If jsonfile read empty-workspace records as `local` now, items the app just saved would vanish from its lists. That would break the jsonfile, desktop and mobile acceptance drivers and the real desktop app. Once the app resolves `local` from `workspace.Local`, task 6 only affects data saved before workspaces existed. Order is now 7, 8, 9, 10, 11, 12, 6, 13… (task 6 was later dropped: see below).
- 2026-11-01: task 8 gains a third cycle, (c): an empty workspace ID from `WithID` is refused with `ErrNoWorkspace`, the same as a missing one. It came up in task 8's green; see the decision under (c).
- 2026-11-01: task 10 was done in task 9's green, as agreed. Stamping the resolved workspace on create broke every read that still passed `""`, so fixing the existing tests was task 10.
- 2026-11-01: task 6 dropped. This is the template and has no users, so data saved before workspaces existed doesn't need to read as `local`. Breaking changes are fine at this stage. Products built from the template handle their own compatibility. SPEC.md §4.8 loses its matching rule and example.
