import { queryOptions, useQuery } from "@tanstack/react-query";
import { api, unwrap } from "../../api/client";

const entitlementsKey = ["entitlements"] as const;

/**
 * The workspace's entitlements, shared by every gated control. `Layout`
 * prefetches it when the app starts, so pages rarely wait for it, and the
 * answer is kept for the session (it never goes stale), so a control doesn't
 * change under the user when another page mounts. The server still refuses a
 * feature that's off, should the plan change meanwhile.
 */
export const entitlementsQuery = queryOptions({
  queryKey: entitlementsKey,
  queryFn: async () => unwrap(await api.GET("/api/v1/entitlements")),
  staleTime: Number.POSITIVE_INFINITY,
});

/**
 * What to show of a gated feature: `on`, `locked` (with the plans that have
 * it, for the upgrade hint; empty when there's no hint), `hidden`, or
 * `pending` until the entitlements arrive.
 */
export type FeatureState =
  | { state: "on" }
  | { state: "locked"; availableOn: string[] }
  | { state: "hidden" }
  | { state: "pending" };

/**
 * Whether the workspace's plan has the feature with this key (e.g.
 * "item-export"), and how to show it if not. This only drives the display:
 * the server refuses a feature that's off whatever the page shows. So when the
 * entitlements can't be read, the feature shows locked with no hint, rather
 * than disappearing; so does a key the plans don't know, as the CLI's help
 * shows it.
 */
export function useFeature(key: string): FeatureState {
  const { data, isError } = useQuery(entitlementsQuery);
  if (isError) return { state: "locked", availableOn: [] };
  if (data === undefined) return { state: "pending" };
  const feature = data.features[key];
  if (feature === undefined) return { state: "locked", availableOn: [] };
  if (feature.isOn) return { state: "on" };
  if (feature.whenOff === "hide") return { state: "hidden" };
  return { state: "locked", availableOn: feature.availableOn };
}
