@desktop
Feature: Share local data between the desktop app and the command line
  As a user on my own computer
  I want the desktop app and the command line to work on the same data
  So that I find my data wherever I run them from

  # These scenarios run only through the desktop driver (test/acceptance),
  # which opens the desktop app on a temporary settings folder, with no data
  # folder chosen, and runs the command line beside it in the same way.

  @wip
  Scenario: The command line uses the desktop app's data folder
    When I close the desktop app
    And I ask the command line where my data is
    Then it names the desktop app's data folder

  @wip
  Scenario: The command line can't use the data while the desktop app is open
    When I ask the command line where my data is
    Then I am told the data is already in use

  @wip
  Scenario: The command line uses the data folder I choose
    When I ask the command line where my data is, choosing the data folder "elsewhere"
    Then it names the data folder "elsewhere"

  # sample:start
  @wip
  Scenario: Items added in the desktop app are seen from the command line
    Given an item named "Widget"
    When I close the desktop app
    And I list the items from the command line
    Then the command line shows "Widget"
  # sample:end
