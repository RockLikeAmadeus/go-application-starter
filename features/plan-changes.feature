Feature: Change a workspace's plan
  As someone running the product
  I want an admin to move a workspace to another plan
  So that each workspace gets the plan its owner pays for

  # The scenarios that move a workspace choose plans (@plans), so they run on
  # the drivers whose plans the tests set: memory, jsonfile, sqlite and http.
  # The http driver's admin works on the server's storage, as
  # `starter admin set-plan` does beside a server: the API can't change plans.
  # The one that switches between workspaces (@workspaces) runs on memory,
  # jsonfile and sqlite only. The local edition's refusal runs on the desktop
  # driver, which is the local edition.

  @plans
  Scenario: An admin moves a workspace to another plan
    Given my workspace is on the "free" plan
    When an admin moves my workspace to the "pro" plan
    Then my workspace's plan is "pro"

  @plans
  Scenario: Moving a workspace to a plan that doesn't exist is refused
    Given my workspace is on the "free" plan
    When an admin tries to move my workspace to the "platinum" plan
    Then I am told the "plan" is invalid
    And my workspace's plan is "free"

  @plans @workspaces
  Scenario: Moving one workspace leaves the others on their plans
    Given workspace "Alpha" is on the "free" plan
    And workspace "Beta" is on the "free" plan
    When an admin moves workspace "Alpha" to the "pro" plan
    Then workspace "Alpha"'s plan is "pro"
    And workspace "Beta"'s plan is "free"

  @desktop
  Scenario: The local edition's plan can't be changed
    When I try to move my workspace to the "pro" plan from the command line
    Then I am told the local edition's plan is set when the app is built
    And my workspace's plan is "local"
