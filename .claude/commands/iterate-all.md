---
description: TDD loop — run an /iterate cycle in a fresh subagent for every open task of the feature in progress, one commit per task, until none are left
argument-hint: "[optional: feature slug or docs/wip/ file; default: the most recently committed one with open tasks]"
allowed-tools: Agent, Bash(git add:*), Bash(git commit:*), Bash(git push:*)
---

Run an `/iterate` cycle for each open task of one `docs/wip/` file, without stopping between cycles. I use this when I've reviewed the plan and don't expect to change it along the way.

> $ARGUMENTS

You coordinate. Each cycle runs in its own subagent, so it starts from a clean context, as it would after `/clear`. Stay small: don't read the code, the plan or test output yourself. Rely on what each subagent reports, apart from the review and the sweep below, which read only the lines their searches find.

**Pick the file.** If the arguments name a feature, use `docs/wip/<slug>.md`. Otherwise use the only file in `docs/wip/`, or, when there are several, the one with open tasks (`- [ ]`) that was committed to most recently:

```
for f in docs/wip/*.md; do grep -q '^- \[ \]' "$f" && echo "$(git log -1 --format=%ct -- "$f") $f"; done | sort -n | tail -1
```

Say which file you picked before the first cycle, and stay on it. If it has no open tasks, STOP and say so.

**Check the working tree.** Before the first cycle, note `HEAD`'s short hash, as the loop's starting point, and run `git status --short`. If there are changes, ask me with the AskUserQuestion tool whether to include them, leave them out, or stop. If I leave them out, name those paths in each subagent's prompt, stage only the paths its report lists (`git add -- <paths>`), and never `git add -A`.

**Loop.** Repeat for as long as the file has an open task:

1. When only one open task is left, run the sweep below first.
2. Start a `general-purpose` subagent with the Agent tool, with the prompt below, and wait for its report.
3. Act on the report's first word:
   - `QUESTION`: ask me with the AskUserQuestion tool. Send my answer to the same subagent with SendMessage (load it with ToolSearch if it's deferred), and wait again.
   - `BLOCKED`: STOP the loop, leaving its changes unstaged and uncommitted.
   - `DONE`: go on to step 4.
   - Anything else, or no report at all: see "When a cycle is cut off" below.
4. Review the cycle, as below.
5. Check with `grep` that the task is ticked. Write the message to a file in the scratchpad, ending it with the attribution lines this session gives for commits. Then commit and push as `/iterate` says: `git add -A`; `git commit -F <file>`; `git push`, as three separate, plain commands.
6. Print one line: the task, the commit's short hash, and how far the `@wip` scenarios got. Go straight on to the next task. Don't end your turn between cycles.

**Review each cycle before committing it.** A `DONE` report can be wrong, so check four things. Don't read the whole diff: search it, and read only the hunks the searches point at. Search the changes to tracked files with `git diff -U0 HEAD`, and new files by passing `git ls-files -o --exclude-standard` to `grep`.
- **Paths:** `git status --short` lists the same paths as the report.
- **Imports:** for each new import of one of this module's packages (`go list -m` gives the module path), check it against the architecture in `CLAUDE.md`. depguard only guards the core, so look for an adapter importing another adapter, a package outside `internal/bootstrap` choosing a storage adapter, and template code importing sample code.
- **Deferred work:** search the added lines for `until`, `for now`, `temporary`, `later` and `task [0-9]`. Each note that leaves something for later must name an open task in the file that does it.
- **Commit message:** for everything the message says this cycle removes, renames or replaces, `grep` that the old form is gone.

If something is wrong, send it to the same subagent with SendMessage, to fix it, run the finishing checks again and report again. If it needs a decision from me rather than a fix, ask me with AskUserQuestion, as for `QUESTION`. Commit only once the review is clean.

**Sweep before the last task.** The last task often only tidies up or updates the docs, so anything an earlier cycle left behind has to be found before it starts. Search for promises that no remaining task will keep:
- in this loop's commit messages: `git log <starting point>..HEAD --format=%B`, searched for `until`, `later`, `for now`, `temporar`, `remove` and `task [0-9]`;
- in the tree: `git grep -nIiE 'until .{0,40}task [0-9]|for now|temporar'`, leaving out `docs/plans/` and other `docs/wip/` files. The file's own notes count.

For each promise, check that the last task covers it, or that it points to another feature's task. If any don't, ask me with AskUserQuestion whether to add tasks for them, and add the tasks I agree to as unticked tasks before the last one. Then carry on with the loop, which now does those tasks first.

**When a cycle is cut off.** A subagent can end without a report, or with one that doesn't start with `DONE`, `QUESTION` or `BLOCKED`. This happens when it reaches a usage limit, an API error or a time limit. Its work may be half done, so:
1. Don't commit, stage or discard anything, don't start a new subagent, and don't wait or schedule a retry yourself.
2. Run `git status --short`. Then STOP and tell me the task, the subagent's ID, what it said last, and which paths it left changed.
3. When I say to carry on, resume the same subagent with SendMessage. Tell it that it was interrupted, to check the tree and the task's notes, and to carry on from where it stopped. If it can't be resumed, start a new subagent with the usual prompt and this added: "A cycle for this task was interrupted and left these changes: <paths>. They are part of this cycle: check them against the task, and keep or finish them. Report `BLOCKED` if you can't tell how far it got." Either way, the cycle goes through the review before it's committed.

If you are cut off yourself, I'll start you again. Then the file, `git log` and `git status --short` show where the loop is. If there are changes and no subagent you can resume, follow step 3 as for a subagent that can't be resumed.

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
- a subagent reports `BLOCKED` or is cut off, a check after `DONE` fails, or a commit or push fails. Never force-push, pull or rewrite history, and don't start another cycle on top of it;
- I ask you to stop.

**Don't go beyond the tasks.** Removing the `@wip` tags, deleting the `docs/wip/` file, and marking a step done in `docs/plans/` stay with me unless they are tasks in the file. `/iterate`'s permission to commit covers each cycle only.

**Finish.** Show me, in this order:
- why the loop stopped, and which tasks remain, if any;
- a list of the cycles: each task, with its commit's short hash and subject;
- every decision the subagents took without asking;
- anything the reviews sent back, and what the sweep found;
- the last cycle's checks and the `@wip` scenarios' progress. When no task is left, run `task accept -- --godog.tags=@wip` and report it, with the suggested next step: removing the tags if they all pass, or what is still failing.
