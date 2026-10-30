import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Home } from "./Layout";

describe("Home", () => {
  it("says there is nothing here yet", () => {
    render(<Home />);

    expect(screen.getByRole("heading", { name: "Home" })).toBeInTheDocument();
    expect(screen.getByText("Nothing here yet.")).toBeInTheDocument();
  });
});
