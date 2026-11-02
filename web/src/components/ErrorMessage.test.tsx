import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ErrorMessage } from "./ErrorMessage";

describe("ErrorMessage", () => {
  it("shows the error's message as an alert", () => {
    render(<ErrorMessage error={new Error("name must not be empty")} />);

    expect(screen.getByRole("alert")).toHaveTextContent("name must not be empty");
  });
});
