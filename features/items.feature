Feature: Manage items
  As a user
  I want to keep a list of named items
  So that I can track the things I care about

  # Business-rule scenarios. These run through the core's driving port
  # (ports.ItemService) against several adapter combinations; see
  # test/acceptance. Browser journeys live in web/e2e/features instead.

  Scenario: Create an item
    When I create an item named "Widget"
    Then the items are:
      | Widget |

  Scenario: Item names are trimmed
    When I create an item named "  Widget  "
    Then the items are:
      | Widget |

  Scenario Outline: Reject invalid item names
    When I try to create an item named <name>
    Then I am told the "name" is invalid
    And there are no items

    Examples:
      | name  |
      | ""    |
      | "   " |

  Scenario: Items are listed oldest first
    Given an item named "First"
    And an item named "Second"
    And an item named "Third"
    Then the items are:
      | First  |
      | Second |
      | Third  |

  Scenario: Rename an item
    Given an item named "Old name"
    When I rename "Old name" to "New name"
    Then the items are:
      | New name |

  Scenario: Delete an item
    Given an item named "Keep"
    And an item named "Discard"
    When I delete "Discard"
    Then the items are:
      | Keep |

  Scenario: Deleting a missing item fails
    When I try to delete an item that does not exist
    Then I am told the item was not found
