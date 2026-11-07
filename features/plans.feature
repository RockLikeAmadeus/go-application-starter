@plans
Feature: Plans and limits
  As someone running the product
  I want each workspace's plan to set its limits
  So that paid plans can offer more than free ones

  # These scenarios choose each workspace's plan, so they run only on the
  # drivers whose plans the tests set: memory, jsonfile and http (see
  # test/acceptance). The desktop and mobile apps are on the local plan
  # (desktop.feature, mobile.feature).

  Scenario: A workspace can see its plan and its limits
    Given the "free" plan allows 3 items, and blocks new ones when over
    And my workspace is on the "free" plan
    Then my workspace's plan is "free"
    And its item limit is 3, which blocks new ones when over

  Scenario: Items can be created up to the limit
    Given the "free" plan allows 3 items, and blocks new ones when over
    And my workspace is on the "free" plan
    When I create an item named "First"
    And I create an item named "Second"
    And I create an item named "Third"
    Then the items are:
      | First  |
      | Second |
      | Third  |

  Scenario Outline: At the limit, a new item is refused
    Given the "free" plan allows 3 items, and <behaviour> when over
    And my workspace is on the "free" plan
    And an item named "First"
    And an item named "Second"
    And an item named "Third"
    When I try to create an item named "Fourth"
    Then I am told the limit of 3 items on the "free" plan is reached
    And the items are:
      | First  |
      | Second |
      | Third  |

    Examples:
      | behaviour                 |
      | blocks new ones           |
      | makes them read only      |

  Scenario: A plan without an item limit allows any number of items
    Given the "pro" plan has no limits
    And my workspace is on the "pro" plan
    And an item named "First"
    And an item named "Second"
    And an item named "Third"
    When I create an item named "Fourth"
    Then the items are:
      | First  |
      | Second |
      | Third  |
      | Fourth |

  Scenario: Over a limit that blocks new ones, items can still be changed and deleted
    Given the "pro" plan has no limits
    And the "free" plan allows 3 items, and blocks new ones when over
    And my workspace is on the "pro" plan
    And an item named "First"
    And an item named "Second"
    And an item named "Third"
    And an item named "Fourth"
    When my workspace moves to the "free" plan
    And I rename "First" to "Renamed"
    And I delete "Second"
    Then the items are:
      | Renamed |
      | Third   |
      | Fourth  |

  Scenario: Over a limit that makes them read only, items can't be changed
    Given the "pro" plan has no limits
    And the "free" plan allows 3 items, and makes them read only when over
    And my workspace is on the "pro" plan
    And an item named "First"
    And an item named "Second"
    And an item named "Third"
    And an item named "Fourth"
    When my workspace moves to the "free" plan
    And I try to rename "First" to "Renamed"
    Then I am told items are read only while over the limit of 3 on the "free" plan
    And the items are:
      | First  |
      | Second |
      | Third  |
      | Fourth |

  Scenario: Over a limit that makes them read only, items can still be deleted
    Given the "pro" plan has no limits
    And the "free" plan allows 3 items, and makes them read only when over
    And my workspace is on the "pro" plan
    And an item named "First"
    And an item named "Second"
    And an item named "Third"
    And an item named "Fourth"
    When my workspace moves to the "free" plan
    And I delete "Fourth"
    Then the items are:
      | First  |
      | Second |
      | Third  |

  Scenario: Back within a read-only limit, items can be changed again
    Given the "pro" plan has no limits
    And the "free" plan allows 3 items, and makes them read only when over
    And my workspace is on the "pro" plan
    And an item named "First"
    And an item named "Second"
    And an item named "Third"
    And an item named "Fourth"
    When my workspace moves to the "free" plan
    And I delete "Fourth"
    And I rename "First" to "Renamed"
    Then the items are:
      | Renamed |
      | Second  |
      | Third   |
