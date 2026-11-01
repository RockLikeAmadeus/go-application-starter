# WIP: plans and limits

Step 2 of [`docs/plans/editions-and-accounts.md`](../plans/editions-and-accounts.md).

- Spec: `SPEC.md` §4.9 "Plans and limits"
- Scenarios: [`features/plans.feature`](../../features/plans.feature) (`@plans`, sample), plus "is on the local plan" in [`features/desktop.feature`](../../features/desktop.feature) and [`features/mobile.feature`](../../features/mobile.feature) (template)
- Decision record: [ADR 5](../adr/0005-hosted-and-local-editions.md), "Features and limits come from plans"

## Decisions

- **Limits only in this step.** On/off features (and show-locked or hide) come with their first user, CSV export, in step 3. Nothing in this step demands them, so writing them now would be code without a failing test.
- **Over-limit behaviour, per limit:**
  - `block-new`: at or over the limit, creating is refused; changing and deleting work.
  - `read-only`: creating is refused at or over the limit. Over it (count > max), changing is refused too. Deleting always works, so a workspace can get back within the limit itself, and then changing works again.
  - A plan without a given limit has no limit on that thing.
- **Errors:** one domain error, `*domain.LimitError{Limit, Max, ReadOnly}`. "the limit of 3 items is reached" (create) or "items are read only while over the limit of 3" (change). HTTP: 403 with a problem body carrying the limit, max and read-only flag; `httpclient` maps it back.
- **Domain shapes (template):** `domain.Entitlements{Plan, Limits map[string]Limit}`, `domain.Limit{Name, Max, WhenOver}`, `domain.OverLimit` (`BlockNew`, `ReadOnly`). The rules sit on `Entitlements`: `CheckCreate(limit, count)` and `CheckChange(limit, count)`, both nil when the plan has no such limit. `domain.NewLimit` validates (max ≥ 0, a known `WhenOver`).
- **Ports (template):**
  - driven `ports.Entitlements`: `For(ctx, workspace) (domain.Entitlements, error)`;
  - driving `ports.EntitlementsService`: `CurrentEntitlements(ctx)`, which resolves the workspace first (fail closed, as every use case).
- **Config adapter, `internal/adapters/plans` (template):**
  - **Format:** JSON that allows full-line `//` comments (`.jsonc`, as `tsconfig.json`). Stdlib `encoding/json` after dropping those lines, so no new dependency. The comments let the file carry `sample:start`/`sample:end` markers: `--remove-sample` strips the items limit, which plain JSON couldn't do. The sample entry is the only one in its `limits` object, so stripping it leaves valid JSON.
  - **Shape:** `{"plans": {"<name>": {"limits": {"<limit>": {"max": 3, "whenOver": "block-new"}}}}}`.
  - **Shipped file, embedded:** `local` (no limits), `free` (items: 3, `block-new`, inside a sample block), `pro` (no limits). `STARTER_PLANS_FILE` replaces it with a file.
  - **Which plan a workspace is on:** the adapter takes a function from workspace to plan name. The local edition's answers `bootstrap.LocalPlan` for every workspace. The acceptance harness answers from a map its steps set. Step 4 (`ChangePlan`) replaces this with stored assignments.
- **Local plan chosen at build time:** `var LocalPlan = "local"` in `internal/bootstrap`, set with `-ldflags "-X …/internal/bootstrap.LocalPlan=<name>"`. `NewEmbedded` fails fast if the catalog has no such plan.
- **Counting:** `CreateItem` and `RenameItem` count with `FindAll`, so no new repository query (and no new isolation test). A product that needs a cheaper count adds a `Count` query with its own isolation case (`docs/workspaces.md`).
- **Not in this step:** usage counts in the API ("4 of 5"); a hosted default plan (step 4 or 8 sets plans for hosted workspaces); UI for limits (the SPA, CLI and TUI already show error messages).
- **Acceptance:**
  - `@plans` scenarios run on memory, jsonfile and http, where the harness builds the core: it owns a test catalog and plan map that the steps fill in, and the http driver's server gets the same. Desktop and mobile build through `bootstrap`, so they run only "is on the local plan".
  - "my workspace moves to the X plan" changes the harness's map for now; step 4 can make it call `ChangePlan`.
  - Steps for the item limit go in a new sample file, `test/acceptance/plan_steps.go`. The template step "my workspace's plan is X" goes in a template file.

## Tasks

