---
description: Start a new feature — BDD outer loop (spec + failing Gherkin scenarios, no implementation)
argument-hint: <short description of the feature>
---

You are starting the **BDD outer loop** for this feature:

> $ARGUMENTS

Follow these steps exactly, then STOP. Do **not** write production code.

1. **Understand.** Read `SPEC.md`, `CLAUDE.md`, and the existing `features/*.feature` files. If the feature is ambiguous, ask me up to 3 focused questions before writing anything.
2. **Spec.** Add or update a section in `SPEC.md` describing the capability in plain business language: purpose, rules, and examples. Keep it short.
3. **Scenarios.** Write Gherkin in `features/<capability>.feature`:
   - Use the domain's language (see the glossary in SPEC.md), not UI or HTTP terms.
   - One behaviour per scenario. Cover the happy path, the business rules, and the important failures.
   - Prefer `Scenario Outline` + `Examples` for rule variations.
   - Tag the new scenarios `@wip`.
   - If a scenario or the spec quotes user-facing text, agree it with me first (`.claude/rules/user-facing-text.md`). Otherwise, describe the meaning and leave the wording to the cycle that writes it.
   - Only add a browser journey under `web/e2e/features/` if the feature has user-visible UI wiring that the Godog layer cannot prove. Keep those few and thin.
   - Likewise, only add a terminal journey under `test/terminal/features/` if the feature has CLI or TUI wiring the Godog layer cannot prove (options, output, exit status, keys). Keep those few and thin too.
4. **Prove they fail.** Run `task accept -- --godog.tags=@wip` and confirm the new scenarios fail or are undefined for the *right* reason. Do not write step definitions that contain business logic.
5. **Plan the inner loop.** Write a numbered TDD task list: the smallest sequence of unit-test-sized behaviours (domain → app/use case → ports → adapters → UI) that will turn the scenarios green. Each item should be a good argument for `/red`. If the feature adds a stored entity, or a new query on one, its data belongs to a workspace: plan the steps from the checklist in `docs/workspaces.md` (the entity's `WorkspaceID`, a port embedding `ports.WorkspaceRepository[T]`, `storagetest.RunWorkspaceIsolation` in each adapter, the fail-closed use-case test).
6. **Save the plan.** Write `docs/wip/<feature-slug>.md` (see `.claude/rules/work-tracking.md`). It should contain:
   - links to the spec section and the feature file(s);
   - the key decisions behind the plan, so a fresh session doesn't re-derive them;
   - the task list as checkboxes;
   - an empty "Changes to the plan" log.

End by showing me the scenarios and the task list, and suggest the first `/red` command.
