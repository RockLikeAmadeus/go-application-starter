@sign-in-provider
Feature: Sign in through a sign-in provider
  As someone using a hosted copy of the product
  I want to sign in with an account I already have at a sign-in provider
  So that the product never needs my password

  # These scenarios run on the oidc driver only: the hosted edition over HTTP,
  # signing in through a fake sign-in provider inside the tests. The rules
  # that hold for every way of signing in are in sign-in.feature, which runs
  # on the oidc driver too. The test plans start new workspaces on "free".

  @wip
  Scenario: Changing my name at the provider keeps my account
    Given "Ada" has signed in before
    And an admin has moved "Ada"'s workspace to the "pro" plan
    When "Ada" changes their name at the sign-in provider to "Ada Lovelace" and signs in again
    Then I am signed in as "Ada Lovelace"
    And my workspace's plan is "pro"

  @wip
  Scenario: Two people with the same name get their own accounts
    Given "Sam" has signed in before
    And an admin has moved "Sam"'s workspace to the "pro" plan
    When another person named "Sam" signs in
    Then I am signed in as "Sam"
    And my workspace's plan is "free"

  @wip
  Scenario: Creating an account starts at the provider's sign-up page
    When I create an account at the sign-in provider as "Grace"
    Then the sign-in provider showed me its sign-up page first
    And I am signed in as "Grace"
    And my workspace's plan is "free"

  @wip
  Scenario: A provider that can't start at sign-up shows its sign-in page instead
    Given the sign-in provider can't start at its sign-up page
    When I create an account at the sign-in provider as "Grace"
    Then the sign-in provider showed me its sign-in page first
    And I am signed in as "Grace"

  @wip
  Scenario: Cancelling at the provider doesn't sign me in
    When I cancel signing in at the sign-in provider
    Then I am told that signing in didn't work
    And I am not signed in

  @wip
  Scenario: Coming back from the provider without starting to sign in here is refused
    When someone comes back from the sign-in provider without having started to sign in on this server
    Then they are told that signing in didn't work
    And they are not signed in
