@workspaces
Feature: Keep each workspace's data separate
  As someone using a product that many people share
  I want my workspace's data kept apart from everyone else's
  So that nobody else can see or change it

  # These scenarios run only on the memory, jsonfile, sqlite, and postgres drivers (see
  # test/acceptance). Until sign-in exists, nothing over HTTP can pick a
  # workspace, and the desktop and mobile apps are the local edition, with one
  # workspace. The local edition's single workspace is covered by every other
  # feature, which runs without naming a workspace.

  Scenario: Each workspace lists only its own items
    Given workspace "Alpha" has an item named "Widget"
    And workspace "Beta" has an item named "Gadget"
    Then workspace "Alpha" has the items:
      | Widget |
    And workspace "Beta" has the items:
      | Gadget |

  Scenario Outline: Another workspace's item can't be reached
    Given workspace "Alpha" has an item named "Widget"
    When someone in workspace "Beta" tries to <action> the item "Widget" from workspace "Alpha"
    Then I am told the item was not found
    And workspace "Alpha" has the items:
      | Widget |

    Examples:
      | action |
      | view   |
      | rename |
      | delete |

  Scenario Outline: Nothing can be done without a workspace
    When someone with no workspace tries to <action>
    Then they are told they have no workspace

    Examples:
      | action                          |
      | list the items                  |
      | create an item named "Widget"   |
