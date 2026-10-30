@wip
Feature: Rename the project
  # Project tooling, not product behaviour: see docs/rename.md.
  # Each scenario runs the rename tool on a throwaway copy of the project as it
  # stands, whatever it is currently called.

  Background:
    Given a copy of the project

  Scenario Outline: All names come from the new brand name
    When I rename it to "<name>"
    Then the app is branded "<name>"
    And the command-line tool is called "<command>"
    And settings are read from environment variables starting "<prefix>"
    And the web package is called "<command>-web"

    Examples:
      | name         | command      | prefix        |
      | Zenith       | zenith       | ZENITH_       |
      | Zenith Notes | zenith-notes | ZENITH_NOTES_ |

  Scenario: No trace of the old names is left
    When I rename it to "Zenith Notes" with module "example.com/zenith/notes"
    Then no file mentions the old names

  Scenario: The module path stays unless I ask to change it
    # The old module path may contain the old slug; it must survive intact.
    When I rename it to "Zenith Notes"
    Then the project's Go module is unchanged

  Scenario: Move the module too
    When I rename it to "Zenith Notes" with module "example.com/zenith/notes"
    Then the project's Go module is "example.com/zenith/notes"

  Scenario: The project remembers its new names
    When I rename it to "Zenith Notes" with module "example.com/zenith/notes"
    Then the project records module "example.com/zenith/notes" and name "Zenith Notes"

  Scenario: Rename more than once
    When I rename it to "Zenith"
    And I commit the changes
    And I rename it to "Orbit"
    Then the app is branded "Orbit"
    And no file mentions "Zenith"

  Scenario: The renamed project is healthy
    When I rename it to "Zenith Notes" with module "example.com/zenith/notes"
    Then the project builds and its Go tests pass

  Scenario: Preview the changes first
    When I preview renaming it to "Zenith Notes"
    Then it lists the changes to "web/index.html"
    And it lists the command-line tool's move to "cmd/zenith-notes"
    And the project is unchanged

  Scenario Outline: Refuse invalid input and change nothing
    When I try to rename it with name "<name>" and module "<module>"
    Then it fails saying "<message>"
    And the project is unchanged

    Examples:
      | name   | module       | message                             |
      |        |              | --name is required                  |
      | Zenith | not a module | invalid module path                 |
      | !!!    |              | name must contain a letter or digit |

  Scenario: Refuse to rename to the current names
    When I try to rename it to its current name
    Then it fails saying "nothing to rename"
    And the project is unchanged

  Scenario: Refuse to run with uncommitted changes
    Given the copy has uncommitted changes
    When I try to rename it with name "Zenith" and module ""
    Then it fails saying "commit or stash your changes first"
    And the project is unchanged
