# 2. Spec-first REST with OpenAPI

- Status: accepted
- Date: 2026-10-29

## Context

The web UI needs an API. Options considered: REST + OpenAPI, ConnectRPC/gRPC, GraphQL.

## Decision

Spec-first REST, with `api/openapi.yaml` as the single source of truth:

- Go server interface: `oapi-codegen` strict server on `net/http`.
- Go client (used by the CLI/TUI remote mode): `oapi-codegen` client.
- TypeScript types: `openapi-typescript`, called through `openapi-fetch`.
- Errors: RFC 9457 problem documents with a machine-readable `type`.

## Consequences

- The frontend and backend cannot drift; CI fails on stale generated code.
- The API is universally consumable (curl, other languages).
- Revisit if streaming or many internal service-to-service calls appear (ConnectRPC), or if clients need flexible graph queries (GraphQL). Adding either would be a new driving adapter; the core would not change.
