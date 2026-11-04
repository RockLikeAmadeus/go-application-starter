import { screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { renderWithClient } from "../../test/render";
import { warningsApi } from "../../test/warningsApi";
import { WarningsNotice } from "./WarningsNotice";

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

    await waitFor(() => expect(fetch).toHaveBeenCalled());
    await waitFor(() => expect(client.isFetching()).toBe(0));
    expect(screen.queryByRole("region", { name: "Warnings" })).not.toBeInTheDocument();
  });
});
