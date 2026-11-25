@server
Feature: Tell when the server is ready
  As someone running the server
  I want it to say whether it can reach its data
  So that users are only sent to a server that can serve them

  # These scenarios run only through the server driver (test/acceptance),
  # which wires the server as cmd/server does, on a temporary data folder.
  # Those that remove it (@removes-data-folder) don't run on Windows, which
  # refuses to delete files the server still has open.

  Scenario: A server that can reach its data is ready
    Then the server says it is ready

  @removes-data-folder
  Scenario: A server that can't reach its data isn't ready
    When the server's data folder is removed
    Then the server says it isn't ready, because its data is unavailable

  @removes-data-folder
  Scenario: A server whose data comes back is ready again
    Given the server's data folder is removed
    When the server's data folder is put back
    Then the server says it is ready
