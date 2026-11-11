---
description: TDD refactor — improve structure without changing behaviour; all tests stay green
argument-hint: "[optional: area or smell to focus on]"
---

TDD **REFACTOR** step. Focus: $ARGUMENTS

Rules:

0. Run `task wip:next` first. If a feature is in progress, it prints the open tasks and the decisions behind them. Start from the **For refactor:** note that `/green` left under the current item.
1. Run `task test` and `task accept` first. If anything is red, STOP and tell me: refactoring only happens on green. While the feature's scenarios are still `@wip`, "`task accept` green" means every scenario outside them: `task accept -- --godog.tags='~@wip'`. The `@wip` ones are expected to fail until the feature is done.
2. Improve the code touched in the current red/green cycle (and the tests themselves if useful). Look for duplication, naming, long functions, misplaced responsibilities, and violations of the hexagonal dependency rule in CLAUDE.md.
3. **No behaviour changes**: no new features, no new tests asserting new behaviour, no changes to user-facing text or to public API contracts (`api/openapi.yaml`) unless I asked.
4. Make small steps, re-running the relevant tests after each one. If a step turns a test red, revert that step.
5. Finish with `task lint`, `task test`, and `task accept` (outside `@wip`, as in step 1), all green. Report how far the `@wip` scenarios got (`task accept -- --godog.tags=@wip`), without chasing them.
6. Update the `docs/wip/` file: tick off or annotate the item this step advanced, and if the plan changed, revise the list and log why. Replace the **For refactor:** note with one line saying what was done, and what was deliberately left and why. Details belong in the commit message; only a decision or plan change that a later step needs gets more, in the file's Decisions or "Changes to the plan" section.
7. STOP. Summarise what you changed and why. If nothing was worth changing, say so; that's a valid outcome.
