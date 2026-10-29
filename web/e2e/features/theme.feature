@wip
Feature: Choose a theme
  # Presentation only: the theme never reaches the core, so these browser
  # scenarios are the specification (there is no Godog counterpart).

  Scenario Outline: A first visit follows the system preference
    Given my system prefers <system>
    When I open the app for the first time
    Then the app uses the <theme> theme

    Examples:
      | system           | theme |
      | a dark scheme    | dark  |
      | a light scheme   | light |
      | no colour scheme | dark  |

  Scenario Outline: Switch theme
    Given my system prefers <system>
    And I open the app for the first time
    When I switch to the <chosen> theme
    Then the app uses the <chosen> theme

    Examples:
      | system         | chosen |
      | a dark scheme  | light  |
      | a light scheme | dark   |

  Scenario: A chosen theme is remembered over the system preference
    Given my system prefers a dark scheme
    And I open the app for the first time
    And I switch to the light theme
    When I come back to the app later
    Then the app uses the light theme
