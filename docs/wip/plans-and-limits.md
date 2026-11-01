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
5. [ ] **ports + app, `EntitlementsService.CurrentEntitlements`:** returns the current workspace's entitlements from `ports.Entitlements`; with no workspace, `ErrNoWorkspace` and the port isn't asked. Add a `testsupport` fake for `ports.Entitlements`.
6. [ ] **plans adapter, parse:** a plans document gives each plan's entitlements (name, limits with name, max and `WhenOver`).
7. [ ] **plans adapter, comments:** full-line `//` comments are ignored.
8. [ ] **plans adapter, bad config:** invalid JSON, an invalid limit (through `NewLimit`), and unknown fields are errors that name the plan and limit.
9. [ ] **plans adapter, `ports.Entitlements`:** `For` answers the plan the plan-of function names; an unknown plan name is an error.
10. [ ] **plans adapter, shipped file:** the embedded default loads and has a `local` plan with no limits.
11. [ ] **app, `CreateItem` at the limit** is refused with `LimitError` and saves nothing. `NewItemService` takes `ports.Entitlements`; existing tests use an unlimited fake.
12. [ ] **app, `RenameItem` over a `read-only` limit** is refused and saves nothing; over `block-new` it works.
13. [ ] **app, `DeleteItem` over a `read-only` limit** works (likely already passing; keep as a guard).
14. [ ] **app, entitlements errors:** when `ports.Entitlements` fails, `CreateItem` and `RenameItem` return the error and save nothing.
15. [ ] **API, `GET /api/v1/entitlements`:** `api/openapi.yaml` + `task gen`; the handler returns the current entitlements (`httpapi.Options.Entitlements`).
16. [ ] **httpclient, `CurrentEntitlements`** maps the response back to `domain.Entitlements`.
17. [ ] **API, limit errors:** `problem.go` maps `*LimitError` to 403 with limit, max and read-only (the item endpoints' responses in `openapi.yaml`, inside sample blocks).
18. [ ] **httpclient** maps that 403 back to `*domain.LimitError`.
19. [ ] **wiring:** `config.PlansFile` (`STARTER_PLANS_FILE`), `bootstrap.LocalPlan`, and `NewEmbedded` building the plans adapter for the local plan into `ItemService` and a new `App.Entitlements`; an unknown local plan or a bad file fails at start. `cmd/server`, `desktop.Open` and `mobile` pass `App.Entitlements` to `httpapi`.
20. [ ] **acceptance:** the harness's `entitlements` driving port, test catalog and plan map (all drivers); step definitions (`plan_steps.go`, sample; the plan-name step, template). Turns the `@wip` scenarios green.
21. [ ] **template hygiene:** `features/plans.feature` and `plan_steps.go` in `SampleFiles`; the items limit, its error responses and its app code inside sample blocks; `task test:init`. Add a starter step for template users (under the TODO.md item's "Steps the new file must give the developer using the template"): define your plans in the plans file, and pick the local plan with `-ldflags`.
22. [ ] **docs:** CLAUDE.md (an entitlements rule, `LimitError` in the domain errors, `internal/adapters/plans` in Architecture, a Gotcha for `LocalPlan` and `STARTER_PLANS_FILE`), README's configuration table, `features/README.md` (the `@plans` tag), ADR 5 (limits built; features in step 3), the editions plan's step 2.
23. [ ] **finish:** `task check`, remove the `@wip` tags, delete this file.

## Changes to the plan
