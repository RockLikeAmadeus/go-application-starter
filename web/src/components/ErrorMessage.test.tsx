import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ErrorMessage } from "./ErrorMessage";

describe("ErrorMessage", () => {
  it("shows the error's message as an alert", () => {
    render(<ErrorMessage error={new Error("The name cannot be empty.")} />);

    expect(screen.getByRole("alert")).toHaveTextContent("The name cannot be empty.");
  });
});
