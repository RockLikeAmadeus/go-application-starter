# Init asks for the app's description and publisher

TODO.md item: **Init asks for the app's description and publisher.**

- Spec: [docs/template.md, "Make it mine"](../template.md#make-it-mine). Init is template tooling, so its spec lives there, not in SPEC.md.
- Scenarios: [`scripts/initproject/features/init.feature`](../../scripts/initproject/features/init.feature), the three `@wip` ones. Run with `task test:init -- --godog.tags=@wip`.

## Decisions

- **Flags, not prompts** (init runs non-interactively in its acceptance tests and scripts), and **required**, like `--module` and `--name`, so no project starts with the placeholders (chosen 2026-11-03).
- **`--description`, `--publisher`, `--email`.** No separate `--maintainer`: the Linux maintainer is `<publisher> <<email>>`. The Windows `CompanyName` is the publisher, and `LegalCopyright` is `© <publisher>` (no year, which would go stale).
- **Windows `FileDescription` stays the brand name** (rename already writes it): Task Manager shows it as the process's name, so a sentence would look odd. The description goes only to nfpm's `description` and the desktop file's `Comment`.
- **`Categories=Utility;` stays a manual step** for template users (there's no good flag for a freedesktop category list).
- **Validation:** the description and publisher refuse double quotes, backslashes and line breaks (they land in YAML, JSON, a desktop file and makensis's `-D` arguments, see the Taskfile's `package:desktop:windows`), the same rule as `project.Slug` has for the name. The publisher also refuses `<` and `>` (they would break `Name <email>`). The email must be a bare address (`net/mail`), not `Name <addr>`.
- **The template's files carry placeholder lines that init replaces by key**: nfpm.yaml gains a `description:` and app.desktop a `Comment=` (so the template's own packages are described too). Their text must not mention the template's name, or `task rename` and "no trace of the template's name" interfere. A missing key is a template bug, and init fails before touching a file (every check runs first, `Run`'s rule).
- The edits happen **after** the rename pass, so a description that happens to contain the template's name isn't rewritten. They live in `scripts/initproject` (init-only; rename doesn't take these details).
- The acceptance steps that don't name the details pass sample ones (`packageDetails` in `acceptance_test.go`). So until task 1 is green, **every** init scenario fails with "unknown flag: --description", not just the `@wip` ones.

## Tasks

1. [x] `ParseArgs` accepts `--description`, `--publisher` and `--email` into `Options`, and requires each (`--description is required`, …). Update `run_test.go`'s argument lists to pass them.
   - Red: `TestParseArgsRequiresEachPackageDetail` (and the updated `TestParseArgs` happy cases) failed with "unknown flag: --description".
   - Green: the three flags and their "is required" checks. Tests pass them with `withDetails`.
   - Refactor: one loop over `required` replaces the five "is required" checks; `withDetails`/`wantDetails` moved to `helpers_test.go`, used by three test files.
2. [x] `ParseArgs` refuses double quotes, backslashes and line breaks in the description and publisher (`description must not contain double quotes or backslashes`), `<` or `>` in the publisher, and an email that isn't a bare address (`invalid email address`).
   - Red: `TestParseArgsRefusesUnsafePackageDetails` failed: ParseArgs accepted every bad value (error = nil). Line breaks say `… must be one line`.
   - Green: ParseArgs checks the description, publisher and email after the module.
   - Refactor: `checkDetails` holds the detail checks, with `checkLine` for the description's and publisher's shared rule (fixing the stray backslash). Slug keeps its own check: refusing line breaks there would change the name's rule.
3. [x] A pure function sets nfpm.yaml's `description:` and `maintainer:` lines (quoted so any allowed text is valid YAML), keeping everything else; it fails when either key is missing.
   - Red: `TestSetNFPMDetails` failed: the stub returned the text unchanged and no error for a missing key.
   - Green: `SetNFPMDetails(text, Options)` in `packaging.go` rewrites both top-level lines, quoted.
   - Refactor: `replaceLine(text, prefix, line)` holds the find-and-replace (ready for task 4's `Comment=`); doc comments added.
4. [x] A pure function sets the desktop file's `Comment=` line; it fails when the key is missing.
   - Red: `TestSetDesktopComment` failed: the stub returned the text unchanged and no error for a missing key.
   - Green: `SetDesktopComment(text, Options)` sets `Comment=` through `replaceLine`, leaving `Comment[xx]=` alone.
   - Refactor: doc comment; `runSetterTests`/`setterTest` hold the table loop both setter tests share (ready for task 5).
5. [x] A pure function sets info.json's `CompanyName` and `LegalCopyright` (`© <publisher>`), keeping its layout, `FileDescription` and the `@VERSION@` placeholders; JSON-escaped, no HTML escaping (`&` stays `&`). Fails when a key is missing.
   - Red: `TestSetInfoJSON` failed: the stub returned the text unchanged and no error for a missing key.
   - Green: `SetInfoJSON(text, Options)` sets both values through `setJSONString` (indent and comma kept, encoder with HTML escaping off).
   - Refactor: nothing worth changing (doc comments already written).
6. [x] `Run` applies 3–5 to the three files after renaming, and refuses before touching any file when one is missing or lacks a key. Add the placeholder `description:` and `Comment=` lines to the template's files.
   - Red: `TestRunSetsThePackageDetails` failed: files unchanged after Run, and no error for a missing `Comment=`.
   - Green: `setPackageDetails` checks the three files before anything is touched, then writes after the rename; placeholder lines added to nfpm.yaml and app.desktop; the other Run tests get minimal packaging files (`writePackagingFiles`).
   - Refactor: nothing worth changing (the `write` flag is small and documented).
7. [ ] Docs and steps: README's "Using this as a template" command, the Taskfile `init` task's `desc`, TODO.md's template-user steps (drop the nfpm description/maintainer and info.json steps, keep `Categories` in the desktop file one), `test/packaging/nsis_test.go`'s "placeholders" comment, and the desktop packaging plan's mentions of these placeholders. Remove the `@wip` tags, run `task test:init`, `task test:rename` and `task check`, tick the TODO item, delete this file.

## Changes to the plan

(none yet)
