---
paths:
  - "TODO.md"
  - "docs/wip/**"
  - "docs/plans/**"
  - "features/**"
---

# Tracking features and backlog items

## Feature task lists (`docs/wip/`)

`/feature` saves its inner-loop task list to `docs/wip/<feature-slug>.md`, which is tracked by git, so it survives a cleared context. It also holds the feature's links, the decisions behind the plan, and a log of plan changes.
- When resuming a feature, run `task wip:next` first. It prints the file's links, decisions and open tasks, and leaves out the finished tasks and the plan-change log. Read those only when the task needs them.
- To update an item, find it with `grep -n` and read just those lines, rather than the whole file.
- Tick an item only when its refactor is done: `task wip:next` hides ticked items, and with them any **For refactor:** note.
- After each `/red`, `/green`, or `/refactor`, tick off or annotate the item it advanced. If the plan changes, revise the list and note why.
- `/green` leaves a **For refactor:** note under its item, listing the refactoring it spotted. `/refactor` starts from that note and replaces it with what it did.
- Delete the file in the same change that removes the `@wip` tags. `/feature` makes this, and updating the plan, the feature's last task (docs and finish).

## Multi-feature plans (`docs/plans/`)

`TODO.md` is the backlog. When an item needs more than one `/feature`, its plan lives in `docs/plans/<item-slug>.md`, which is also tracked and temporary:
- The plan holds the item's decisions, an ordered list of steps (each one a `/feature`) with their status, and a log of plan changes. The TODO item links to it.
- When resuming the item, run `task wip:next` first if the current step has a `docs/wip/` file. Then read only the plan sections the task needs: find them with `grep -n '^#'`. Read the whole plan when starting a step or changing the plan.
- When a step's feature finishes, mark the step done in the plan, with a one-line summary and links, and log any plan changes.
- Delete the plan in the same change that ticks the TODO item. Move any follow-ups still open into `TODO.md` first.
- A plan made in plan mode lives outside the repo, so copy it into `docs/plans/` before starting the first step.

## Pull requests

Each feature has its own branch, `feature/<feature-slug>`, which `/feature` creates from an up-to-date `main`. Its cycles commit there, and its pull request opens after its last task (`/iterate` on that task, or the end of `/iterate-all`). I review and merge it.

- Open it with `gh pr create --base main --title <title> --body-file <file in the scratchpad>`, then give me its link.
- The title is the feature in the commit style: `feat: <what the feature does>`.
- The description is short: one paragraph on what the feature does, with links to its `SPEC.md` section and feature file(s); the commits (`git log --format='- %h %s' main..HEAD`); the last checks and how far the `@wip` scenarios got; and anything the docs-and-finish task left undone (`@wip` tags still in place, the `docs/wip/` file or plan not yet updated), so I can see it before merging. End it with the attribution lines this session gives for pull requests.
- If `gh pr create` fails (for example, `gh` isn't signed in), say why and leave the branch as it is. Never merge or close a pull request.
