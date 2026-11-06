# Keep plain-text data in a synced folder

Step 5 of [the editions plan](../plans/editions-and-accounts.md).

- Spec: [SPEC.md §4.16](../../SPEC.md#416-keep-plain-text-data-in-a-synced-folder)
- Scenarios: [`features/synced-plain-text.feature`](../../features/synced-plain-text.feature) (`@desktop`, 14 scenarios, tagged `@wip` at the top of the file)
- Decisions: [ADR 5](../adr/0005-hosted-and-local-editions.md) "Synced data folders", [ADR 3](../adr/0003-plain-text-storage.md) (to update)

## Decisions

Agreed with the user:
- **The mark is written only when the folder looks synced:** `syncedfolder.Detect` finds a service and `ignore_synced_folder` is off. Every plain-text copy still reads other devices' marks. So the default folder, servers and Android have no background writes, and a sync setup we can't detect gets no warning.
- **Both warnings are checked when the data is opened.** They go in `App.Warnings`, so every front end shows them with no new plumbing. A desktop app left open sees a new conflict copy the next time it opens.

Decided in planning:
- **Layout:** `<dir>/<table>/<key>.json`, with an escaped key, holding `{"version": 2, "key": "<key>", "record": {...}}`. The version is per file, because a sync service can bring files across at different times. The v1 `items.json` is ignored, not migrated (the template has no users).
- **A conflict copy is a file that doesn't match its own key.** A copy keeps the original's contents, so its `key` differs from its file name. That covers every service's naming (Dropbox, Syncthing, OneDrive, Google Drive's `(1)`, iCloud's ` 2`) with no list of name patterns, and also a user's `copy of ….json`. Reading a table skips these files, so nothing shows twice. This is why the key is in the file: plans records don't hold their own.
- **A marked file per device, not one shared lock file.** The plan said one lock file, but two devices rewriting it would make a conflict copy in exactly the case we're warning about. So it's `<dir>/.devices/<device>.json`, holding the device and the time, and each device writes only its own.
- **A device is `os.Hostname()`.** Two computers with the same name look like one, which is acceptable for a best guess. The desktop app and the CLI on one computer are one device. If the name can't be read, the app leaves no mark.
- **Timing:** renewed every minute, stale after five, to allow for slow sync services. Times are compared on this device's clock, so a skewed clock can miss or invent a warning.
- **Where it lives:** the conflict scan is `jsonfile.Store`'s, since only it knows the layout. Marks are in `internal/adapters/syncedfolder`, beside `Detect`, and `internal/bootstrap` composes them. Neither adapter imports the other.
- **The scan** looks one level down, in folders that aren't hidden, at `.json` files that decode as records. So `.devices`, `.stfolder`, `config.toml` and other files in a shared data folder aren't conflict copies.
- **Template first:** all of this is template code. The conflict copy step writes a record in a made-up table, so the template's scenarios need no sample. The sample has two thin scenarios, in a sample block.
- **Kept:** the local flock (`.lock`), which still stops two programs on one computer, and SQLite's looks-synced warning. Heartbeats are plain text only.

Known limit: iCloud can replace a file it has moved off the device with a hidden placeholder (`.<name>.json.icloud`). Reading skips it, so that record looks missing until it downloads. Note this in the docs (task 15).

## Tasks

Each is one `/red` → `/green` → `/refactor` cycle.

One file per record (`jsonfile`):
- [x] 1. A record written under a table and key reads back from `<table>/<key>.json`, in the version 2 format. Replaces `readTable`/`writeTable` and `table_test.go`, keeping the atomic write.
  - Red: `TestRecordIsKeptInItsOwnVersion2FileAndReadsBack` failed: `things/1.json` doesn't exist.
  - Green: `writeRecord` and `readRecords` keep each record in `<table>/<key>.json`, version 2. Reading has no version check or filters yet (tasks 3, 4, 7).
  - Refactor: the atomic write moved to `replaceFile`, shared with `writeTable`. The `record` helpers stay in `store.go` until the old ones go.
- [x] 2. Keys that can't be file names (`/`, `:`) round-trip.
  - Red: `TestKeyThatCannotBeAFileNameRoundTrips` failed: `writeRecord` couldn't create the temp file, as the `/` in the key named a folder that doesn't exist.
  - Green: `writeRecord` names the file with `url.QueryEscape(key)`.
  - Refactor: the escaping moved to `recordFileName`, ready for task 7 to compare a file's name with its key.
- [x] 3. A missing table is empty.
  - Red: `TestTableWithNoRecordsFolderIsEmpty` failed: `readRecords` returned the "no such file or directory" error.
  - Green: `readRecords` returns an empty table when the folder doesn't exist.
  - Refactor: nothing. The same check in `readTable` goes with it in tasks 5 and 6.
- [x] 3b. A file with an unknown version is refused.
  - Red: `TestRecordWithAnUnknownVersionIsRefused` failed: `readRecords` returned no error for version 99.
  - Green: `readRecords` refuses a file whose version isn't `recordFormatVersion`.
  - Refactor: nothing. `readTable`'s matching check goes in tasks 5 and 6.
- [x] 4. Reading a table skips what isn't a `.json` file, such as leftover temp files.
  - Red: `TestReadingATableSkipsWhatIsNotAJSONFile` failed: `readRecords` tried to decode `.1.json.icloud`.
  - Green: `readRecords` skips folders and names without the `.json` extension.
  - Refactor: nothing. `readRecords` is still short; tasks 7 and 8 will show whether the per-file decode should be its own function.
- [x] 5. Deleting a record removes its file, and no other.
  - Red: `TestDeletingARecordRemovesOnlyItsFile` failed: the stub left the record in place.
  - Green: `deleteRecord` removes `recordFileName(key)` from the table's folder.
  - Refactor: nothing. The path join in `writeRecord` and `deleteRecord` is one short line, so a helper wouldn't simplify it.
- [x] 5b. `ItemRepository` keeps each item in its own file, and saving one leaves the others untouched (contract and isolation suites stay green). Sample thin use. Define the "each item has its own file" step.
  - Red: `TestEachItemIsKeptInItsOwnFileAndSavingOneLeavesTheOthersUntouched` failed: `items/b.json` doesn't exist. Its `items.json` assertion went from `TestItemsSurviveReopen`.
  - Green: `ItemRepository` uses `writeRecord`, `readRecords` and `deleteRecord`; the step checks `items/<id>.json` for each item, with `desktopAppsDataFolder` shared with `itNamesTheDesktopAppsDataFolder`.
  - Refactor: nothing. `FindByID` and `Delete` still read the whole table, so that task 7's skip covers them; a one-file read isn't worth a second path yet.
- [ ] 6. `WorkspacePlanRepository` keeps each workspace's plan in its own file.

Conflict copies (`jsonfile`, then `bootstrap`):
- [ ] 7. Reading a table skips a file whose key doesn't match its name, as a table test over each service's naming.
- [ ] 8. `Store.ConflictCopies` lists them across tables, sorted, as paths relative to the data folder, leaving out hidden folders and files that aren't records.
- [ ] 9. `bootstrap` puts a conflict copy warning in `App.Warnings` on plain text, naming the files, one or many, regardless of `ignore_synced_folder`. Define the conflict copy steps; turns the first 5 scenarios green.
- [ ] 10. Define the item conflict copy step; turns the sample's scenario green.

Marks (`syncedfolder`, then `bootstrap`):
- [ ] 11. `OtherDevices` returns other devices with a recent mark, and leaves out this device, old marks and unreadable files.
- [ ] 12. A mark is written atomically to `.devices/<device>.json`, is renewed on each tick, and is removed on stopping.
- [ ] 13. `bootstrap` warns about other devices on plain text, naming them. Define the device steps; turns the 3 warning scenarios green.
- [ ] 14. `bootstrap` marks the folder only when it looks synced and `ignore_synced_folder` is off, and `App.Close` stops the mark before the store closes. Define the mark steps; turns the last 4 scenarios green.

Finish:
- [ ] 15. Docs: ADR 3 (layout, version 2, what "one file per domain type" became), ADR 5 if a detail changed, `docs/data-locations.md` (including `workspace_plans.json` and the iCloud limit), `docs/configuration.md` (`ignore_synced_folder` also stops the mark), `.claude/rules/storage.md`, the `jsonfile` package comment.
- [ ] 16. Run `task test:init`, `task check` and `task accept`, remove `@wip`, delete this file, and mark step 5 done in the plan.

Run the init suite at tasks 5b, 9 and 10 too: they add code next to the sample's. Until task 16, `task test:init` itself is red: its `go test ./test/...` has no tag filter, so the undefined `@wip` steps fail in strict mode. So copy the working tree to the scratchpad, delete `features/synced-plain-text.feature`, commit in the copy, and run `go test ./scripts/initproject/acceptance/ -count=1 -timeout 30m` there. Task 16 runs it for real.

## Changes to the plan

- Task 1: `readTable`/`writeTable` and `table_test.go` can't go until the repositories stop using them, so `writeRecord`/`readRecords` come in beside them, and tasks 5 and 6 remove the old ones.
- Task 3 split: the missing table is task 3, and refusing an unknown version is task 3b, as each needs its own test.
- Task 5 split: `deleteRecord` is task 5 and moving `ItemRepository` is 5b. No template repository deletes, so the helper needs a template test of its own, or `unused` fails once `--remove-sample` drops the items. `ItemRepository` needs it first, so it comes first. Run `task test:init` at 5b, not 5.
- Task 5b: `task test:init` is red until the `@wip` tags go (it was at HEAD too), so tasks 5b, 9 and 10 check on a copy without the feature file, as the note under the task list says.
