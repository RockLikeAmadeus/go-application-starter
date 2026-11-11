@sign-in
Feature: Sign in in the browser
  # UI journey only: proves the browser -> /auth/ -> session cookie -> API
  # wiring with the dev sign-in. The rules are specified in
  # /features/sign-in.feature and run with Godog. These scenarios run in
  # their own Playwright project, against a server with sign-in on.

  Scenario: Sign in, use the app, and sign out
    Given I am signed out
    When I sign in as "Browser Ada"
    Then I see the items page
    And I am shown as signed in as "Browser Ada"
    When I sign out
    Then I am asked to sign in
