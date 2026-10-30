import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Layout } from "./Layout";

function renderLayout() {
  const router = createMemoryRouter([
    { element: <Layout />, children: [{ index: true, element: <p>Page</p> }] },
  ]);
  return render(<RouterProvider router={router} />);
}

describe("Layout", () => {
  beforeEach(() => {
    // jsdom has no matchMedia; the theme reads the system colour-scheme preference.
    vi.stubGlobal("matchMedia", () => ({ matches: false }));
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
    document.documentElement.removeAttribute("style");
    document.documentElement.removeAttribute("class");
    document.documentElement.removeAttribute("data-style");
  });

  it("groups the Style picker and Dark mode switch under Appearance in the sidebar", () => {
    renderLayout();

    const appearance = within(screen.getByRole("complementary")).getByRole("group", {
      name: "Appearance",
    });
    expect(within(appearance).getByRole("combobox", { name: "Style" })).toBeInTheDocument();
    expect(within(appearance).getByRole("switch", { name: "Dark mode" })).toBeInTheDocument();
  });

  it("offers a closed Menu button for phone screens", () => {
    renderLayout();

    expect(screen.getByRole("button", { name: "Menu" })).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("opens the navigation and Appearance controls in a dialog from the Menu button", async () => {
    const user = userEvent.setup();
    renderLayout();

    const menu = screen.getByRole("button", { name: "Menu" });
    await user.click(menu);

    expect(menu).toHaveAttribute("aria-expanded", "true");
    const dialog = within(screen.getByRole("dialog"));
    expect(dialog.getByRole("navigation")).toBeInTheDocument();
    expect(dialog.getByRole("group", { name: "Appearance" })).toBeInTheDocument();
  });

  it("closes the menu when a link in it is chosen", async () => {
    const user = userEvent.setup();
    renderLayout();
    const menu = screen.getByRole("button", { name: "Menu" });
    await user.click(menu);

    await user.click(within(screen.getByRole("dialog")).getByRole("link", { name: "Starter" }));

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(menu).toHaveAttribute("aria-expanded", "false");
  });
  // sample:start

  it("shows the app name and an Items link in the sidebar navigation", () => {
    renderLayout();

    const sidebar = within(screen.getByRole("complementary"));
    expect(sidebar.getByText("Starter")).toBeInTheDocument();
    const nav = within(sidebar.getByRole("navigation"));
    expect(nav.getByRole("link", { name: "Items" })).toHaveAttribute("href", "/");
  });
  // sample:end
});
