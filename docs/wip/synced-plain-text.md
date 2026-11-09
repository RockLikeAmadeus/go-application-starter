# Keep plain-text data in a synced folder

Step 5 of [the editions plan](../plans/editions-and-accounts.md).

- Spec: [SPEC.md §4.16](../../SPEC.md#416-keep-plain-text-data-in-a-synced-folder)
- Scenarios: [`features/synced-plain-text.feature`](../../features/synced-plain-text.feature) (`@desktop`, 14 scenarios, tagged `@wip` at the top of the file)
- Decisions: [ADR 5](../adr/0005-hosted-and-local-editions.md) "Synced data folders", [ADR 3](../adr/0003-plain-text-storage.md) (to update)

## Decisions

Agreed with the user:
- **The mark is written only when the folder looks synced:** `syncedfolder.Detect` finds a service and `ignore_synced_folder` is off. Every plain-text copy still reads other devices' marks. So the default folder, servers and Android have no background writes, and a sync setup we can't detect gets no warning.
- **Both warnings are checked when the data is opened.** They go in `App.Warnings`, so every front end shows them with no new plumbing. A desktop app left open sees a new conflict copy the next time it opens.
- **Plain text is never used in the hosted edition,** so a plain-text data folder only ever holds one workspace. Files need no per-workspace folders, and nothing here needs to perform well across many workspaces. `jsonfile` still takes the workspace on every call and runs the contract and isolation suites, because the ports are shared and the workspace scenarios run on it.
- **A one-record lookup reads only that record's file.** Reading the whole table to find one record costs one file open per record. `FindAll` still reads everything, which is the format's limit; large datasets belong on SQLite.

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
- [x] 6. `WorkspacePlanRepository` keeps each workspace's plan in its own file.
  - Red: `TestEachWorkspacesPlanIsKeptInItsOwnFileAndSavingOneLeavesTheOthersUntouched` failed: `workspace_plans/ws-b.json` doesn't exist.
  - Green: `WorkspacePlanRepository` uses `writeRecord` and `readRecords`.
  - Refactor: removed `readTable`, `writeTable`, `tableFile`, `fileFormatVersion`, `Store.path` and their three tests, as nothing calls them now. The package comment still describes the old layout; task 15 rewrites it.

Conflict copies (`jsonfile`, then `bootstrap`):
- [x] 7. Reading a table skips a file whose key doesn't match its name, as a table test over each service's naming.
  - Red: `TestReadingATableSkipsAFileWhoseKeyDoesNotMatchItsName` failed for all six names: the copy came back as record 1.
  - Green: `readRecords` skips a file whose name isn't `recordFileName(file.Key)`.
  - Refactor: `readRecords`' doc comment explains the skip. The per-file decode stays inline until task 7b gives it a second caller.
- [x] 7b. `readRecord` returns one record from its own file without reading any other.
  - Red: `TestReadRecordReturnsOneRecordWithoutReadingAnyOther` failed: the stub returned an empty record.
  - Green: `readRecord` reads and decodes `recordFileName(key)` alone. No not found, key or version checks yet (7b2–7b4).
  - Refactor: `readRecordFile` holds the read and decode for `readRecords` and `readRecord`. The version check stays in `readRecords` until 7b4 moves it in.
- [x] 7b2. `readRecord` returns not found for a missing file.
  - Red: `TestReadRecordOfAMissingFileIsNotFound` failed: the error was the "no such file or directory" one, not `domain.ErrNotFound`.
  - Green: `readRecord` turns a missing file into `domain.ErrNotFound`.
  - Refactor: nothing. `readRecord` is four lines; 7b3 and 7b4 add to it, so its shape is better judged after them.
- [x] 7b3. `readRecord` returns not found for a file whose key doesn't match its name.
  - Red: `TestReadRecordOfAFileWhoseKeyDoesNotMatchItsNameIsNotFound` failed: `readRecord` returned no error for a file holding key 1 under the name `2.json`.
  - Green: `readRecord` returns `domain.ErrNotFound` when the file's key isn't the one asked for, checked after other errors so that a decode failure stays an error.
  - Refactor: `readRecord`'s doc comment covers the key check. The body stays as it is until 7b4 adds its last check.
- [x] 7b4. `readRecord` refuses a file with an unknown version.
  - Red: `TestReadRecordOfAFileWithAnUnknownVersionIsRefused` failed: `readRecord` returned no error for version 99.
  - Green: `readRecord` refuses a file whose version isn't `recordFormatVersion`.
  - Refactor: the version check moved into `readRecordFile`, so both readers share it. It still runs before the key check.
- [x] 7c. `WorkspacePlanRepository.PlanOf` reads only that workspace's file.
  - Red: `TestPlanOfReadsOnlyThatWorkspacesFile` failed: `PlanOf` stopped at the undecodable `ws-b.json`.
  - Green: `PlanOf` uses `readRecord`.
  - Refactor: nothing. `PlanOf` is three lines. Only the sample's repository calls `readRecords` in production now, so the init suite ran on a copy here too and stayed green; `table_test.go` keeps it used.
- [x] 7d. `ItemRepository.FindByID` reads only that item's file. Sample thin use.
  - Red: `TestFindByIDReadsOnlyThatItemsFile` failed: `FindByID` stopped at the undecodable `b.json`.
  - Green: `FindByID` uses `readRecord`, and still treats another workspace's item as not found.
  - Refactor: the test explains its undecodable file, as 7c's does. `lookup` stays for `Delete` until 7e, which can remove it. The init suite ran on a copy.
- [x] 7e. `ItemRepository.Delete` reads only that item's file. Sample thin use.
  - Red: `TestDeleteReadsOnlyThatItemsFile` failed: `Delete` stopped at the undecodable `b.json`.
  - Green: `Delete` uses `readRecord`, and still treats another workspace's item as not found.
  - Refactor: `workspaceRecord` replaces `lookup` and holds the workspace check for `FindByID` and `Delete`. The init suite ran on a copy.
- [x] 8. `Store.ConflictCopies` lists them across tables, sorted, as paths relative to the data folder.
  - Red: `TestConflictCopiesListsThemAcrossTables` failed: the stub returned nothing.
  - Green: `ConflictCopies` walks each folder in the data directory and lists the `.json` files that don't match their key. `os.ReadDir` gives the order, table then name, so there's no separate sort.
  - Refactor: `conflictCopiesIn` holds the per-table loop; `isJSONFile` and `recordFile.matches` are shared with `readRecords`. An undecodable file still stops the scan until 8c.
- [x] 8b. `ConflictCopies` leaves out hidden folders.
  - Red: `TestConflictCopiesLeavesOutHiddenFolders` failed: the copy in `.stversions` was listed.
  - Green: `ConflictCopies` skips folders whose name starts with a dot.
  - Refactor: `ConflictCopies`' doc comment says hidden folders aren't tables. A helper for the one-line check wouldn't simplify it.
- [x] 8c. `ConflictCopies` leaves out files that aren't records.
  - Red: `TestConflictCopiesLeavesOutFilesThatAreNotRecords` failed for all five kinds: four stopped the scan with an error, and a file with no key was listed.
  - Green: `readRecordFile` marks decode and version failures with `errNotARecord`, which the scan skips; other read errors still stop it. A file with no key isn't listed.
  - Refactor: `recordFile.isConflictCopy` holds the key check, and `ConflictCopies`' doc comment covers these files. `readRecords` and `readRecord` still refuse such files inside a table.
- [x] 8d. Benchmarks for `readRecords`, `readRecord` and `ConflictCopies` at 1,000 and 10,000 records. There's no red, as they specify no behaviour. Note the numbers here, for task 15's rough record count in ADR 3.
  - Red: none, and no green. `table_bench_test.go` ran first time, and each benchmark fails if it reads the wrong number of records or finds a copy.
  - Numbers (laptop, Intel Ultra 7 155U, small records, files in the page cache): `readRecords` 6 ms at 1,000 and 67 ms at 10,000; `ConflictCopies` 5 ms and 63 ms; `readRecord` 5 µs at both.
  - Refactor: the setup writes the files directly, as `writeRecord`'s fsync took it past three minutes; the run now takes 9 s.
- [x] 9. `bootstrap` puts a conflict copy warning in `App.Warnings` on plain text, naming the files, one or many, regardless of `ignore_synced_folder`. Define the conflict copy steps; turns the first 5 scenarios green.
  - Red: `TestApp_WarnsAboutConflictCopiesOnPlainText` failed for all three cases: `Warnings` was empty.
  - Green: `conflictCopyWarning` in `bootstrap` adds the warning on plain text. A failed scan gives no warning until 9b. `synced_plain_text_steps.go` defines the conflict copy steps.
  - Refactor: `warnsSaying` is shared by the warning steps, and `conflictCopyPath` by the file steps. The two messages stay as whole sentences, as joining them would make the wording harder to read. The init suite ran on a copy.
- [x] 9b. A failed conflict copy scan is a warning saying so, and the app keeps working.
  - Red: `TestApp_WarnsWhenConflictCopiesCantBeCheckedAndKeepsWorking` failed: `Warnings` was empty with an unreadable table folder.
  - Green: `conflictCopyWarning` returns a warning carrying the scan's error.
  - Refactor: `conflictCopyWarning`'s doc comment covers a failed scan. SPEC.md has no wording for this warning, and the scenarios don't cover it, as an unreadable folder is unit-test detail.
- [x] 10. Define the item conflict copy step; turns the sample's scenario green.
  - Red: none, and no green. The scenario failed only on its undefined step, as `readRecords` already skips copies. It's kept as a guard: with the skip removed it fails with `["Widget on another device"]`.
  - The step writes a copy with different contents, Syncthing-named so it's read after the original: an identical or earlier copy passes even with the skip removed, as records are held by key. The scenario's stray `| name |` row is gone: `the items are:` has no header.
  - Refactor: `registerItemSteps`' comment names this feature. The two item file paths stay as they are, as they differ in the part a helper would hide. The init suite ran on a copy.

Marks (`syncedfolder`, then `bootstrap`):
- [x] 11. `OtherDevices` returns other devices with a recent mark.
  - Red: `TestOtherDevicesReturnsDevicesWithARecentMark` failed: the stub returned nothing.
  - Green: `OtherDevices` lists the device named in each file in `.devices`. No filters yet (11b–11d).
  - Refactor: nothing. The loop is short, and 11c and 11d decide the shape of its read and decode. Lint accepts the two parameters that 11b and 11c will use.
- [x] 11b. `OtherDevices` leaves out this device.
  - Red: `TestOtherDevicesLeavesOutThisDevice` failed: the result was `["desktop" "laptop"]`.
  - Green: `OtherDevices` skips a mark whose device is `thisDevice`.
  - Refactor: nothing. The check reads the device in the file, not its name, so a conflict copy of our own mark is skipped too. 11c and 11d decide the loop's shape.
- [x] 11c. `OtherDevices` leaves out marks five minutes old or more.
  - Red: `TestOtherDevicesLeavesOutMarksFiveMinutesOldOrMore` failed: the result was `["laptop" "phone" "tablet"]`.
  - Green: `OtherDevices` skips a mark at least `markStaleAfter` old, reading the mark's `time`.
  - Refactor: `OtherDevices`' doc comment says recent, and notes the clock limit. The one condition stays until 11d reshapes the loop.
- [x] 11d. `OtherDevices` leaves out files that aren't marks.
  - Red: `TestOtherDevicesLeavesOutFilesThatAreNotMarks` failed for no device and a device that isn't text: the result was `["laptop" ""]`. The other three kinds already fell out as stale, with a zero time.
  - Green: `OtherDevices` skips a mark with no device.
  - Refactor: `readMark` checks the read and decode errors, which had worked only because a zero time is stale, and says whether the file is a mark.
- [x] 11e. A folder with no marks has no other devices.
  - Red: none, and no green. `TestOtherDevicesOfAFolderWithNoMarksIsEmpty` passed first time, as a missing `.devices` reads as no entries. It's kept as a guard: it fails with `["ghost"]` when the list starts non-empty.
  - Refactor: `OtherDevices`' doc comment says why the read error is ignored: the marks are a best guess.
- [x] 12. `WriteMark` leaves a device's mark in `.devices/<device>.json`.
  - Red: `TestWriteMarkLeavesTheDevicesMarkInItsOwnFile` failed: `.devices/desktop.json` doesn't exist.
  - Green: `WriteMark` creates `.devices` and writes the device and time, with a plain `os.WriteFile` until 12b.
  - Refactor: nothing. The two `filepath.Join` lines are too short for a helper, and 12b reshapes the write.
- [x] 12b. Writing a mark replaces the old file atomically.
  - Red: `TestWriteMarkReplacesTheOldFileRatherThanWritingOverIt` failed: the mark was written over in place.
  - Green: `WriteMark` writes a temp file beside the mark and renames it over.
  - Refactor: `replaceMark` holds the temp file and rename. It stays apart from `jsonfile.replaceFile`, as neither adapter imports the other and a mark needs no fsync.
- [x] 12c. A mark is renewed on each tick.
  - Red: `TestKeepMarkRenewsTheMarkOnEachTick` failed: nothing took the tick.
  - Green: `KeepMark` writes the mark on each tick in a goroutine, dated by the tick; its `stop` waits for the goroutine.
  - Refactor: `readMarkFile` in the tests replaces the two copies of the read and decode. `KeepMark` stays as it is until 12d adds removal.
- [x] 12d. A mark is removed on stopping.
  - Red: `TestKeepMarkRemovesTheMarkOnStopping` failed: the mark was still there after stopping.
  - Green: `stop` removes the mark once the goroutine has ended.
  - Refactor: `markPath` holds the mark's path, and `KeepMark`'s doc comment covers removal and why a failure is ignored.
- [x] 13. `bootstrap` warns about another device on plain text, naming it. Define the device steps; turns the 3 warning scenarios green.
  - Red: `TestApp_WarnsWhenAnotherDeviceIsUsingPlainText` failed: `Warnings` was empty.
  - Green: `otherDeviceWarning` names the first device from `OtherDevices`; device steps defined; the 3 warning scenarios pass.
  - Refactor: `commandLineWarnsSaying` replaces the three copies of the command line check. Listing every device waits for 13b.
- [x] 13b. The warning names every other device.
  - Red: `TestApp_WarningNamesEveryOtherDevice` failed: the warning named `laptop` but not `phone`.
  - Green: `otherDeviceWarning` joins every device from `OtherDevices`, with ", " like the conflict copy list.
  - Refactor: only its doc comment. The two warning tests share their setup; left until 13c shows what it needs.
- [x] 13c. This device's own mark gives no warning.
  - Red: `TestApp_ThisDevicesOwnMarkGivesNoWarning` failed: the warning named this computer.
  - Green: `otherDeviceWarning` passes `os.Hostname()` as this device; if it can't be read, every mark counts as another device's.
  - Refactor: `warningsWithMarksFrom` holds the warning tests' setup; made-up devices are now `work-laptop` and `work-phone`, so a computer named `laptop` doesn't fail them.
- [x] 14. `bootstrap` marks a plain-text folder that looks synced on opening.
  - Red: `TestApp_MarksAPlainTextFolderThatLooksSynced` failed: no device had a mark, want this computer.
  - Green: `markFolder` writes this device's mark once on opening plain text, with no check yet that the folder looks synced (14b, 14c) and no renewing or removal (14d). With no computer name it leaves no mark.
  - Refactor: nothing. `markFolder` is short, and the test's `.dropbox` setup, also in the SQLite test, is left until 14c shows whether a helper fits. The init suite ran on a copy.
- [x] 14b. A folder that doesn't look synced gets no mark.
  - Red: `TestApp_LeavesNoMarkInAFolderThatDoesNotLookSynced` failed: this computer had a mark in a folder with no sync service.
  - Green: `markFolder` returns before marking unless `syncedfolder.Detect` finds a service.
  - Refactor: only `markFolder`'s doc comment, which now says it marks only a folder that looks synced. The `.dropbox` setup is left until 14c shows whether a helper fits.
- [x] 14c. `ignore_synced_folder` means no mark.
  - Red: `TestApp_LeavesNoMarkWhenTheSyncedFolderIsIgnored` failed: this computer had a mark with the setting on.
  - Green: `markFolder` returns before marking when `ShouldIgnoreSyncedFolder` is set.
  - Refactor: `dropboxDataDir` holds the `.dropbox` setup that three tests copied. `markFolder`'s doc comment now names the setting.
- [x] 14d. The mark is renewed every minute, and `App.Close` stops it before the store closes.
  - Red: `TestApp_CloseRemovesTheMark` failed: this computer still had a mark after `Close`.
  - Green: `markFolder` keeps the mark with `KeepMark` on a one-minute ticker and returns its stop; jsonfile's `a.close` stops it, then closes the store.
  - Refactor: dropped the overwritten `a.close = store.Close`; `markFolder`'s early returns share `noMark`, and its doc comment covers renewing and stopping. The init suite ran on a copy.
- [ ] 14e. Define the mark steps; turns the last 4 scenarios green.

Finish:
- [ ] 15. Docs: ADR 3 (layout, version 2, what "one file per domain type" became, a rough record count from task 8d), ADR 5 (plain text is never used in the hosted edition, and any other detail that changed), `docs/data-locations.md` (including `workspace_plans.json` and the iCloud limit), `docs/configuration.md` (`ignore_synced_folder` also stops the mark), `.claude/rules/storage.md`, the `jsonfile` package comment.
- [ ] 16. Run `task test:init`, `task check` and `task accept`, remove `@wip`, delete this file, and mark step 5 done in the plan.

Run the init suite at tasks 5b, 7d, 7e, 9 and 10 too: they add code next to the sample's. Until task 16, `task test:init` itself is red: its `go test ./test/...` has no tag filter, so the undefined `@wip` steps fail in strict mode. So copy the working tree to the scratchpad, delete `features/synced-plain-text.feature`, commit in the copy, and run `go test ./scripts/initproject/acceptance/ -count=1 -timeout 30m` there. Task 16 runs it for real.

## Changes to the plan

- Task 1: `readTable`/`writeTable` and `table_test.go` can't go until the repositories stop using them, so `writeRecord`/`readRecords` come in beside them, and tasks 5 and 6 remove the old ones.
- Task 3 split: the missing table is task 3, and refusing an unknown version is task 3b, as each needs its own test.
- Task 5 split: `deleteRecord` is task 5 and moving `ItemRepository` is 5b. No template repository deletes, so the helper needs a template test of its own, or `unused` fails once `--remove-sample` drops the items. `ItemRepository` needs it first, so it comes first. Run `task test:init` at 5b, not 5.
- Task 5b: `task test:init` is red until the `@wip` tags go (it was at HEAD too), so tasks 5b, 7d, 7e, 9 and 10 check on a copy without the feature file, as the note under the task list says.
- Tasks 7b–7d and the benchmarks (now 8d) added after task 6, from a question about the cost of opening many files. `readRecords` opens every file even to find one record, so lookups get their own read, and benchmarks replace our estimates. This reverses 5b's note that `FindByID` and `Delete` keep the full read: a direct read can't pick up a conflict copy, because it has a different name, so only the key check is needed. The benchmarks come after 8 so that they cover the scan.
- Task 7b split: reading is 7b, a missing file 7b2 and a mismatched key 7b3, as each needs its own test. 7b4 is new: without it, moving lookups onto `readRecord` in 7c and 7d would drop task 3b's version refusal for them.
- Task 7d split: `FindByID` is 7d and `Delete` is 7e, as each needs its own test.
- Task 9 split: the warning is 9, and a failed scan is 9b, as it needs its own test. The spec says the app keeps working, so a failed scan warns rather than stopping it.
- Task 8 split: listing is 8, hidden folders 8b and files that aren't records 8c, as each needs its own test. The benchmarks move from 8b to 8d.
- Task 11 split: recent marks are 11, this device 11b, old marks 11c and unreadable files 11d, as each needs its own test.
- Task 13 split: one device is 13, several 13b and this device's own mark 13c, as each needs its own test.
- Task 11d split: files that aren't marks are 11d, and a folder with no marks is 11e, as each needs its own test.
- Task 12 split: writing is 12, the atomic replace 12b, renewing 12c and removing 12d, as each needs its own test.
- Task 14 split: marking is 14, an unsynced folder 14b, the setting 14c, renewing and stopping 14d and the steps 14e, as each needs its own test.
