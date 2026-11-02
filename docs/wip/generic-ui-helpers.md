# WIP: Move generic UI helpers out of the sample

- Backlog: `TODO.md`, "Do first", **Move generic UI helpers out of the sample**.
- Spec: [docs/template.md](../template.md), "Make it mine": the `--remove-sample` rule's new sub-bullet and the new example. Not `SPEC.md`: this is template tooling, not product behaviour.
- Scenarios: [scripts/initproject/features/init.feature](../../scripts/initproject/features/init.feature), the two `@wip` scenarios. Run them with `task test:init -- --godog.tags=@wip` (about 10 s).
- Temporary: delete this file in the change that removes the `@wip` tags and ticks the TODO item.

## Decisions
- **The outer check is the init tool's Gherkin, not `features/`.** Nothing a user of the app sees changes; what changes is what a project keeps after `task init --remove-sample`. The steps (in `scripts/initproject/acceptance/acceptance_test.go`) are thin file checks plus a `golangci-lint run ./...` on the trimmed copy.
- **Shown failing (2026-11-02):**
  - "Removing the sample keeps what the front ends share": `internal/adapters/tui/styles.go does not contain "errorStyle"` (the styles sit in a sample block). The web step is skipped after it; `web/src/components/ErrorMessage.tsx` doesn't exist yet either.
  - "The project without the sample passes the Go linters": `test/acceptance/plans.go:50:21: func (*testPlans).define is unused`. An existing bug the new scenario found: `define` is template machinery whose only callers are the sample's `plan_steps.go`. The TUI styles would add the same failure if simply moved out of the block, which is why they ended up in it.
- **The TUI styles stay used through small template helpers, not dummy references.** `errorLine(err error) string` (errorStyle, "Error: …") and `selectedRow(text string) string` (selectedStyle, "> …") live in `styles.go`, outside the sample block, with tests in a new `styles_test.go`. golangci-lint counts test files, so the helpers' own tests keep them (and `danger`) used once the Items screen is gone. `items.go` calls the helpers.
- **The web app's shared components go in `web/src/components/`** (new folder). `web/src/features/` is per capability and `web/src/lib/` holds non-component helpers. `ErrorMessage` moves there unchanged (`role="alert"`, `text-danger`) with its own test; `ItemsPage` imports it. TypeScript and ESLint don't flag an unused export, so no extra care is needed on the web side.
- **`define` gets a template caller:** a generic step `the "<plan>" plan has no limits` in `entitlements_steps.go` (template), which `features/plans.feature` uses instead of the sample's `the "<plan>" plan has no item limit`. The method value referenced in `sc.Step` counts as a use.
- **Web lint isn't run on the trimmed copy:** it needs `node_modules` in the copy, and ESLint wouldn't catch this class of problem anyway.

## Tasks
1. [x] TUI: `errorLine(err)` renders "Error: <message>" in the error style. Test in `internal/adapters/tui/styles_test.go`. Move `danger` and `errorStyle` out of the sample block; `items.go` uses `errorLine`.
   - Red: `TestErrorLineShowsTheMessageInTheErrorStyle` (`styles_test.go`, package `tui`, since the helper is unexported) failed with `errorLine = "", want "\x1b[38;2;220;38;38mError: boom\x1b[m"`. It compares with `errorStyle.Render`, so the colour is checked without hard-coding escape codes.
   - Green: `errorLine` returns `errorStyle.Render("Error: " + err.Error())`.
   - Refactor: `danger` and `errorStyle` moved out of the sample block into the template styles, `errorLine` got a doc comment and normal formatting, and `items.go` calls it. `@wip` progress: the first scenario now fails on `selectedStyle` (task 2), and the trimmed project's lint reports only `define` (task 4), so the helper's test keeps `errorStyle` and `danger` in use, as planned.
2. [x] TUI: `selectedRow(text)` renders "> <text>" in the selected style. Move `selectedStyle` out of the sample block; `items.go` uses `selectedRow`. The sample block in `styles.go` should then be empty: remove its markers.
   - Red: `TestSelectedRowMarksTheTextInTheSelectedStyle` failed with `selectedRow = "", want "\x1b[1;38;2;124;58;237m> Milk\x1b[m"`.
   - Green: `selectedRow` returns `selectedStyle.Render("> " + text)`.
   - Refactor: `selectedStyle` joined the template styles (next to `itemStyle`, the unselected row), the now-empty sample block and its markers went, `selectedRow` got a doc comment and normal formatting, and `items.go` calls it. `@wip` progress: the first scenario's TUI step passes; it now fails on `web/src/components/ErrorMessage.tsx does not exist` (task 3). Lint on the trimmed project still reports only `define` (task 4).
3. [x] Web: `ErrorMessage` in `web/src/components/ErrorMessage.tsx` shows the error's message as an alert (`ErrorMessage.test.tsx`). `ItemsPage` imports it and drops its local copy.
   - Red: "shows the error's message as an alert" (`web/src/components/ErrorMessage.test.tsx`) failed with `Unable to find an accessible element with the role "alert"` against a stub returning `null`.
   - Green: `ErrorMessage` renders the Items page's markup unchanged: `<p role="alert" className="text-sm text-danger">` with the message.
   - Refactor: `ItemsPage` imports the shared `ErrorMessage` and its local copy is gone; its own alert test still passes. `@wip` progress: "Removing the sample keeps what the front ends share" passes; only the lint scenario fails, on `define` (task 4).
4. [x] Acceptance: template step `the "<plan>" plan has no limits` in `test/acceptance/entitlements_steps.go`, calling `testPlans.define`; `features/plans.feature` uses it, and the sample's "has no item limit" step goes. Red is the `@wip` lint scenario.
   - Red: besides the `@wip` lint scenario, `features/plans.feature`'s five `Given the "pro" plan has no item limit` became `… has no limits`, and `task accept` failed on the undefined step (strict mode), on each driver that runs `@plans`.
   - Green: `thePlanHasNoLimits` in `entitlements_steps.go` (template) defines the plan with no limits through `testPlans.define`, and refuses with `errFixedPlans` on drivers whose plans are fixed, like the other plan steps.
   - Refactor: the sample's `thePlanHasNoItemLimit` and its registration went from `plan_steps.go`, and `registerEntitlementsSteps`' comment now covers defining plans. `@wip` progress: both init scenarios pass.
5. [ ] Docs: CLAUDE.md (Architecture: `web/src/components/` for shared web components; TUI conventions: shared styles and helpers in `styles.go`, outside sample blocks), and `docs/template.md`'s "Leave something behind" paragraph (`task test:init` now also lints the trimmed project).
6. [ ] Finish: full `task test:init`, `task check`, remove the `@wip` tags, tick the TODO item, delete this file.

## Changes to the plan
