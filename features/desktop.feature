@desktop
Feature: Use the desktop app
  As a user on my own computer
  I want the app to keep my data on my computer
  So that I can use it without a server or a connection

  # These scenarios run only through the desktop driver (test/acceptance),
  # which opens the desktop app's wiring on a temporary data folder. The
  # desktop driver also runs every other feature's scenarios.

  # sample:start
  Scenario: Items are still there after the desktop app is reopened
    Given an item named "Widget"
    When I close the desktop app and open it again
    Then the items are:
      | Widget |
  # sample:end

  Scenario: Only one copy of the desktop app can use a data folder
    When I try to open a second copy of the desktop app on the same data folder
    Then I am told the data is already in use

  Scenario: The desktop app is on the local plan
    Then my workspace's plan is "local"

  # sample:start
  @wip
  Scenario: The desktop app can export items
    Given an item named "Widget"
    Then item export is on for my workspace
    When I export the items
    Then the export lists:
      | Widget |
  # sample:end
