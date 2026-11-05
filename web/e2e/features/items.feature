Feature: Items in the browser
  # UI journeys only: prove the browser -> API -> core wiring.
  # Business rules are specified in /features and run with Godog.

  Scenario: Add an item
    Given I am on the items page
    When I add an item named "Browser widget"
    Then I see "Browser widget" in the item list

  Scenario: Rename an item
    Given I am on the items page
    And I have added an item named "Before rename"
    When I rename "Before rename" to "After rename"
    Then I see "After rename" in the item list
    And I do not see "Before rename" in the item list

  Scenario: Delete an item
    Given I am on the items page
    And I have added an item named "Short-lived"
    When I delete "Short-lived"
    Then I do not see "Short-lived" in the item list

  Scenario: See validation errors
    Given I am on the items page
    When I add an item named ""
    Then I see the error "name must not be empty"

  Scenario: Export the items
    Given I am on the items page
    And I have added an item named "Exported widget"
    When I export the items
    Then the downloaded file lists "Exported widget"
