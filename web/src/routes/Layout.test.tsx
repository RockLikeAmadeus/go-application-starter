import { act, render, screen, within } from "@testing-library/react";
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

  it("marks the menu closed when the browser closes the dialog, as Escape does", async () => {
    const user = userEvent.setup();
    renderLayout();
    const menu = screen.getByRole("button", { name: "Menu" });
    await user.click(menu);

    // jsdom can't press Escape on a modal dialog, so close it the way the browser would.
    const dialog = screen.getByRole<HTMLDialogElement>("dialog");
    act(() => dialog.close());

    expect(menu).toHaveAttribute("aria-expanded", "false");
  });

  it("handles Escape itself instead of letting the browser close the menu at once", async () => {
    const user = userEvent.setup();
    renderLayout();
    const menu = screen.getByRole("button", { name: "Menu" });
    await user.click(menu);

    // Escape on a modal dialog fires a cancelable `cancel` before the browser closes it.
    const cancel = new Event("cancel", { cancelable: true });
    act(() => screen.getByRole("dialog").dispatchEvent(cancel));

    expect(cancel.defaultPrevented).toBe(true);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(menu).toHaveAttribute("aria-expanded", "false");
  });

  it("keeps the menu on screen, marked as closing, until its closing transition ends", async () => {
    const user = userEvent.setup();
    renderLayout();
    await user.click(screen.getByRole("button", { name: "Menu" }));
    const dialog = screen.getByRole("dialog");
    // jsdom applies no stylesheet, so give the drawer the transition its CSS would.
    dialog.style.transitionDuration = "0.2s";

    await user.click(dialog);

    expect(dialog).toHaveAttribute("data-closing");
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    act(() => {
      dialog.dispatchEvent(new Event("transitionend"));
    });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(dialog).not.toHaveAttribute("data-closing");
  });

  it("ignores transitions that end inside the menu while it is closing", async () => {
    const user = userEvent.setup();
    renderLayout();
    await user.click(screen.getByRole("button", { name: "Menu" }));
    const dialog = screen.getByRole("dialog");
    dialog.style.transitionDuration = "0.2s";

    await user.click(dialog);
    act(() => {
      within(dialog)
        .getByRole("navigation")
        .dispatchEvent(new Event("transitionend", { bubbles: true }));
    });

    expect(screen.getByRole("dialog")).toHaveAttribute("data-closing");
  });

  it("still closes the menu if its closing transition never reports that it ended", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    try {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderLayout();
      await user.click(screen.getByRole("button", { name: "Menu" }));
      const dialog = screen.getByRole("dialog");
      dialog.style.transitionDuration = "0.2s";

      await user.click(dialog);
      act(() => {
        vi.advanceTimersByTime(1000);
      });

      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    } finally {
      vi.useRealTimers();
    }
  });

  it("closes the menu when its backdrop is clicked", async () => {
    const user = userEvent.setup();
    renderLayout();
    const menu = screen.getByRole("button", { name: "Menu" });
    await user.click(menu);

    // A click on the backdrop targets the <dialog> element itself, not its content.
    await user.click(screen.getByRole("dialog"));

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