1. [x] **domain, limit checks:** `Entitlements.CheckCreate` refuses at or over a limit with `*LimitError`, allows below it, and allows anything when the plan has no such limit (table test).
    - **Red:** `TestEntitlements_CheckCreate` in `domain/entitlements_test.go`: below, at, over, at a read-only limit, and a plan without the limit. The three refusals failed with `nil`. Compile stubs: `entitlements.go` (`OverLimit`, `Limit`, `Entitlements`) and `LimitError` in `errors.go`, with an empty `Error()` (task 3).
    - **Green:** `CheckCreate` looks the limit up; missing or `count < Max` is allowed, otherwise `&LimitError{Limit, Max}`.
    - **Refactor:** doc comments on the `OverLimit` constants and `CheckCreate`. Left `Limit.Name` duplicating its map key: the config adapter (task 6) fills both, and callers with a `Limit` in hand will need its name. `LimitError.Error()` stays empty until task 3 tests its messages.
2. [x] **domain, change checks:** `CheckChange` refuses only over a `read-only` limit (`ReadOnly: true`); allows at a `read-only` limit, over a `block-new` one, and with no limit (table test).
    - **Red:** `TestEntitlements_CheckChange`: over and at a read-only limit, over a block-new one, and a plan without the limit. Only "over a read-only limit" failed, with `nil`, against a stub that always allowed.
    - **Green:** `CheckChange` refuses with `&LimitError{Limit, Max, ReadOnly: true}` only when the limit exists, is `ReadOnly`, and `count > Max`.
    - **Refactor:** a doc comment on `CheckChange`, and the two tests' identical assertions moved into an `assertLimitError` helper. Left the two lookups in `CheckCreate` and `CheckChange` as they are: two lines each, and they read more plainly than a shared helper would.
3. [x] **domain, `LimitError` messages** for the two cases.
    - **Red:** `TestLimitError_Message`: "the limit of 3 items is reached" and "items are read only while over the limit of 3". Both failed with `""`.
    - **Green:** `Error()` formats one or the other on `ReadOnly`, with the limit's name as the thing counted.
    - **Refactor:** documented `LimitError`'s fields, including that a limit's name is the plural noun its messages use, so limits are named that way (`items`). Nothing else was worth changing.
4. [x] **domain, `NewLimit` validation:** negative max and an unknown `WhenOver` are refused.
    - **Red:** `TestNewLimit`: a valid limit is built as given; a negative max (field `max`), an unknown behaviour and none at all (field `whenOver`) are refused with `*ValidationError`, the domain's existing error for a broken rule. All four failed against a stub returning `Limit{}, nil`.
    - **Green:** `NewLimit` checks `max >= 0` and `whenOver` is `BlockNew` or `ReadOnly`, then builds the `Limit`.
    - **Refactor:** a doc comment on `NewLimit`: build limits through it, never as a literal, and its fields are the plans file's keys. Left the other domain types as they are: nothing else changed this cycle.
5. **ports + app, `EntitlementsService.CurrentEntitlements`** (split in two: one test per cycle):
   - [x] **5a.** It returns the current workspace's entitlements from `ports.Entitlements`.
     - **Red:** `TestCurrentEntitlements_AreTheCurrentWorkspacesPlan` (`app/entitlements_service_test.go`), with a test-local `plansByWorkspace` fake answering two workspaces differently. It failed with empty entitlements against a stub. Compile stubs: `ports.EntitlementsService` (driving) and `ports.Entitlements` (driven), and `app.EntitlementsService`.
     - **Green:** resolve the workspace and ask `ports.Entitlements.For` with it. The resolver's error is ignored for now: 5b's test demands the early return.
     - **Refactor:** doc comments on `NewEntitlementsService` and `CurrentEntitlements`, as `ItemService` has. The ports sit outside the sample block, beside `WorkspaceResolver`, since they're template. Nothing else was worth changing.
   - [x] **5b.** With no workspace, it returns `ErrNoWorkspace` and `ports.Entitlements` isn't asked.
     - **Red:** `TestCurrentEntitlements_WithoutAWorkspaceIsRefusedBeforeAskingThePlans`, with `testsupport.NoWorkspace` and a counting `unaskedPlans` fake (as `untouchedRepo` in the item tests). It failed: no error, and the plans asked once.
     - **Green:** return the resolver's error before asking the plans.
     - **Refactor:** nothing worth changing. The service reads like `ItemService`'s use cases, and its two test fakes each serve one test, as the item tests' do.
