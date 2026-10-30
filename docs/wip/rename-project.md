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
9. [x] Rename's `ParseArgs`: `--name` is required, `--module` is optional and checked, `--dry-run` is accepted, unknown flags are refused.
   - **Red:** `TestParseArgs` in `scripts/rename/args_test.go`, a table test that mirrors init's. It fails on every row: `ParseArgs` is a stub that returns zero `Options` and no error.
     - Compile-only stubs: `scripts/rename/args.go` (`Options{Module, Name, DryRun}`, stub `ParseArgs`) and `scripts/rename/main.go` (an empty `func main`, which a `main` package needs to build). Item 14 fills in `main.go`.
   - Green: `ParseArgs` in `scripts/rename/args.go`, a near copy of init's using pflag. `--name` is required, and `--module` is checked with `module.CheckPath` only when it's given. `task test:go` passes, but it doesn't run `./scripts/rename/` yet; item 14 adds that. No `@wip` scenario can pass yet, since `main.go` is still empty.
   - Refactor: the module path check and its "invalid module path" wrapping are now `project.CheckModule` (`scripts/internal/project/module.go`), used by both tools' `ParseArgs`. It's covered by both `TestParseArgs` tables, so it has no test of its own.
   - Deliberately left:
     - The two flag sets stay separate: they differ in flags and in which are required.
     - The flag help texts aren't printed anywhere, in either tool, because output is discarded. Leave them unless `main.go` (item 14) shows usage.
10. [x] Rename's `NewNames` keeps the current module when `--module` is omitted, and refuses with "nothing to rename" when the new names equal the current ones.
    - **Red (keep module):** `TestNewNamesKeepsTheCurrentModuleWhenNoneIsGiven` in `scripts/rename/names_test.go`. The current names are "Acme" (`example.com/acme/tasks`). Renaming to "Zenith Notes" with no module must give the full set of new names on the current module, with `App` equal to the slug. It fails: `NewNames(opts, current)` in `scripts/rename/names.go` is a stub returning zero `Names`.
    - Green (keep module): `NewNames` slugs the name with `project.Slug` (passing on its error) and always keeps `current.Module`. **`--module` is ignored for now**: no unit test asks for it yet. It needs its own `/red` (new item 10a), or `--module` would silently do nothing.
    - For refactor: moved to item 10a, since it was better done after that.
    - "Nothing to rename" gets its own `/red` next.
    - **Red (nothing to rename):** `TestNewNamesRefusesTheCurrentNames` in `scripts/rename/names_test.go`. Renaming "Acme" (`example.com/acme/tasks`) to "Acme" with no module must fail with an error mentioning "nothing to rename". It fails: `NewNames` returns no error.
      - Comparing the whole `Names` is enough. In the template, renaming to "Starter" would still change the app name (its deploy name differs from its slug), so that run isn't refused, and it shouldn't be.
    - Green (nothing to rename): after building the new names, `NewNames` returns "nothing to rename" if they equal `current`.
    - Refactor: the three `NewNames` tests share a package-level `acme` fixture in `names_test.go`. `NewNames`'s doc comment now says when it fails.
    - Deliberately left: `NewNames` itself. It is short, and its two rules (keep the module, refuse no change) are rename's own.
10a. [x] **New:** rename's `NewNames` uses `--module` when it's given.
    - **Red:** `TestNewNamesUsesTheModuleWhenOneIsGiven` in `scripts/rename/names_test.go`. With current module `example.com/acme/tasks` and `--module example.com/zenith/notes`, the new module must be the given one. It fails: `NewNames` always keeps the current module.
    - Green: `NewNames` uses `opts.Module` when it's given, and `current.Module` otherwise.
    - Refactor: the name derivation is now `project.NewNames(module, name)` in `scripts/internal/project/names.go`. Rename's `NewNames` only picks the module and delegates to it.
      - Init's `NewNames` was a pure wrapper, so `scripts/initproject/names.go` was deleted with plain `rm`, and init's `Run` calls `project.NewNames(opts.Module, opts.Name)` directly. This supersedes item 1's note that `NewNames(Options)` stays in init.
      - `project.NewNames` has no test of its own: init's `Run` tests and rename's `NewNames` tests cover it.
    - Deliberately left: rename keeps its own `NewNames(opts, current)`, since choosing the module and (next) refusing "nothing to rename" are rename's rules, not init's.
