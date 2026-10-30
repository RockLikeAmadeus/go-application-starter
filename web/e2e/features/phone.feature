Feature: Use on a phone
  # Presentation only: the layout never reaches the core, so these browser
  # scenarios are the specification (there is no Godog counterpart).
  # @phone scenarios run on emulated phones (see playwright.config.ts);
  # the rest run on a desktop browser.

  @wip @phone
  Scenario: The menu starts closed on a phone
    When I open the app on my phone
    Then I can open the menu
    And the navigation is hidden

  @wip @phone
  Scenario: Open the menu
    Given I open the app on my phone
    When I open the menu
    Then I see the navigation
    And I see the Appearance controls

  @wip @phone
  Scenario Outline: Close the menu
    Given I open the app on my phone
    And I open the menu
    When I <dismiss>
    Then the navigation is hidden

    Examples:
      | dismiss                       |
      | go to the home page from it   |
      | press Escape                  |
      | tap outside the menu          |

  @wip @phone
  Scenario: Change the style from the menu
    Given I open the app on my phone
    And I open the menu
    When I switch to the Soft style
    Then the app uses the Soft style
    And I see the navigation

  @wip @phone
  Scenario Outline: Pages fit a phone screen
    Given my system prefers a <theme> scheme
    And I open the app on my phone
    And I open the menu
    And I switch to the <style> style
    When I press Escape
    Then the page does not scroll sideways

    Examples:
      | style    | theme |
      | Crisp    | light |
      | Crisp    | dark  |
      | Terminal | light |
      | Terminal | dark  |
      | Soft     | light |
      | Soft     | dark  |

  @wip
  Scenario: A wide screen shows the navigation without a menu
    When I open the app for the first time
    Then I see the navigation
    And there is no menu
