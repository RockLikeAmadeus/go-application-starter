# WIP: Rename project (`task rename`)

Temporary: delete this file in the change that removes the `@wip` tags.

- **Spec:** [`docs/rename.md`](../rename.md)
- **Scenarios:** [`scripts/rename/features/rename.feature`](../../scripts/rename/features/rename.feature) (`@wip`), plus "The project remembers its names" in [`scripts/initproject/features/init.feature`](../../scripts/initproject/features/init.feature)
- **Run them:** `go test ./scripts/rename/acceptance/ -count=1 --godog.tags=@wip` and `go test ./scripts/initproject/acceptance/ -count=1 --godog.tags=@wip`. Neither is in `task accept`, because they are tooling, not product.
- **Origin:** TODO.md, "Do first" #2.

## Decisions

- **The current names live in `project.json`** at the repo root: `{module, name, slug, app}`. The environment variable prefix and the web package name (`<slug>-web`) are derived, not stored.
  - `app` is the deploy name, which differs from the slug only in the template (`go-application-starter` vs `starter`). After init or a rename, `app == slug`.
  - Init reads the template's names from this file instead of hardcoded constants.
  - Both tools write the record explicitly (JSON marshal) and skip it during text rewriting.
- **Shared code lives in `scripts/internal/project`**, which survives init. Init deletes only `scripts/initproject/` and `docs/template.md`. That library and the rename tool must contain **no literal template names**, because a rename rewrites every tracked file, including its own source.
- **Replacement is global, by whole word**, as in init. To guard against an old slug that is a common word, there is `--dry-run`, which lists the changes and needs no clean tree, plus the clean-tree guard for real runs.
- **`--name` is required. `--module` is optional** and keeps the current module path when omitted.
- **An unchanged module path must survive intact even when it contains the old slug.** The template's module ends in `go-application-starter`. So the bare-word pass must not touch the module path. The scenario "The module path stays unless I ask to change it" covers this.
- **Rename scenarios never run init.** They copy the project as it stands and read the old names from its `project.json`, so they keep working after init has deleted itself.

## Task list

1. [x] **Refactor** (no behaviour change): move `Slug`, `EnvPrefix`, `Names`, `ShouldRewrite`, `CheckClean`, `FollowUp` and git listing from `scripts/initproject` into `scripts/internal/project`, with their tests. Init stays green.
   - Done. The command plumbing (`Execute`, `Installed`) moved too, as `Git`/`GitFiles`.
   - `NewNames(Options)` stays in init for now, because it takes init's `Options`.
   - `task test:go` and CI's unit step now also run `./scripts/internal/...`, outside the template block.
2. [ ] `Names` gains `App`, and `Rewrite(text, from, to)` replaces one project's names with another's, including a deploy app name that differs from the slug.
3. [ ] `Rewrite` leaves an unchanged module path intact even when it contains the old slug.
4. [ ] `Rewrite` quotes regexp metacharacters in the brand and slug.
5. [ ] `RenamePath(path, from, to)` moves `cmd/<old slug>/` to `cmd/<new slug>/`.
6. [ ] `ReadRecord` / `WriteRecord` read and write `project.json` (a round trip gives the same names). Add `project.json` with the template's names.
7. [ ] `RenameFile(dir, rel, from, to, extra)`: move init's `renameFile` here, with an optional extra transform (init uses it to strip blocks).
8. [ ] Init's `Run` reads its old names from the record and writes the new record. Delete the `template*` constants.
9. [ ] Rename's `ParseArgs`: `--name` is required, `--module` is optional and checked, `--dry-run` is accepted, unknown flags are refused.
10. [ ] Rename's `NewNames` keeps the current module when `--module` is omitted, and refuses with "nothing to rename" when the new names equal the current ones.
11. [ ] Rename's `Run` with `--dry-run` reports changed files, lines and path moves, writes nothing, and needs no clean tree.
12. [ ] Rename's `Run` does the rename: clean-tree guard, every tracked file except `project.json`, then writes the record.
13. [ ] Rename's `Run` returns `go mod tidy` as a follow-up only when the module changed.
14. [ ] Wiring and docs:
    - `scripts/rename/main.go`;
    - `rename` (outside the template block) and `test:rename` tasks in `Taskfile.yml`;
    - CI steps for the unit tests;
    - README pointer and CLAUDE.md Commands row;
    - `docs/template.md` notes (`project.json`, shared library);
    - check `.golangci.yml`.

    Then remove `@wip`, tick TODO.md, and delete this file.

## Changes to the plan

_None yet._
