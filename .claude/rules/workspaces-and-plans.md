---
paths:
  - "internal/core/**"
  - "internal/adapters/**"
  - "internal/bootstrap/**"
  - "internal/testsupport/**"
  - "test/acceptance/**"
  - "features/**"
  - "docs/workspaces.md"
  - "docs/features.md"
  - "docs/billing.md"
---

# Workspaces, plans, and domain errors (ADR 5)

## Workspaces

Data belongs to a workspace. Use cases get the current one from `ports.WorkspaceResolver` and pass it **explicitly** to every repository call. A workspace-scoped entity's repository port embeds `ports.WorkspaceRepository[T]`, treats another workspace's record as missing (`ErrNotFound`), and every adapter's contract tests run `storagetest.RunWorkspaceIsolation`, plus an isolation case for each extra query. With no workspace, a use case returns `domain.ErrNoWorkspace` before touching storage. `bootstrap` wires `workspacesFor(cfg)`: `workspace.Local`, or `workspace.FromContext` with sign-in on (the hosted edition, ADR 9), where the session middleware sets the workspace; tests use `testsupport.FixedWorkspace` or `NoWorkspace`. **Adding an entity: follow the checklist in `docs/workspaces.md`.**

## Plans and entitlements

Plans set a workspace's limits and turn its features on or off. A use case that a limit gates asks `ports.EntitlementsProvider` for the workspace's plan, counts, and checks with the domain's rules before saving: `Entitlements.CheckCreate` before adding one, `CheckChange` before changing one. Deleting is never limited, so a workspace over a read-only limit can get back within it. If the plan or the count can't be looked up, refuse the change: a failed lookup must never mean "no limits". Front ends read the plan through `ports.EntitlementsService` (`GET /api/v1/entitlements`). A use case that a feature gates checks `Entitlements.CheckFeature(key)` first, and refuses when the lookup fails; front ends only show the feature's state (`FeatureGate`, `cli.Gated`, `gatedKeyHelp`) and never decide it. `bootstrap` wires the local plan; tests use `testsupport.FixedEntitlements`. **Gating a feature: follow the checklist in `docs/features.md`.**

The local edition's plan is `bootstrap.LocalPlan`: `local`, unless a build sets another with `-ldflags "-X <module>/internal/bootstrap.LocalPlan=<plan>"`. The plans come from `internal/adapters/plans/plans.jsonc`, built into the binary, unless `STARTER_PLANS_FILE` names another file. An unknown local plan or a bad plans file fails at start. The file is JSON with full-line `//` comments only: a comment after a value is a parse error.

## Changing a plan

`ports.BillingService.ChangePlan(ctx, workspace, plan)` is the core's only billing concept, and the one use case that takes the workspace as an argument, because an admin or a payment provider acts on someone else's workspace. `starter admin set-plan` is its reference adapter, and it refuses `--server`: no API changes plans until roles exist. A payment provider is a driving adapter that calls it. A plan the catalog (`ports.PlanCatalog`) doesn't have is a `*domain.ValidationError` on `plan`, and a failed lookup refuses too. It checks no limits: a smaller plan keeps the data, which the over-limit rules then govern.

Each workspace's plan is stored through `ports.WorkspacePlanRepository`, which has its own `storagetest` contract suite and doesn't embed `WorkspaceRepository[T]`, since a plan is one value per workspace. `plans.StoredPlanOf` reads it into a `plans.Provider`, answering a default for a workspace with none stored and passing on any other error.

**The local edition's plan stays fixed:** it never reads a stored plan, and `bootstrap.NewBilling` refuses with `ErrLocalPlanIsFixed` (or `ErrNoPlanChangesOverAServer` with `--server`), opening no storage. **Stored plans are on with sign-in:** the hosted edition's provider reads `plans.StoredPlanOf`, falling back to the plans file's required `newWorkspacePlan`, with hints on, and `NewBilling` returns the real service on the configured storage. **Connecting a payment provider: follow `docs/billing.md`.**

## Domain errors

Domain errors (`domain.ErrNotFound`, `*domain.ValidationError`, `*domain.LimitError`, `*domain.FeatureError`, `domain.ErrNoWorkspace`) are the shared language. `httpapi/problem.go` maps them to HTTP, and `httpclient` maps them back. `ErrNoWorkspace` is a 401 `no_workspace` problem ("Please sign in to continue."), as a request with no valid session has no workspace.
