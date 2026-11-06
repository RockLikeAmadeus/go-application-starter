# Formal wording for user-facing text

Replaces the existing user-facing text with the wording agreed with the user on 2026-11-06. It goes before task 14 of [synced plain text](synced-plain-text.md).

- Rule: [`.claude/rules/user-facing-text.md`](../../.claude/rules/user-facing-text.md)
- Spec: [SPEC.md §4.16](../../SPEC.md#416-keep-plain-text-data-in-a-synced-folder) quotes some of this text

## Decisions

Agreed with the user:
- **The wording below is final.** Use it exactly, including punctuation. Anything not listed needs agreeing first. Gherkin can keep the domain term "conflict copy", as it isn't shown to users.
- **A conflict is settled by deleting the copy, or replacing the original with it.** The old "keep the version you want and delete the other" was wrong: deleting the original leaves a copy whose name doesn't match its key, so the app skips it and the record disappears. `SPEC.md` lines 380 and 389 repeat this, so correct them with the warning.
- **The data folder lock's `--server` ending is for the CLI and TUI only.** The desktop app and the server have no such option.
- **The Oxford comma is mandatory,** including in lists the code joins, such as three or more plans. Today's `PlansPhrase` and `FeatureGate` don't add it.
- **Left as they are:** "Move a workspace to another plan", "Launch the interactive terminal UI", the `starter info` labels, "Nothing here yet.", "Page not found.", "No items yet.", "New item name", the `item` commands' help lines, and the TUI's "(locked)" notes other than the several-plans form.

Decided in planning:
- **The plan name in limit errors crosses the API.** `httpclient` rebuilds `LimitError` from the problem response, so `plan` goes in `api/openapi.yaml` as well as on the error.
- **`httpclient` strips the field name to recover a validation message.** The new pattern starts with "The {field}", so that has to change too.

## Agreed wording

Warnings (`internal/bootstrap`):
- **Another device, one:** Your data folder appears to be in use on {devices}. Running the application from two devices at once while plain-text storage is enabled can result in lost changes: please close the application on your other device before continuing.
- **Another device, several:** Your data folder appears to be in use on {devices}. Running the application from multiple devices at once while plain-text storage is enabled can result in lost changes: please close the application on your other devices before continuing.
- **{devices}, several:** "work-laptop and work-phone"; "work-laptop, work-phone, and home-pc".
- **One conflict copy:** Your data folder contains 1 conflicted copy, {name}, which appears to have been created by a file sync service. A change may be missing from your data. Please compare it with the original file, then either delete it or use it to replace the original before continuing.
- **Several conflict copies:** Your data folder contains {n} conflicted copies that appear to have been created by a file sync service: {names}. Changes may be missing from your data. Please compare each one with its original file, then either delete it or use it to replace the original before continuing.
- **Scan failed:** Your data folder could not be checked for conflicted copies, so a change may be missing from your data. The following error occurred: {error}
- **SQLite in a synced folder:** Your data folder appears to be synced by {service}. A file sync service can damage SQLite data: please move your data folder out of the synced location before continuing. Alternatively, you can switch to plain-text storage, but your existing data will not be transferred automatically.

Plans:
- **Feature, one plan:** {feature} is not included in your current plan. To use this feature, please upgrade to the {plan} plan.
- **Feature, several plans:** {feature} is not included in your current plan. To use this feature, please upgrade to one of the following plans: pro or team.
- **Feature, no plan:** {feature} is not included in your current plan.
- **Limit reached:** You have reached the limit of {max} {limit} on the {plan} plan.
- **Read only:** Your {limit} are read-only because you have exceeded the limit of {max} {limit} on the {plan} plan.
- **Web note:** Available on the pro plan / Available on the following plans: pro or team / Not available
- **CLI note:** (not included in your current plan; available on the pro plan) / (not included in your current plan). Several plans: "available on the following plans: pro or team".
- **TUI note:** (locked: available on the pro plan) / (locked). Several plans: "available on the following plans: pro or team".

Errors:
- **Local plan:** The plan for the local edition is set when the application is built and cannot be changed.
- **Over a server:** A plan cannot be changed through a server connection.
- **Internal:** An unexpected error occurred on the server. Please try again later.
- **Not found:** The requested record could not be found.
- **Validation:** The name cannot be empty. / The name cannot be longer than 200 characters. / The specified plan does not exist. Other fields follow the same pattern.
- **Unknown setting:** The settings file {path} contains an unknown setting: "colour".
- **Unreadable settings:** The settings file {path} could not be read: {error}
- **Storage setting:** The storage setting "disk" is not recognized. Please use "memory", "jsonfile", or "sqlite".
- **Log format setting:** The log format setting "xml" is not recognized. Please use "text" or "json".
- **Data folder locked, CLI and TUI:** Your data folder is in use by another program on this computer ({path}). Please close the other program and try again, or connect to a running server with --server.
- **Data folder locked, elsewhere:** Your data folder is in use by another program on this computer ({path}). Please close the other program and try again.
- **Other server error:** The server returned an error ({status}): {detail}

Command line:
- **`info` help:** Show the data location and the current plan
- **`admin` help:** Commands for product administrators
- **`admin set-plan` output:** Workspace {workspace} is now on the {plan} plan.
- **`item delete` output:** Item {id} has been deleted.
- **Prefixes:** Warning: {text} and Error: {text}, the second in `starter`, the server and the desktop app.

## Tasks

Each is one `/red` → `/green` → `/refactor` cycle.

- [x] 1. The one conflict copy warning, with `SPEC.md`'s rule corrected.
  - Red: `TestApp_WarnsAboutConflictCopiesOnPlainText` pins the one-copy wording; its two one-copy cases failed on the old text.
  - Green: `conflictCopyWarning` uses the agreed text; `SPEC.md`'s rule and example corrected.
  - Refactor: nothing changed; the test spells the text out on purpose, and tasks 1b and 1c rewrite the rest of the function.
- [x] 1b. The several conflict copies warning.
  - Red: `TestApp_WarnsAboutConflictCopiesOnPlainText/several_copies` pins the agreed wording; failed on the old text.
  - Green: `conflictCopyWarning` uses the agreed text.
  - Refactor: nothing changed. The two wordings stay whole so each reads as written; nothing else quotes the old text.
- [x] 1c. The scan failed warning.
  - Red: `TestApp_WarnsWhenConflictCopiesCantBeCheckedAndKeepsWorking` pins the agreed wording; failed on the old text.
  - Green: `conflictCopyWarning` uses the agreed text.
  - Refactor: nothing changed. The code comments keep the domain term "conflict copy"; nothing else quotes the old text.
- [x] 2. The one other device warning.
  - Red: `TestApp_WarnsWhenAnotherDeviceIsUsingPlainText` pins the agreed wording; failed on the old text.
  - Green: `otherDeviceWarning` uses the agreed text for one device; `SPEC.md`'s example and the acceptance step's expected text updated.
  - Refactor: nothing changed. Task 2b rewrites the other branch; the Gherkin's "looks in use" stays, as users don't read it.
- [x] 2b. The two other devices warning.
  - Red: `TestApp_WarningNamesEveryOtherDevice` pins the agreed wording; failed on the old text.
  - Green: `otherDeviceWarning` uses the agreed text, joining the names with " and ".
  - Refactor: nothing changed. Task 2c replaces the join, so a helper now would be guesswork.
- [x] 2c. The warning for three or more other devices, with the Oxford comma.
  - Red: `TestApp_WarningListsThreeOtherDevicesWithTheOxfordComma` pins the agreed join; failed with "home-pc and work-laptop and work-phone".
  - Green: `otherDeviceWarning` joins three or more names with commas and ", and ".
  - Refactor: moved the join into `deviceList`. Left separate from `PlansPhrase`: plans end in "or", and tasks 4 and 5 rewrite it.
- [x] 3. The SQLite synced folder warning.
  - Red: `TestApp_WarnsWhenSQLiteDataLooksSynced` pins the agreed wording; failed on the old text.
  - Green: `syncedFolderWarning` text replaced; acceptance and terminal steps' expected text and `SPEC.md`'s example updated.
  - Refactor: nothing changed. Left the old text in the `Layout.test.tsx` and `styles_test.go` fixtures, which are arbitrary strings; the Gherkin's "looks synced" stays, as users don't read it.
- [x] 4. The feature error, in its one-plan form.
  - Red: `TestEntitlements_CheckFeature/off_and_locked` pins the agreed wording; failed on the old text.
  - Green: `FeatureError.Error` returns the agreed text for one plan.
  - Refactor: nothing changed. Left the separate one-plan branch, as 4b and 4c rewrite the others.
- [x] 4b. The feature error, in its several-plans form.
  - Red: `TestEntitlements_CheckFeature/off_and_locked,_on_two_plans` pins the agreed wording; failed on the old text.
  - Green: `FeatureError.Error` returns the agreed text for several plans.
  - Refactor: extracted `planList`, shared with `PlansPhrase`, and the shared opening sentence. Left the missing Oxford comma for task 5, which also fixes this error through `planList`.
- [ ] 4c. The feature error, in its no-plan form.
- [ ] 5. The web, CLI and TUI plan notes, with the Oxford comma in lists of three or more plans (`planList` in `domain`, so the feature error gets it too).
- [ ] 6. Limit errors name the plan, through `api/openapi.yaml` and `httpclient`.
- [ ] 7. The local plan, over a server, internal and not found errors.
- [ ] 8. The validation pattern, with `httpclient` recovering the message.
- [ ] 9. The settings errors.
- [ ] 10. The data folder locked error, with the `--server` ending for the CLI and TUI only.
- [ ] 11. The other server error.
- [ ] 12. The `info` and `admin` help, and the `set-plan` and `item delete` output.
- [ ] 13. The `Warning:` and `Error:` prefixes.
- [ ] 14. Update the docs that quote old text, run `task test:init` and `task check`, and delete this file.

## Changes to the plan

- Task 1 split: the one-copy, several-copies and scan-failed warnings each need their own test, so they are tasks 1, 1b, and 1c.
- Task 2 split: the one-device and several-device warnings each need their own test, so they are tasks 2 and 2b.
- Task 2b split: two devices and three or more each need their own test, so they are tasks 2b and 2c.
- Task 4 split: the one-plan, several-plans, and no-plan feature errors each need their own test, so they are tasks 4, 4b, and 4c.
