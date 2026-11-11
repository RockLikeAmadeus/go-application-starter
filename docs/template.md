# Using this template

This repository is a template: a starter project you copy when you begin a new app. A fresh copy still carries the template's own name ("Starter") and a small demo capability, **Item** (a simple list you can add to, rename, and delete from). `task init` cleans that up in one step:

- **Without `--remove-sample`:** everything is renamed from the template's name to yours.
- **With `--remove-sample`:** the Item demo is deleted as well, leaving an empty app that still builds, passes its tests, and is ready for your first `/feature`.

Either way, the tool then **deletes itself**, because you only need it once. To rename the project again later, use `task rename` (see [docs/rename.md](rename.md)), which stays.

> This page and the init tool describe the template, not your product. Product intent lives in `SPEC.md`. The init tool deletes this page, and itself, once it has run.

## Make it mine

```sh
task init -- --module github.com/you/tasks --name "Acme Tasks" \
  --description "Tasks for busy teams" --publisher "Acme Ltd" --email dev@acme.example   # keep the sample Item
task init -- --module github.com/you/tasks --name "Acme Tasks" \
  --description "Tasks for busy teams" --publisher "Acme Ltd" --email dev@acme.example \
  --remove-sample
```

**Purpose:** start a real project without hunting down every place the template's own name appears, or the placeholders its desktop packages carry.

**Rules**
- `--module` (the Go module path) and `--name` (the brand name people see) are both required.
- So are the details the desktop packages show, so no project ships the template's placeholders:
  - `--description`: one line about the app. It describes the Linux packages (`apt show`, software centres) and the app's menu entry (the desktop file's `Comment`).
  - `--publisher`: the person or company that publishes the app. It is the Windows program's company, its copyright reads `© <publisher>`, and with `--email` it maintains the Linux packages: `Acme Ltd <dev@acme.example>`.
  - `--email`: where people reach the publisher about the packages.
  - The Windows program's description stays the brand name: Task Manager shows it as the process's name.
  - The menu entry's category (`Categories=Utility;`) is still yours to choose by hand.
- Every other name comes from the brand name's **slug**: lower case, with runs of anything other than letters and digits turned into single hyphens. "Acme Tasks" becomes `acme-tasks`.
  - The brand name replaces "Starter" in the UI and the docs.
  - The slug names the command-line tool (`cmd/<slug>`), the Docker image, the deploy configs, and the web package (`<slug>-web`).
  - The slug in upper case, with underscores for hyphens, prefixes the environment variables: `ACME_TASKS_STORAGE`.
  - The slug without its hyphens ends the app ID that the Android app and the desktop app's Linux window use: `com.example.acmetasks`.
- No trace of the template's name is left: not its module path, not "starter", and not `STARTER_`.
- The sample **Item** capability stays unless `--remove-sample` is given. With it, Item is removed from every layer and from `SPEC.md`, and the project still builds and passes its tests.
  - What the front ends share stays, so your first screen doesn't have to rebuild it: the terminal app's error and selection styles, and the web app's error message. The project still passes the Go linters.
- The tool removes itself when it finishes: its code, this page, and the `init` task.
- It refuses to run, and changes nothing, when:
  - the input is invalid (a flag is missing, the module path is invalid, the name has no letters or digits, the name, description, or publisher contains a double quote, a backslash, or a line break, the publisher contains `<` or `>`, or the email isn't an address);
  - the working tree has uncommitted changes. That way, `git` can always undo a run.

**Examples**
- `--name "Acme"` → brand "Acme", command `acme`, variables `ACME_…`.
- `--name "Acme Tasks"` → brand "Acme Tasks", command `acme-tasks`, variables `ACME_TASKS_…`.
- `--remove-sample` → the Items screens go, but the TUI's `errorStyle` and `selectedStyle` and the web app's `ErrorMessage` stay.
- `--publisher "Acme Ltd" --email dev@acme.example` → Linux maintainer `Acme Ltd <dev@acme.example>`, Windows company "Acme Ltd", copyright "© Acme Ltd".
- `--name "!!!"` → refused: the name needs a letter or digit.
- `--email "not an address"` → refused.
- Uncommitted changes → refused: commit or stash first.

## What happens when you run it

1. **Checks.** It makes sure both flags are given and valid, and that you have no uncommitted changes. If anything is wrong, it stops before touching a file.
2. **Renames.** It replaces the Go module path, the brand, the command-line tool's name, and the `STARTER_…` environment variables in every tracked file, and moves `cmd/starter` to `cmd/<slug>`. It records the new names in `project.json`, where `task rename` finds them later.
3. **Describes the desktop packages.** It writes the description, publisher, and email into `build/desktop/linux/nfpm.yaml`, `build/desktop/linux/app.desktop` and `build/desktop/windows/info.json`.
4. **Removes the sample** (only with `--remove-sample`). It deletes the Item files and the Item parts of shared files, then removes any folders that leaves empty.
5. **Removes itself.** Its code, this page, and the `init` and `test:init` tasks go.
6. **Tidies up.** It runs `go mod tidy`. After removing the sample, it first regenerates the API code from the trimmed `api/openapi.yaml`, then regenerates the web client's types with `pnpm`.
   - If pnpm is not installed, or fails (for example, `web/node_modules` is missing or Node is too old), it does not stop. It prints a note telling you to run `task gen` once that is fixed.

Afterwards, look over `git diff`, run `task check`, and commit. If you don't like the result, `git reset --hard && git clean -fd` puts everything back.

## Keeping the template ready for init

This section is for people changing the template itself. The tool can only remove what it knows about, so sample and template-only code is marked in two ways.

**Whole files and folders** that exist only for the sample are listed in `SampleFiles` in `scripts/initproject/manifest.go`. When you add a sample-only file, add it there too, or it will be left behind.

**Parts of shared files** are wrapped in marker comments. Each marker goes on a line of its own, in whatever comment syntax the file uses:

```go
// sample:start
a.Items = app.NewItemService(items, system.Clock{}, system.UUIDs{})
// sample:end
```

- `sample:start` / `sample:end` wrap code that is removed with `--remove-sample`. You'll find them in the startup code, the API spec, the CLI, the TUI, the web app's router and sidebar, `SPEC.md`, `CLAUDE.md`, `.claude/rules/tui.md`, and the README.
- `template:start` / `template:end` wrap things that only make sense in the template, such as the `task init` task and the README's "Using this as a template" section. These are always removed.
- Imports don't need markers: in Go files, unused imports are tidied away automatically.
- A missing or mistyped marker makes the tool fail with the line number, rather than silently deleting the rest of a file.

**Code shared with `task rename`** lives in `scripts/internal/project`, which survives init. A rename rewrites every tracked file, that package included, so it must never mention the template's own names. The init tool reads them from `project.json` instead.

**Leave something behind that still works.** When code is removed, whatever is left must still build, lint, and pass its tests. For example, the TUI and the web app fall back to a "Nothing here yet." home screen when there is no Items page, and `features/README.md` keeps the `features/` folder from disappearing. Anything a later screen will need too, such as the TUI's error and selection styles or the web app's `ErrorMessage`, belongs in template code that the sample uses, not in the sample. Watch for code that only the sample calls: golangci-lint's `unused` check fails on it once the sample is gone. To check, run `task test:init`: its scenarios remove the sample, then run the Go build, tests, and linters on the result and check that the shared helpers are still there.

## Specification

The behaviour above is specified in [`scripts/initproject/features/init.feature`](../scripts/initproject/features/init.feature). Run it with `task test:init`. Each scenario runs the tool on a throwaway copy of the template, so it is slow and is not part of `task test`.
