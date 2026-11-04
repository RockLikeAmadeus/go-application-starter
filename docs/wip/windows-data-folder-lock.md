# WIP: lock the data folder on Windows

- TODO item: "Lock the data folder on Windows" in [`TODO.md`](../../TODO.md).
- Spec: no change. [docs/data-locations.md](../data-locations.md) ("One process per folder, with `jsonfile`") already states the rule, with a Windows caveat that this removes.
- Business scenario: the existing desktop-driver scenario that a second copy fails with `jsonfile.ErrLocked` (`test/acceptance/desktop_steps.go`). It runs on Linux only, so no new Gherkin: this makes the same rule hold on Windows.
- Unit test: `TestOpenFailsWhenDirectoryIsLocked` in `internal/adapters/storage/jsonfile/store_test.go`, which skips itself on Windows today.
- Temporary: delete this file in the same change that ticks the TODO item.

## Decisions (agreed 2026-11-04)

- **No `/feature`.** The behaviour and its scenario already exist; this is a platform port. Tasks run with `/iterate`.
- **Tested by a Windows CI job** (`windows-latest` in `.github/workflows/ci.yml`) that runs the jsonfile package's unit tests on every push. Nothing here can run Windows code (no Windows machine, no Wine), so locally a cycle can only cross-compile and lint it: `GOOS=windows go vet ./internal/adapters/storage/jsonfile/` and `GOOS=windows go test -c`.
- **Seeing red:** the unskipped test can fail only on Windows. Showing that needs a push of the red commit. Ask before pushing anything other than `/iterate`'s own commit.
- **`LockFileEx` from `golang.org/x/sys/windows`** (already in `go.mod` as an indirect dependency; it becomes direct), with `LOCKFILE_EXCLUSIVE_LOCK|LOCKFILE_FAIL_IMMEDIATELY` on the `.lock` file, mirroring `lock_unix.go`. `ERROR_LOCK_VIOLATION` → `ErrLocked`. No third-party flock library: it would wrap the same call. Windows locks are mandatory, not advisory, but only `.lock` is locked, and nothing reads it.
- `lock_other.go` stays as the no-op for platforms that are neither Unix nor Windows (`//go:build !unix && !windows`).

## Tasks

- [x] 1. A Windows CI job runs the jsonfile package's unit tests (`go test ./internal/adapters/storage/jsonfile/`), with the lock test still skipped, so the package is known to pass on Windows before the lock lands. Infrastructure only, checked on GitHub after the push.
  - No red/green (CI only): `go-windows` job in `ci.yml`; the `image` job now needs it too. Cross-compiled the package's tests for Windows locally; refactor: nothing.
- [ ] 2. Unskip `TestOpenFailsWhenDirectoryIsLocked`: a second `Open` of a locked folder fails with `ErrLocked` on Windows too. Green: `lock_windows.go` with `LockFileEx`; `lock_other.go` narrowed to `!unix && !windows`.
  - Watch: on Windows an open file can't be deleted, so once the store holds `.lock` open, `TestStoreIsNotReadyOnceItsDataDirectoryIsRemoved` (`os.RemoveAll` while open) will likely fail there. Decide then whether to skip it on Windows (a folder in use can't be removed there, so that state can't happen) or change how it's set up.
- [ ] 3. The lock is released on `Close`, so the folder can be opened again (check an existing test covers it on Windows; add one if not).
- [ ] 4. Docs: drop the Windows caveat in `docs/data-locations.md`; ADR 3's `flock` line says Windows uses `LockFileEx`; check CLAUDE.md's Gotcha ("on Unix only: the lock is `flock`") and fix it; tick the TODO item and delete this file.

## Changes to the plan
