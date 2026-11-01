# 3. Plain-text storage: one JSON file per domain type

- Status: accepted (for the walking skeleton)
- Date: 2026-10-29

## Context

We want zero-infrastructure persistence that is human-readable and easy to replace later with SQL or NoSQL.

## Decision

The `jsonfile` adapter stores each domain type in one file in a data directory (`data/items.json`), much like one table per file:

```json
{ "version": 1, "records": { "<id>": { "id": "...", "workspaceId": "...", "name": "...", ... } } }
```

- Writes go to a temp file, then fsync, then an atomic rename.
- An in-process RW mutex serialises access.
- An advisory `flock` on `data/.lock` stops a second process from opening the same directory.
- On-disk records are separate structs from domain entities, so the format can evolve. `version` allows migrations.
- Every adapter must pass `storagetest.RunItemRepositoryContract`.
- Each record carries the workspace it belongs to (ADR 5). Lookups return only the given workspace's records, and treat another workspace's as missing.

YAML was considered. JSON was chosen because it needs only the standard library and has no typing ambiguities.

## Consequences

- Every write rewrites the whole file. That is fine up to roughly 10k records per type, and beyond that it is the signal to migrate.
- **Single writer only**: run exactly one server instance per volume (the Fly config enforces one machine). The CLI/TUI must use `--server` while the server holds the directory.
- Migrating to a database means: add an adapter, pass the contract suite, and switch it in `internal/bootstrap`. The core does not change.
