---
paths:
  - "internal/core/**"
  - "internal/adapters/**"
  - "internal/bootstrap/**"
  - "internal/testsupport/**"
  - "test/acceptance/**"
  - "features/**"
  - "docs/workspaces.md"
---

# Workspaces, plans and domain errors (ADR 5)

## Workspaces

Data belongs to a workspace. Use cases get the current one from `ports.WorkspaceResolver` and pass it **explicitly** to every repository call. A workspace-scoped entity's repository port embeds `ports.WorkspaceRepository[T]`, treats another workspace's record as missing (`ErrNotFound`), and every adapter's contract tests run `storagetest.RunWorkspaceIsolation`, plus an isolation case for each extra query. With no workspace, a use case returns `domain.ErrNoWorkspace` before touching storage. `bootstrap` wires `workspace.Local`; tests use `testsupport.FixedWorkspace` or `NoWorkspace`. **Adding an entity: follow the checklist in `docs/workspaces.md`.**

## Plans and entitlements

Plans set a workspace's limits. A use case that a limit gates asks `ports.Entitlements` for the workspace's plan, counts, and checks with the domain's rules before saving: `Entitlements.CheckCreate` before adding one, `CheckChange` before changing one. Deleting is never limited, so a workspace over a read-only limit can get back within it. If the plan or the count can't be looked up, refuse the change: a failed lookup must never mean "no limits". Front ends read the plan through `ports.EntitlementsService` (`GET /api/v1/entitlements`). `bootstrap` wires the local plan; tests use `testsupport.FixedEntitlements`.

The local edition's plan is `bootstrap.LocalPlan`: `local`, unless a build sets another with `-ldflags "-X <module>/internal/bootstrap.LocalPlan=<plan>"`. The plans come from `internal/adapters/plans/plans.jsonc`, built into the binary, unless `STARTER_PLANS_FILE` names another file. An unknown local plan or a bad plans file fails at start. The file is JSON with full-line `//` comments only: a comment after a value is a parse error.

## Domain errors

Domain errors (`domain.ErrNotFound`, `*domain.ValidationError`, `*domain.LimitError`, `*domain.FeatureError`, `domain.ErrNoWorkspace`) are the shared language. `httpapi/problem.go` maps them to HTTP, and `httpclient` maps them back. (`ErrNoWorkspace` isn't mapped yet: nothing over HTTP can lack a workspace until sign-in exists.)
