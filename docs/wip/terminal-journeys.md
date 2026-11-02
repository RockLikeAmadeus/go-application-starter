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
1. [x] `term`: `Start` runs a program in a pseudo-terminal (80×24) and `Screen()` shows what it has printed (adds `creack/pty`, `charmbracelet/x/vt`). First red: `sh -c 'printf hello'` → the screen shows "hello".
   - Red: `TestScreenShowsWhatTheProgramPrinted` (`test/terminal/term/term_test.go`) polls `Screen()` for 5 s; it failed with an empty screen.
   - Green: `Start(*exec.Cmd)` (an `*exec.Cmd`, so the runner can set the scenario's `HOME` and env) runs it with `pty.StartWithSize`; a goroutine copies the pty's output into a `vt.Emulator` under the `Terminal`'s own mutex (`vt.SafeEmulator` doesn't guard `String()`). `Screen()` is `String()` under that mutex; `Close()` closes the pty.
   - Refactor: the size is named once (`cols`, `rows`); the copy loop is `showOutput`; `Close`'s doc no longer claims it stops the program. Journeys: still 9 undefined (no steps yet, as expected).
   - Left for later tasks: the emulator's input side (`vt.Emulator.Read`, an unbuffered pipe) is unread, so a program that asks the terminal something (Bubble Tea does) blocks the emulator's `Write`, and the copy loop with it, while holding the lock. Task 3 must forward it to the pty. `Close()` doesn't stop or wait for the program (task 4). `test/terminal/term` isn't in `task test:go` yet: add `./test/terminal/term/` there in task 12 (it's fast, and a unit test).
2. [x] `term`: `WaitFor(text, timeout)` returns once the screen shows text; on timeout, its error includes the screen as it was.
   - Red: `TestWaitForFailsWithTheScreenWhenTheTextNeverShows` (`sh -c 'printf hello; sleep 5'`, wait 300 ms for "goodbye": the error must name `"goodbye"` and show "hello"). It failed because the stub returned no error. The test is on the timeout, because a happy-path test would pass against a stub that returns `nil`.
   - Green: `WaitFor` polls `Screen()` every 10 ms until it contains the text, or returns an error with the text, the timeout and the screen.
   - Refactor: task 1's test now uses `WaitFor` instead of its own loop, so it also covers `WaitFor`'s happy path. The start-and-clean-up lines are a `start(t, script)` test helper. Journeys: still 9 undefined.
3. [x] `term`: `Type(text)` and `Press(key)` (`enter`, `down`, `up`, `j`, `k`, `q`, …) send keys as a terminal would, through the emulator's encoding. Split over two cycles (2026-11-02): typing and Enter first, then the other named keys.
   - [x] 3a. Red: `TestTypedTextAndEnterReachTheProgram` (`read line; printf 'got:%s' "$line"`; type "hi", press Enter, wait for "got:hi"). It failed with an empty screen (no-op stubs).
     Green: a goroutine copies the emulator's input side (`emu.Read`) to the pty, which also answers programs that ask the terminal something (task 1's left-over). `Type` is `emu.SendText`; `Press` sends `vt.KeyEnter` whatever the name, under the lock (`SendKey` reads the emulator's modes).
     Refactor: the forwarding goroutine is the `sendInput` method, beside `showOutput`, with why it must run (the emulator blocks until its answers are read). Left: `Press`'s hard-coded Enter (3b adds the names); `Close` leaves `sendInput` blocked on the emulator, so task 4 should close the emulator too when it makes `Close` stop the program. Journeys: still 9 undefined.
   - [x] 3b. `Press` sends `down`, `up`, `j`, `k`, `q` (and any other key the journeys use) by name, through `emu.SendKey`, so the cursor-keys mode is respected.
     Red: `TestPressSendsTheKeyAsATerminalWould`, a table over `enter`, `down`, `up`, `j`, `k`, `q`: a raw-mode shell (`stty raw -echo`) prints the bytes it gets with `od`. All but `enter` failed, getting `0d` (Enter) from the hard-coded `Press`.
     Green: `Press` looks the name up in a map of named keys (`enter`, `down`, `up`) and otherwise sends the name's first character as the key.
     Refactor: the map is the package-level `namedKeys`, and `Press`'s doc says it takes one of those or a single character. Left: the first-character fallback, which 3c replaces. Journeys: still 9 undefined.
   - [x] 3c. `Press` with a name it doesn't know (e.g. `"pgdn"`, or a multi-character typo) fails loudly instead of sending its first character. Added 2026-11-02: 3b's single test covers the known names; this is a separate behaviour.
     Red: `TestPressFailsOnAKeyItDoesNotKnow` (`Press("pgdn")` must return an error naming `"pgdn"`). `Press` now returns an `error`; with the stub's `nil`, it failed with "returned no error".
     Green: a name that isn't in `namedKeys` and isn't a single character returns an "unknown key" error, and nothing is sent.
     Refactor: the other tests check `Press`'s error, so a bad name fails at once instead of timing out in `WaitFor`. Journeys: still 9 undefined.
4. [x] `term`: `Wait()` returns the program's exit status; `AltScreen()` says whether the terminal is still on the alternate screen. Split over cycles (2026-11-02), as task 3 was: `Wait` takes a timeout, as `WaitFor` does, so a journey whose program never exits fails instead of hanging the run.
   - [x] 4a. `Wait(timeout)` returns the program's exit status.
     Red: `TestWaitReturnsTheExitStatus` (`exit 3`). The stub returned 0: "exit status = 0, want 3".
     Green: the `Terminal` keeps the `*exec.Cmd`; `Wait` calls `cmd.Wait` and turns an `*exec.ExitError` into its exit code. The timeout isn't used yet (4b). (A first green also had the timeout branch; it was cut back, since no test asked for it.)
     Refactor: nothing worth changing (`Wait` is six lines; its unused timeout is 4b's). Journeys: still 9 undefined.
   - [x] 4b. `Wait` returns an error, with the screen, when the program doesn't exit within the timeout.
     Red: `TestWaitFailsWithTheScreenWhenTheProgramKeepsRunning` (`printf hello; sleep 5`, `Wait(300ms)`): it blocked the whole 5 s and returned no error.
     Green: `Wait` runs `cmd.Wait` in a goroutine and races it against the timeout; on timeout the error shows the screen.
     Refactor: `WaitFor` and `Wait` build their errors with one helper, `errorWithScreen` ("…; the screen shows:" and the screen). Journeys: still 9 undefined. Left for 4d: after a timeout, the goroutine still waits on the program, so a second `Wait` would call `cmd.Wait` twice ("Wait was already called"); `Close` should end the program, and a later `Wait` should reuse the first one's result.
   - [x] 4c. `AltScreen()` says whether the terminal is on the alternate screen.
     Red: `TestAltScreenIsOnWhileTheProgramUsesIt` (the program switches with `ESC [?1049h` and prints "full-screen"). The stub's `false` failed it.
     Green: `AltScreen` is the emulator's `IsAltScreen`, under the lock.
     Refactor: nothing worth changing (three lines, shaped like `Screen`). Journeys: still 9 undefined.
   - [x] 4d. `Close` stops the program if it's still running.
     Red: `TestCloseStopsAProgramThatIgnoresHangUp` (`trap '' HUP; printf ready; sleep 30`; after `Close`, `Wait(2s)` must return). Closing the pty alone hangs the program up, which this one ignores, so `Wait` timed out.
     Green: `Close` kills the program (ignoring the error, which only says it has already exited), then closes the pty.
     Refactor: nothing worth changing (two lines). Journeys: still 9 undefined.
   - [x] 4e. A second `Wait` (e.g. after a timeout, or after `Close`) returns the first one's result instead of `cmd.Wait`'s "Wait was already called" (left over from 4b).
     Red: `TestWaitAgainAfterATimeoutReturnsTheExitStatus` (`read line; printf bye; exit 4`: a `Wait` that times out, Enter, then a second `Wait` must return 4). Its first version, without "bye" and the 100 ms pause, passed most runs: two `cmd.Wait` calls raced to reap the program, and the loser got "waitid: no child processes" (1–3 failures in 10 under `-race`). With the pause, the first `Wait` has reaped it, and it failed 10 times in 10 with "exec: Wait was already called".
     Green: `Start` waits for the program once, in a goroutine that stores `cmd.Wait`'s result and closes `exited`; `Wait` selects on `exited` or the timeout.
     Refactor: that goroutine is the `waitForExit` method, beside `showOutput` and `sendInput`. The test's pause stays: it's what made the red reliable, and it costs 100 ms. Journeys: still 9 undefined.
5. [x] Runner: build `starter` once (in `TestMain`), a fresh computer per scenario; scenario "The command line says where my data is" passes (`--data-dir`).
   - Red: the scenario is the test. The steps (`test/terminal/steps.go`: `journey`, its `run`, `env` and `succeeds` helpers, and the scenario's three steps) were written first, with the computer (a temp folder as `HOME`, `XDG_CONFIG_HOME` under it, no `STARTER_*`, `TERM=xterm-256color`, the programs run in it). With `starter` not built, the When step failed: "exec: no command".
   - Green: `TestMain` builds `cmd/starter` once into a temp folder (`go build -o`), and removes it after the run. The scenario passes: 1 passed, 8 undefined.
   - Refactor: `TestMain` reports a failed build and exits 1 instead of panicking, through a `build(out, pkg)` helper that task 6 reuses for the server. `env` drops the real `HOME`, `XDG_CONFIG_HOME`, `TERM` and `STARTER_*` and puts the scenario's own first, with `filepath.Join`. Note: `go test -run` doesn't pick one scenario (Godog runs them all); use `--godog.tags` or a feature path with a line number. Journeys: 1 passed, 8 undefined.
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
- 2026-11-02: tasks 3 and 4 split into one cycle per behaviour (3a–3c, 4a–4e), since each named several. 4b (`Wait`'s timeout) and 3c (unknown keys) got their own tests. 4e (a second `Wait`) came out of 4b.
- 2026-11-02, 4d: dropped "`Close` closes the emulator, so `sendInput` ends". `vt.Emulator.Close` sets a `closed` flag that `Read` checks without a lock, so closing it while `sendInput` reads is a data race (`-race` fails). The cost is one idle goroutine per terminal, in a test process: accepted.
