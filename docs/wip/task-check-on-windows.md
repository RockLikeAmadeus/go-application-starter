# Make `task check` pass on Windows

- Backlog: TODO.md §1 "Make `task check` pass on Windows" (its "Agreed 2026-11-25" line).
- Spec: SPEC.md §4.10 "Tell when the server is ready", the sentence on `@removes-data-folder` scenarios and Windows.
- Scenarios: no new ones. The three that remove the data folder are tagged `@removes-data-folder`: two in [`features/readiness.feature`](../../features/readiness.feature) and one in [`features/sqlite-storage.feature`](../../features/sqlite-storage.feature).

## Decisions (agreed 2026-11-25)
- **No `@wip` scenarios.** The feature changes how the tests run on Windows, not what the product does. The outer proof is the Windows run itself: `task test:go` and `task accept` failed on Windows on 2026-11-25 (the failures are below, under each task), and the feature is done when `task check` passes on Windows and the widened `go-windows` CI job is green. The docs-and-finish task has no `@wip` tags to remove.
- **Tests that remove a data folder the app holds open are skipped on Windows,** with the reason given. Windows refuses to delete an open file, so the situation can't be set up, and Linux CI still proves the behaviour. In Go, `t.Skip` on `runtime.GOOS == "windows"`. In Godog, the server driver leaves out `@removes-data-folder` on Windows, the way it leaves out `@postgres` without a database (`driversFor` in `test/acceptance/drivers.go`). The tag names what the scenarios do, not the OS (`naming.md`).
- **Every other failure is fixed, not skipped,** unless the thing it checks doesn't exist on Windows (Unix file modes). Then the check is skipped on Windows, with the reason, and the rest of the test still runs.
- **The token file's privacy on Windows** comes from the settings folder (`%AppData%\starter`), which Windows lets only the user (and administrators) read. Go's file modes mean little there (a mode is 666 or 444). So the mode checks skip on Windows, and the docs say where the privacy comes from. Check it by hand with `icacls` before saying so.
- **`task check` skips `e2e:terminal` on Windows and says so** (per-command `platforms:` in the Taskfile), rather than stopping. This is developer output, not user-facing text.
- **CI:** the existing `go-windows` job (today only the jsonfile tests) grows to run what `task test:go` and `task accept` run, with `-race`. The race detector needs cgo and a 64-bit C compiler; check whether `windows-latest` has one (MinGW `gcc`) before adding a setup step. Postgres tests skip there, since Windows runners can't start Linux service containers, and the Linux job covers them. Lint, Vitest, browser e2e, and terminal journeys stay Linux only.
- No client work: nothing a user sees changes.

## Tasks
1. [ ] Acceptance drivers: on Windows, the server driver leaves out `@removes-data-folder` scenarios, and elsewhere it runs them. `driversFor` takes the OS, and a case in `drivers_test.go` covers each. Proof: `task accept` on Windows has no failing `server` scenario.
   - Failing on Windows: `TestFeatures/server`, 3 scenarios (`unlinkat …\.lock` or `data.db`: "The process cannot access the file because it is being used by another process").
2. [ ] Skip the Go tests that remove an open data folder on Windows, with the reason: `TestServerIsNotReadyOnceItsDataFolderIsRemoved` (`cmd/server`), `TestOpenIsNotReadyOnceTheDataFolderIsRemoved` (`desktop`), `TestStoreIsReadyUntilItsDataFolderIsRemoved` (`sqlite`), and `TestNewEmbedded_IsNotReadyOnceTheDataFolderIsRemoved` (`bootstrap`, both subtests). Consider one small test helper for the skip if four copies of the reason read badly.
3. [ ] Tests that leave storage open when their temporary folder is removed close it first: `TestOpenUsesTheSettingsFilesPlansFile` (`desktop`) and `TestNewBilling_WithSignInOnChangesAPlanThatTheAppsEntitlementsShow` (`bootstrap`, built only with `-tags devsignin`). Look for the cause in what each opens (an `App` or a billing service whose store isn't closed), not in the test's cleanup order alone.
   - Failing: `TempDir RemoveAll cleanup: unlinkat …\starter\.lock` (desktop), and `…\002\data.db` (bootstrap).
4. [ ] `TestWriteMarkReplacesTheOldFileRatherThanWritingOverIt` (`syncedfolder`) proves the replacement in a way that holds on Windows too. `os.SameFile` reads identity when called there, so comparing two `os.Stat` results can't tell. For example, hold the old file open and check it still reads the old mark, if Windows allows the rename over it.
   - Failing: `devicemarks_test.go:162: the mark was written over in place…` (a false alarm).
5. [ ] `tokenstore` file tests: the mode checks skip on Windows, with the reason, and the rest of each test still runs: `TestFile_KeepsEachServersTokenApartInAPrivateFileUnderItsNormalisedAddress` and `TestFile_MakesAnExistingWiderFilePrivateWhenItSavesATokenInIt`.
   - Failing: `file mode = 666; want 600`.
6. [ ] `TestRenameFileRewritesAndMovesAFile` (`scripts/internal/project`): the executable-bit check skips on Windows, which has none, and the rest still runs.
   - Failing: `cmd/zenith/run.sh mode = -rw-rw-rw-, want -rwxr-xr-x`.
7. [ ] `TestLoginSaysWhoIsSignedInAndWhereTheTokenIsKept` (`cli`) fakes the browser in a way that runs on Windows. Today it's a `#!/bin/sh` script using `curl`, which Windows can't start, so the sign-in waits forever and the package times out after 10 minutes. For example, the test binary as the browser (Go's helper-process pattern), or an opener the test passes in. Check other tests that set `BROWSER` the same way.
   - Failing: `panic: test timed out after 10m0s`, in its `credential_store` subtest.
8. [ ] `TestInfoNamesTheSettingsFile` (`cli`), "none, when the default one can't be created": make the settings folder unusable in a way that holds on Windows, where `chmod 0o500` doesn't stop creating files. For example, a file where the folder should be.
   - Failing: the output names `…\starter\config.toml` instead of `Settings file: none`.
9. [ ] `TestMainRunsInAGitRepository` (`scripts/initproject`) builds the tool with `.exe` on Windows, so it can start it.
10. [ ] Run `task test:go` on Windows from start to finish once tasks 2 to 9 are done, and fix anything they hid. The packages that timed out or stopped early may still hide failures: `cli` timed out at task 7, and `test:go` stopped after its first command. `gen:check`, `lint`, `test:web`, and `e2e` already pass on Windows (2026-11-25).
11. [ ] Taskfile: `task check` on Windows skips `e2e:terminal` and says why, instead of stopping. Then `task check` passes on Windows from start to finish.
12. [ ] CI: the `go-windows` job runs the Go unit tests (with `-race`, the `devsignin` tests, and the init tool's tests) and the acceptance tests, as `task test:go` and `task accept` do. Set up a C compiler only if `windows-latest` lacks one. Rename the job's display name. Check the job goes green on the branch.
13. [ ] Docs and finish: `docs/ci-cd.md` (the workflows table says "five jobs" and leaves out `go-windows`; say what the Windows job runs and why, in the doc's style), README Prerequisites (the 64-bit C compiler `-race` needs on Windows, if not there already), the token file's privacy on Windows in `docs/sign-in.md`, and CLAUDE.md's Commands row for `task e2e:terminal` if needed. Tick the TODO.md item with a "Done" note, and delete this file. No `@wip` tags to remove.

## Changes to the plan
