import { screen, waitFor, within } from "@testing-library/react";
import type { QueryClient } from "@tanstack/react-query";
import { afterEach, describe, expect, it, type Mock, vi } from "vitest";
import { renderWithClient } from "../../test/render";
import { warningsApi } from "../../test/warningsApi";
import { WarningsNotice } from "./WarningsNotice";

/** Waits until the notice has asked for the warnings and had its answer. */
async function settled(fetch: Mock, client: QueryClient) {
  await waitFor(() => expect(fetch).toHaveBeenCalled());
  await waitFor(() => expect(client.isFetching()).toBe(0));
}

describe("WarningsNotice", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("shows each of the app's warnings", async () => {
    vi.stubGlobal("fetch", warningsApi("First warning.", "Second warning."));

    renderWithClient(<WarningsNotice />);

    const warnings = await screen.findByRole("region", { name: "Warnings" });
    expect(within(warnings).getByText("First warning.")).toBeInTheDocument();
    expect(within(warnings).getByText("Second warning.")).toBeInTheDocument();
  });

  it("shows nothing when there are no warnings", async () => {
    const fetch = warningsApi();
    vi.stubGlobal("fetch", fetch);

    const { client } = renderWithClient(<WarningsNotice />);

    await settled(fetch, client);
    expect(screen.queryByRole("region", { name: "Warnings" })).not.toBeInTheDocument();
  });

  it("shows nothing when the request fails", async () => {
    const fetch = vi.fn(async () => new Response(null, { status: 500 }));
    vi.stubGlobal("fetch", fetch);

    const { client } = renderWithClient(<WarningsNotice />);

    await settled(fetch, client);
    expect(screen.queryByRole("region", { name: "Warnings" })).not.toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
