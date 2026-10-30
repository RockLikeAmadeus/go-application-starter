@mobile @wip
Feature: Use the mobile app
  As a user with an Android phone
  I want the app to keep my data on my phone
  So that I can use it without a server or a connection

  # These scenarios run only through the mobile driver (test/acceptance),
  # which opens the mobile app's wiring on a temporary data folder and sends
  # every request over the same bridge the Android app uses. The mobile
  # driver also runs every other feature's scenarios, except the desktop app's.

  # sample:start
  Scenario: Items are still there after the mobile app is reopened
    Given an item named "Widget"
    When I close the mobile app and open it again
    Then the items are:
      | Widget |
  # sample:end
