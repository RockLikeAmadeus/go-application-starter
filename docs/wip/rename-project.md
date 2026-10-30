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
    - Refactor (skip): none was needed. Init's duplicate `ShouldRewrite` check went away when init switched to `RenameFile` in 7c.
    - **Red (errors):** `TestRenameFileNamesTheFileInItsErrors`. When `extra` fails on `docs/guide.md`, the error must still wrap the cause (`errors.Is`) and name the file. It fails: the error is the bare `"unbalanced block"`.
    - Green (errors): `RenameFile` wraps `extra`'s error as `fmt.Errorf("%s: %w", rel, err)`.
    - Refactor: I tried switching init onto `RenameFile`, with an `extra` that removes blocks and runs goimports. That turned `TestRunRemovesTheSampleOnlyWhenAsked/kept_by_default` red, so I reverted it, and the code is as it was.
      - The cause: `RenameFile` runs `format.Source` on every `.go` file, which init doesn't do. The test's `bootstrap.go` fixture isn't gofmt-formatted, and formatting adds a blank line.
      - Formatting is the fix for init's latent import-order bug, so the switch is a behaviour change. It needs its own red/green (item 7c).
    - Still open: a `format.Source` failure inside `RenameFile` isn't wrapped with `rel`; init wraps all of its transform errors. Add a `/red` only if it matters. Real files are gofmt-clean, which lint enforces.
7c. [x] **New:** init formats Go files after renaming, so a new module path that sorts differently keeps imports in order. `/red` in `scripts/initproject/run_test.go`: a module such as `zenith.dev/x` must leave `github.com/...` imports sorted first. `/green`: switch init's `renameFile` onto `project.RenameFile`, with `extra` doing block removal plus goimports for `--remove-sample`, and init's own `ShouldRewrite` check, `Rewrite` call and wrapping removed. The `kept_by_default` fixture in `TestRunRemovesTheSampleOnlyWhenAsked` must become gofmt-clean in the same step: its expected output assumes no formatting.
    - **Red:** `TestRunKeepsGoImportsSortedWhenTheModuleSortsDifferently` in `scripts/initproject/run_test.go`. `cmd/starter/main.go` imports the template's module and `github.com/spf13/cobra`. Init to `zenith.dev/notes` must leave cobra first. It fails: `zenith.dev/notes/internal/cli` stays in the template module's old place, ahead of cobra.
    - Green: init's `renameFile` now calls `project.RenameFile`, with a new `removeBlocks(rel, text, removeSample)` as `extra`. The old `transform` and init's own error wrapping are gone.
      - With your approval, the `kept_by_default` fixture in `TestRunRemovesTheSampleOnlyWhenAsked` gained a blank line, so it's gofmt-clean. It had assumed init never formats files.
      - Init's non-WIP acceptance first failed on this file: a note here quoted the sample block's end marker, and init treats every tracked file's markers as real. I reworded the note. **Never write those marker strings in prose in tracked files.**
    - Refactor:
      - `renameFile` is gone. `Run` calls `project.RenameFile` directly, with a `strip` closure built once, before the loop.
      - The private `removeBlocks` is now `stripBlocks`, so it no longer reads like the exported `RemoveBlocks`.
    - Deliberately left:
      - Init's `Rewrite`/`RenamePath` wrappers in `rewrite.go` are now used only by `rewrite_test.go`. They, their tests and `template` go together in item 8, when `ReadRecord` replaces `template`. Deleting them now would drop the template-specific cases (module path, deploy app name) before item 8 has a replacement.
      - A `format.Source` failure in `RenameFile` still isn't wrapped with `rel`. That's a behaviour change, so it would need a `/red`.
