Feature: Items in a terminal
  # Wiring journeys only: the sample's commands and screen, run in a terminal
  # as a user would. Item rules are specified in /features/items.feature.

  Scenario: Add an item with the command line
    When I add an item named "Terminal widget" with the command line
    Then the command line lists "Terminal widget"

  Scenario: The command line reports a mistake
    When I add an item named "   " with the command line
    Then it fails, saying "The name cannot be empty."

  Scenario: Add an item in the terminal app
    Given I open the terminal app
    When I add an item named "Typed widget" in the terminal app
    Then the terminal app lists "Typed widget"

  Scenario Outline: Move through the items with <down> and <up>
    Given I have added the items "First" and "Second"
    And I open the terminal app
    When I press <down>
    Then "Second" is selected
    When I press <up>
    Then "First" is selected

    Examples:
      | down | up   |
      | j    | k    |
      | down | up   |

  Scenario: Export the items with the command line
    Given I have added the items "First" and "Second"
    When I export the items with the command line
    Then the export lists "First" and then "Second"

  Scenario Outline: The command line shows item export <shown> when the plan doesn't have it
    Given a plans file whose local plan doesn't have item export, which is <shown> when off
    When I ask the command line for help
    Then item export is <listed> in the help
    When I export the items with the command line
    Then it fails, saying "Item export is not included in your current plan."

    Examples:
      | shown        | listed                          |
      | shown locked | listed as disabled on your plan |
      | hidden       | not listed                      |
