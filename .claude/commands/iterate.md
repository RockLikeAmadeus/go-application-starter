---
description: TDD loop — one red → green → refactor cycle for the next task, without stopping between steps, then commit and push it
argument-hint: "[optional: behaviour to specify; default: the next open task in docs/wip/]"
allowed-tools: Bash(git add:*), Bash(git commit:*), Bash(git push:*), Bash(gh pr view:*), Bash(gh pr create:*)
---

One full **red → green → refactor** cycle, for:

> $ARGUMENTS

If that's empty, take the first unticked task in the `docs/wip/` file of the feature in progress.

Run the three steps below in order, each exactly as its own command would (`.claude/commands/red.md`, `green.md` and `refactor.md`), with every rule they set, except their "STOP" at the end: go straight on to the next step instead. Don't merge the steps: the test must fail for the right reason before any production code is written, and refactoring starts only once everything is green.

0. **Prepare.** Run `git status`, and note whether the working tree is clean: the cycle commits its own changes at the end. Run `task wip:next`, read only the sections of the plan (`docs/plans/`) the task needs, if any, as `.claude/rules/work-tracking.md` says, and read the code the task touches. If no task is open, STOP and say so.
1. **Red** (`red.md`): exactly one new failing test, failing for the expected reason. Record it in the `docs/wip/` file.
2. **Green** (`green.md`): the minimum production code to pass it, with the full unit suite green. Record it, and leave the **For refactor:** note.
3. **Refactor** (`refactor.md`), starting from that note. While the feature's scenarios are still `@wip`, "`task accept` green" means every scenario outside them: `task accept -- --godog.tags='~@wip'`. Report how far the `@wip` ones got (`task accept -- --godog.tags=@wip`), without chasing them. Replace the note with what was done.

**When to stop and ask me.** Ask only when the decision is genuinely mine, then carry on with my answer. Ask with the AskUserQuestion tool, and make each question stand on its own. Stop to ask when:
- the task is ambiguous, or the decisions in the `docs/wip/` file don't settle how to do it;
- making it pass needs more than the task describes, or changes the plan (a new port, a changed API contract, a dropped or reordered task);
- an existing test has a bug, or something is red that this cycle didn't break;
- a refactoring would change behaviour or a public contract;
- the cycle adds or changes user-facing text (`.claude/rules/user-facing-text.md`): always, before the test that pins it. This is never a conventional answer.

Anything with a conventional answer, or that the code, CLAUDE.md or the `docs/wip/` file settles, isn't a question: decide, and say so in the summary.

**When a task covers two behaviours that each need their own test**, don't ask: split it. Keep the first behaviour as task N and add the other as Nb (then Nc…) right after it, one test each, log the split under "Changes to the plan" in the `docs/wip/` file, and run this cycle on task N.

**When the new test passes straight away.** If it passes because the behaviour already exists (not because the test is wrong: check it can fail, e.g. that it waits for or asserts the right thing), don't ask me: keep it as a guard. Skip green, since no production code is demanded, and still run refactor. Record in the `docs/wip/` file that there was no red, and why. Commit it as `test:` (or `feat:` if refactor changed production code). A test that passes only sometimes is not this case: make it fail reliably for the right reason, as `red.md` says.

**Finish.** `task lint`, `task test` and the acceptance run above, all green.

**Run each check once per state of the code.** While working, run the narrowest test that covers the change. Run `task test` and the acceptance run once after green: refactor's opening check reuses those results instead of running them again. Refactor's small steps re-run only the tests they touch. At the finish, run only what the refactor could have affected since green: if it changed nothing, that's `task lint`; if it changed only one package's tests, that package's tests and `task lint`. Otherwise run the full finishing checks. Slow, separate checks (`task test:init`, `task e2e:terminal`) run once, when the cycle touched what they cover.

**Keep the `docs/wip/` notes short**, as `red.md`, `green.md` and `refactor.md` say: one line per step. This cycle's commit message holds the detail.

**Commit and push.** Running `/iterate` is my explicit request to commit its cycle, so when the cycle went cleanly, stage everything (`git add -A`), commit, and push to the current branch's upstream. Run them as three separate, plain commands (`git add -A`; `git commit -F <message file in the scratchpad>`; `git push`, or `git push -u origin HEAD` when the branch has no upstream yet, as on a feature's first cycle), not chained with `&&` or piped, so they match the `allowed-tools` above and don't prompt. "Cleanly" means:
- the cycle ran to the end: red failed for the right reason (or the test passed straight away and was kept as a guard, as above), then green, then refactor;
- the finishing checks above are all green;
- nothing is left open for me: any question I answered during the cycle is settled.

The commit message follows the repository's style (`feat: …`, or `test:`/`refactor:` when that's all the cycle did). It names the task, and says what the test specifies and what the code now does. End it with the attribution lines this session gives for commits.

Don't commit, and say why, when:
- anything above isn't clean (the cycle stopped early, or a check is red);
- the working tree had changes before the cycle started that aren't part of it. Ask whether to include them, leave them out, or not commit.

Never pull, force-push or rewrite history. If the push is refused (for example, the remote has moved on), leave the commit in place and tell me.

**Open the pull request after the last task.** When this cycle's commit was pushed and the `docs/wip/` file has no open task left (`grep -c '^- \[ \]'` gives 0), open the feature's pull request, as `.claude/rules/work-tracking.md` says under "Pull requests". Running `/iterate` on the last task is my request to open it. Don't open one from `main`, or when the branch already has one (`gh pr view --json url` finds it): report that one's link instead.

Then STOP after this one cycle, and show me, in this order:
- the test, and the failure it showed before green (or why it passed straight away);
- the production change, in 1–3 lines;
- what the refactor changed, or why nothing was worth changing;
- the checks: the finishing checks' results and the `@wip` scenarios' progress;
- the commit and push: its hash and its full message, after the checks and before the next task (not only in an opening line), or why there wasn't one;
- the next open task, as the suggested next `/iterate` (or `/red`); or, after the last task, the pull request's link, or why there wasn't one.