8. [x] Init's `Run` reads its old names from the record and writes the new record. Delete the `template*` constants.
   - **Red (read):** `TestRunReadsTheOldNamesFromTheProjectRecord` in `scripts/initproject/run_test.go`. `project.json` names "Old Thing" (`example.com/old/thing`, `old-thing`), and `README.md` mentions its brand, env prefix, module and command path. Init to "Acme" must replace all of them. It fails: init still uses the hardcoded `template` names, so `README.md` is unchanged.
   - Heads-up for `/green`: the other `Run` tests write no `project.json`. Once `Run` reads it, they'll fail with a missing-file error.
     - The simplest fix is a test helper that writes a `project.json` with the template's names into each test's dir. That's a test change, so it needs your approval.
     - The other route, falling back to the template's names when the file is missing, is behaviour no test demands, and it would keep the `template` names alive.
   - Green (read): after its checks, `Run` calls `project.ReadRecord(dir)` and passes those names to `project.RenameFile` instead of `template`. A missing record is an error.
     - With your approval, a new `writeTemplateRecord(t, dir)` helper in `run_test.go` writes the template's `project.json`. Every `Run` test that gets past the checks calls it, and so does `TestMainRunsInAGitRepository` (before its git commit).
     - Init's non-WIP acceptance passes. Init's `@wip` scenario passes 3 of its 4 steps; only "can still be renamed" fails, since `scripts/rename` doesn't exist yet.
   - Refactor:
     - The hardcoded template names are gone. `scripts/initproject/rewrite.go` (`template` and the `Rewrite`/`RenamePath` wrappers) and its `rewrite_test.go` were deleted with plain `rm`.
     - Their one case not already covered, "whole words only", moved into `project`'s `TestRewriteReplacesOneProjectsNamesWithAnothers`, as "Old Notesy and old-notesy stay".
     - The test helpers `writeFile`, `readFile` and `writeTemplateRecord` moved from `run_test.go` to a new `helpers_test.go`, since `main_test.go` uses them too.
   - Deliberately left: init still rewrites `project.json` as text, which happens to give the right record. The next `/red` makes it write the record explicitly and skip the file.
   - **Red (write):** `TestRunRecordsTheNewNames` in `scripts/initproject/run_test.go`. The old record names a project "App", so its names are also JSON keys, and `project.json` is listed. After init to "Acme Tasks", `project.ReadRecord` must give the new names. It fails: the text rewrite renames the `"app"` key too, so the app name is lost.
     - First draft: the test used the name `Acme "Pro"`, where the text rewrite leaves the quote unescaped. You then chose to reject quotes and backslashes in names (item 8a), so I switched to a case that stays valid.
   - Green (write): after renaming the files, `Run` calls `project.WriteRecord(dir, n)`, which overwrites whatever the text rewrite made of `project.json`. Init's non-WIP acceptance passes. The init `@wip` scenario passes 3 of 4 steps, as before.
   - Refactor:
     - The record's file name is now exported as `project.RecordFile`, and init's loop skips that file instead of renaming its text and then overwriting it. Item 12's rename loop can use it too.
     - `Run`'s doc comment now says it replaces the names in `project.json` and records the new ones there.
   - Writing the new record is a separate `/red` after this one.
8a. [x] **New:** names with double quotes or backslashes are refused, because they would break JSON, Go strings and other quoted text the rename touches.
    - **Red:** two rows in `project`'s `TestSlug`: `Acme "Pro"` and `Acme\Pro`, each wanting "name must not contain double quotes or backslashes".
    - Green: `Slug` refuses them first. Init and rename both get this, since both derive the slug from the name.
    - Apostrophes stay allowed ("Bob's Tasks"). The brand only appears in double-quoted strings, YAML plain scalars, HTML and JSX, where an apostrophe is safe.
    - Refactor:
      - `Slug`'s doc comment explains the refusal.
      - `docs/rename.md` and `docs/template.md` list it among the invalid inputs.
    - Deliberately left: an example row in the rename feature's "refuse invalid input" outline. That adds an acceptance check, so it belongs with item 14's wiring, or a `/feature` touch-up if you want it.
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
- Item 8a was added at your request while doing item 8: refuse names with double quotes or backslashes. Item 8's write test changed its example to match.
- Item 7c was split out of 7b's refactor. Moving init onto `RenameFile` makes init format Go files, which changes the output for unformatted input, so it gets a red/green rather than a refactor.
- Item 7b was split out of 7a's refactor note. Moving init onto `RenameFile` needs two behaviours `RenameFile` doesn't have yet, so it can't be a pure refactor.
- Item 4: only the brand is tested for regexp metacharacters. A slug can't contain any, because `Slug` only produces `[a-z0-9-]`.
