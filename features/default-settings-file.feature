Feature: Create a default settings file
  As a user on my own computer
  I want the app to create a settings file when I don't have one
  So that I have a file to edit, holding the settings as they are

  # The @desktop scenarios run only through the desktop driver
  # (test/acceptance), which opens the desktop app on a temporary settings
  # folder with no settings file, and runs the command line beside it. The
  # @mobile scenario runs only through the mobile driver.

  @desktop @wip
  Scenario Outline: The desktop app creates a settings file holding the defaults
    Then my settings file sets "<setting>" to <value>

    Examples:
      | setting              | value                         |
      | storage              | "jsonfile"                    |
      | data_dir             | the desktop app's data folder |
      | ignore_synced_folder | false                         |

  @desktop @wip
  Scenario Outline: The created settings file leaves out the settings with no default
    Then my settings file doesn't set "<setting>"

    Examples:
      | setting    |
      | server_url |
      | plans_file |
      | port       |
      | log_format |

  @desktop @wip
  Scenario: The created settings file changes nothing
    When I close the desktop app
    And I ask the command line where my data is
    Then it names the desktop app's data folder

  @desktop @wip
  Scenario: The command line creates the settings file when there is none
    Given I close the desktop app
    And I delete my settings file
    When I ask the command line where my data is
    Then my settings file sets "storage" to "jsonfile"

  @desktop @wip
  Scenario: A settings file that's already there is kept as it is
    Given I replace my settings file with one choosing the data folder "work-data"
    When I close the desktop app and open it again
    And I close the desktop app
    And I ask the command line where my data is
    Then my settings file is as I left it

  @desktop @wip
  Scenario Outline: No settings file is created when another one is named
    Given I close the desktop app
    And I delete my settings file
    And I name a settings file of my own <how>
    When I ask the command line where my data is
    Then I have no settings file in my settings folder

    Examples:
      | how                    |
      | in the environment     |
      | with a command option  |

  @mobile @wip
  Scenario: The mobile app creates no settings file
    Then the mobile app has created no settings file
