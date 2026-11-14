@remote-mode
Feature: Use the desktop and mobile apps on a hosted server
  As someone with an account on a hosted copy of the product
  I want the desktop and mobile apps to work on my own workspace there
  So that I can use the apps on the same data as the web app and the command line

  # These scenarios run on the http and oidc drivers, whose servers sign
  # people in. The desktop and mobile apps in them are the apps' own wiring,
  # opened in-process on a temporary settings folder and set to the driver's
  # server; the mobile app's requests go over the bridge the Android app
  # uses. The apps sign in through a browser the tests drive, with the dev
  # sign-in on the http driver and the fake sign-in provider on the oidc
  # driver. The test plans start new workspaces on "free".

  @wip
  Scenario Outline: Signing in from an app works on my workspace
    Given "Ada" has signed in before
    And an admin has moved "Ada"'s workspace to the "pro" plan
    When I sign in from the <app> as "Ada"
    Then the <app> is signed in as "Ada"
    And the <app> says my workspace's plan is "pro"

    Examples:
      | app         |
      | desktop app |
      | mobile app  |

  @wip
  Scenario Outline: An app on a server signs in through the browser
    When I open the <app> on the server
    Then the <app> offers to sign in through the browser

    Examples:
      | app         |
      | desktop app |
      | mobile app  |

  @wip
  Scenario Outline: An app can't see anything without signing in
    When the <app> tries to see my workspace's plan without signing in
    Then the <app> tells me to sign in

    Examples:
      | app         |
      | desktop app |
      | mobile app  |

  @wip
  Scenario Outline: Creating an account from an app signs me in to a new workspace
    When I create an account from the <app> as "Grace"
    Then the <app> is signed in as "Grace"
    And the <app> says my workspace's plan is "free"

    Examples:
      | app         |
      | desktop app |
      | mobile app  |

  @wip
  Scenario Outline: Cancelling a sign-in leaves the app signed out
    Given the <app> is waiting for me to sign in in the browser
    When I cancel signing in from the <app>
    Then the <app> is not signed in
    And the <app> keeps no session for the server

    Examples:
      | app         |
      | desktop app |
      | mobile app  |

  @wip
  Scenario Outline: Signing out from an app ends its session
    Given I have signed in from the <app> as "Ada"
    When I sign out from the <app>
    And the <app> tries to see my workspace's plan
    Then the <app> tells me to sign in
    And the <app> keeps no session for the server

    Examples:
      | app         |
      | desktop app |
      | mobile app  |

  @wip
  Scenario: Signing in from the command line doesn't sign the desktop app in
    Given I have signed in from the command line on this computer as "Ada"
    When the desktop app tries to see my workspace's plan
    Then the desktop app tells me to sign in

  @wip
  Scenario: Signing out from the desktop app leaves the command line signed in
    Given I have signed in from the command line on this computer as "Ada"
    And I have signed in from the desktop app as "Ada"
    When I sign out from the desktop app
    Then the command line is signed in as "Ada"

  @wip
  Scenario Outline: An app built for the server opens on it with no setting
    Given the <app> was built for the server, with no server setting
    When I sign in from the <app> as "Ada"
    Then the <app> is signed in as "Ada"

    Examples:
      | app         |
      | desktop app |
      | mobile app  |

  # sample:start
  @wip
  Scenario Outline: What I add from an app is in my workspace on the web
    Given I have signed in from the <app> as "Ada"
    When I create an item named "Widget" from the <app>
    And I sign in as "Ada" in the browser
    Then the items are:
      | Widget |

    Examples:
      | app         |
      | desktop app |
      | mobile app  |
  # sample:end
