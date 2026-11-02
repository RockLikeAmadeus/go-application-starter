# WIP: `/readyz` checks storage without the sample

- Backlog: [TODO.md](../../TODO.md), "`/readyz` checks storage without the sample".
- Spec: [SPEC.md §4.10](../../SPEC.md#410-tell-when-the-server-is-ready), and the **Ready** glossary entry.
- Scenarios: [`features/readiness.feature`](../../features/readiness.feature) (`@server`, `@wip`).
- Temporary: delete this file in the same change that removes the `@wip` tags.

## Decisions
- **The problem:** `/readyz`'s only check is `opts.Items.ListItems`, inside a sample block (`internal/adapters/httpapi/server.go`). After `--remove-sample` it always answers "ready", even when storage is broken. With no `Items` set, it panics on a nil interface and the recoverer turns that into a 500.
- **The check comes from the storage adapter, not an entity:** `httpapi.Options` gets `Ready func(context.Context) error`. Nil means always ready (memory storage). `/readyz` answers 503 `storage unavailable` when it fails, and logs the error so whoever runs the server can see why. The sample `ListItems` check goes.
- **What "ready" means for jsonfile:** `(*jsonfile.Store).Ready` creates and removes a temp file in the data folder. That catches a missing folder, a read-only remount and a full disk, which a plain `os.Stat` would miss. It runs on every probe and isn't cached, so a folder that comes back is ready again (the third scenario). A probe every few seconds writing an empty temp file is cheap. A future SQL adapter's `Ready` is its ping.
- **bootstrap owns the choice:** `App.Ready` is set by `NewEmbedded`: the store's `Ready` for jsonfile, nil for memory. `NewRemote` leaves it nil (the CLI and TUI serve no probes). `cmd/server` and `desktop.Open` pass `App.Ready` into `httpapi.Options`, so the desktop app's `/readyz` is honest too, although nothing probes it today.
- **Gherkin on a new `server` driver:** readiness is something whoever runs the server sees, so the scenarios are tagged `@server` and run only on a new `server` driver. It wires the server as `cmd/server` does (`bootstrap.NewEmbedded` with jsonfile on a temp data folder, then `httpapi.NewHandler` with `Ready`) and serves it with `httptest`. Every other driver gets `~@server`. The memory, jsonfile and http drivers have no data folder to remove, and desktop and mobile aren't servers.
- **No sample anywhere in this feature:** the feature file, its steps and the server driver have no `sample:start` blocks. `task test:init` runs `go test ./test/...` on a copy made with `--remove-sample`, so it proves `/readyz` works without the sample. That's the TODO item's "a test that runs without the sample". The server driver wires no `Items`, because these scenarios don't touch items.
- **The steps ask over HTTP** (`GET /readyz` with a plain `http.Client`), because readiness isn't a driving port: it's the HTTP adapter's probe. "Removing the data folder" is `os.RemoveAll` on the driver's temp folder, and "putting it back" is `os.MkdirAll`. The lock file goes with the folder. That's fine for these scenarios, and the store's `Close` must still not fail the cleanup (check in item 6).

## Tasks
1. [ ] jsonfile: an open `Store` is ready (`Store.Ready(ctx)` returns nil) — `store_test.go`.
2. [ ] jsonfile: a `Store` whose data folder has been removed isn't ready, and the error says why.
3. [ ] httpapi: `/readyz` answers 503 `storage unavailable` when `Options.Ready` fails, and logs the error — `server_test.go`, with no `Items` set.
4. [ ] httpapi: `/readyz` answers 200 `ready` with no `Ready` and no `Items` (today it panics into a 500). This removes the sample `ListItems` check.
5. [ ] bootstrap: `NewEmbedded` on jsonfile gives an `App.Ready` that fails once the data folder is removed. On memory, it's ready — `bootstrap_test.go`.
6. [ ] acceptance: the `server` driver (`@server` only, and `~@server` on the others) and the readiness steps, thin and with no sample block. The scenarios go green here. Check that cleanup tolerates the removed folder.
7. [ ] cmd/server: pass `App.Ready` into `httpapi.Options`. Pull the handler's options into a small function with a test in `cmd/server` (`/readyz` is 503 after the data folder is removed), so the wiring is test-driven.
8. [ ] desktop: `desktop.Open` passes `App.Ready` too — `open_test.go`: `/readyz` is 503 after the data folder is removed.
9. [ ] Docs: `features/README.md`, the drivers comment in `test/acceptance/drivers.go` and CLAUDE.md's `features/ + test/acceptance` line name the `@server` tag and the server driver. CLAUDE.md's Architecture/Gotchas mention `Options.Ready` if useful.
10. [ ] Finish: `task check`, `task test:init`, remove the `@wip` tags, tick the TODO.md item and delete this file.

## Changes to the plan
