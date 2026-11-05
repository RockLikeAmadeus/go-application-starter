import { screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { entitlementsApi } from "../../test/entitlementsApi";
import { renderWithClient } from "../../test/render";
import { FeatureGate } from "./FeatureGate";

const exportFeature = { name: "Export", availableOn: [] };

describe("FeatureGate", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it.each<{
    when: string;
    fetch: () => ReturnType<typeof vi.fn>;
    isPending?: boolean;
    want: { control: "enabled" | "disabled" | "absent"; note?: string };
  }>([
    {
      when: "the plan has it",
      fetch: () =>
        entitlementsApi({ export: { ...exportFeature, isOn: true, whenOff: "show-locked" } }),
      want: { control: "enabled" },
    },
    {
      when: "it's locked, with the plans that have it",
      fetch: () =>
        entitlementsApi({
          export: {
            ...exportFeature,
            isOn: false,
            whenOff: "show-locked",
            availableOn: ["pro", "team"],
          },
        }),
      want: { control: "disabled", note: "Available on the pro or team plans" },
    },
    {
      when: "it's locked with no hint",
      fetch: () =>
        entitlementsApi({ export: { ...exportFeature, isOn: false, whenOff: "show-locked" } }),
      want: { control: "disabled", note: "Not available" },
    },
    {
      when: "it's hidden",
      fetch: () => entitlementsApi({ export: { ...exportFeature, isOn: false, whenOff: "hide" } }),
      want: { control: "absent" },
    },
    {
      when: "the entitlements haven't arrived (a same-size placeholder)",
      fetch: () => vi.fn(() => new Promise<Response>(() => {})),
      isPending: true,
      want: { control: "disabled" },
    },
  ])("shows the control $want.control when $when", async ({ fetch, isPending = false, want }) => {
    const api = fetch();
    vi.stubGlobal("fetch", api);

    const { client } = renderWithClient(
      <FeatureGate feature="export">
        {(control) => <button {...control}>Export</button>}
      </FeatureGate>,
    );

    await waitFor(() => expect(api).toHaveBeenCalled());
    await waitFor(() => expect(client.isFetching()).toBe(isPending ? 1 : 0));
    const button = screen.queryByRole("button", { name: "Export" });
    if (want.control === "absent") {
      expect(button).not.toBeInTheDocument();
      return;
    }
    expect(button).toBeInTheDocument();
    expect(button).toHaveProperty("disabled", want.control === "disabled");
    if (want.note) expect(button).toHaveAccessibleDescription(want.note);
    else expect(button).not.toHaveAccessibleDescription();
  });
});
