import { screen, within } from "@testing-library/react";
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
});
