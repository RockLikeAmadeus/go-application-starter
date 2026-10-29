# 1. Hexagonal architecture with a Go core

- Status: accepted
- Date: 2026-10-29

## Context

The application must support several user interfaces (web, CLI, TUI, and possibly more later) and several storage technologies over its life, without rewriting business logic.

## Decision

Use ports and adapters:

- `internal/core/domain`: entities and business rules; stdlib only.
- `internal/core/ports`: interfaces. *Driving* ports (e.g. `ItemService`) are what UIs call. *Driven* ports (e.g. `ItemRepository`, `Clock`, `IDGenerator`) are what the core needs.
- `internal/core/app`: use cases implementing the driving ports.
- `internal/adapters/*`: everything that touches the outside world.
- `internal/bootstrap`: the single composition root.

The dependency rule (core never imports adapters) is enforced by `depguard` in CI.

`ports.ItemService` is also implemented by an HTTP client adapter. The CLI and TUI therefore run unchanged against either the embedded core or a remote server.

## Consequences

- Each UI and storage backend can be added or replaced in isolation.
- Acceptance tests run the same Gherkin through several adapter combinations cheaply.
- There is more boilerplate (DTO ↔ domain mapping in each adapter) than in a layered CRUD app.
