---
description: TDD loop — run an /iterate cycle in a fresh subagent for every open task of the feature in progress, one commit per task, until none are left
argument-hint: "[optional: feature slug or docs/wip/ file; default: the most recently committed one with open tasks]"
allowed-tools: Agent, Bash(git add:*), Bash(git commit:*), Bash(git push:*)
---

Run an `/iterate` cycle for each open task of one `docs/wip/` file, without stopping between cycles. I use this when I've reviewed the plan and don't expect to change it along the way.

> $ARGUMENTS

You coordinate. Each cycle runs in its own subagent, so it starts from a clean context, as it would after `/clear`. Stay small: don't read the code, the plan or test output yourself, and rely on what each subagent reports.

**Pick the file.** If the arguments name a feature, use `docs/wip/<slug>.md`. Otherwise use the only file in `docs/wip/`, or, when there are several, the one with open tasks (`- [ ]`) that was committed to most recently:

```
for f in docs/wip/*.md; do grep -q '^- \[ \]' "$f" && echo "$(git log -1 --format=%ct -- "$f") $f"; done | sort -n | tail -1
```

Say which file you picked before the first cycle, and stay on it. If it has no open tasks, STOP and say so.

**Check the working tree.** Run `git status --short` before the first cycle. If there are changes, ask me with the AskUserQuestion tool whether to include them, leave them out, or stop. If I leave them out, name those paths in each subagent's prompt, stage only the paths its report lists (`git add -- <paths>`), and never `git add -A`.

**Loop.** Repeat for as long as the file has an open task:

1. Start a `general-purpose` subagent with the Agent tool, with the prompt below, and wait for its report.
2. Act on the report's first word:
   - `QUESTION`: ask me with the AskUserQuestion tool. Send my answer to the same subagent with SendMessage (load it with ToolSearch if it's deferred), and wait again.
   - `BLOCKED`: STOP the loop, leaving its changes unstaged and uncommitted.
   - `DONE`: go on to step 3.
3. Check with `grep` that the task is ticked. Write the message to a file in the scratchpad, ending it with the attribution lines this session gives for commits. Then commit and push as `/iterate` says: `git add -A`; `git commit -F <file>`; `git push`, as three separate, plain commands.
4. Print one line: the task, the commit's short hash, and how far the `@wip` scenarios got. Go straight on to the next task. Don't end your turn between cycles.

Prompt for each subagent, with `<file>` filled in:

> Run one TDD cycle for the first unticked task in `<file>`. Read `.claude/commands/iterate.md` and follow it, reading `red.md`, `green.md` and `refactor.md` beside it, with every rule they set, except for these changes:
>
> - Wherever they say `task wip:next`, run `task wip:next -- <file>`.
> - You can't reach me directly. Where they say to ask me, don't use AskUserQuestion. End your turn with a report starting `QUESTION`: each question standing on its own, with its options and your recommendation. I'll answer by message, and you carry on from where you stopped.
> - Don't stage, commit or push. Stop after the finishing checks, leaving everything unstaged. The coordinator commits.
> - If the coordinator names changes that aren't part of this cycle, leave them alone and don't ask about them.
>
> End with a report. Its first word is `DONE` if the cycle ran to the end and the finishing checks are green, or `BLOCKED` if it can't finish, whatever the reason, including anything red that you can't fix. For `BLOCKED`, say why and what state the tree is in. For `DONE`, keep it short: the task; how far the `@wip` scenarios got; any decision you took without asking; every path you changed; and the full commit message `iterate.md` describes, without attribution lines.

The `docs/wip/` file and the git history hold the loop's state, not this conversation. If the context has been compacted, carry on from the file: the next task is its first unticked one.

**STOP the loop, and say why, when:**
- no open task is left;
- a subagent reports `BLOCKED`, a check after `DONE` fails, or a commit or push fails. Never force-push, pull or rewrite history, and don't start another cycle on top of it;
- I ask you to stop.

**Don't go beyond the tasks.** Removing the `@wip` tags, deleting the `docs/wip/` file, and marking a step done in `docs/plans/` stay with me unless they are tasks in the file. `/iterate`'s permission to commit covers each cycle only.

**Finish.** Show me, in this order:
- why the loop stopped, and which tasks remain, if any;
- a list of the cycles: each task, with its commit's short hash and subject;
- every decision the subagents took without asking;
- the last cycle's checks and the `@wip` scenarios' progress. When no task is left, run `task accept -- --godog.tags=@wip` and report it, with the suggested next step: removing the tags if they all pass, or what is still failing.