11. [x] Rename's `Run` with `--dry-run` reports changed files, lines and path moves, writes nothing, and needs no clean tree.
    - **Red:** `TestRunWithDryRunReportsTheChangesAndWritesNothing` in `scripts/rename/run_test.go`. The project "Acme" has `README.md` (`# Acme` plus an unrelated line), `cmd/acme/main.go` and its record, and the tree is dirty. A dry run to "Zenith" must not fail, and its output must:
      - name `README.md` and show the changed line as `- # Acme` / `+ # Zenith`, but not the unrelated line;
      - show the move from `cmd/acme/main.go` to `cmd/zenith/main.go`.
      - Afterwards, the files, the record and the tree must be as they were. It fails: `Run(dir, files, args, porcelain, out)` in `scripts/rename/run.go` is a stub that writes no output.
      - New test helpers `writeFile` and `readFile` in `scripts/rename/helpers_test.go`, copied from init's.
      - The dirty tree pins "needs no clean tree" now, so item 12's clean-tree guard can't break dry runs unnoticed.
    - Green: `Run` parses the args, reads the record, derives the new names, then for each listed file rewrites its text in memory with `project.Rewrite`. It prints the file name and each changed line as `- old` / `+ new`, and any move from `project.RenamePath` as `old -> new`. It writes nothing.
      - Naive on purpose: it ignores `--dry-run` (every run is a dry run until item 12), doesn't skip `project.json` or files `ShouldRewrite` rejects, and compares lines by index (safe, since a rename never adds or removes newlines).
    - Refactor: the loop body is now `preview(out, dir, rel, old, n)`, and the line comparison is `printChangedLines(out, before, after)`, whose comment explains why lines line up by index. `Run`'s doc comment describes `files` and `out`.
    - Deliberately left:
      - The skips (`project.json`, `ShouldRewrite`) and honouring `--dry-run` are behaviour, so they wait for item 12's tests. A dry run must list exactly what a real run changes.
      - The `- ` / `+ ` format stays as the test pins it.
