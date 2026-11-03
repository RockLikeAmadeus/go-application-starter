Feature: Keep data in SQLite
  As a user, or someone running the server
  I want my data kept in a SQLite database
  So that several programs can use it at the same time

  # Every other feature's business scenarios also run on SQLite storage (the
  # sqlite driver in test/acceptance). These cover what only SQLite changes:
  # the desktop ones run on the desktop driver, the server ones on the server
  # driver.

  @desktop
  Scenario: The command line can use SQLite data while the desktop app is open
    Given the desktop app keeps its data in SQLite
    When I ask the command line where my data is
    Then it names the desktop app's data folder

  # sample:start
  @desktop
  Scenario: Items added in the desktop app are seen from the command line while it is open
    Given the desktop app keeps its data in SQLite
    And an item named "Widget"
    When I list the items from the command line
    Then the command line shows "Widget"

  @desktop
  Scenario: Items kept in SQLite are still there after the desktop app is reopened
    Given the desktop app keeps its data in SQLite
    And an item named "Widget"
    When I close the desktop app and open it again
    Then the items are:
      | Widget |
  # sample:end

  @server
  Scenario: A server keeping its data in SQLite is ready
    Given the server keeps its data in SQLite
    Then the server says it is ready

  @server
  Scenario: A server whose SQLite data is gone isn't ready
    Given the server keeps its data in SQLite
    When the server's data folder is removed
    Then the server says it isn't ready, because its data is unavailable
