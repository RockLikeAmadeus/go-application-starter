Feature: Make the template mine
  # Template tooling, not product behaviour: see docs/template.md.
  # Each scenario runs the init tool on a throwaway copy of the template.
  # Steps that don't name the description, publisher or email give the tool
  # sample ones, since it needs them.

  Background:
    Given a fresh copy of the template

  Scenario: Give the project its own module path
    When I make it mine with module "example.com/acme/tasks" and name "Acme"
    Then the project's Go module is "example.com/acme/tasks"

  Scenario Outline: Other names come from the brand name
    When I make it mine with module "example.com/acme/tasks" and name "<name>"
    Then the app is branded "<name>"
    And the command-line tool is called "<command>"
    And settings are read from environment variables starting "<prefix>"

    Examples:
      | name       | command    | prefix      |
      | Acme       | acme       | ACME_       |
      | Acme Tasks | acme-tasks | ACME_TASKS_ |

  Scenario: No trace of the template's name is left
    When I make it mine with module "example.com/acme/tasks" and name "Acme"
    Then no file mentions the template's name

  Scenario: The sample stays unless I ask to remove it
    When I make it mine with module "example.com/acme/tasks" and name "Acme"
    Then the project has the sample Item capability

  Scenario: Remove the sample
    When I make it mine with module "example.com/acme/tasks" and name "Acme", removing the sample
    Then the project has no sample Item capability
    And the specification has no sample Item capability

  Scenario: Removing the sample keeps what the front ends share
    When I make it mine with module "example.com/acme/tasks" and name "Acme", removing the sample
    Then the terminal app keeps its error and selection styles
    And the web app keeps its error message

  Scenario: The project without the sample passes the Go linters
    When I make it mine with module "example.com/acme/tasks" and name "Acme", removing the sample
    Then the project's Go code passes the linters

  Scenario Outline: The new project is healthy
    When I make it mine with module "example.com/acme/tasks" and name "Acme"<removal>
    Then the project builds and its Go tests pass

    Examples:
      | removal               |
      |                       |
      | , removing the sample |

  Scenario: The desktop packages describe the app and name its publisher
    When I make it mine with:
      | module      | example.com/acme/tasks |
      | name        | Acme Tasks             |
      | description | Tasks for busy teams   |
      | publisher   | Acme Ltd               |
      | email       | dev@acme.example       |
    Then the Linux packages are described as "Tasks for busy teams"
    And the Linux packages are maintained by "Acme Ltd <dev@acme.example>"
    And the app's menu entry is described as "Tasks for busy teams"
    And the Windows program is published by "Acme Ltd", with copyright "© Acme Ltd"
    And the Windows program is still described as "Acme Tasks"

  Scenario: No placeholder publisher is left in the desktop packages
    When I make it mine with module "example.com/acme/tasks" and name "Acme"
    Then the desktop packages name no placeholder publisher

  Scenario: The project remembers its names
    When I make it mine with module "example.com/acme/tasks" and name "Acme Tasks"
    Then the project records module "example.com/acme/tasks" and name "Acme Tasks"
    And the project can still be renamed

  Scenario: The init tool removes itself
    When I make it mine with module "example.com/acme/tasks" and name "Acme"
    Then the init tool and its documentation are gone

  Scenario Outline: Refuse invalid input and change nothing
    When I try to make it mine with module "<module>" and name "<name>"
    Then it fails saying "<message>"
    And the project is unchanged

    Examples:
      | module                 | name | message                             |
      |                        | Acme | --module is required                |
      | example.com/acme/tasks |      | --name is required                  |
      | not a module           | Acme | invalid module path                 |
      | example.com/acme/tasks | !!!  | name must contain a letter or digit |

  Scenario: Refuse to run with uncommitted changes
    Given the copy has uncommitted changes
    When I try to make it mine with module "example.com/acme/tasks" and name "Acme"
    Then it fails saying "commit or stash your changes first"
    And the project is unchanged

  Scenario Outline: Refuse missing or invalid package details and change nothing
    # An empty value leaves the detail out.
    When I try to make it mine with <detail> "<value>"
    Then it fails saying "<message>"
    And the project is unchanged

    Examples:
      | detail      | value          | message                                                    |
      | description |                | --description is required                                  |
      | publisher   |                | --publisher is required                                    |
      | email       |                | --email is required                                        |
      | email       | not an address | invalid email address                                      |
      | publisher   | Acme <Ltd>     | publisher must not contain < or >                          |
      | publisher   | Acme\\Ltd      | publisher must not contain double quotes or backslashes    |
      | description | Say "hi"       | description must not contain double quotes or backslashes  |
