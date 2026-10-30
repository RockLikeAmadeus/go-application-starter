Feature: Install the web app
  # Presentation and caching only: nothing here reaches the core, so these
  # browser scenarios are the specification (there is no Godog counterpart).
  # They run on desktop Chromium, where Playwright supports service workers.

  @wip
  Scenario: The app can be installed
    When I open the app for the first time
    Then the app can be installed as "Starter"

  @wip
  Scenario: The app opens without a connection
    Given I have opened the app before
    When I open the app without a connection
    Then I see the navigation
    And the app says it is offline

  @wip
  Scenario: The offline notice goes when the connection comes back
    Given I have opened the app before
    And I open the app without a connection
    And the app says it is offline
    When the connection comes back
    Then the app no longer says it is offline
