# WIP: features that plans turn on or off (editions plan, step 3)

- Plan: [`docs/plans/editions-and-accounts.md`](../plans/editions-and-accounts.md), step 3.
- Spec: [SPEC.md §4.9](../../SPEC.md) "Plans, features and limits" (rules: features).
- Scenarios: [`features/plan-features.feature`](../../features/plan-features.feature) (`@plans`: memory, jsonfile, sqlite, http); the local plan's export in [`features/desktop.feature`](../../features/desktop.feature) and [`features/mobile.feature`](../../features/mobile.feature) (sample blocks).
- Journeys: [`web/e2e/features/items.feature`](../../web/e2e/features/items.feature) "Export the items"; [`test/terminal/features/items.feature`](../../test/terminal/features/items.feature) "Export the items with the command line" and the locked/hidden help outline.
- ADR: [ADR 5](../adr/0005-hosted-and-local-editions.md) "Features and limits come from plans".

## Decisions (agreed 2026-11-04)
- **The machinery is what counts, not the export.** Everything that gates a feature is template code and survives `--remove-sample`: the domain types, the plans file's `features`, the API fields and problem, the httpclient mapping, and the front-end helpers (web, CLI, TUI), each with its own unit tests. CSV export of items is only the thin sample that runs it end to end, all inside sample files or blocks.
- **Shipped plans:** the `local` and `pro` plans have `item-export`, while `free` shows it locked. It's in a sample block of `plans.jsonc`; the `features` section itself is template.
- **The plans file:** a top-level `"features": {"<feature>": {"name": "Item export", "whenOff": "show-locked" | "hide"}}`, and each plan lists the features it turns on: `"features": ["item-export"]`. A plan naming an unknown feature, a missing or blank `name`, or a bad `whenOff` fails when the file is read. A feature a plan doesn't list is off.
- **Feature keys and display names:** the key (`item-export`) is what code and the API use. The required `name` ("Item export") is what people read, so template users choose its wording and capitals. `FeatureError` and the hints use the name ("Item export isn't on your plan").
- **Fail closed in the core:** `CheckFeature` refuses a feature that's off, and also one the catalog doesn't know. The use case refuses when the entitlements can't be looked up. The front ends only *display* gating, so when they can't read the entitlements they avoid jarring changes rather than hide things (see Web and CLI below).
- **The upgrade hint** names the plans that have the feature, worked out from the catalog when it's looked up (so the harness's `define` stays right). It leaves out the local edition's plan (`bootstrap.LocalPlan`). A hidden feature has no hint. **The local edition gives no hints:** bootstrap wires the plans adapter with hints off, and the acceptance harness turns them on. A locked feature with no hint says only that it isn't available.
- **Domain shape:** `Entitlements.Features map[string]Feature`, where `Feature{Key, Name, IsOn, WhenOff, AvailableOn}` is built with `NewFeature`, `FeatureOffBehavior` is `show-locked` or `hide`, and `CheckFeature(key)` returns a `*FeatureError{Feature, Name, AvailableOn}`. That's a 403 problem with `feature`, `name` and `availableOn`, and `httpclient` maps it back.
- **Sample use case:** `ItemService.ExportItems(ctx) ([]byte, error)` checks `item-export`, then encodes the workspace's items with a domain function: CSV with the header `name,createdAt` (RFC 3339), oldest first. Its endpoint is `GET /api/v1/items/export` (`text/csv`).
- **Front ends:**
  - **CLI:** a gated command carries the feature key as a Cobra annotation. The help lists gated commands and notes the locked ones as disabled on the current plan, with the hint ("disabled on your plan: available on the pro plan"). A `hide` feature isn't listed, honouring the template user's choice (confirmed 2026-11-04). The help **doesn't open the app**: that would create the data folder, run migrations and fail on a locked folder. It reads only the entitlements, through a light bootstrap path (the local edition: the plans file and `LocalPlan`, no storage; `--server`: `GET /api/v1/entitlements`). If that fails, gated commands are listed with no note: the core still refuses them. Running a gated command that's off fails through the core's `FeatureError`. The sample is `starter item export` (CSV on stdout).
  - **TUI:** a template helper builds a gated key's help (empty when hidden, marked locked with the hint when locked), with its own test. The TUI gets no sample export, because it can't download a file.
  - **Web:** a template `useFeature(key)` hook (on / locked + hint / hidden), and a gate component that renders its children, a locked version with the hint, or nothing. The sample is an Export button on the Items page that downloads the CSV through `apiFetch`, so it goes over the bridge on Android.
  - **No pop-in:** `Layout` starts the entitlements request when the app starts, and keeps the answer for the session (a long `staleTime`). A page with gated controls includes the entitlements in its loading state, so its first paint already has every control in its final state. The gate component also has a placeholder of the same size (a disabled control) for anything rendered before the entitlements arrive. If they fail to load, the feature shows locked with no hint (the server refuses anyway).
