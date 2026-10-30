import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { StylePicker } from "./StylePicker";

describe("StylePicker", () => {
  it("shows a Style select listing every style, with the current one selected", () => {
    render(<StylePicker style="terminal" onChange={() => {}} />);

    const picker = screen.getByRole("combobox", { name: "Style" });
    expect(
      within(picker)
        .getAllByRole("option")
        .map((option) => option.textContent),
    ).toEqual(["Crisp", "Terminal", "Soft"]);
    expect(picker).toHaveDisplayValue("Terminal");
  });

  it("reports the chosen style", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<StylePicker style="crisp" onChange={onChange} />);

    await user.selectOptions(screen.getByRole("combobox", { name: "Style" }), "Soft");

    expect(onChange).toHaveBeenCalledWith("soft");
  });
});
