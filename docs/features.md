# Features: turning something on or off by plan

A **feature** is something a workspace's plan turns on or off, like exporting items. This guide explains how gating works, then gives a checklist for gating a new feature. The decision behind it is [ADR 5](adr/0005-hosted-and-local-editions.md). The business rules are in `SPEC.md` §4.9, "Plans, features and limits". Limits, which count things, work differently: see `.claude/rules/workspaces-and-plans.md`.

## How it works

```
plans.jsonc ──► plans adapter ──► Entitlements.Features ──► use case: CheckFeature(key)  (refuses)
                                         │
                                         └──► GET /api/v1/entitlements ──► front ends  (only show)
```

1. **The plans file defines the feature.** You add it to the top-level `features` section with a key, a `name` that people read, and `whenOff`. Each plan lists the keys it turns on. A feature a plan doesn't list is off for that plan. The file fails to load if a plan lists an unknown key, a `name` is missing or blank, or `whenOff` isn't `show-locked` or `hide`.
2. **The core refuses, and the front ends only show.** The use case gets the workspace's entitlements and calls `entitlements.CheckFeature(key)` before doing anything else. That returns a `*domain.FeatureError` ("Item export is not included in your current plan. To use this feature, please upgrade to the pro plan.") when the feature is off, or when the key isn't in the catalog. If the entitlements can't be looked up, the use case refuses too. A front end never decides on its own whether a feature can be used.
3. **The error is the same everywhere.** Over HTTP it's a 403 problem with `feature`, `name` and `availableOn`, and `httpclient` turns it back into the same `*FeatureError`. The CLI and TUI print it the same way in `--server` mode as with local data.
4. **The upgrade hint** (`Feature.AvailableOn`) lists the plans that have the feature, leaving out the local edition's plan. A hidden feature has no hint. The local edition gives no hints at all, so a locked feature there just says it isn't available.

### What each front end shows

| `whenOff` | Web (`FeatureGate`) | CLI help (`cli.Gated`) | TUI help line (`gatedKeyHelp`) |
|---|---|---|---|
| on | the control | listed as usual | the key's help |
| `show-locked` | the control disabled, with "Available on the pro plan" (or "Not available") | listed, "(not included in your current plan; available on the pro plan)" | "x export (locked: available on the pro plan)" |
| `hide` | nothing | not listed | nothing |
| can't read the entitlements | disabled, "Not available" | listed as usual (the core still refuses) | |

The web app avoids things appearing late (pop-in). `Layout` prefetches `entitlementsQuery` when the app starts, and the answer is kept for the session. `FeatureGate` also renders the control disabled until the answer arrives.

## Gating a new feature

The template's sample, item export, follows every step below. While it's still in your project, read its code alongside this guide: `ExportItems` in `internal/core/app/item_service.go`, `item export` in `internal/adapters/cli/item.go`, the Export button in `web/src/features/items/ItemsPage.tsx`, and `features/plan-features.feature`. The steps use a made-up `widget-sharing`.

### 1. The plans file

In `internal/adapters/plans/plans.jsonc`, add the feature to `features`, and its key to the `features` list of each plan that has it:

```jsonc
"features": {
	"widget-sharing": {"name": "Widget sharing", "whenOff": "show-locked"}
},
"plans": {
	"local": {"limits": {}, "features": ["widget-sharing"]},
	"pro":   {"limits": {}, "features": ["widget-sharing"]}
}
```

The key is a contract: code and the API use it, so don't reword it later. Change the `name` whenever you like. Decide whether the local edition's plan has the feature. If it doesn't, local users see it locked with no hint, or don't see it at all.

### 2. The use case: check first

```go
const widgetSharing = "widget-sharing"

func (s *WidgetService) ShareWidget(ctx context.Context, id string) error {
	workspace, err := s.workspaces.CurrentWorkspace(ctx)
	if err != nil {
		return err
	}
	entitlements, err := s.entitlementsProvider.For(ctx, workspace)
	if err != nil {
		return err // a failed lookup never means "allowed"
	}
	if err := entitlements.CheckFeature(widgetSharing); err != nil {
		return err
	}
	// … the work itself
}
```

Check before you touch storage. Unit-test three cases with `testsupport.FixedEntitlements`: on, off (a `*domain.FeatureError` for this key), and the lookup failing. Also add the use case to the fail-closed table, so it refuses when there's no workspace.

### 3. The API

Nothing to add for gating itself. `GET /api/v1/entitlements` already lists every feature, and `problemFor` already maps `*FeatureError`. A new endpoint for the feature's work keeps the shared `default` problem response, which covers the 403.

### 4. The front ends

- **Web:** wrap the control in `FeatureGate` and spread the props it hands you:

  ```tsx
  <FeatureGate feature="widget-sharing">
    {(control) => <button type="button" {...control} onClick={share}>Share</button>}
  </FeatureGate>
  ```

  Include `entitlementsQuery` in the page's loading state, next to the page's own query, so the first paint already has the control in its final state. For anything other than a single control, `useFeature(key)` returns `on`, `locked` (with `availableOn`), `hidden` or `pending`.
- **CLI:** wrap the command: `Gated("widget-sharing", &cobra.Command{…})`. The help notes it or hides it by the plan, reading only the entitlements and never opening storage. Running the command while the feature is off fails through the core's error. You don't need to check anything in the command.
- **TUI:** build the key's entry in the screen's help line with `gatedKeyHelp("s share", entitlements.Features["widget-sharing"])`, and leave it out when the result is empty. The entitlements come from `ports.EntitlementsService` (`bootstrap.App`'s `Entitlements`). As with the CLI, the core refuses the key's action when the feature is off.

### 5. Tests

- **Acceptance:** scenarios tagged `@plans` run on the memory, jsonfile, sqlite, postgres, and http drivers. The harness's `testPlans` turns hints on and can turn a feature on or off for a scenario. The helpers that do it (`setFeature`, `setWhenOff`) live in the sample's `plan_feature_steps.go`, so copy them out before removing the sample.
- **Journeys:** the sample has one browser journey (`web/e2e/features/items.feature`, "Export the items") and terminal journeys (`test/terminal/features/items.feature`). Add one only if the feature's wiring isn't already proven by them.

### Checklist

- [ ] The feature is in `plans.jsonc`'s `features` section, with a lasting key, a `name` and `whenOff`, and each plan that has it lists its key.
- [ ] Every use case the feature gates checks the entitlements first, and refuses when they can't be looked up. Unit tests cover on, off and a failed lookup.
- [ ] The web control sits in a `FeatureGate`, and its page waits for `entitlementsQuery` in its loading state.
- [ ] A CLI command for it is wrapped in `Gated`. A TUI key for it shows its help through `gatedKeyHelp`.
- [ ] `@plans` scenarios cover the feature on and off.

## Pitfalls

- **Checking in a front end only.** Hiding a button isn't gating: the API, the CLI and other clients can still call the use case. The use case must call `CheckFeature`.
- **Treating a failed lookup as "on".** The core refuses. A front end that can't read the entitlements shows the feature locked (web) or plainly (CLI help), and leaves refusing to the core.
- **Opening the app to show the CLI's help.** That would create the data folder, run migrations and fail on a locked folder. `Gated` and the help use `bootstrap.NewEntitlementsService`, which opens no storage.
- **Renaming a key.** Every plan's list, the use case's constant and the front ends' gates name it. Change the `name` instead.
