# Renaming the project

Projects often start under a working title. `task rename` changes the project's names later, in one step, as often as you like. It does the same renaming `task init` did when the project was made from the template, but it is not deleted after it runs.

```sh
task rename -- --name "Zenith Notes" --dry-run                     # see what would change
task rename -- --name "Zenith Notes"                               # rebrand, keep the Go module path
task rename -- --name "Zenith Notes" --module github.com/you/zenith # rebrand and move the module
```

**Purpose:** rebrand without hunting down every place the old name appears.

**Rules**
- The project's current names are recorded in `project.json` at the repository root: the Go module path, the brand name, the slug, and the deploy app name. The tool reads the old names from there and writes the new ones back. Don't edit it by hand.
- `--name` (the new brand name) is required. `--module` is optional; without it the Go module path stays as it is.
- The new names are derived from the brand name exactly as `task init` derives them:
  - The **slug** is the brand name in lower case, with runs of anything other than letters and digits turned into single hyphens: "Zenith Notes" becomes `zenith-notes`. It names the command-line tool (`cmd/<slug>`), the Docker image, the deploy configs and the web package (`<slug>-web`).
  - The **environment variable prefix** is the slug in upper case, with underscores for hyphens: `ZENITH_NOTES_STORAGE`.
- No trace of the old names is left in any tracked file: not the old brand, slug or environment variable prefix, and not the old module path when `--module` is given.
- Names are replaced as whole words everywhere. If the old slug is an ordinary word (a project called "Tasks" has the slug `tasks`), unrelated uses of that word are replaced too. **Run with `--dry-run` first.** It lists every change, file by file and line by line, and changes nothing.
- It refuses to run, and changes nothing, when:
  - the input is invalid (`--name` is missing, the module path is invalid, or the name has no letters or digits);
  - the new names are the same as the current ones;
  - the working tree has uncommitted changes. That way, `git` can always undo a run. A dry run doesn't need a clean tree.

**Examples**
- "Acme" → `--name "Zenith Notes"` → brand "Zenith Notes", command `zenith-notes`, variables `ZENITH_NOTES_…`, module unchanged.
- "Acme" → `--name "Zenith"` → `--name "Orbit"`: renaming twice leaves no trace of "Acme" or "Zenith".
- `--name "Acme"` when the project is already called "Acme" → refused: nothing to rename.
- Uncommitted changes → refused: commit or stash first.

Afterwards, look over `git diff`, run `task check`, and commit. To undo a run, use `git reset --hard && git clean -fd`.

## Specification

The behaviour above is specified in [`scripts/rename/features/rename.feature`](../scripts/rename/features/rename.feature). Run it with `task test:rename`. Each scenario runs the tool on a throwaway copy of the project, so it is slow and is not part of `task test`.