6. [x] **plans adapter, parse:** a plans document gives each plan's entitlements (name, limits with name, max and `WhenOver`).
    - **Red:** `TestParse_GivesEachPlansEntitlements` (`adapters/plans/plans_test.go`): a document with `free` (items: 3, read-only) and `pro` (no limits) gives a `plans.Catalog` (plan name → `domain.Entitlements`). It failed with an empty catalog against a stub.
    - **Green:** `Parse` unmarshals into a private `document` struct and builds each plan's entitlements. The JSON error is ignored and limits are built as literals: tasks 8 (bad config) and 8's `NewLimit` check demand otherwise.
    - **Refactor:** named the anonymous structs (`planRecord`, `limitRecord`), as jsonfile names its on-disk records, with a comment on why they're separate from the domain types; a doc comment on `Parse` showing the format. Left the ignored JSON error and the literal limits for task 8, whose tests demand them.
7. [x] **plans adapter, comments:** full-line `//` comments are ignored.
    - **Red:** `TestParse_IgnoresFullLineComments`: a document with `//` comment lines, including sample markers around the only limit, parses as the same plan without them. It failed with an empty catalog: the comments made the JSON invalid, and the error is still ignored (task 8).
    - **Green:** `Parse` empties every line whose first non-blank characters are `//`, keeping the line so JSON error positions still match the file, then unmarshals.
    - **Refactor:** moved the stripping into `stripComments`, so `Parse` reads as strip → decode → build. `Parse`'s doc comment now says full-line comments are allowed (for explanations and sample markers) and end-of-line ones aren't.
8. [x] **plans adapter, bad config:** invalid JSON, an invalid limit (through `NewLimit`), and unknown fields are errors that name the plan and limit.
    - **Red:** `TestParse_RefusesABadPlansFile`: invalid JSON, a negative max, an unknown `whenOver`, and an unknown field (`maximum`). Each must fail, with the plan and limit named for an invalid limit, and the field named for an unknown one. All four were accepted.
    - **Green:** decode with `DisallowUnknownFields`, wrapping the error in "plans file: …"; build each limit with `domain.NewLimit`, wrapping its error with the plan and limit. The standard library's unknown-field error names the field but not where it is; good enough, since field names are few.
    - **Refactor:** the per-plan loop moved into `planRecord.entitlements(name)`, so `Parse` reads as strip → decode → build each plan. Errors still read `plans file: plan "free", limit "items": max must not be negative`. Left alone: a second JSON value after the first is still accepted. Refusing it would be new behaviour, so it needs its own test; it isn't worth one in a template, where the file is the developer's own.
