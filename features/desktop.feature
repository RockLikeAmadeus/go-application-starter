@desktop @wip
Feature: Use the desktop app
  As a user on my own computer
  I want the app to keep my items on my computer
  So that I can use it without a server or a connection

  # These scenarios run only through the desktop driver (test/acceptance),
  # which opens the desktop app's wiring on a temporary data folder. The
  # desktop driver also runs every scenario in items.feature.

  Scenario: Items are still there after the desktop app is reopened
    Given an item named "Widget"
    When I close the desktop app and open it again
    Then the items are:
      | Widget |

  Scenario: Only one copy of the desktop app can use a data folder
    When I try to open a second copy of the desktop app on the same data folder
    Then I am told the data is already in use
