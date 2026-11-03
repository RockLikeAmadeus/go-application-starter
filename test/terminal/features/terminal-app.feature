Feature: The terminal app
  # Wiring journeys only: they open the built `starter tui` in a terminal and
  # work it with the keyboard, as a user would. What each screen does is
  # specified in /features and by the TUI's unit tests.

  Scenario: Open and quit the terminal app
    When I open the terminal app
    Then it says how to quit
    When I quit the terminal app
    Then the terminal shows what it showed before

  @wip
  Scenario: The terminal app warns when my SQLite data is in a synced folder
    Given my data is kept in SQLite in a folder synced by Dropbox
    When I open the terminal app
    Then it warns that my data folder looks synced by "Dropbox"
