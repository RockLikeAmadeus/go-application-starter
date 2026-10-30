---
description: TDD green — write the minimum production code to pass the failing test, then stop
argument-hint: "[optional: test name]"
---

TDD **GREEN** step. $ARGUMENTS

Rules:

0. If a `docs/wip/` file exists for the feature in progress, read it first. It holds the task list and the decisions behind it.
1. Find the currently failing test (the one just written by `/red`, or the one named above). Run it to see the failure.
2. Write the **simplest production code that makes it pass**. Hard-coding, duplication, and naive code are acceptable here. Do not add behaviour that no test demands. Do not refactor.
3. Do **not** modify the test, unless it has an outright bug. If it does, stop and tell me.
4. Run the full unit suite for the affected side (`task test:go` and/or `task test:web`). Every test must pass. If you broke something, fix the production code minimally.
5. If an `@wip` acceptance scenario may now pass, run `task accept -- --godog.tags=@wip` and report the result. Do not chase it further.
6. Update the `docs/wip/` file: tick off or annotate the item this step advanced, and if the plan changed, revise the list and log why.
7. STOP. Summarise the change in 1–3 lines and list any obvious refactoring opportunities for `/refactor`. Do not refactor.
