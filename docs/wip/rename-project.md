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
2. [x] `Names` gains `App`, and `Rewrite(text, from, to)` replaces one project's names with another's, including a deploy app name that differs from the slug.
   - Green: `scripts/internal/project/rewrite.go`. It doesn't use `regexp.QuoteMeta` yet; that's left for item 4.
   - Refactor: init's `Rewrite` now delegates to `project.Rewrite(text, template, n)`. The template's names are a single `template` `project.Names` value in `scripts/initproject/rewrite.go`, which item 8 replaces with `ReadRecord`. Init's `NewNames` sets `App = slug`.
3. [x] `Rewrite` leaves an unchanged module path intact even when it contains the old slug.
   - Green: after the compound pass, the text is split on `to.Module`, and the brand and slug are replaced only in the pieces between those occurrences.
   - Refactor: the patterns are now compiled once (`wholeWord`), and the splitting lives in a small `outside(text, keep, f)` helper. `wholeWord` is where item 4's `regexp.QuoteMeta` goes.
4. [x] `Rewrite` quotes regexp metacharacters in the brand. (The slug needs no test: `Slug` only produces `[a-z0-9-]`, and none of those are special.)
   - Green: `wholeWord` wraps the name in `regexp.QuoteMeta`. Tested by `TestRewriteTreatsTheOldBrandLiterally`.
5. [x] `RenamePath(path, from, to)` moves `cmd/<old slug>/` to `cmd/<new slug>/`.
   - Green: `project.RenamePath`. Refactor: init's `RenamePath` now delegates to it with the `template` names, so init's `rewrite.go` holds only those names and two one-line wrappers.
6. [x] `ReadRecord` / `WriteRecord` read and write `project.json` (a round trip gives the same names). Add `project.json` with the template's names.
   - Green: `scripts/internal/project/record.go` and the repo-root `project.json`, both done.
   - The init `@wip` step "records module … and name …" now passes, but only by accident: init's text rewrite happens to fix up `project.json`. Item 8 makes init write it explicitly. "Can still be renamed" fails until `scripts/rename` exists.
7. [x] `RenameFile(dir, rel, from, to, extra)`: move init's `renameFile` here, with an optional extra transform (init uses it to strip blocks).
   - **Red:** `TestRenameFileRewritesAndMovesAFile` in `scripts/internal/project/renamefile_test.go`, against a stub in `renamefile.go`. `extra` gets `(rel, original text)` before the rewrite. The test file is a shell script, so Go formatting doesn't come into it.
   - Green: `RenameFile` in `scripts/internal/project/renamefile.go`.
   - Refactor: the file handling is now `project.RewriteFile(dir, rel, newRel, transform)`, which `RenameFile` builds on, and `RemoveEmptyParents` is exported.
     - Init's `renameFile` is a thin wrapper over `RewriteFile`, and its `removeAll` uses the shared `RemoveEmptyParents`, so init has no file-moving code of its own.
   - Still to do: init can't call `RenameFile` yet, because it runs goimports *after* the rewrite and `extra` runs *before*. Once 7a formats Go files after the rewrite, init's goimports can move into `extra`. At that point, add tests for a `ShouldRewrite` skip and for wrapping errors with `rel`, and switch init over.
   - `RenameFile` still calls `extra` unconditionally. The rename tool needs nil to be allowed, so test it there.
7a. [x] **New:** after a rewrite, Go files are formatted (`go/format`, which sorts imports), because a changed module path can reorder imports.
    - Green: after the rewrite, `RenameFile` runs `.go` files through `format.Source`. Tested by `TestRenameFileKeepsGoImportsSorted`.
    - Refactor: init's wrong comment in `transform` now says only removal is tidied there, and that a module path which sorts differently can still unsort imports, which `project.RenameFile` handles. Init's latent bug itself is left: fixing it changes behaviour, and it goes away when init switches to `RenameFile`.
    - Deliberately left:
      - **Switching init to `RenameFile`** isn't a pure refactor. `RenameFile` lacks the `ShouldRewrite` skip and the `rel` error wrapping, so each needs a `/red` first (new item 7b). After that, init's block removal and goimports (for `--remove-sample`) move into `extra`. A binary-looking `.go` file is handled once the `ShouldRewrite` skip comes first.
      - **An `isGo` helper:** not done. There's one check in each package, and sharing it would mean exporting a one-liner.
    - Init's comment "Renaming alone never breaks formatting" is wrong for a module such as `zenith.dev/x`, which would sort after `github.com/spf13/cobra`.
    - It goes in `RenameFile`, so both tools get it.
    - Init keeps its goimports pass (unused-import removal) only for `--remove-sample`.
