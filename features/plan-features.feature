@plans
Feature: Features that plans turn on or off
  As someone running the product
  I want each plan to turn features on or off
  So that paid plans can offer features that free ones don't

  # Item export is the sample feature: it proves the gating that every
  # feature uses. These scenarios choose each workspace's plan, so they run
  # only on the drivers whose plans the tests set: memory, jsonfile, sqlite
  # and http (see test/acceptance). The desktop and mobile apps are on the
  # local plan (desktop.feature, mobile.feature).

  @wip
  Scenario: A plan with the feature lets the workspace use it
    Given the "pro" plan has item export
    And my workspace is on the "pro" plan
    And an item named "First"
    And an item named "Second"
    When I export the items
    Then the export lists:
      | First  |
      | Second |

  @wip
  Scenario Outline: A plan without the feature refuses it
    Given the "pro" plan has item export
    And the "free" plan doesn't have item export
    And item export is <shown> when off
    And my workspace is on the "free" plan
    And an item named "First"
    When I try to export the items
    Then I am told item export isn't on my plan<hint>

    Examples:
      | shown        | hint                                   |
      | shown locked | , and is available on the "pro" plan   |
      | hidden       |                                        |

  @wip
  Scenario Outline: A workspace can see whether its plan has the feature
    Given the "pro" plan has item export
    And the "free" plan doesn't have item export
    And item export is <shown> when off
    And my workspace is on the "<plan>" plan
    Then item export is <state> for my workspace

    Examples:
      | shown        | plan | state                                |
      | shown locked | pro  | on                                   |
      | hidden       | pro  | on                                   |
      | shown locked | free | locked, available on the "pro" plan  |
      | hidden       | free | hidden                               |

  @wip
  Scenario: Moving to a plan with the feature turns it on
    Given the "pro" plan has item export
    And the "free" plan doesn't have item export
    And my workspace is on the "free" plan
    And an item named "First"
    When my workspace moves to the "pro" plan
    And I export the items
    Then the export lists:
      | First |

  @wip
  Scenario: Moving to a plan without the feature turns it off
    Given the "pro" plan has item export
    And the "free" plan doesn't have item export
    And my workspace is on the "pro" plan
    When my workspace moves to the "free" plan
    And I try to export the items
    Then I am told item export isn't on my plan, and is available on the "pro" plan
