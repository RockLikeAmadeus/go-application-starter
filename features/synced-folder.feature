Feature: Warn when SQLite data is in a synced folder
  As a user who keeps my data in SQLite
  I want to be warned when my data folder looks synced by a file-sync service
  So that the sync service doesn't damage my data without my knowing

  # The desktop scenarios run on the desktop driver, which opens the desktop
  # app on a temporary settings folder and runs the command line beside it.
  # The server scenario runs on the server driver. Which paths and files mean
  # a synced folder is checked by unit tests; these use one of each kind.

  @desktop
  Scenario Outline: The desktop app warns when its SQLite data is in a synced folder
    Given the desktop app keeps its data in SQLite in a folder synced by <service>
    Then the desktop app warns that its data folder looks synced by "<service>"

    Examples:
      | service   |
      | Dropbox   |
      | Syncthing |
      | OneDrive  |

  @desktop
  Scenario: The command line warns when its SQLite data is in a synced folder, and still works
    Given the desktop app keeps its data in SQLite in a folder synced by Dropbox
    When I ask the command line where my data is
    Then the command line warns that its data folder looks synced by "Dropbox"
    And it names the desktop app's data folder

  @desktop
  Scenario: SQLite data outside a synced folder gets no warning
    Given the desktop app keeps its data in SQLite
    Then the desktop app gives no warning

  @desktop
  Scenario: Plain-text data in a synced folder gets no warning
    Given the desktop app keeps its data as plain text in a folder synced by Dropbox
    Then the desktop app gives no warning

  @desktop
  Scenario: The user can turn the warning off
    Given the desktop app keeps its data in SQLite in a folder synced by Dropbox
    And I turn off the synced folder warning
    When I close the desktop app and open it again
    Then the desktop app gives no warning

  @desktop
  Scenario: The command line gives no warning once the user turns it off
    Given the desktop app keeps its data in SQLite in a folder synced by Dropbox
    And I turn off the synced folder warning
    When I ask the command line where my data is
    Then the command line gives no warning

  @server
  Scenario: A server keeping its SQLite data in a synced folder warns
    Given the server keeps its data in SQLite in a folder synced by Dropbox
    Then the server warns that its data folder looks synced by "Dropbox"