12. [x] Rename's `Run` does the rename: clean-tree guard, every tracked file except `project.json`, then writes the record.
    - Split into one `/red` each: (a) the rename itself, (b) the clean-tree guard, (c) skipping `project.json` and files `ShouldRewrite` rejects, in both the real run and the dry run.
    - **Red (rename):** `TestRunRenamesTheProject` in `scripts/rename/run_test.go`. Renaming "Acme" to "Zenith" (no `--dry-run`, clean tree) must rewrite `README.md` (brand and env prefix), move `cmd/acme/main.go` to `cmd/zenith/main.go` with its text renamed, remove `cmd/acme`, and record the new names. It fails: `Run` only previews, so `README.md` is unchanged and `cmd/zenith/main.go` doesn't exist.
    - Green (rename): without `--dry-run`, `Run` passes every listed file to `project.RenameFile` with a pass-through `keep` extra, then calls `project.WriteRecord`. With `--dry-run`, it previews as before. `project.json` is still renamed as text first and then overwritten, which is harmless until (c).
    - Refactor: `Run` now only checks and chooses. The two loops are `previewAll(out, dir, files, old, n)` and `renameAll(dir, files, old, n)`, which also writes the record. `Run`'s doc comment mentions `--dry-run`.
    - Deliberately left:
      - The pass-through `keep` closure stays. Letting `RenameFile` take nil would change `project`'s behaviour, and one line of closure isn't worth a `/red`. That supersedes item 7's note that the rename tool needs nil allowed.
      - The two `Run` tests each write the `acme` record first; with only two, a helper isn't worth it yet.
    - **Red (clean tree):** `TestRunRefusesUncommittedChangesAndChangesNothing` in `scripts/rename/run_test.go`. With porcelain ` M README.md`, a real rename to "Zenith" must fail with "commit or stash your changes first" and leave `README.md` and the record unchanged. It fails: `Run` renames anyway.
    - Green (clean tree): after the dry-run branch, `Run` calls `project.CheckClean(porcelain)` before `renameAll`.
    - Refactor: the three `Run` tests start from `newAcmeProject(t)` in `scripts/rename/helpers_test.go`, which makes a temp dir holding the `acme` record. The `acme` fixture moved there from `names_test.go`, since both test files use it.
    - Deliberately left: `Run` itself, which reads cleanly as checks, then a preview or a guarded rename.
    - **Red (skips):** `TestRunWithDryRunListsOnlyFilesARealRunRewrites` in `scripts/rename/run_test.go`. A dry run to "Zenith" with a new module must not list `go.sum` (a lock file `ShouldRewrite` rejects) or `project.json` (written, not rewritten). It fails: the preview lists both, with their changed lines.
      - Only the dry run is tested. A real run already leaves both right: `RenameFile` skips what `ShouldRewrite` rejects, and `WriteRecord` overwrites whatever the text rewrite made of `project.json`. Skipping the record in `renameAll` too is a refactor, like init's.
    - Green (skips): `previewAll` skips `project.RecordFile`, and `preview` lists a file's changed lines only when `ShouldRewrite` accepts it (its move is still listed).
    - Refactor:
      - `renameAll` skips `project.RecordFile` too, as init's loop does, so the preview and the real run touch the same files. `WriteRecord` writes it afresh.
      - `preview` asks `ShouldRewrite` before rewriting.
      - The doc comments say `previewAll` mirrors `renameAll`, and that the record is written, not renamed.
    - Deliberately left: the record skip appears in both loops. Filtering `files` once in `Run` would mean copying the caller's slice, which costs more than the two lines it saves.
13. [x] Rename's `Run` returns `go mod tidy` as a follow-up only when the module changed.
    - **Red:** `TestRunAsksForGoModTidyOnlyWhenTheModuleChanges` in `scripts/rename/run_test.go`, a two-row table. Keeping the module must return no follow-up; `--module example.com/zenith/notes` must return `go mod tidy`. The "module changed" row fails: `Run` always returns nil. The "module kept" row passes already and guards the "only".
    - Green: after `renameAll`, `Run` returns `go mod tidy` when `n.Module != old.Module`, and nil otherwise.
    - Refactor: the command is a named `tidy` var with a comment saying why it follows a module change, and `Run`'s doc comment now matches init's: what `files` and `porcelain` are, that a dry run needs no clean tree, and what it returns.
    - Deliberately left: init has its own `tidy` var. Sharing one through `project` would export a three-word slice for no real gain.
14. [ ] Wiring and docs:
    - `scripts/rename/main.go`;
    - `rename` (outside the template block) and `test:rename` tasks in `Taskfile.yml`;
    - CI steps for the unit tests;
    - README pointer and CLAUDE.md Commands row;
    - `docs/template.md` notes (`project.json`, shared library);
    - check `.golangci.yml`.

    Then remove `@wip`, tick TODO.md, and delete this file.

## Changes to the plan

- Item 10a was added during item 10's green. Keeping the current module was the minimum to pass, so using a given `--module` needs a test of its own.
- Item 7a was added while planning item 7. Changing the module path can leave Go imports out of order, which `gofmt` and lint reject. Init has the same latent bug; it hasn't shown up because `example.com/...` happens to sort where the template's module did.
- Item 8a was added at your request while doing item 8: refuse names with double quotes or backslashes. Item 8's write test changed its example to match.
- Item 7c was split out of 7b's refactor. Moving init onto `RenameFile` makes init format Go files, which changes the output for unformatted input, so it gets a red/green rather than a refactor.
- Item 7b was split out of 7a's refactor note. Moving init onto `RenameFile` needs two behaviours `RenameFile` doesn't have yet, so it can't be a pure refactor.
- Item 4: only the brand is tested for regexp metacharacters. A slug can't contain any, because `Slug` only produces `[a-z0-9-]`.
