---
description: TDD refactor — improve structure without changing behaviour; all tests stay green
argument-hint: "[optional: area or smell to focus on]"
---

TDD **REFACTOR** step. Focus: $ARGUMENTS

Rules:

1. Run `task test` and `task accept` first. If anything is red, STOP and tell me: refactoring only happens on green.
2. Improve the code touched in the current red/green cycle (and the tests themselves if useful). Look for duplication, naming, long functions, misplaced responsibilities, and violations of the hexagonal dependency rule in CLAUDE.md.
3. **No behaviour changes**: no new features, no new tests asserting new behaviour, no changes to public API contracts (`api/openapi.yaml`) unless I asked.
4. Make small steps, re-running the relevant tests after each one. If a step turns a test red, revert that step.
5. Finish with `task lint`, `task test` and `task accept`, all green.
6. STOP. Summarise what you changed and why. If nothing was worth changing, say so; that's a valid outcome.
