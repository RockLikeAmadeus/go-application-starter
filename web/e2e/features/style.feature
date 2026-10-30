Feature: Choose a style
  # Presentation only: the style never reaches the core, so these browser
  # scenarios are the specification (there is no Godog counterpart).

  Scenario: A first visit uses the default style
    When I open the app for the first time
    Then the app uses the Crisp style

  Scenario Outline: Switch style
    Given I open the app for the first time
    When I switch to the <style> style
    Then the app uses the <style> style

    Examples:
      | style    |
      | Terminal |
      | Soft     |

  Scenario: A chosen style is remembered
    Given I open the app for the first time
    And I switch to the Terminal style
    When I come back to the app later
    Then the app uses the Terminal style

  Scenario: Changing the style keeps the theme
    Given my system prefers a dark scheme
    And I open the app for the first time
    When I switch to the Soft style
    Then the app uses the Soft style
    And the app uses the dark theme
