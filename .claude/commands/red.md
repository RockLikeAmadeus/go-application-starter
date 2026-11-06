---
description: TDD red — write ONE failing unit test for the next behaviour, then stop
argument-hint: <behaviour to specify>
---

TDD **RED** step for:

> $ARGUMENTS

Rules:

0. Run `task wip:next` first. If a feature is in progress, it prints the open tasks and the decisions behind them.
1. Write **exactly one** new test (one `func TestXxx` or `t.Run` case in Go, or one `it(...)` in Vitest) that specifies the behaviour above. Pick the innermost layer where the behaviour belongs (domain > app > adapter > UI); see the architecture rules in CLAUDE.md.
2. **Agree user-facing text first.** If the test pins user-facing text, agree it with me before writing the test, as `.claude/rules/user-facing-text.md` says. The wording I agree goes in the test as it stands.
3. Do **not** change production code. The only exception is the minimum needed for the test to *compile*: an empty function, method, or type returning a zero value. A compile failure is not a valid "red".
4. Run the narrowest test command (e.g. `go test ./internal/core/domain/ -run TestXxx` or `pnpm --dir web test -- -t "name"`).
5. Confirm it fails **for the expected reason**: an assertion failure that matches the missing behaviour. If it passes or fails for another reason, fix the test, not the code.
6. Update the `docs/wip/` file: tick off or annotate the item this step advanced, and if the plan changed, revise the list and log why. Keep the note to one line: the test's name and how it failed. Details belong in the commit message; only a decision or plan change that a later step needs gets more, in the file's Decisions or "Changes to the plan" section.
7. STOP. Show me the test, the failure output, and a one-line description of what `/green` must do. Do not continue to green.
