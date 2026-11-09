# 3. Plain-text storage: one JSON file per record

- Status: accepted (for the walking skeleton). Revised in step 5 of the editions plan: one file per record, in place of one file per domain type.
- Date: 2026-10-29

## Context

We want zero-infrastructure persistence that is human-readable and easy to replace later with SQL or NoSQL.

The data folder may also sit in a folder that a file-sync service keeps in sync, used from one device at a time (ADR 5, SPEC.md §4.16). A sync service settles a clash by keeping both files, so the smaller each file is, the less a clash touches.

## Decision

The `jsonfile` adapter keeps each domain type in its own folder of a data directory, much like a table, and each record in its own file in that folder:

```
<dir>/<table>/<key>.json   {"version": 2, "key": "<key>", "record": { "id": "...", "workspaceId": "...", "name": "...", ... }}
<dir>/.lock                advisory lock held while a Store is open
```

- The file name is the key, escaped with `url.QueryEscape` so that any key (with `/` or `:`, say) is one file name on every platform.
- The key is in the file as well as in its name, because not every record holds its own (a workspace's plan doesn't). It is how conflict copies are found (below).
- The version is per file, because a sync service can bring files across at different times. A file with another version is refused. Version 1 (one `<table>.json` per domain type, holding every record) is ignored, not migrated: the template had no users.
- Writes go to a temp file beside the record's, then fsync, then an atomic rename. Deleting a record removes its file.
- Reading one record (`FindByID`, `Delete`, a workspace's plan) opens only that record's file. Listing a table (`FindAll`) opens every file in its folder.
- Reading a table skips what isn't a `.json` file (temp files, a sync service's placeholders) and every conflict copy.
- An in-process RW mutex serialises access.
- A lock on `data/.lock` stops a second process from opening the same directory: an advisory `flock` on Unix, `LockFileEx` on Windows.
- On-disk records are separate structs from domain entities, so the format can evolve. `version` allows migrations.
- Every adapter must pass `storagetest.RunItemRepositoryContract`.
- Each record carries the workspace it belongs to (ADR 5). Lookups return only the given workspace's records, and treat another workspace's as missing. Plain text is only ever used by the local edition, so a data folder holds one workspace's records, and the files need no folder per workspace.

### Conflict copies

A sync service that finds two versions of a file keeps both, renaming one: `6f1c (Sam's conflicted copy 2026-11-06).json` (Dropbox), `6f1c.sync-conflict-20261106-120000-ABCDEFG.json` (Syncthing), `6f1c-DESKTOP-ABC123.json` (OneDrive), `6f1c (1).json` (Google Drive), `6f1c 2.json` (iCloud), and so on. A copy keeps the original's contents, so **a file whose key doesn't match its own name is a conflict copy**, whichever service named it, with no list of name patterns. That also covers a user's `copy of 6f1c.json`.

`Store.ConflictCopies` lists them, looking one level down, in folders that aren't hidden, at `.json` files that decode as records. So `.devices`, `.stfolder`, `config.toml` and other programs' files in a shared folder aren't conflict copies. `internal/bootstrap` warns about them (SPEC.md §4.16).

YAML was considered. JSON was chosen because it needs only the standard library and has no typing ambiguities.

## Consequences

- A write touches only its record's file, so a clash in a synced folder is limited to that record.
- Listing a table opens every file in it. On a laptop (Intel Ultra 7 155U, small records, files in the page cache), reading a table takes about 6 ms at 1,000 records and 67 ms at 10,000, and the conflict copy scan about the same. Reading one record takes about 5 µs at either size (`table_bench_test.go`, `task bench`). That is fine up to roughly 10,000 records in all; beyond that, or on a slow disk, it is the signal to move to SQLite.
- **Single writer only**: run exactly one server instance per volume (the Fly config enforces one machine). The CLI/TUI must use `--server` while the server holds the directory.
- Migrating to a database means: add an adapter, pass the contract suite, and switch it in `internal/bootstrap`. The core does not change.
- That's what SQLite storage did ([ADR 7](0007-sqlite-storage.md)): with `storage = "sqlite"`, several programs can share a data folder, so the single-writer limit applies to plain text only. Plain text stays the default.
- iCloud can replace a file it has moved off the device with a hidden placeholder (`.<name>.json.icloud`). Reading skips it, so that record looks missing until iCloud downloads it again.
