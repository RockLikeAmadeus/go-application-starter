@desktop
Feature: Keep settings in a settings file
  As a user on my own computer
  I want to keep the app's settings in a file
  So that I don't have to set them each time I start it

  # These scenarios run only through the desktop driver (test/acceptance),
  # which opens the desktop app on a temporary settings folder and runs the
  # command line beside it.

  @wip
  Scenario: The settings file chooses the data folder
    Given my settings file chooses the data folder "work-data"
    When I close the desktop app
    And I ask the command line where my data is
    Then it names the data folder "work-data"

  @wip
  Scenario: The desktop app uses the settings file's data folder
    Given my settings file chooses the data folder "work-data"
    When I close the desktop app and open it again
    Then the desktop app keeps its data in "work-data"

  @wip
  Scenario: A setting chosen when starting overrides the settings file
    Given my settings file chooses the data folder "work-data"
    And I start with the data folder "scratch" chosen in the environment
    When I close the desktop app
    And I ask the command line where my data is
    Then it names the data folder "scratch"

  @wip
  Scenario: The command line says where a setting came from
    Given my settings file chooses the data folder "work-data"
    And I start with the data folder "scratch" chosen in the environment
    When I close the desktop app
    And I ask the command line where my data is
    Then it says the data folder was chosen by "STARTER_DATA_DIR"

  @wip
  Scenario: A settings file with a setting the app doesn't know stops it
    Given my settings file has the setting "colour" set to "blue"
    When I close the desktop app
    And I ask the command line where my data is
    Then I am told the settings file is invalid, because of "colour"