- **Sample files for `task init`:** `features/plan-features.feature` and its steps file go in `scripts/initproject/manifest.go`'s `SampleFiles`. The desktop and mobile scenarios are in sample blocks.

## Tasks
1. [x] domain: `NewFeature` builds a feature, and refuses a blank `name` (Field `name`) and a `whenOff` other than `show-locked` or `hide` (Field `whenOff`) with a `*ValidationError`.
   - Red: `TestNewFeature` (valid + 4 invalid cases) failed on the stub's zero `Feature` and nil error.
   - Green: `NewFeature(key, name, whenOff)` checks the trimmed name and `whenOff`; `IsOn`/`AvailableOn` left for task 2 and the adapter.
   - Refactor: doc comments on `FeatureOffBehavior`, its constants, `Feature` and `NewFeature`, matching `NewLimit`'s; the name isn't trimmed when stored (no test asks).
2. [x] domain: `Entitlements.CheckFeature`: on → nil; off and locked → `*FeatureError` with `AvailableOn`; off and hidden → `*FeatureError` without; unknown → `*FeatureError` without. Test `FeatureError.Error()`'s text, with a hint and without.
   - Red: `TestEntitlements_CheckFeature` (6 cases, each checking the error and its text) failed on the stub's nil; `TestNewFeature` now uses `reflect.DeepEqual` (`AvailableOn` makes `Feature` incomparable).
   - Green: `CheckFeature` refuses off and unknown features (an unknown one's `Name` is its key); `Error()` is "<Name> isn't on your plan", plus ", and is available on the pro plan" / "the pro or team plans".
   - Refactor: one `FeatureError` built, hint added only for `ShowLocked`; doc comments on `CheckFeature`, `FeatureError` and the new `Feature` fields.
3. [x] plans adapter: `Parse` reads `features` (key, `name`, `whenOff`) and each plan's `features` list into `Entitlements.Features` (every catalog feature, `IsOn` per plan). It refuses an unknown feature, a missing `name` and a bad `whenOff`.
   - Red: `TestParse_ReadsFeatures` (catalog + 3 refusals) failed on `unknown field "features"`; the older Parse tests now want an empty `Features` map, as `Limits` is.
   - Green: `Parse` builds the file's `features` with `NewFeature`, gives each plan a copy with `IsOn` set for those it lists, and refuses an unlisted key.
   - Refactor: `buildFeatures` helper, `maps.Clone` for each plan's copy, `Parse`'s doc shows the features; errors read `feature "x": …` like limits'.
4. [x] plans adapter: `For` fills `AvailableOn` (sorted plan names that have the feature, without the local plan) when hints are on, and leaves it empty when they're off.
   - Red: `TestEntitlements_FillsTheUpgradeHintWhenHintsAreOn` (on / off) failed with `AvailableOn = nil, want {"pro", "team"}`; the adapter got stub `ShouldGiveHints` and `LocalPlan` fields.
   - Green: with `ShouldGiveHints`, `For` gives each of the plan's features (in a cloned map) the sorted catalog plans that have it on, leaving out `LocalPlan`.
   - Refactor: `withHints` (copies the map) and `plansWith` helpers; doc comments on `ShouldGiveHints`, `LocalPlan` and `For`. Bootstrap and the harness are left to tasks 5 and 6.
5. [x] bootstrap (template): a light entitlements-only path that opens no storage, for the CLI's help; the local edition wires no hints (a locked feature has no `AvailableOn`).
   - Red: `TestNewEntitlements_ReadsTheLocalPlanWithoutStorageOrHints` (a plans file with `export` locked on local; no data folder made) failed on the stub's nil service.
   - Green: `NewEntitlementsService(cfg)` wraps `localEntitlementsProvider` in `app.NewEntitlementsService` with `workspace.Local{}`; no storage opened.
   - Refactor: doc comments only (`NewEntitlementsService` says what it skips; `localEntitlementsProvider` says it gives no hints). The `--server` branch waits for task 13, whose help needs it.
5b. [x] plans.jsonc (+ bootstrap test in a sample block): the template's `features` section; `item-export` ("Item export", sample block) on for local and pro, locked for free.
   - Red: `TestNewEntitlements_ShippedPlansHaveItemExport` (local / pro / free, through `LocalPlan`) failed: no `item-export` in the shipped file.
   - Green: `plans.jsonc` gets a `features` section and each plan a `features` list; `item-export` and the plans' entries are in sample blocks.
   - Refactor: nothing; the file's explanatory comments are task 19's.
   - `task test:init` is red only on the `@wip` plan-features scenarios (undefined steps) in the generated project; expected until tasks 6, 9 and 19 (user, 2026-11-04).
6. [x] acceptance harness: hints on in `testPlans`; steps for "has / doesn't have item export", "is shown locked / hidden when off" and "item export is … for my workspace" (in a sample steps file).
   - Red: the "can see whether its plan has the feature" outline, with the new steps (`plan_feature_steps.go`, sample, with `testPlans.setFeature` / `setWhenOff`): the locked row failed with `AvailableOn = [], want ["pro"]`.
   - Green: `testPlans.For` sets `ShouldGiveHints` and `LocalPlan`; the outline passes on memory, jsonfile and sqlite (http, desktop and mobile wait for the API's features, tasks 10–12). `plan-features.feature` and its steps file are in `SampleFiles` now (from task 19).
   - Refactor: `withFeature` (copy a plan's features with one replaced) shared by `setFeature` and `setWhenOff`; the helpers moved from `plans.go` to the sample steps file, since the template without the sample left them unused (golangci-lint).
   - `task test:init`: the sample-removed project passes; the kept-sample one fails only on the `@wip` export steps, still undefined until task 9.
7. [x] domain (sample): encode items as CSV, with header `name,createdAt`, RFC 3339 and CSV quoting.
   - Red: `TestItemsCSV` (a plain name and one with a comma and quotes) failed on the stub's empty output.
   - Green: `ItemsCSV(items) []byte` writes the header and a row per item with `encoding/csv` and `time.RFC3339`, in the order given (the app sorts, task 8).
   - Refactor: a doc comment only (order given, why write errors are ignored); the function is small and in the sample file `item.go`.
8. [x] app (sample): `ExportItems` returns `ErrNoWorkspace` first (in the fail-closed table), refuses with `FeatureError` when off, refuses when the entitlements lookup fails, and returns the items' CSV, oldest first.
   - Red: `TestExportItems` (on → CSV / off → `FeatureError` / lookup fails) and an `ExportItems` row in the no-workspace table failed on the stub's nil, nil.
   - Green: `ExportItems` (on `ItemService`, not the port yet: task 9's drivers need it there) resolves the workspace, looks up the entitlements, checks `item-export`, and encodes `FindAll`'s items, already oldest first by the port's contract.
   - Refactor: an `itemExport` constant beside `itemsLimit`, and a doc comment; the lookup-then-check isn't a helper yet, with one gated use case.
9. [x] acceptance (sample): export steps ("I export the items", "I try to…", "the export lists", "I am told item export isn't on my plan…"). The memory, jsonfile and sqlite drivers go green.
   - Red: none. The export steps (`plan_feature_steps.go`, with `world.export` in a sample block) passed straight away on memory, jsonfile and sqlite, since task 8 built `ExportItems`; on http's zero-value stub they fail as they should (no header; no `FeatureError`). Kept as a guard.
   - Compile minimum: `ExportItems` on `ports.ItemService` (sample block), and a `nil, nil` stub in `httpclient` until task 12.
   - Refactor: only the errorlint fix (`%w`). `test:init`: the kept-sample project still fails only on the `@wip` export scenarios (http, desktop, mobile), as accepted under 5b.
10. [x] API: `openapi.yaml` gets `features` in `Entitlements` (by key: `name`, `on`, `whenOff`, `availableOn`; hidden ones too); `task gen`; httpapi serves them.
   - Red: `TestEntitlementsIncludeEveryFeatureByKey` (locked with a hint, on, hidden) failed with `features = <nil>`.
   - Green: `Entitlements.features` (required) and a `Feature` schema in `openapi.yaml`; `toEntitlementsDTO` maps each feature, a nil `AvailableOn` as `[]`. The old entitlements test now wants `"features":{}`.
   - Refactor: `toFeatureDTO` (keeps `availableOn` a list, never null), beside the limits loop.
10b. [x] API: the feature problem (403 with `feature`, `name` and `availableOn`); httpapi maps `FeatureError` to it.
   - Red: `TestAFeatureErrorIsAForbiddenProblemNamingTheFeature` (with a hint / without: `availableOn` `[]`) failed with status 500, want 403.
   - Green: `feature`, `name` and `availableOn` on `Problem` in `openapi.yaml`; `problemFor` maps `*FeatureError` to a 403 `feature` problem (`ProblemFeature`), a nil `AvailableOn` as `[]`.
   - Refactor: `availableOnDTO` (nil → `[]`) shared by `toFeatureDTO` and the problem. The plans rule's domain-error list gets `FeatureError` in task 12, once httpclient maps it back.
11. [x] API + httpapi (sample): `/items/export` (`text/csv`) in `openapi.yaml`, and the export handler.
   - Red: `TestExportServesTheItemsAsCSV` (new sample file `httpapi/items_test.go`) failed with status 404: `export` was read as an item ID.
   - Green: `GET /api/v1/items/export` (`exportItems`, `text/csv`) in `openapi.yaml`'s sample block, `task gen`; the handler serves `ExportItems`' bytes. The new test file is in `SampleFiles`.
   - Refactor: nothing; the handler is three lines like its neighbours. ServeMux's precedence keeps `export` from matching `{id}`, which the test guards.
12. [x] httpclient: reads the features in the entitlements (every field, a hint-less `availableOn` back to nil).
   - Red: `TestCurrentEntitlements_ReadsTheFeatures` (on with a hint, locked, hidden with none) failed with `Features = map[]`; the round-trip test now wants an empty `Features` map.
   - Green: `toEntitlements` maps each API feature to a `domain.Feature` (key, name, on, whenOff), an empty `availableOn` back to nil. The http driver's "can see whether its plan has the feature" outline now passes.
   - Refactor: `toFeature` helper with a doc comment (why `[]` becomes nil); the service's doc mentions features.
12b. [x] httpclient: maps the feature problem back to `*FeatureError` (`feature`, `name`, `availableOn`); the plans rule's domain-error list gets `FeatureError`.
   - Red: `TestAFeatureErrorComesBackAsAFeatureError` (with a hint / without) failed with `server error (403): Export isn't on your plan…`, not a `*FeatureError`.
   - Green: `toError` maps a `feature` problem to `*FeatureError` (`feature`, `name`, `availableOn`), an empty `availableOn` back to nil.
   - Refactor: `fromAvailableOn` (`[]` → nil) shared by `toFeature` and `toError`; `FeatureError` added to the plans rule's domain-error list.
12c. [x] httpclient (sample): `ExportItems` over `GET /items/export`. The http, desktop and mobile drivers go green.
   - Red: `TestExportItems_ReadsTheCSV` (a quoted name through a real server) failed with `ExportItems = ""`, the stub's nil.
   - Green: `ExportItems` calls `ExportItemsWithResponse`, returns the body on 200 and `toError` otherwise.
   - Refactor: `newClient(t, plan)` test helper shared with `TestRoundTrip`; a doc comment on `ExportItems`. Every `@wip` acceptance scenario now passes on every driver.
   - `task test:init`: the removed-sample project passes; the kept-sample one fails only on the `@wip` terminal journeys (undefined steps until task 14).
13. [x] CLI (template): the gated-command helper and the help: on → listed plainly; locked → listed as disabled on your plan, with the hint, or without one in the local edition; hidden → not listed; entitlements unreadable → listed with no note. The help opens no storage.
   - Red: `TestHelpNotesGatedCommandsByThePlan` (template `gated_test.go`, 5 cases, no data folder made) failed on locked, hidden and the server's hint: the stub `Gated` leaves the command listed plainly.
   - Green: `Gated` sets a `feature` annotation; the root's help func loads the config (`loadConfig`, split from `openApp`), reads `bootstrap.NewEntitlementsService` (now with a `--server` branch) and hides or notes gated children; `domain.PlansPhrase` split from `FeatureError.Error` for the hint.
   - Refactor: `note(cmd, feature)` pulled out of the help loop; doc comments on `Gated`, `noteGatedCommands`, `readEntitlements` and `NewEntitlements`. Also gofmt'd `plans.go` (left by 89b94ec; user, 2026-11-05).
14. [x] CLI (sample): `starter item export`, gated by `item-export`, prints the CSV on stdout.
   - Red: `TestExportPrintsTheItemsAsCSV` (a quoted name, jsonfile) failed: no `export` subcommand, so `item`'s help was printed.
   - Green: `item export` (`Gated("item-export", …)`) writes `ExportItems`' bytes to stdout; a locked plan fails through the core's `FeatureError`.
   - Refactor: nothing; the key stays a literal, like the app's `itemExport`, since keys are the plans file's contract.
14b. [x] terminal journeys (sample): the steps (including a plans-file step that sets `STARTER_PLANS_FILE`); `task e2e:terminal` goes green.
   - Red: none. The export and help steps (`item_steps.go`, with a plans-file step setting `STARTER_PLANS_FILE`) passed straight away, since tasks 13 and 14 built the behaviour; swapping the outline's plans fails both rows as it should. Kept as a guard.
   - Refactor: nothing; the steps are short and in the sample file. `task test:init` now passes in full.
15. [x] TUI (template): the gated key help helper and its test, outside sample blocks.
   - Red: `TestGatedKeyHelpShowsTheKeyByItsFeaturesState` (template `tui/gated_test.go`: on, locked with / without a hint, hidden) failed on the stub's empty string.
   - Green: `gatedKeyHelp(binding, feature)` returns the binding, "… (locked: available on the pro plan)", "… (locked)", or "" when hidden.
   - Refactor: a `switch` shaped like the CLI's `note`, and a doc comment. Not shared with `note`: the wording differs per front end, and only the rule's four branches repeat.
16. [x] web (template): `useFeature`, stubbing `fetch` of `/api/v1/entitlements`: on / locked + hint / hidden / loading → pending / failed → locked without a hint.
   - Red: `useFeature.test.ts` (`it.each`, 5 states, `entitlementsApi` fetch fake in `web/src/test`) failed on the stub's `pending` in 4 rows.
   - Green: `useFeature(key)` queries `/api/v1/entitlements` and maps the key's feature to `on` / `locked` + `availableOn` / `hidden`; `pending` before the answer; `locked` with no hint on an error.
   - Refactor: an `entitlementsKey` constant and doc comments. Lint (`noUncheckedIndexedAccess`) wanted a key the plans don't know handled: a sixth row failed with a TypeError, then it shows locked with no hint, as the CLI's help does.
16b. [x] web (template): `Layout` starts the entitlements request when the app starts.
   - Red: `Layout` "starts reading the plan's features when the app starts…" (`entitlementsApi` fetch fake) failed: only `/api/v1/connection` and `/api/v1/warnings` were requested.
   - Green: `entitlementsQuery` (`queryOptions`) exported from `useFeature.ts`, used by `useFeature` and by `Layout`'s `usePrefetchQuery`.
   - Refactor: a doc comment on `entitlementsQuery`; the query stays in `useFeature.ts`, where 16c's `staleTime` goes. `task test:init` passes.
16c. [x] web (template): the entitlements answer is kept for the session (a long `staleTime`).
   - Red: `useFeature` "keeps the entitlements for the session…" (a second hook on the same client) failed: fetch called 2 times, want 1. `renderHookWithClient` takes an optional client and returns it.
   - Green: `entitlementsQuery` gets `staleTime: Number.POSITIVE_INFINITY`.
   - Refactor: `entitlementsQuery`'s doc comment says why it never goes stale; nothing else worth changing.
17. [ ] web (template): the gate component: children / locked with hint / locked without hint ("Not available") / nothing / a same-size disabled placeholder while pending.
18. [ ] web (sample): the Items page's Export button downloads the CSV; e2e steps; `task e2e` goes green. Check the download by hand in `task run:desktop` (WebKitGTK).
19. [ ] docs: `docs/features.md` (a checklist for gating a new feature), linked from CLAUDE.md's plans rule; ADR 5's step 3 line; the plans file comments; a template-user step for defining their own features if needed; `task test:init`.
20. [ ] Finish: `task check`, `task accept`, `task e2e`, `task e2e:terminal`; remove `@wip`; mark step 3 done in the plan; delete this file.

## Changes to the plan
- 2026-11-04 (user review, before the first `/red`):
  - Features get a required display `name` in the plans file, so template users choose the wording; messages use it.
  - CLI help lists locked commands as disabled on the plan, instead of hiding them, and reads only the entitlements, never opening storage (opening the app for help would create the folder, run migrations and fail on a lock).
  - The web app avoids pop-in: it prefetches the entitlements, gated pages wait for them in their loading state, and the gate has a same-size placeholder. Failing to load shows the feature locked, not hidden.
- 2026-11-04 (user, during task 5): task 5 split in two, one test each: 5 is the template light path and no hints; 5b is the shipped plans file's `features` and the sample `item-export`, whose test sits in a sample block.
- 2026-11-04 (user, during task 10): task 10 split in two, one test each: 10 serves the features in the entitlements; 10b maps `FeatureError` to its problem. `/items/export` moves to task 11, whose handler test needs it.
- 2026-11-04 (user, during task 12): task 12 split in three, one test each: 12 reads the features; 12b maps the feature problem back; 12c is the sample `ExportItems`, which turns the http, desktop and mobile drivers green.
- 2026-11-05 (user, during task 14): task 14 split in two, one test each: 14 is the CLI's `item export` command; 14b is the terminal journey steps and `task e2e:terminal`.
- 2026-11-05 (user, during task 16): task 16 split in two, one test each: 16 is `useFeature`'s states; 16b is `Layout`'s prefetch with a long `staleTime`. `/iterate` now splits such tasks without asking.
- 2026-11-05 (`/iterate`, task 16b): task 16b split in two, one test each: 16b is `Layout`'s prefetch; 16c keeps the answer for the session (a long `staleTime`).
