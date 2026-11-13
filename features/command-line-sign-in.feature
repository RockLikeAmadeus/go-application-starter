@command-line-sign-in
Feature: Sign in from the command line
  As someone with an account on a hosted copy of the product
  I want to sign in from the command line
  So that the command line and the terminal app work on my own workspace

  # These scenarios run on the http and oidc drivers, whose servers sign
  # people in. The command line in them is a client the tests run in-process:
  # it signs in through a browser the tests drive, with the dev sign-in on the
  # http driver and the fake sign-in provider on the oidc driver. The test
  # plans start new workspaces on "free".

  @wip
  Scenario: Signing in from the command line works on my workspace
    Given "Ada" has signed in before
    And an admin has moved "Ada"'s workspace to the "pro" plan
    When I sign in from the command line as "Ada"
    Then the command line is signed in as "Ada"
    And the command line says my workspace's plan is "pro"

  @wip
  Scenario: The command line can't see anything without signing in
    When the command line tries to see my workspace's plan without signing in
    Then the command line tells me to sign in

  @wip
  Scenario: Signing out from the command line ends its session
    Given I have signed in from the command line as "Ada"
    When I sign out from the command line
    And the command line tries to see my workspace's plan
    Then the command line tells me to sign in
    And the command line keeps no session for the server

  @wip
  Scenario: Signing out from the command line leaves my browser signed in
    Given I am signed in as "Ada" in the browser
    And I have signed in from the command line as "Ada"
    When I sign out from the command line
    Then I am still signed in as "Ada" in the browser

  @wip
  Scenario: A sign-in code works only once
    Given I have signed in from the command line as "Ada"
    When someone tries to use the same sign-in code again
    Then they are not given a session

  # sample:start
  @wip
  Scenario: What I add from the command line is in my workspace on the web
    Given I have signed in from the command line as "Ada"
    When I create an item named "Widget" from the command line
    And I sign in as "Ada" in the browser
    Then the items are:
      | Widget |
  # sample:end
