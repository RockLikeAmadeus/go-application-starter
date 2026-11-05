Feature: The command line in a terminal
  # Wiring journeys only: they run the built `starter` program in a terminal,
  # as a user would, to prove its options, output and status reach the core
  # and back. Business rules are specified in /features and run with Godog.

  Scenario: The command line says where my data is
    Given my data is in the folder "work-data"
    When I ask the command line where my data is
    Then it says my data is in the folder "work-data"

  Scenario: The command line works on a server's data
    Given a server is running
    When I ask the command line where my data is, connected to the server
    Then it says my data is on the server

  Scenario: A command the command line doesn't know fails
    When I give the command line the command "frobnicate"
    Then it fails, saying "unknown command"

  @wip
  Scenario: The command line can't change the local edition's plan
    When I ask the command line to move the workspace "local" to the "pro" plan
    Then it fails, saying "set when the app is built"
