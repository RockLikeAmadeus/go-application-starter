Feature: Signing in from the command line, in a terminal
  # Wiring journeys only: they run the built `starter` program and a built
  # server with the dev sign-in, and stand in for the browser with a helper
  # that the command line opens instead. Business rules are specified in
  # /features/command-line-sign-in.feature and run with Godog.

  @wip
  Scenario: Signing in from the command line opens the browser and says who I am
    Given a server with the dev sign-in is running
    When I sign in to the server from the command line as "Ada"
    Then the command line says I am signed in as "Ada"
    And the command line says I am signed in as "Ada" on the server

  @wip
  Scenario: The command line can't sign in to a server that doesn't sign people in
    Given a server is running with sign-in off
    When I try to sign in to the server from the command line
    Then it fails, saying the server doesn't sign people in

  @wip
  Scenario: The terminal app tells me to sign in
    Given a server with the dev sign-in is running
    When I open the terminal app, connected to the server, without signing in
    Then it tells me to sign in from the command line