7b. [x] **New:** `RenameFile` leaves files `ShouldRewrite` skips untouched (one `/red`), and wraps its errors with `rel` (another `/red`). Moving init onto `RenameFile` turned out to change behaviour, so it moved to 7c.
    - **Red (skip):** `TestRenameFileLeavesSkippedFilesUntouched` in `renamefile_test.go`. A `go.sum` that holds the old module path must come back unchanged, and `extra` must not be called for it. It fails: `RenameFile` rewrites it, and calls `extra`.
    - Green (skip): `RenameFile`'s transform returns the content unchanged when `ShouldRewrite` is false, before `extra`, the rewrite and formatting. A skipped file is still moved by `RenamePath`, as in init.
    - **For refactor:** none yet. The `ShouldRewrite` check is now duplicated in init's `transform`, but that goes away when init switches to `RenameFile`, after the error-wrapping `/red`. A binary file named `.go` is now skipped before formatting, which settles that part of 7a's note.
    - **Red (errors):** `TestRenameFileNamesTheFileInItsErrors`. When `extra` fails on `docs/guide.md`, the error must still wrap the cause (`errors.Is`) and name the file. It fails: the error is the bare `"unbalanced block"`.
    - Green (errors): `RenameFile` wraps `extra`'s error as `fmt.Errorf("%s: %w", rel, err)`.
    - Refactor: I tried switching init onto `RenameFile`, with an `extra` that removes blocks and runs goimports. That turned `TestRunRemovesTheSampleOnlyWhenAsked/kept_by_default` red, so I reverted it, and the code is as it was.
      - The cause: `RenameFile` runs `format.Source` on every `.go` file, which init doesn't do. The test's `bootstrap.go` fixture isn't gofmt-formatted, and formatting adds a blank line.
      - Formatting is the fix for init's latent import-order bug, so the switch is a behaviour change. It needs its own red/green (item 7c).
    - Still open: a `format.Source` failure inside `RenameFile` isn't wrapped with `rel`; init wraps all of its transform errors. Add a `/red` only if it matters. Real files are gofmt-clean, which lint enforces.
7c. [ ] **New:** init formats Go files after renaming, so a new module path that sorts differently keeps imports in order. `/red` in `scripts/initproject/run_test.go`: a module such as `zenith.dev/x` must leave `github.com/...` imports sorted first. `/green`: switch init's `renameFile` onto `project.RenameFile`, with `extra` doing block removal plus goimports for `--remove-sample`, and init's own `ShouldRewrite` check, `Rewrite` call and wrapping removed. The `kept_by_default` fixture in `TestRunRemovesTheSampleOnlyWhenAsked` must become gofmt-clean in the same step: its expected output assumes no formatting.
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

- Item 7a was added while planning item 7. Changing the module path can leave Go imports out of order, which `gofmt` and lint reject. Init has the same latent bug; it hasn't shown up because `example.com/...` happens to sort where the template's module did.
- Item 7c was split out of 7b's refactor. Moving init onto `RenameFile` makes init format Go files, which changes the output for unformatted input, so it gets a red/green rather than a refactor.
- Item 7b was split out of 7a's refactor note. Moving init onto `RenameFile` needs two behaviours `RenameFile` doesn't have yet, so it can't be a pure refactor.
- Item 4: only the brand is tested for regexp metacharacters. A slug can't contain any, because `Slug` only produces `[a-z0-9-]`.
