import { waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { entitlementsApi } from "../../test/entitlementsApi";
import { renderHookWithClient } from "../../test/render";
import { type FeatureState, useFeature } from "./useFeature";

const exportFeature = { name: "Export", availableOn: [] };

describe("useFeature", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it.each<{ when: string; fetch: () => unknown; want: FeatureState }>([
    {
      when: "the plan has it",
      fetch: () =>
        entitlementsApi({ export: { ...exportFeature, isOn: true, whenOff: "show-locked" } }),
      want: { state: "on" },
    },
    {
      when: "it's off and shown locked",
      fetch: () =>
        entitlementsApi({
          export: { ...exportFeature, isOn: false, whenOff: "show-locked", availableOn: ["pro"] },
        }),
      want: { state: "locked", availableOn: ["pro"] },
    },
    {
      when: "it's off and hidden",
      fetch: () => entitlementsApi({ export: { ...exportFeature, isOn: false, whenOff: "hide" } }),
      want: { state: "hidden" },
    },
    {
      when: "the plans don't know it",
      fetch: () => entitlementsApi({}),
      want: { state: "locked", availableOn: [] },
    },
    {
      when: "the entitlements haven't arrived",
      fetch: () => vi.fn(() => new Promise<Response>(() => {})),
      want: { state: "pending" },
    },
    {
      when: "the entitlements can't be read",
      fetch: () => vi.fn(async () => new Response(null, { status: 500 })),
      want: { state: "locked", availableOn: [] },
    },
  ])("is $want.state when $when", async ({ fetch, want }) => {
    vi.stubGlobal("fetch", fetch());

    const { result } = renderHookWithClient(() => useFeature("export"));

    await waitFor(() => expect(result.current).toEqual(want));
  });
});