9. **plans adapter, `ports.Entitlements`** (split in two, as task 5):
   - [x] **9a.** `For` answers the plan that the plan-of function names for the workspace.
     - **Red:** `TestEntitlements_AnswersThePlanTheWorkspaceIsOn`: two workspaces on two plans; `For(ws-a)` must give `free`'s entitlements. It failed with empty entitlements against a stub. Compile stub: `plans.Entitlements{Catalog, PlanOf}`, asserted to satisfy `ports.Entitlements`.
     - **Green:** `For` asks `PlanOf` and looks the plan up in the catalog. `PlanOf`'s error and a missing plan are ignored until 9b.
     - **Refactor:** doc comments on the fields and `For`, saying who supplies `PlanOf`: the local edition (its one plan), the acceptance tests (a map), and later `ChangePlan`'s stored plans. Nothing else was worth changing.
   - [x] **9b.** An unknown plan name, or an error from `PlanOf`, is an error.
     - **Red:** `TestEntitlements_RefusesAPlanItCantAnswer`: a plan the catalog doesn't have (`gold`) must give an error naming it; `PlanOf`'s error must come back as is. Both returned `nil`.
     - **Green:** `For` returns `PlanOf`'s error, then refuses a missing plan with `plan "gold" isn't in the plans file`.
     - **Refactor:** `For`'s doc comment now says how it fails. Nothing else: `PlanOf`'s error stays unwrapped, as use cases return the resolver's, since wrapping would add nothing a caller needs yet.
10. [x] **plans adapter, shipped file:** the embedded default loads and has a `local` plan with no limits.
    - **Red:** `TestDefault_HasALocalPlanWithNoLimits`: `plans.Default()` loads the shipped file without error, and it has a `local` plan with no limits. It failed against a stub returning an empty catalog.
    - **Green:** `internal/adapters/plans/plans.jsonc`, embedded with `go:embed`, holding only `local`; `Default` parses it. `free` and `pro` wait for task 21, where the sample limit and its markers are checked by `task test:init` (see the log).
    - **Refactor:** doc comments on `Default` and the embedded variable. `STARTER_PLANS_FILE` isn't mentioned yet: task 19 adds it. The `.jsonc` keeps tabs: Prettier only checks `web/`.
11. [x] **app, `CreateItem` at the limit** is refused with `LimitError` and saves nothing. `NewItemService` takes `ports.Entitlements`; existing tests use an unlimited fake.
    - **Red:** `TestCreateItem_AtTheLimitIsRefusedAndSavesNothing`: on a plan allowing 3 items (block-new), a fourth must be refused with `*LimitError{Limit: "items", Max: 3}`, leaving 3 items. It failed with `nil`: the fourth was created. To compile, `NewItemService` takes a `ports.Entitlements` after the resolver; the existing tests, the TUI and httpclient tests and the acceptance harness pass the new `testsupport.FixedEntitlements{}` (its zero value has no limits), and `bootstrap` passes the plans adapter on the shipped catalog with `"local"` for every workspace (see the log).
    - **Green:** `CreateItem` asks the entitlements for the workspace, counts its items with `FindAll`, and returns `CheckCreate("items", count)`'s error before building the item. Both lookups' errors are ignored until task 14 demands them.
    - **Refactor:** `bootstrap`'s comments now say where the entitlements come from. Left the `"items"` literal: it's used once; task 12 uses it again, which is when a shared name is worth it. Left `FixedEntitlements` as the one fake for both an unlimited and a limited plan, like `FixedWorkspace`.
12. [x] **app, `RenameItem` over a `read-only` limit** is refused and saves nothing; over `block-new` it works.
    - **Red:** `TestRenameItem_OverALimit`, a table test: four items made on an unlimited plan, then a second service on the same repository with a limit of 3. Over a read-only limit, renaming must be refused with `*LimitError{ReadOnly: true}` and leave the name; over a block-new one, it must work. The read-only case failed with `nil`. The block-new case passed from the start, as task 2's table did: it guards against checking renames with `CheckCreate`.
    - **Green:** `RenameItem` finds the item, then counts the workspace's items and returns `CheckChange("items", count)`'s error before renaming.
    - **Refactor:** the shared lookup moved into `itemUsage` (the entitlements and the item count) and the limit's name into the `itemsLimit` constant, in `item_service.go`. The tests' plan moved into `freePlan(whenOver)`. Lookup errors are still ignored: task 14 demands them, and `itemUsage` is where they'll go.
13. [x] **app, `DeleteItem` over a `read-only` limit** works (likely already passing; keep as a guard).
    - **Red:** `TestDeleteItem_OverAReadOnlyLimitWorks`: four items, then a read-only limit of 3; deleting one must work. It passed straight away, since `DeleteItem` has no limit check. Kept as a guard, at Alec's choice: it pins that deleting always works, which is how a workspace gets back within a read-only limit.
    - **Green:** nothing to do.
    - **Refactor:** the rename and delete tests' shared setup (four items made on an unlimited plan, then a service on `freePlan(whenOver)` over the same repository) moved into `overTheFreePlan(t, whenOver)`.
14. [x] **app, entitlements errors:** when `ports.Entitlements` fails, `CreateItem` and `RenameItem` return the error and save nothing.
    - **Red:** `TestUseCases_WhenThePlanCantBeLookedUpAreRefusedAndSaveNothing`, a table test over `CreateItem` and `RenameItem` with a test-local `unavailablePlans` fake: each must return its error and leave the one existing item as it was. Both failed with `nil`: the error was dropped and the empty plan let them through.
    - **Green:** `itemUsage` returns `For`'s error, and both use cases return it before checking the limit. Unwrapped, as `CreateItem` returns the resolver's.
    - **Refactor:** the setup shared by three tests (items made on an unlimited plan, then a service on another plan) became `serviceWithItems(t, entitlements, names...)`; `overTheFreePlan` now calls it. The count's error is task 14b's (see the log).
14b. [x] **app, count errors:** when counting the items (`FindAll`) fails, `CreateItem` and `RenameItem` return the error, wrapped as `count items: …`, and save nothing.
    - **Red:** `TestUseCases_WhenTheItemsCantBeCountedAreRefusedAndSaveNothing`, a table test over both use cases with `uncountableRepo`, which wraps a memory repository and fails only `FindAll`. Each must return `count items: <the error>` and leave the existing item as it was, checked on the memory repository underneath. Both failed with `nil`.
    - **Green:** `itemUsage` returns `FindAll`'s error wrapped as `count items: …`, as storage errors are wrapped elsewhere (`save item: …`).
    - **Refactor:** the two error tests' table of calls became the shared `limitedUseCases`. Left the new test's setup apart from `serviceWithItems`: it needs the memory repository underneath, which the helper hides.
15. [x] **API, `GET /api/v1/entitlements`:** `api/openapi.yaml` + `task gen`; the handler returns the current entitlements (`httpapi.Options.Entitlements`).
    - **Red:** `TestEntitlementsAreTheCurrentWorkspacesPlanAndLimits` (`httpapi/server_test.go`), with a test-local `fixedEntitlements` driving port: the body must decode to `{"plan":"free","limits":{"items":{"max":3,"whenOver":"block-new"}}}`. It failed with `{"limits":null,"plan":""}`. To compile: the endpoint and its `Entitlements` and `Limit` schemas in `openapi.yaml` (template, outside the sample blocks; limits keyed by name and `whenOver` an enum, as in the plans file), `task gen`, `Options.Entitlements`, and an empty handler in `httpapi/entitlements.go`.
    - **Green:** `Handlers` keeps the driving port, and `GetEntitlements` maps the domain entitlements to the generated types. Its error goes back as every handler's does, to the problem mapping already tested.
    - **Refactor:** the mapping moved into `toEntitlementsDTO`, beside the items' `toDTO`; a doc comment on `Options.Entitlements`. Nothing serves it yet: `cmd/server`, `desktop.Open` and mobile pass it in task 19.
16. [ ] **httpclient, `CurrentEntitlements`** maps the response back to `domain.Entitlements`.
17. [ ] **API, limit errors:** `problem.go` maps `*LimitError` to 403 with limit, max and read-only (the item endpoints' responses in `openapi.yaml`, inside sample blocks).
18. [ ] **httpclient** maps that 403 back to `*domain.LimitError`.
19. [ ] **wiring:** `config.PlansFile` (`STARTER_PLANS_FILE`), `bootstrap.LocalPlan`, and `NewEmbedded` building the plans adapter for the local plan into `ItemService` and a new `App.Entitlements`; an unknown local plan or a bad file fails at start. `cmd/server`, `desktop.Open` and `mobile` pass `App.Entitlements` to `httpapi`.
20. [ ] **acceptance:** the harness's `entitlements` driving port, test catalog and plan map (all drivers); step definitions (`plan_steps.go`, sample; the plan-name step, template). Turns the `@wip` scenarios green.
21. [ ] **template hygiene:** `features/plans.feature` and `plan_steps.go` in `SampleFiles`; the items limit, its error responses and its app code inside sample blocks; `task test:init`. Add a starter step for template users (under the TODO.md item's "Steps the new file must give the developer using the template"): define your plans in the plans file, and pick the local plan with `-ldflags`.
22. [ ] **docs:** CLAUDE.md (an entitlements rule, `LimitError` in the domain errors, `internal/adapters/plans` in Architecture, a Gotcha for `LocalPlan` and `STARTER_PLANS_FILE`), README's configuration table, `features/README.md` (the `@plans` tag), ADR 5 (limits built; features in step 3), the editions plan's step 2.
23. [ ] **finish:** `task check`, remove the `@wip` tags, delete this file.

## Changes to the plan

- 2026-11-01: task 5 split into 5a (answers the workspace's entitlements) and 5b (fails closed), one test each. The `testsupport` fake for `ports.Entitlements` moves to task 11, where `ItemService` first needs one; 5a's fake is local to its test.
- 2026-11-01: the shipped plans file holds only `local` after task 10, which is all its test demands. `free` (with the sample item limit inside sample markers) and `pro` are added in task 21, which checks `--remove-sample` strips the limit and leaves valid JSON.
- 2026-11-01: task 11 wires the plans adapter into `bootstrap.NewEmbedded` early, because `NewItemService` now needs a `ports.Entitlements` and production code can't use a `testsupport` fake: the shipped catalog, with `"local"` as a literal for every workspace, and `Default`'s error returned. Task 19 still adds `LocalPlan`, `STARTER_PLANS_FILE`, failing fast on an unknown plan, and `App.Entitlements`.
- 2026-11-01: task 14b added, at Alec's choice: `itemUsage` still ignored `FindAll`'s error, so a storage failure would count 0 items and let a create through the limit. It needs its own test.
- 2026-11-01: task 9 split the same way: 9a (answers the workspace's plan) and 9b (an unknown plan, or `PlanOf` failing, is an error).
