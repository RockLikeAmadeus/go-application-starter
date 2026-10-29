import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ThemeToggle } from "./ThemeToggle";

describe("ThemeToggle", () => {
  it("shows a Dark mode switch that is on only for the dark theme", () => {
    const { rerender } = render(<ThemeToggle theme="dark" onChange={() => {}} />);
    expect(screen.getByRole("switch", { name: "Dark mode" })).toBeChecked();

    rerender(<ThemeToggle theme="light" onChange={() => {}} />);
    expect(screen.getByRole("switch", { name: "Dark mode" })).not.toBeChecked();
  });

  it("reports the other theme when the switch is flipped", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const { rerender } = render(<ThemeToggle theme="dark" onChange={onChange} />);

    await user.click(screen.getByRole("switch", { name: "Dark mode" }));
    expect(onChange).toHaveBeenLastCalledWith("light");

    rerender(<ThemeToggle theme="light" onChange={onChange} />);
    await user.click(screen.getByRole("switch", { name: "Dark mode" }));
    expect(onChange).toHaveBeenLastCalledWith("dark");
  });
});
