# Workspaces: keeping each workspace's data apart

Every piece of data belongs to a **workspace**, and a request only ever sees or changes its own workspace's data. This guide explains how that works, then walks through adding a new entity that follows the rule. The decision behind it is [ADR 5](adr/0005-hosted-and-local-editions.md); the business rules are in `SPEC.md`, "Keep each workspace's data separate".

**Why every entity needs it:** the same core runs in two editions. The **local edition** (CLI, TUI, desktop and mobile apps, and a self-hosted server) has exactly one workspace. The **hosted edition** serves many users from one server, each in their own workspace. If one entity skips the rule, the hosted edition leaks that entity's data between users, and nothing else in the stack would stop it. So the rule is built in at the core, and checked by tests every storage backend must pass.

## How it works

```
driving adapter ──► use case ──► WorkspaceResolver.CurrentWorkspace(ctx) ──► workspace
 (HTTP, CLI, …)        │                                                       │
                       └──────────► repository.FindAll(ctx, workspace) ◄───────┘
```

1. **A use case asks for the current workspace** through a driven port, `ports.WorkspaceResolver`. It never reads the context itself, and a driving adapter never passes a workspace as a parameter. The one exception is `ChangePlan` ([`docs/billing.md`](billing.md)), where an admin or a payment provider acts on someone else's workspace.
2. **It passes the workspace explicitly to every repository call** (`FindByID`, `FindAll`, `Delete`). `Save` uses the workspace the record carries, which the use case set when it created the record.
3. **A repository treats another workspace's record as missing.** It returns `domain.ErrNotFound`, exactly as for an ID that doesn't exist. The caller can't even tell whether the record exists, so IDs never leak another workspace's data.
4. **No workspace means no data.** If the resolver can't name a workspace, it returns `domain.ErrNoWorkspace`, and the use case returns that before touching storage. This is "fail closed": a bug that loses the workspace refuses the request rather than showing everyone's data.

### The resolvers

`internal/core/ports` declares the port. `internal/adapters/workspace` implements it:

| Resolver | Used by | Answers |
|---|---|---|
| `workspace.Local{}` | every local-edition front end, wired in `internal/bootstrap` | always `domain.LocalWorkspace` (`"local"`) |
| `workspace.FromContext{}` | the hosted edition (a server with the `sign_in` setting on), wired in `internal/bootstrap`; the acceptance tests | the ID that `workspace.WithID(ctx, id)` put in the context, or `ErrNoWorkspace` if it's missing or empty |

`internal/bootstrap` picks the resolver from the `sign_in` setting, in `workspacesFor`. With sign-in on, the server's session middleware reads the session cookie and calls `workspace.WithID` with the signed-in account's workspace. A request with no valid session gets no workspace, so the use cases refuse it with `ErrNoWorkspace`. [docs/sign-in.md](sign-in.md) explains sign-in. Front ends in the local edition don't set anything: `Local` answers for them.

For tests, `internal/testsupport` has two fakes: `testsupport.FixedWorkspace("ws-a")` always resolves to its own value, and `testsupport.NoWorkspace{}` always returns `ErrNoWorkspace`.

## Adding a workspace-scoped entity

The template's sample entity, `Item`, follows every step below, so while the sample is still in your project, read its files alongside this guide (`internal/core/domain/item.go`, `internal/core/app/item_service.go`, the storage adapters' `item_repository.go`). The steps use a made-up `Widget`.

### 1. Domain: the entity carries its workspace

```go
type Widget struct {
	ID          string
	WorkspaceID WorkspaceID
	Name        string
	// …
}

func NewWidget(workspace WorkspaceID, id, name string, now time.Time) (Widget, error) {
	// validate, then:
	return Widget{ID: id, WorkspaceID: workspace, Name: name /* … */}, nil
}
```

The workspace is set once, at creation, and never changes. Nothing outside the use case's create path should set it.

### 2. Port: embed `ports.WorkspaceRepository`

```go
// WidgetRepository is the driven port for persisting Widgets.
type WidgetRepository interface {
	ports.WorkspaceRepository[domain.Widget]
	// Add anything else the entity needs, and give every query a workspace:
	// FindByName(ctx context.Context, workspace domain.WorkspaceID, name string) (domain.Widget, error)
}
```

`WorkspaceRepository[T]` gives you `Save`, and `FindByID`, `FindAll` and `Delete` taking the workspace. **Every extra query method takes the workspace too**, and returns only that workspace's records.

### 3. Use cases: resolve first, then pass it on

```go
type WidgetService struct {
	repo       ports.WidgetRepository
	workspaces ports.WorkspaceResolver
	clock      ports.Clock
	ids        ports.IDGenerator
}

func (s *WidgetService) CreateWidget(ctx context.Context, name string) (domain.Widget, error) {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return domain.Widget{}, err
	}
	w, err := domain.NewWidget(workspace, s.ids.NewID(), name, s.clock.Now())
	// … save and return
}

func (s *WidgetService) RenameWidget(ctx context.Context, id, name string) (domain.Widget, error) {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return domain.Widget{}, err
	}
	w, err := s.repo.FindByID(ctx, workspace, id) // another workspace's widget → ErrNotFound
	// … change it and Save: it keeps its own workspace
}
```

Every use case opens with the same resolve-or-return block, before any other call. A change to an existing record always starts by finding it **in the current workspace**, so a request can't change another workspace's record by guessing its ID.

### 4. Storage adapters: filter every read, check every delete

Each adapter keeps the workspace with the record and checks it on every lookup:

