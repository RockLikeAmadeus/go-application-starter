Feature: Keep data in Postgres
  As someone running the hosted server
  I want its data kept in a Postgres database
  So that several copies of the server, on different machines, can share it

  # Every other feature's business scenarios also run on Postgres storage (the
  # postgres driver in test/acceptance). These cover what only Postgres
  # changes, on the server driver. All of them need a Postgres database
  # (STARTER_TEST_DATABASE_URL, from `task db:up`). Without one they are
  # skipped locally, and CI always runs them.

  @server @postgres @wip
  Scenario: A server keeping its data in Postgres is ready
    Given the server keeps its data in Postgres
    Then the server says it is ready

  @server @postgres @wip
  Scenario: A server whose Postgres database can't be reached isn't ready
    Given the server keeps its data in Postgres
    When the server's database can't be reached
    Then the server says it isn't ready, because its data is unavailable

  @server @postgres @wip
  Scenario: A server is ready again once its Postgres database is back
    Given the server keeps its data in Postgres
    And the server's database can't be reached
    When the server's database can be reached again
    Then the server says it is ready

  # sample:start
  @server @postgres @wip
  Scenario: Servers sharing a Postgres database see each other's items
    Given two servers keep their data in the same Postgres database
    When an item named "Widget" is added through the first server
    Then the second server lists "Widget"
  # sample:end
