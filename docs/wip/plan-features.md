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
- **Domain shape:** `Entitlements.Features map[string]Feature`, where `Feature{Key, Name, On, WhenOff, AvailableOn}` is built with `NewFeature`, `FeatureOff` is `show-locked` or `hide`, and `CheckFeature(key)` returns a `*FeatureError{Feature, Name, AvailableOn}`. That's a 403 problem with `feature`, `name` and `availableOn`, and `httpclient` maps it back.
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
   - Green: `NewFeature(key, name, whenOff)` checks the trimmed name and `whenOff`; `On`/`AvailableOn` left for task 2 and the adapter.
   - Refactor: doc comments on `FeatureOff`, its constants, `Feature` and `NewFeature`, matching `NewLimit`'s; the name isn't trimmed when stored (no test asks).
2. [x] domain: `Entitlements.CheckFeature`: on → nil; off and locked → `*FeatureError` with `AvailableOn`; off and hidden → `*FeatureError` without; unknown → `*FeatureError` without. Test `FeatureError.Error()`'s text, with a hint and without.
   - Red: `TestEntitlements_CheckFeature` (6 cases, each checking the error and its text) failed on the stub's nil; `TestNewFeature` now uses `reflect.DeepEqual` (`AvailableOn` makes `Feature` incomparable).
   - Green: `CheckFeature` refuses off and unknown features (an unknown one's `Name` is its key); `Error()` is "<Name> isn't on your plan", plus ", and is available on the pro plan" / "the pro or team plans".
   - Refactor: one `FeatureError` built, hint added only for `ShowLocked`; doc comments on `CheckFeature`, `FeatureError` and the new `Feature` fields.
3. [x] plans adapter: `Parse` reads `features` (key, `name`, `whenOff`) and each plan's `features` list into `Entitlements.Features` (every catalog feature, `On` per plan). It refuses an unknown feature, a missing `name` and a bad `whenOff`.
   - Red: `TestParse_ReadsFeatures` (catalog + 3 refusals) failed on `unknown field "features"`; the older Parse tests now want an empty `Features` map, as `Limits` is.
   - Green: `Parse` builds the file's `features` with `NewFeature`, gives each plan a copy with `On` set for those it lists, and refuses an unlisted key.
   - Refactor: `buildFeatures` helper, `maps.Clone` for each plan's copy, `Parse`'s doc shows the features; errors read `feature "x": …` like limits'.
4. [x] plans adapter: `For` fills `AvailableOn` (sorted plan names that have the feature, without the local plan) when hints are on, and leaves it empty when they're off.
   - Red: `TestEntitlements_FillsTheUpgradeHintWhenHintsAreOn` (on / off) failed with `AvailableOn = nil, want {"pro", "team"}`; the adapter got stub `Hints` and `LocalPlan` fields.
   - Green: with `Hints`, `For` gives each of the plan's features (in a cloned map) the sorted catalog plans that have it on, leaving out `LocalPlan`.
   - Refactor: `withHints` (copies the map) and `plansWith` helpers; doc comments on `Hints`, `LocalPlan` and `For`. Bootstrap and the harness are left to tasks 5 and 6.
5. [x] bootstrap (template): a light entitlements-only path that opens no storage, for the CLI's help; the local edition wires no hints (a locked feature has no `AvailableOn`).
   - Red: `TestNewEntitlements_ReadsTheLocalPlanWithoutStorageOrHints` (a plans file with `export` locked on local; no data folder made) failed on the stub's nil service.
   - Green: `NewEntitlements(cfg)` wraps `localEntitlements` in `app.NewEntitlementsService` with `workspace.Local{}`; no storage opened.
   - Refactor: doc comments only (`NewEntitlements` says what it skips; `localEntitlements` says it gives no hints). The `--server` branch waits for task 13, whose help needs it.
5b. [x] plans.jsonc (+ bootstrap test in a sample block): the template's `features` section; `item-export` ("Item export", sample block) on for local and pro, locked for free.
   - Red: `TestNewEntitlements_ShippedPlansHaveItemExport` (local / pro / free, through `LocalPlan`) failed: no `item-export` in the shipped file.
   - Green: `plans.jsonc` gets a `features` section and each plan a `features` list; `item-export` and the plans' entries are in sample blocks.
   - Refactor: nothing; the file's explanatory comments are task 19's.
   - `task test:init` is red only on the `@wip` plan-features scenarios (undefined steps) in the generated project; expected until tasks 6, 9 and 19 (user, 2026-11-04).
6. [ ] acceptance harness: hints on in `testPlans`; steps for "has / doesn't have item export", "is shown locked / hidden when off" and "item export is … for my workspace" (in a sample steps file).
7. [ ] domain (sample): encode items as CSV, with header `name,createdAt`, RFC 3339 and CSV quoting.
8. [ ] app (sample): `ExportItems` returns `ErrNoWorkspace` first (in the fail-closed table), refuses with `FeatureError` when off, refuses when the entitlements lookup fails, and returns the items' CSV, oldest first.
9. [ ] acceptance (sample): export steps ("I export the items", "I try to…", "the export lists", "I am told item export isn't on my plan…"). The memory, jsonfile and sqlite drivers go green.
10. [ ] API: `openapi.yaml` gets `features` in `Entitlements`, the feature problem (403 with `feature` and `availableOn`) and `/items/export` (`text/csv`); `task gen`; httpapi maps `FeatureError` and serves the features.
11. [ ] httpapi (sample): the export handler.
12. [ ] httpclient: reads the features in the entitlements, maps the feature problem back to `*FeatureError`, and `ExportItems` (sample). The http, desktop and mobile drivers go green.
13. [ ] CLI (template): the gated-command helper and the help: on → listed plainly; locked → listed as disabled on your plan, with the hint, or without one in the local edition; hidden → not listed; entitlements unreadable → listed with no note. The help opens no storage.
14. [ ] CLI (sample): `starter item export`; the terminal journey steps (including a plans-file step that sets `STARTER_PLANS_FILE`); `task e2e:terminal` goes green.
15. [ ] TUI (template): the gated key help helper and its test, outside sample blocks.
16. [ ] web (template): `useFeature`, stubbing `fetch` of `/api/v1/entitlements`: on / locked + hint / hidden / loading → pending / failed → locked without a hint. `Layout` prefetches the entitlements, with a long `staleTime`.
17. [ ] web (template): the gate component: children / locked with hint / locked without hint ("Not available") / nothing / a same-size disabled placeholder while pending.
18. [ ] web (sample): the Items page's Export button downloads the CSV; e2e steps; `task e2e` goes green. Check the download by hand in `task run:desktop` (WebKitGTK).
19. [ ] docs: `docs/features.md` (a checklist for gating a new feature), linked from CLAUDE.md's plans rule; ADR 5's step 3 line; the plans file comments; the manifest's `SampleFiles`; a template-user step for defining their own features if needed; `task test:init`.
20. [ ] Finish: `task check`, `task accept`, `task e2e`, `task e2e:terminal`; remove `@wip`; mark step 3 done in the plan; delete this file.

## Changes to the plan
- 2026-11-04 (user review, before the first `/red`):
  - Features get a required display `name` in the plans file, so template users choose the wording; messages use it.
  - CLI help lists locked commands as disabled on the plan, instead of hiding them, and reads only the entitlements, never opening storage (opening the app for help would create the folder, run migrations and fail on a lock).
  - The web app avoids pop-in: it prefetches the entitlements, gated pages wait for them in their loading state, and the gate has a same-size placeholder. Failing to load shows the feature locked, not hidden.
- 2026-11-04 (user, during task 5): task 5 split in two, one test each: 5 is the template light path and no hints; 5b is the shipped plans file's `features` and the sample `item-export`, whose test sits in a sample block.
