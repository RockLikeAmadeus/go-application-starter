---
description: Move finished TODO.md items, with no open sub-tasks, into docs/backlog-history.md
argument-hint: "[optional: --commit to commit and push the move to main]"
allowed-tools: Bash(git add:*), Bash(git commit:*), Bash(git push:*)
---

Move the finished items out of the backlog. Arguments: $ARGUMENTS

1. Find the candidates: `grep -n '^- \[x\]' TODO.md`. An item is its top-level `- [x]` line plus every indented line under it, up to the next top-level item, section heading, or a blank line followed by one.
2. Read just those items (by line range, not the whole file). Move an item only when none of its lines contains an open checkbox (`- [ ]`). Leave an item whose text says work is still left (such as "Left:" or "is next"), even if it is ticked, and tell me about it.
3. Before moving each item, `grep -n` TODO.md and `docs/plans/` for references to it ("above", "below", or its title). Reword a reference that would now dangle, so it points to `docs/backlog-history.md`.
4. Append each item to `docs/backlog-history.md`, word for word, under the heading of the TODO.md section it came from (`## Do first`, `## Other`, and so on), adding the heading at the end of the file if it isn't there yet. Keep the items' order within a section.
   - Rewrite the item's relative links for the `docs/` folder: `docs/x` becomes `x`, and any other path `p` becomes `../p`. Check that each link in the moved text resolves from `docs/`.
   - Within the moved text, change "above" or "below" that pointed to a TODO.md item into a reference to TODO.md.
5. Remove the moved items from TODO.md. Leave every section heading in TODO.md, even if it is now empty, and leave no run of more than one blank line.
6. Check with `grep -c '^- \[x\]' TODO.md` and `git diff --stat`: the lines removed from TODO.md should match the lines added to the history, apart from new headings.
7. Leave everything unstaged, unless the arguments include `--commit`. Then, on an up-to-date `main` only, stage `TODO.md`, `docs/backlog-history.md`, and any reworded files, commit them as `docs: move finished backlog items to the history`, ending the message with the attribution lines this session gives for commits, and push.
8. STOP. List the items moved (their titles only), any ticked items left in place and why, and any references reworded.