- **memory:** a `lookup(workspace, id)` helper that treats a record in another workspace as missing, used by `FindByID` and `Delete`; `FindAll` keeps only the workspace's records.
- **jsonfile:** the on-disk record gets a `workspaceId` field. `FindByID` and `Delete` read the one record with `readRecord` and treat another workspace's as missing (the sample's `workspaceRecord`), and `FindAll` keeps only the workspace's records from `readRecords`.
- **sqlstore (SQLite and Postgres):** written once, for both engines ([ADR 8](adr/0008-postgres-storage.md)). A new goose migration in `internal/adapters/storage/sqlstore/migrations/` (the next number) creates the table, with a `workspace_id` column in the primary key (`PRIMARY KEY (workspace_id, id)`), and the workspace in any unique index that should be per workspace (`UNIQUE (workspace_id, name)`). The repository in `sqlstore` takes a `*sql.DB` and puts `WHERE workspace_id = $1` on **every** query, including updates and deletes, in SQL both engines run (`$n` placeholders, `INSERT … ON CONFLICT (…) DO UPDATE SET …`). Item's `00002_items.sql` and `item_repository.go` are the pattern. Its contract test runs on SQLite and on Postgres (`TestItemRepositoryContractOnSQLite`, `…OnPostgres`, which skips without `STARTER_TEST_DATABASE_URL`).

Then run the isolation suite from the adapter's contract tests:

```go
storagetest.RunWorkspaceIsolation(t,
	func(t *testing.T) ports.WorkspaceRepository[domain.Widget] { return newRepo(t) },
	storagetest.RecordKind[domain.Widget]{
		New:       func(recordWorkspace domain.WorkspaceID, id string) domain.Widget { return domain.Widget{ID: id, WorkspaceID: recordWorkspace, Name: "W"} },
		ID:        func(w domain.Widget) string { return w.ID },
		Workspace: func(w domain.Widget) domain.WorkspaceID { return w.WorkspaceID },
	})
```

It checks that a record keeps its workspace, that `FindAll` lists only the given workspace's records, and that `FindByID` and `Delete` treat another workspace's record as missing. **It doesn't know about your extra query methods:** give each one its own isolation test in the entity's contract suite.

### 5. Wiring

- `internal/bootstrap`, `NewEmbedded`: build the service with `workspacesFor(cfg)`, beside `system.Clock{}` and `system.UUIDs{}`. It answers `workspace.Local{}` in the local edition and `workspace.FromContext{}` with sign-in on.
- The acceptance harness (`test/acceptance/drivers.go`): the memory, jsonfile, sqlite, and postgres drivers build services with `workspace.FromContext{}` (sqlite and postgres share `harness.useSQLStore`), and the http driver with `workspace.Local{}`, or with `workspace.FromContext{}` when it runs as the hosted edition (`startHostedHTTP`). A `Before` hook in `steps.go` already puts every scenario in the local workspace, so your feature's scenarios don't need to name one.

### 6. Tests

- **Use cases:** build the service with `testsupport.FixedWorkspace(…)` and check new records get that workspace. Add a fail-closed table test: with `testsupport.NoWorkspace{}` and a repository that counts its calls, every use case returns `ErrNoWorkspace` and the repository is never called. Item's `TestUseCases_WithoutAWorkspaceAreRefusedBeforeStorage` is the pattern.
- **Storage:** the isolation suite above, plus a case for each extra query method.
- **Acceptance (optional):** the entity's business scenarios run in the local workspace without saying so. If the entity has workspace rules of its own, add scenarios tagged `@workspaces`. They run only on the memory, jsonfile, sqlite, postgres, and http drivers. On memory, jsonfile, sqlite, and postgres, their steps build a context with `workspace.WithID`. On http, each workspace is a signed-in account's personal workspace, so the steps sign in as that account.

### Checklist

- [ ] The entity has a `WorkspaceID`, set only by its constructor.
- [ ] Its repository port embeds `ports.WorkspaceRepository[T]`, and every extra query takes the workspace.
- [ ] Every use case resolves the workspace first, and passes it to every repository call.
- [ ] Every storage adapter filters every read and checks the workspace on every delete and lookup.
- [ ] Every adapter's contract tests call `storagetest.RunWorkspaceIsolation`, plus an isolation case for each extra query.
- [ ] A fail-closed table test covers every use case.
- [ ] `bootstrap` passes `workspacesFor(cfg)`, and the acceptance harness passes the right resolver for each driver.

## When data isn't per workspace

Some data belongs to the whole product, not to a workspace: the plans config, built-in reference data, the list of workspaces itself. Don't force it into a workspace. Give it a repository that doesn't embed `WorkspaceRepository`, and say in its port's doc comment why it isn't scoped. Keep such data read-only to ordinary requests, or a request in one workspace could change what every other workspace sees.

Some data is one value per workspace rather than a list of records, such as which plan a workspace is on. It's still kept apart by workspace, but `WorkspaceRepository[T]` doesn't fit it. Give it its own port and contract suite, as `ports.WorkspacePlanRepository` has, with a case proving that saving one workspace's value leaves another's alone.

## Pitfalls

- **Taking the workspace from the request.** A workspace ID in an API body, query string, or CLI flag would let anyone pick any workspace. Only the resolver decides. (`starter admin set-plan` names a workspace, but it runs beside the storage, so whoever can run it already has the data.)
- **A new query without the workspace.** Search, count, "find by name", exports: each one is a new way to leak. Give it the workspace and its own isolation test.
- **Returning "forbidden" for another workspace's record.** That tells the caller the record exists. Return `ErrNotFound`.
- **A cache or index shared across workspaces.** Key it by workspace.
- **Over HTTP, a missing workspace is a 401.** `httpapi/problem.go` maps `ErrNoWorkspace` to a 401 `no_workspace` problem, which asks the person to sign in, and `httpclient` maps it back. So a new handler needs no sign-in check of its own: calling the use case is enough.
