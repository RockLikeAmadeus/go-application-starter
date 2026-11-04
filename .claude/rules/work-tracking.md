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
- When resuming a feature, read its `docs/wip/` file first.
- After each `/red`, `/green` or `/refactor`, tick off or annotate the item it advanced. If the plan changes, revise the list and note why.
- `/green` leaves a **For refactor:** note under its item, listing the refactoring it spotted. `/refactor` starts from that note and replaces it with what it did.
- Delete the file in the same change that removes the `@wip` tags.

## Multi-feature plans (`docs/plans/`)

`TODO.md` is the backlog. When an item needs more than one `/feature`, its plan lives in `docs/plans/<item-slug>.md`, which is also tracked and temporary:
- The plan holds the item's decisions, an ordered list of steps (each one a `/feature`) with their status, and a log of plan changes. The TODO item links to it.
- When resuming the item, read the plan first, then the current step's `docs/wip/` file, if one exists.
- When a step's feature finishes, mark the step done in the plan, with a one-line summary and links, and log any plan changes.
- Delete the plan in the same change that ticks the TODO item. Move any follow-ups still open into `TODO.md` first.
- A plan made in plan mode lives outside the repo, so copy it into `docs/plans/` before starting the first step.
