# WIP: Terminal journeys for the CLI and TUI

- Backlog: `TODO.md`, "Do first", **Implement BDD layer for CLI and TUI clients**.
- Spec: [SPEC.md §4.13](../../SPEC.md#413-use-the-command-line-and-the-terminal-app), and the glossary's **Command line** and **Terminal app**. Sample: §4.1's "terminal journeys" link.
- Scenarios: [test/terminal/features/command-line.feature](../../test/terminal/features/command-line.feature) and [terminal-app.feature](../../test/terminal/features/terminal-app.feature) (template), [items.feature](../../test/terminal/features/items.feature) (sample). Runner: `test/terminal/terminal_test.go`. Run with `go test ./test/terminal/ -count=1 --godog.tags=@wip` (a task comes in task 12).
- Temporary: delete this file in the change that removes the `@wip` tags and ticks the TODO item.

## Decisions
- **Tools (chosen 2026-11-02): Godog + `creack/pty` + `charmbracelet/x/vt`.** Same runner and Gherkin as `test/acceptance`, all Go. `creack/pty` (v1.1.24) runs the real built program in a pseudo-terminal; `x/vt` (v0.1.0, by Bubble Tea's authors) turns its output into a screen: `SafeEmulator`, `String()` (plain text), `SendKey`/`SendText` (keys encoded as a terminal would, read back with `Read`), `IsAltScreen()`. A small wait-until-the-screen-shows helper replaces go-expect.
  - Rejected: **Netflix/go-expect + hinshun/vt10x** (named in the TODO): no release since early 2022, and vt10x predates escape sequences Bubble Tea v2 and Lip Gloss v2 emit. **VHS**: made for demo GIFs, needs ttyd, ffmpeg and a browser, tapes aren't Gherkin (a fine later item for README demos, not tests). **Microsoft tui-test**: TypeScript with no Gherkin of its own, newer APIs in beta. **teatest**: in-process, so it doesn't prove the built binary, its flags or the terminal, and covers the TUI only.
- **Scope (chosen 2026-11-02): thin wiring journeys**, like `web/e2e`. Business rules stay in `/features` against the driving ports. These prove that the built `starter` program, in a real terminal, reaches the core and back: options, output, exit status, keys, the alternate screen, and `--server`.
- **Where:** `test/terminal/` (package `terminal`): `features/`, the runner, the steps, and a small helper package for the terminal (`test/terminal/term`, unit-tested). Not in `task test` (it builds binaries); its own task, added to `task check` and CI.
- **Template first:** the terminal helper, the build, the per-scenario computer and the server are template code, with template journeys that survive sample removal (`starter info`, `--server … info`, an unknown command, opening and quitting the TUI, which shows "q quit" on both the home and the Items screen). The sample's journeys (`items.feature`, `item_steps.go`) go in `scripts/initproject`'s `SampleFiles`, and their step registration in a sample block.
- **Each scenario gets its own computer:** a temp folder as `HOME` and `XDG_CONFIG_HOME`, with no `STARTER_*` variables from the real environment, so journeys never touch the user's data and can run in parallel later.
- **The programs are built once** per run (`go build` of `cmd/starter` and `cmd/server` into a temp folder), as the user would run them. No `embedweb`: the journeys don't need the SPA.
- **Unix only:** pseudo-terminals here are Unix (`creack/pty`); the package gets a `unix` build tag. CI runs on Linux. Windows (ConPTY) is out of scope.
- **"Fails"** in a step means a non-zero exit status and the message on the terminal (in a pseudo-terminal, standard error and output share the screen).
- **"Selected"** in the TUI is the row `selectedRow` marks with `> `.
- **Shown failing (2026-11-02):** all 9 scenarios undefined (31 steps), since no steps exist yet. `task accept`, `task test` and `task lint` are unaffected. `task test:init` fails until the journeys pass: its scenarios run `go test ./internal/... ./test/...` on the generated project, which includes `test/terminal` (as any feature's failing `@wip` scenarios in `test/acceptance` would). With `--remove-sample` it will also need task 12's `SampleFiles` entries.

## Tasks
1. [ ] `term`: `Start` runs a program in a pseudo-terminal (80×24) and `Screen()` shows what it has printed (adds `creack/pty`, `charmbracelet/x/vt`). First red: `sh -c 'printf hello'` → the screen shows "hello".
2. [ ] `term`: `WaitFor(text, timeout)` returns once the screen shows text; on timeout, its error includes the screen as it was.
3. [ ] `term`: `Type(text)` and `Press(key)` (`enter`, `down`, `up`, `j`, `k`, `q`, …) send keys as a terminal would, through the emulator's encoding.
4. [ ] `term`: `Wait()` returns the program's exit status; `AltScreen()` says whether the terminal is still on the alternate screen.
5. [ ] Runner: build `starter` once (in `TestMain`), a fresh computer per scenario; scenario "The command line says where my data is" passes (`--data-dir`).
6. [ ] Template: start the built server on a free port with memory storage and wait for `/readyz`; scenario "The command line works on a server's data" passes.
7. [ ] Template: scenario "A command the command line doesn't know fails" passes (exit status and message).
8. [ ] Template: scenario "Open and quit the terminal app" passes ("q quit", `q`, exit 0, off the alternate screen).
9. [ ] Sample: `item_steps.go` (in `SampleFiles`, registered in a sample block); scenarios "Add an item with the command line" and "The command line reports a mistake" pass.
10. [ ] Sample: scenario "Add an item in the terminal app" passes (`a`, type, `enter`).
11. [ ] Sample: scenario outline "Move through the items" passes, both rows (`j`/`k`, `down`/`up`).
12. [ ] Wiring: `task e2e:terminal`, in `task check` and CI; `unix` build tag; `SampleFiles` entries; `task test:init` (the trimmed project's terminal journeys still run and pass).
13. [ ] Docs: CLAUDE.md (Commands table, Architecture's `test/terminal`, Testing conventions), `features/README.md` (where terminal journeys live), README's test commands, `docs/template.md` if it lists test layers; tick the TODO item with the tool comparison in one line.
14. [ ] Finish: `task check`, `task test:init`, `task test:rename`, remove `@wip`, tick the TODO item, delete this file.

## Changes to the plan
