---
description: TDD red — write ONE failing unit test for the next behaviour, then stop
argument-hint: <behaviour to specify>
---

TDD **RED** step for:

> $ARGUMENTS

Rules:

1. Write **exactly one** new test (one `func TestXxx` or `t.Run` case in Go, or one `it(...)` in Vitest) that specifies the behaviour above. Pick the innermost layer where the behaviour belongs (domain > app > adapter > UI); see the architecture rules in CLAUDE.md.
2. Do **not** change production code. The only exception is the minimum needed for the test to *compile*: an empty function, method, or type returning a zero value. A compile failure is not a valid "red".
3. Run the narrowest test command (e.g. `go test ./internal/core/domain/ -run TestXxx` or `pnpm --dir web test -- -t "name"`).
4. Confirm it fails **for the expected reason**: an assertion failure that matches the missing behaviour. If it passes or fails for another reason, fix the test, not the code.
5. STOP. Show me the test, the failure output, and a one-line description of what `/green` must do. Do not continue to green.
