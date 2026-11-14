Feature: Signing in from the command line, in a terminal
  # Wiring journeys only: they run the built `starter` program and a built
  # server with the dev sign-in, and stand in for the browser with a helper
  # that the command line opens instead. Business rules are specified in
  # /features/command-line-sign-in.feature and run with Godog.

  Scenario: Signing in from the command line opens the browser and says who I am
    Given a server with the dev sign-in is running
    When I sign in to the server from the command line as "Ada"
    Then the command line says I am signed in as "Ada"
    And the command line says I am signed in as "Ada" on the server

  Scenario: The command line can't sign in to a server that doesn't sign people in
    Given a server is running with sign-in off
    When I try to sign in to the server from the command line
    Then it fails, saying the server doesn't sign people in

  Scenario: The terminal app offers to sign me in
    Given a server with the dev sign-in is running
    When I open the terminal app, connected to the server, without signing in
    Then the terminal app offers to sign me in

  Scenario: Signing in from the terminal app
    Given a server with the dev sign-in is running
    And I have opened the terminal app, connected to the server, without signing in
    When I sign in from the terminal app as "Ada"
    Then the terminal app says I am signed in as "Ada"

  Scenario: Signing out from the terminal app
    Given a server with the dev sign-in is running
    And I have signed in from the terminal app as "Ada"
    When I sign out from the terminal app
    Then the terminal app offers to sign me in
    And the command line says I am not signed in on the server
