---
description: Decide what to work on next, from the feature in progress, the plans, and the backlog, and open any discussion it needs
argument-hint: "[optional: a TODO.md item or topic to consider instead of the next one in order]"
allowed-tools: Bash(task wip:next:*), Bash(git status:*), Bash(git branch:*), Bash(git fetch:*), Bash(git rev-list:*), Bash(gh pr list:*)
---

Work out the next step, and start any discussion it needs. Arguments: $ARGUMENTS

This command usually runs from a cleared context, so gather what it needs from the repository, narrowly (`grep -n`, line ranges), rather than reading whole files. It writes no code, and stages, commits, and pushes nothing. Stop at the first step below that gives an answer.

1. **Where things stand.** Run `git branch --show-current`, `git status --short`, and `task wip:next`.
   - If a `docs/wip/` file exists, a feature is in progress, and finishing it comes first. Show its open tasks and recommend `/iterate` (one cycle) or `/iterate-all` (the rest of them, from an Opus session). If its tasks are all ticked but the file is still there, the docs-and-finish task is left. STOP.
   - If the branch is a `feature/…` branch with no `docs/wip/` file, check for its pull request with `gh pr list --head <branch>`. If one is open, it's waiting for my review and merge; say so, and carry on as if from `main`, noting that a new `/feature` needs a clean, up-to-date `main` first.
   - If the working tree has changes, list them briefly, as they may be unfinished work I meant to come back to.
2. **A plan in progress.** List the plans with `ls docs/plans/`, and their steps with `grep -n '^###' docs/plans/*.md`. A plan whose next step is open (not started, or in progress) gives the next step, unless the step waits on hardware or on something else that isn't done. Read that step's section of the plan, and the plan's decisions it relies on, and go to step 4 with it.
3. **The backlog.** If the arguments name an item, use it. Otherwise, find the first open item in `TODO.md` in section order: list them with `grep -n '^## \|^- \[ \]' TODO.md`, and read the tag key under "How to read this list" (by line range) if you need it. Skip an item that is tagged `waits on hardware`, or whose **Depends on:** line names something still open, and say which ones you skipped and why. Within a section the order is a suggestion, so if a later item in the same section is clearly a better next step (it unblocks more, or is much smaller), say so in one line, but recommend the first one unless the case is strong. If a ticked item still has work left ("Left:", "is next"), it may be the real next step; mention it. If ticked items with no open sub-tasks have built up, suggest `/prune-backlog`.
   Read the chosen item in full, with its **Shapes:** and **Depends on:** lines, and `grep` for the files, ADRs, plans, and `docs/backlog-history.md` entries it mentions, so the discussion rests on what is already built.
4. **The discussion.** `TODO.md` asks for a discussion of whether an item is worth building, and its trade-offs, before any work on it starts. Open it now, following the "Writing and asking" section of `CLAUDE.md`: explain what the item is, the terms behind it, and why it matters for a template.
   - Say what including it would give template users, what it costs (to build, and in weight the template carries: code, docs, and setup steps a template user has to understand or remove), and what leaving it out would mean, including whether a product could add it later without much rework. Be honest when an item looks like bloat, since many are ideas no one has weighed yet.
   - For a `decide` item, the question is whether it belongs in the template at all. For a `?` size, it's also what the item would cover, so it can be sized. For an item without either tag, it's still worth checking that it's valuable as written, and whether it should be smaller.
   - Give a recommendation, with your reasons. Ask the questions you need answered with the question tool, at most 3 at a time, with the trade-offs of each option. If the item is plainly worth building and has nothing to settle, say so briefly and move on.
   - For a plan step, the plan's decisions already settle most of this; discuss only what the step leaves open.
5. **Record the outcome**, once the discussion settles it, and with my agreement, leaving the changes unstaged:
   - A "yes" for a `decide` item: remove the `decide` tag, give the item its size, and add the scope or decisions we agreed to the item.
   - A "no": move the item to `docs/backlog-history.md` under the heading of the section it came from, with a line on why, as `/prune-backlog` describes for moving items (including its link rewriting and dangling references).
   - A changed scope: reword the item.
6. **The way in.** Recommend the command that starts the work, written out in full so I can run it:
   - `S`: a few `/red` → `/green` → `/refactor` cycles, or a docs change, on a branch I create (no `/feature`); name the first behaviour.
   - `M`, or a plan step: `/feature <description>`, with a one-line description drawn from the discussion.
   - `L`: a plan first, in `docs/plans/<item-slug>.md`, made with plan mode or written directly (`.claude/rules/work-tracking.md`); offer to draft it.
   Then STOP.

End with a short summary: where things stand, the item chosen (and any skipped, and why), the outcome of the discussion so far, and the command to run next.
