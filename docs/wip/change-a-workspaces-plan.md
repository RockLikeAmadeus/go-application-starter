# Change a workspace's plan

Step 4 of [`docs/plans/editions-and-accounts.md`](../plans/editions-and-accounts.md).

- Spec: [SPEC.md §4.15](../../SPEC.md) ("Change a workspace's plan"), plus the **Admin** glossary term.
- Scenarios: [`features/plan-changes.feature`](../../features/plan-changes.feature), and the terminal journey "The command line can't change the local edition's plan" in [`test/terminal/features/command-line.feature`](../../test/terminal/features/command-line.feature).
- Lasting decisions: [ADR 5](../adr/0005-hosted-and-local-editions.md) "Billing is a driving adapter".

## Decisions (agreed 2026-11-05)

- **The local edition's plan stays fixed** at `bootstrap.LocalPlan`. It never reads a stored plan, so a local user can't upgrade their own copy. `starter admin set-plan` refuses in the local edition. Until step 8 adds sign-in (and with it the hosted edition), every copy is the local edition, so the command always refuses. Stored plans take effect only where a workspace comes from `workspace.FromContext`, which today means the acceptance drivers.
- **Admin-only means CLI only, no API.** `ChangePlan` is its own driving port (`ports.BillingService`), reached only by `starter admin set-plan WORKSPACE PLAN`, which runs embedded beside the hosted edition's storage. It refuses `--server`. No endpoint changes plans, so `api/openapi.yaml` doesn't change. An admin API waits for sign-in and roles (step 8).
- **Plans are stored in the storage adapters**, through a new driven port, `ports.WorkspacePlanRepository`, with its own `storagetest` contract suite, in memory, jsonfile and sqlite (and Postgres in step 7):
  - `PlanOf(ctx, workspace) (string, error)` returns `domain.ErrNotFound` for a workspace with no stored plan;
  - `SavePlan(ctx, workspace, plan) error` inserts or replaces it;
  - isolation: saving one workspace's plan never changes another's.
  It doesn't embed `WorkspaceRepository[T]`: a plan is one value per workspace, not a list of records.
- **Unknown plans:** `ChangePlan` asks a driven `ports.PlanCatalog` (`HasPlan`), which `plans.Provider` implements. A plan the file lacks is a `*domain.ValidationError` on field `plan` (reusing the step "I am told the "plan" is invalid"). A failed lookup refuses too: fail closed.
- **Explicit workspace:** `ChangePlan(ctx, workspace, plan)` takes the workspace as an argument, not from the resolver, because an admin acts on someone else's workspace. An empty one is `domain.ErrNoWorkspace`, before storage is touched. Workspaces aren't stored entities yet, so any non-empty ID is accepted.
- **No limit checks on a change:** moving to a smaller plan keeps the data, which the over-limit rules of §4.9 then govern (`features/plans.feature`).
- **The plan a workspace reads:** `plans.Provider.PlanOf` gets a helper that reads the stored plan and falls back to a default for a workspace with none stored. Any other error is passed on, never defaulted. The harness's default is `bootstrap.LocalPlan`, as today. The hosted edition's default plan (for a new personal workspace) is chosen in step 8.
- **Wiring:** `bootstrap.NewBilling(cfg)` returns `bootstrap.ErrLocalPlanIsFixed` for every copy for now (and for `--server`, a "not over a server yet" error), without opening storage. Step 8 wires the stored plans there.
- **Harness:** `testPlans.assign`'s map is replaced by each driver's `WorkspacePlanRepository`. "My workspace is on / moves to the … plan" now go through `ChangePlan`, so `features/plans.feature` and `features/plan-features.feature` prove it too. The http driver's admin calls `ChangePlan` in-process on the server's own storage, as `starter admin set-plan` would beside a server.
- **Template code, not sample:** all of it, including the new SQLite migration and the jsonfile file. The thin sample use is the existing item scenarios that move plans. Run `task test:init` in the storage cycles.
- **`docs/billing.md`** is the template users' guide described in the plan (concepts, paid accounts by hand, a Stripe walkthrough, merchant-of-record options, downgrades and grace periods, a paid local licence). It's docs, so it's the last task, not a TDD cycle.

