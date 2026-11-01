---
description: TDD loop — one red → green → refactor cycle for the next task, without stopping between steps
argument-hint: "[optional: behaviour to specify; default: the next open task in docs/wip/]"
---

One full **red → green → refactor** cycle, for:

> $ARGUMENTS

If that's empty, take the first unticked task in the `docs/wip/` file of the feature in progress.

Run the three steps below in order, each exactly as its own command would (`.claude/commands/red.md`, `green.md` and `refactor.md`), with every rule they set, except their "STOP" at the end: go straight on to the next step instead. Don't merge the steps: the test must fail for the right reason before any production code is written, and refactoring starts only once everything is green.

0. **Prepare.** Read the feature's `docs/wip/` file, the plan it belongs to (`docs/plans/`), if any, and the code the task touches. If no task is open, STOP and say so.
1. **Red** (`red.md`): exactly one new failing test, failing for the expected reason. Record it in the `docs/wip/` file.
2. **Green** (`green.md`): the minimum production code to pass it, with the full unit suite green. Record it, and leave the **For refactor:** note.
3. **Refactor** (`refactor.md`), starting from that note. While the feature's scenarios are still `@wip`, "`task accept` green" means every scenario outside them: `task accept -- --godog.tags='~@wip'`. Report how far the `@wip` ones got (`task accept -- --godog.tags=@wip`), without chasing them. Replace the note with what was done.

**When to stop and ask me.** Ask only when the decision is genuinely mine, then carry on with my answer. Ask with the AskUserQuestion tool, and make each question stand on its own. Stop to ask when:
- the task is ambiguous, or the decisions in the `docs/wip/` file don't settle how to do it;
- the new test passes straight away (the behaviour may already exist; I decide whether to keep it as a guard, as `red.md` would report it);
- making it pass needs more than the task describes, or changes the plan (a new port, a changed API contract, a dropped or reordered task);
- an existing test has a bug, or something is red that this cycle didn't break;
- a refactoring would change behaviour or a public contract.

Anything with a conventional answer, or that the code, CLAUDE.md or the `docs/wip/` file settles, isn't a question: decide, and say so in the summary.

**Finish.** `task lint`, `task test` and the acceptance run above, all green. Then STOP after this one cycle, and show me:
- the test, and the failure it showed before green;
- the production change, in 1–3 lines;
- what the refactor changed, or why nothing was worth changing;
- the `@wip` scenarios' progress;
- the next open task, as the suggested next `/iterate` (or `/red`).
