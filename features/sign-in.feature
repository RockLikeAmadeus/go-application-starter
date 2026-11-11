@sign-in
Feature: Sign in to the hosted edition
  As someone using a hosted copy of the product
  I want to sign in to my own account
  So that I work in my own workspace, on its plan, apart from everyone else

  # These scenarios run on the drivers whose storage keeps accounts and
  # sessions: memory, sqlite, postgres, and http. Each signs in with the dev
  # sign-in, which stands in for a real sign-in provider (OIDC comes in part
  # 8b of the editions plan). The test plans start new workspaces on "free".

  @wip
  Scenario: Signing in for the first time gives me a workspace on the starting plan
    When I sign in as "Ada"
    Then I am signed in as "Ada"
    And my workspace's plan is "free"

  @wip
  Scenario: Signing in again brings me back to my workspace
    Given "Ada" has signed in before
    And an admin has moved "Ada"'s workspace to the "pro" plan
    When I sign in as "Ada"
    Then my workspace's plan is "pro"

  @wip
  Scenario: Each account gets its own workspace
    Given "Ada" has signed in before
    And an admin has moved "Ada"'s workspace to the "pro" plan
    When I sign in as "Grace"
    Then my workspace's plan is "free"

  @wip
  Scenario: Nothing can be seen without signing in
    When someone who hasn't signed in tries to see their workspace's plan
    Then they are told to sign in

  @wip
  Scenario: Signing out ends the session
    Given I am signed in as "Ada"
    When I sign out
    And I try to see my workspace's plan
    Then I am told to sign in

  # sample:start
  @wip
  Scenario: What one account adds, another doesn't see
    Given "Ada" has signed in before and added an item named "Widget"
    When I sign in as "Grace"
    Then there are no items
  # sample:end