## Tasks

Core:
1. [x] `app.BillingService.ChangePlan` saves the plan for the given workspace (introduces `ports.BillingService`, `ports.WorkspacePlanRepository`, `ports.PlanCatalog`, and their `testsupport` fakes).
   - Green: `ChangePlan` calls `SavePlan`; the catalog is wired but unused until task 2.
   - Refactor: nothing worth changing; the code mirrors the sibling services and the fakes. `catalog` stays unread until task 2.
2. [x] `ChangePlan` refuses a plan the catalog doesn't have with a `ValidationError` on `plan`, and saves nothing.
   - Red: `TestChangePlan_RefusesAPlanTheCatalogDoesNotHave` fails, `ChangePlan(platinum) = <nil>, want a *ValidationError on plan`.
   - Green: `ChangePlan` asks `catalog.HasPlan` and returns a `ValidationError` on `plan` for an unknown one; the lookup error is ignored until task 3.
   - Refactor: nothing worth changing; the check is three lines and the test matches its sibling. The ignored lookup error stays for task 3 to demand.
3. [x] `ChangePlan` refuses when the catalog can't be read, and saves nothing.
   - Red: `TestChangePlan_RefusesWhenTheCatalogCannotBeRead` fails, `ChangePlan = plan does not exist, want the catalog's error`.
   - Green: `ChangePlan` returns the wrapped lookup error before checking `known`.
   - Refactor: the test's fake is now `unreadableCatalog`, laid out like `unavailablePlans`; no production change.
4. [x] `ChangePlan` refuses an empty workspace with `ErrNoWorkspace`, before touching the catalog or storage.
   - Red: `TestChangePlan_RefusesAnEmptyWorkspaceBeforeTouchingTheCatalogOrStorage` fails, `ChangePlan = look up plan: catalog unavailable, want ErrNoWorkspace`.
   - Green: `ChangePlan` returns `ErrNoWorkspace` first when the workspace is empty.
   - Refactor: the billing tests share `assertStoredPlans`, replacing four copies of the comparison; no production change.

Storage:
5. [x] `storagetest.RunWorkspacePlanRepositoryContract`, run by a memory `WorkspacePlanRepository`: no stored plan is `ErrNotFound`.
   - Red: `TestWorkspacePlanRepositoryContract/PlanOf_with_no_stored_plan_returns_ErrNotFound` fails, `err = <nil>, want ErrNotFound`.
   - Green: memory's `PlanOf` always returns `ErrNotFound`; `SavePlan` does nothing until 5b.
   - Refactor: the `memory` package comment moved from the sample file to this template file, and no longer names items; no code change.
5b. [x] Contract: a saved plan is read back.
   - Red: `TestWorkspacePlanRepositoryContract/PlanOf_returns_the_saved_plan` fails, `PlanOf: not found`.
   - Green: memory keeps the one saved plan in a field, ignoring the workspace until 5d.
   - Refactor: nothing worth changing; `"pro"` is used twice in one short test. The map and mutex wait for 5d, which demands per-workspace storage.
5c. [x] Contract: saving again replaces the plan.
   - No red: `TestWorkspacePlanRepositoryContract/SavePlan_replaces_the_saved_plan` passed straight away, because memory's single field already overwrites. Kept as a guard: a first-save-wins mutation fails it, `plan = "pro", want "free"`. No green.
   - Refactor: the contract's tests share `mustSavePlan` and `assertPlan`, as the item contract does; no production change.
5d. [x] Contract: one workspace's plan isn't another's.
   - Red: `TestWorkspacePlanRepositoryContract/SavePlan_leaves_another_workspace's_plan_alone` fails, `PlanOf("ws-a") = "free", want "pro"`.
   - Green: memory keeps a mutex-guarded map of plans by workspace, replacing the single field.
   - Refactor: nothing worth changing; the repository already matches `ItemRepository`'s layout.
