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
   - Only add a browser journey under `web/e2e/features/` if the feature has user-visible UI wiring that the Godog layer cannot prove. Keep those few and thin.
4. **Prove they fail.** Run `task accept -- --godog.tags=@wip` and confirm the new scenarios fail or are undefined for the *right* reason. Do not write step definitions that contain business logic.
5. **Plan the inner loop.** Write a numbered TDD task list: the smallest sequence of unit-test-sized behaviours (domain → app/use case → ports → adapters → UI) that will turn the scenarios green. Each item should be a good argument for `/red`.
6. **Save the plan.** Write `docs/wip/<feature-slug>.md` (see CLAUDE.md). It should contain:
   - links to the spec section and the feature file(s);
   - the key decisions behind the plan, so a fresh session doesn't re-derive them;
   - the task list as checkboxes;
   - an empty "Changes to the plan" log.

End by showing me the scenarios and the task list, and suggest the first `/red` command.
