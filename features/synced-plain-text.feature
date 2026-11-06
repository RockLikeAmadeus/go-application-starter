@desktop @wip
Feature: Keep plain-text data in a synced folder
  As a user who keeps plain-text data in a folder a file-sync service syncs
  I want to be told about clashes and about other devices using my data
  So that I don't lose changes without knowing

  # These run on the desktop driver, which opens the desktop app on a temporary
  # settings folder and runs the command line beside it. A step that leaves a
  # conflict copy or a device's mark writes the file as a sync service would
  # bring it. What counts as a conflict copy for each sync service, and
  # renewing the mark, are checked by unit tests.

  Rule: The app reports conflict copies, and keeps working

    Scenario: The desktop app warns about a conflict copy
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And a sync service left the conflict copy "note (conflicted copy)" in the data folder
      When I close the desktop app and open it again
      Then the desktop app warns about the conflict copy "note (conflicted copy)"

    Scenario: The command line warns about a conflict copy, and still works
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And a sync service left the conflict copy "note (conflicted copy)" in the data folder
      When I close the desktop app
      And I ask the command line where my data is
      Then the command line warns about the conflict copy "note (conflicted copy)"
      And it names the desktop app's data folder

    Scenario: The warning goes once the user deletes the conflict copy
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And a sync service left the conflict copy "note (conflicted copy)" in the data folder
      And I close the desktop app and open it again
      When I delete the conflict copy "note (conflicted copy)"
      And I close the desktop app and open it again
      Then the desktop app gives no warning

    Scenario: A conflict copy gets a warning even if the folder doesn't look synced
      Given a sync service left the conflict copy "note (conflicted copy)" in the data folder
      When I close the desktop app and open it again
      Then the desktop app warns about the conflict copy "note (conflicted copy)"

    Scenario: Turning off the synced folder warning leaves the conflict copy warning on
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And I turn off the synced folder warning
      And a sync service left the conflict copy "note (conflicted copy)" in the data folder
      When I close the desktop app and open it again
      Then the desktop app warns about the conflict copy "note (conflicted copy)"

    # sample:start
    Scenario: Each item is kept in its own file
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And an item named "Widget"
      And an item named "Gadget"
      Then each item has its own file in the data folder

    Scenario: A conflict copy of an item doesn't show up as another item
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And an item named "Widget"
      And a sync service left a conflict copy of "Widget" in the data folder
      When I close the desktop app and open it again
      Then the items are:
        | Widget |
    # sample:end

  Rule: The app warns when another device seems to be using the data

    Scenario: The desktop app warns when another device is using the data
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And the device "work-laptop" marked the data folder as in use 1 minute ago
      When I close the desktop app and open it again
      Then the desktop app warns that the data folder looks in use on "work-laptop"

    Scenario: The command line warns when another device is using the data, and still works
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And the device "work-laptop" marked the data folder as in use 1 minute ago
      When I close the desktop app
      And I ask the command line where my data is
      Then the command line warns that the data folder looks in use on "work-laptop"
      And it names the desktop app's data folder

    Scenario: A device that stopped long ago doesn't count
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And the device "work-laptop" marked the data folder as in use 1 hour ago
      When I close the desktop app and open it again
      Then the desktop app gives no warning

  Rule: A copy of the app marks a synced folder as in use while it has it open

    Scenario: The desktop app marks a synced folder as in use on this device
      When the desktop app keeps its data as plain text in a folder synced by Dropbox
      Then the data folder is marked as in use on this device

    Scenario: Closing the desktop app removes the mark
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      When I close the desktop app
      Then the data folder isn't marked as in use

    Scenario: A folder that doesn't look synced gets no mark
      Then the data folder isn't marked as in use

    Scenario: Turning off the synced folder warning means no mark
      Given the desktop app keeps its data as plain text in a folder synced by Dropbox
      And I turn off the synced folder warning
      When I close the desktop app and open it again
      Then the data folder isn't marked as in use