6. [x] jsonfile `WorkspacePlanRepository` passes the contract. Run `task test:init`.
   - Red: `TestWorkspacePlanRepositoryContract` (jsonfile) fails all four cases against a stub, e.g. `err = <nil>, want ErrNotFound`.
   - Green: `workspace_plans.json`, one record per workspace ID, through `readTable`/`writeTable`.
   - Refactor: the `Store` package comment no longer names items, `mu` loses its `nolint:unused`, and the `table_test.go` comment no longer assumes a project with no repositories; no behaviour change. The read-modify-write stays in each repository: the only other copy is the sample's.
6b. [x] jsonfile keeps the plans after the store is reopened.
   - No red: `TestWorkspacePlansSurviveReopen` passed straight away, because task 6 already writes through `writeTable`. Kept as a guard: a repository holding plans in memory passes the contract but fails it, `after reopen PlanOf = "", not found, want "pro"`. No green.
   - Refactor: nothing worth changing; the test is short, and `openStore` can't open the first store because it must close before the reopen.
7. [x] sqlite `WorkspacePlanRepository` (a new template migration) passes the contract. Run `task test:init` (the migration comes after the sample's `00002_items.sql`).
   - Red: `TestWorkspacePlanRepositoryContract` (sqlite) fails all four cases against a stub, e.g. `err = <nil>, want ErrNotFound`.
   - Green: `00003_workspace_plans.sql`, one row per workspace, with `INSERT OR REPLACE` and `ErrNotFound` for no row.
   - Refactor: `openStore` moved to the template's `store_test.go`, used by the contract test and the sample's `openItemRepository`, as in jsonfile; no production change.
7b. [x] sqlite keeps the plans after the store is reopened.
   - No red: `TestWorkspacePlansSurviveReopen` passed straight away, because task 7 already writes to the database file. Kept as a guard: a `CREATE TEMP TABLE` migration passes the contract but fails it, `after reopen PlanOf = "", no such table: workspace_plans`. No green.
   - Refactor: nothing worth changing; the test mirrors jsonfile's, and `openStore` can't open the first store because it must close before the reopen.

Plans adapter:
8. [x] `plans.Provider.HasPlan` answers whether its catalog has the plan (`ports.PlanCatalog`).
   - Red: `TestHasPlan_AnswersWhetherTheCatalogHasThePlan` fails, `HasPlan("free") = false, want true`.
   - Green: `HasPlan` looks the plan up in the catalog, and `Provider` is asserted to be a `ports.PlanCatalog`.
   - Refactor: `Provider`'s comment names both ports; no code change. Its `PlanOf` comment ("will store the answers instead") stays for task 9.
9. [x] `plans.StoredPlanOf` reads a `Provider.PlanOf` from a `WorkspacePlanRepository`: the stored plan.
   - Red: `TestStoredPlanOf_AnswersTheWorkspacesStoredPlan` fails, `PlanOf = "", want "pro"`.
   - Green: `StoredPlanOf` returns the repository's `PlanOf`; `fallback` stays unread until 9b.
   - Refactor: `Provider.PlanOf`'s comment now names `StoredPlanOf`; no code change.
9b. [x] `StoredPlanOf` answers the default for a workspace with no stored plan.
   - Red: `TestStoredPlanOf_AnswersTheDefaultForAWorkspaceWithNoStoredPlan` fails, `PlanOf: not found`.
   - Green: `StoredPlanOf` answers `fallback` on any repository error; 9c narrows that to `ErrNotFound`.
   - Refactor: `StoredPlanOf`'s comment names the fallback; no code change. Defaulting on every error stays for 9c to demand.
9c. [x] `StoredPlanOf` passes on any other error, never defaulting.
   - Red: `TestStoredPlanOf_PassesOnAnyOtherErrorInsteadOfTheDefault` fails, `PlanOf = "free", <nil>, want the repository's error`.
   - Green: `StoredPlanOf` answers `fallback` only on `ErrNotFound`, and returns any other error.
   - Refactor: `StoredPlanOf`'s comment says other errors are passed on; no code change.

Wiring:
10. [x] `bootstrap.NewBilling` refuses in the local edition with `ErrLocalPlanIsFixed`, opening no storage.
   - Red: `TestNewBilling_RefusesInTheLocalEditionWithoutOpeningStorage` fails, `err = <nil>, want ErrLocalPlanIsFixed`.
   - Green: `NewBilling` always returns `ErrLocalPlanIsFixed`; 10b adds `--server`.
   - Refactor: `NewBilling`'s and `ErrLocalPlanIsFixed`'s comments say why it refuses; no code change. The ignored `cfg` stays for 10b to demand.
10b. [x] `bootstrap.NewBilling` refuses `--server`.
   - Red: `TestNewBilling_RefusesOverAServer` fails, `err = the local edition's plan is set when the app is built, want ErrNoPlanChangesOverAServer`.
   - Green: `NewBilling` returns `ErrNoPlanChangesOverAServer` when `cfg.ServerURL` is set, before the local refusal.
   - Refactor: `NewBilling`'s comment names the `--server` refusal; no code change.
11. [x] `starter admin set-plan WORKSPACE PLAN` reports `bootstrap.NewBilling`'s refusal as an error.
   - Red: `TestAdminSetPlanReportsThatBillingRefused` fails, `admin set-plan = unknown command "admin" for "starter", want the refusal`.
   - Green: `admin set-plan` loads the settings and returns `newBilling`'s error; `main` passes `bootstrap.NewBilling`.
   - Refactor: the `cli` package comment names the driving ports, not just `ItemService`, and `admin`'s help is plainer; no code change. The ignored settings error stays for 11d.
11b. [x] `admin set-plan` calls `ChangePlan` and says which plan the workspace is now on.
   - Red: `TestAdminSetPlanMovesTheWorkspaceAndSaysWhichPlanItIsOn` fails, `ChangePlan("", ""), want ("ws-a", "pro")`.
   - Green: `admin set-plan` calls `ChangePlan` with its two arguments and prints the new plan; `ChangePlan`'s error is ignored until 11c, the arguments unchecked until 11e.
   - Refactor: the `admin` tests share `runAdmin`; no production change. The ignored `ChangePlan` error stays for 11c, the unchecked arguments for 11e.
11c. [x] `admin set-plan` reports `ChangePlan`'s refusal, and doesn't say the workspace moved.
   - Red: `TestAdminSetPlanReportsChangePlansRefusalAndDoesNotSayTheWorkspaceMoved` fails, `admin set-plan = <nil>, want the refusal`.
   - Green: `admin set-plan` returns `ChangePlan`'s error before printing.
   - Refactor: the `admin` tests share `usingBilling`, replacing two copies of the factory; no production change. The refusing factory stays inline, its only use.
11d. [x] `admin set-plan` reports settings it can't load.
   - Red: `TestAdminSetPlanReportsSettingsItCannotLoadAndChangesNothing` fails, `admin set-plan = <nil>, want the settings error naming "bogus"`, and `ChangePlan("ws-a", "pro"), want no call`.
   - Green: `admin set-plan` returns `loadConfig`'s error before building billing.
   - Refactor: nothing worth changing; the error check is the same three lines as its neighbours, and the test uses the shared helpers.
11e. [x] `admin set-plan` refuses anything but a workspace and a plan.
   - Red: `TestAdminSetPlanRefusesAnythingButAWorkspaceAndAPlan` fails, three arguments `= <nil>, want a usage error`, and fewer panic with `index out of range`.
   - Green: `set-plan` declares `cobra.ExactArgs(2)`.
   - Refactor: nothing worth changing; `ExactArgs` matches the `item` commands, and the test uses the shared helpers.
11f. [x] `admin set-plan --server` reaches `ErrNoPlanChangesOverAServer`, through the real `bootstrap.NewBilling`.
   - No red: `TestAdminSetPlanOverAServerIsRefused` passed straight away, because `loadConfig` already applies `--server` and task 10b refuses it. Kept as a guard: dropping the flag's override fails it, `admin set-plan --server = the local edition's plan is set when the app is built, want ErrNoPlanChangesOverAServer`. No green.
   - Refactor: nothing worth changing; the test is one call through the shared `runAdmin`.

Acceptance:
12. [x] Harness: each driver's `WorkspacePlanRepository`, `plans` reading it with `bootstrap.LocalPlan` as the default, and `h.billing`. "My workspace is on / moves to" go through `ChangePlan`; drop `testPlans.assign`. Bind "an admin moves my workspace". `features/plans.feature` and `plan-features.feature` must stay green. Scenario "An admin moves a workspace to another plan".
   - Red: with the step stubbed, the scenario fails on memory, jsonfile, sqlite and http, `plan = "free", want "pro"`.
   - Green: `testPlans` reads each driver's `WorkspacePlanRepository` through `plans.StoredPlanOf` and is the `PlanCatalog` of a new `h.billing`; the steps call `ChangePlan`; `assign` is gone.
   - Refactor: the step function is `moveMyWorkspaceToThePlan`, and `Provider.PlanOf`'s comment no longer mentions a map; no code change.
12b. [ ] Bind "an admin tries to move my workspace". Scenario "Moving a workspace to a plan that doesn't exist is refused".
12c. [ ] Bind "workspace "X" is on", "an admin moves workspace "X"" and ""X"'s plan is". Scenario "Moving one workspace leaves the others on their plans".
13. [ ] Desktop steps: "I try to move my workspace to the … plan from the command line" runs the CLI in-process, and "I am told the local edition's plan is set when the app is built" checks `ErrLocalPlanIsFixed`.
14. [ ] Terminal step: "I ask the command line to move the workspace … to the … plan" runs `starter admin set-plan`.

Finish:
15. [ ] Docs: `docs/billing.md`; ADR 5's status; `.claude/rules/workspaces-and-plans.md` (ChangePlan, the stored plan, the local plan staying fixed); CLAUDE.md's architecture block, if `plans` or `storage` lines need it; `docs/workspaces.md` if its checklist is affected.
16. [ ] `task check`, `task test:init`, `task e2e:terminal`; remove the `@wip` tags; mark step 4 done in the plan; delete this file.

## Changes to the plan

- 2026-11-05: task 5 split into 5, 5b, 5c and 5d, one contract behaviour per test.
- 2026-11-05: task 6 split into 6 and 6b: the contract, then surviving a reopen.
- 2026-11-06: task 7 split into 7 and 7b: the contract, then surviving a reopen.
- 2026-11-06: task 9 split into 9, 9b and 9c: the stored plan, the default, and other errors.
- 2026-11-06: task 10 split into 10 and 10b: the local edition, then `--server`.
- 2026-11-06: the "not over a server yet" error is `bootstrap.ErrNoPlanChangesOverAServer`.
- 2026-11-06: task 11 split into 11 to 11f, one behaviour per test. The refusal comes first so that the real command, which always refuses, never reaches a nil service between commits.
- 2026-11-06: `cli.NewRootCmd` takes a `BillingFactory`, like its `AppFactory`, because `bootstrap.NewBilling` refuses every copy and tests must substitute it to reach a successful change.
- 2026-11-06: task 12 split into 12, 12b and 12c, one `@wip` scenario each. Their red is the scenario itself, with the step stubbed so it fails on its assertion rather than as undefined.
- 2026-11-05: `task test:init` is expected to fail until task 16 removes the `@wip` tags: "The new project is healthy" (both examples) runs `go test ./test/...` with no `~@wip`, so the undefined `@wip` steps fail under Strict. Agreed: in tasks 6, 7 and 12 it counts as passing if only those two fail.
