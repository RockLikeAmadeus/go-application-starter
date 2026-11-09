---
description: Start a new feature — BDD outer loop (spec + failing Gherkin scenarios, no implementation)
argument-hint: <short description of the feature>
allowed-tools: Bash(git fetch:*), Bash(git switch:*), Bash(git add:*), Bash(git commit:*), Bash(git push:*)
---

You are starting the **BDD outer loop** for this feature:

> $ARGUMENTS

Follow these steps exactly, then STOP. Do **not** write production code. Running `/feature` is my request to commit and push its own changes, in step 7.

0. **Start from main.** Run `git branch --show-current` and `git status --short`. If the branch isn't `main`, or the working tree has changes, STOP and tell me: a feature starts from a clean `main`. Then run `git fetch origin main` and `git rev-list --count HEAD..origin/main`. If `main` is behind `origin/main`, STOP and tell me, as I pull it myself.
1. **Understand.** Read `SPEC.md`, `CLAUDE.md`, and the existing `features/*.feature` files. If the feature is ambiguous, ask me up to 3 focused questions before writing anything.
   Once the feature is clear, choose its slug (used for `docs/wip/<feature-slug>.md` in step 6) and create its branch with `git switch -c feature/<feature-slug>`, before writing anything. Every `/iterate` cycle of the feature commits to this branch, and the last one opens its pull request.
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
   End the list with a **docs and finish** task: update the docs the feature changes, remove the `@wip` tags once every scenario passes, and delete the `docs/wip/` file. If the feature is a step of a `docs/plans/` plan, the task also marks the step done there. If it's the plan's last step, the task instead moves the plan's open follow-ups into `TODO.md`, ticks the TODO item and deletes the plan (`.claude/rules/work-tracking.md`).
6. **Save the plan.** Write `docs/wip/<feature-slug>.md` (see `.claude/rules/work-tracking.md`). It should contain:
   - links to the spec section and the feature file(s);
   - the key decisions behind the plan, so a fresh session doesn't re-derive them;
   - the task list as checkboxes;
   - an empty "Changes to the plan" log.
7. **Commit and push.** Only if the new scenarios fail for the right reason and the scenarios outside them still pass (`task accept -- --godog.tags='~@wip'`), stage everything (`git add -A`), commit, and push the new branch, as three separate, plain commands: `git add -A`; `git commit -F <message file in the scratchpad>`; `git push -u origin HEAD`. The message is `test: specify <the feature> (@wip)`, with a short body naming the spec section, the feature file(s) and the number of tasks, and ending with the attribution lines this session gives for commits. Never pull, force-push or rewrite history. If something isn't clean, or the push is refused, don't commit (or leave the commit in place), and tell me why.

End by showing me the branch, the scenarios, the task list, and the commit's hash (or why there wasn't one), and suggest the first `/iterate` (or `/red`) command.
