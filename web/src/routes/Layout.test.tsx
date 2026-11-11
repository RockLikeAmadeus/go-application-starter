import { onlineManager, useMutation } from "@tanstack/react-query";
import { act, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { connectionApi } from "../test/connectionApi";
import { entitlementsApi } from "../test/entitlementsApi";
import { renderWithClient } from "../test/render";
import { sessionApi } from "../test/sessionApi";
import { warningsApi } from "../test/warningsApi";
import { Layout } from "./Layout";

function renderLayout(page = <p>Page</p>) {
  const router = createMemoryRouter([
    { element: <Layout />, children: [{ index: true, element: page }] },
  ]);
  return renderWithClient(<RouterProvider router={router} />);
}

describe("Layout", () => {
  beforeEach(() => {
    // jsdom has no matchMedia; the theme reads the system colour-scheme preference.
    vi.stubGlobal("matchMedia", () => ({ matches: false }));
    // The offline notice asks whether a connection is required; leave it unanswered.
    vi.stubGlobal(
      "fetch",
      vi.fn(() => new Promise<Response>(() => {})),
    );
  });
  afterEach(() => {
    onlineManager.setOnline(true);
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
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

  it("tells the user when the app is offline, on every page", () => {
    vi.spyOn(navigator, "onLine", "get").mockReturnValue(false);

    renderLayout();

    expect(screen.getByRole("status")).toHaveTextContent("You're offline");
  });

  it("shows the app's warnings, on every page", async () => {
    vi.stubGlobal("fetch", warningsApi("Your data folder looks synced by Dropbox."));

    renderLayout();

    const warnings = await screen.findByRole("region", { name: "Warnings" });
    expect(warnings).toHaveTextContent("Your data folder looks synced by Dropbox.");
  });

  it("starts reading the plan's features when the app starts, before any page asks", async () => {
    const fetch = entitlementsApi({});
    vi.stubGlobal("fetch", fetch);

    renderLayout();

    await waitFor(() =>
      expect(fetch.mock.calls.map(([request]) => new URL(request.url).pathname)).toContain(
        "/api/v1/entitlements",
      ),
    );
  });

  it("asks someone signed out to sign in, with no sidebar or page, when sign-in is on", async () => {
    vi.stubGlobal("fetch", sessionApi({ isSignInOn: true }));

    renderLayout();

    expect(await screen.findByRole("heading", { name: "Sign in" })).toBeInTheDocument();
    expect(screen.getByText("Please sign in to continue.")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Sign in" })).toHaveAttribute("href", "/auth/sign-in");
    expect(screen.queryByText("Page")).not.toBeInTheDocument();
    expect(screen.queryByRole("complementary")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Menu" })).not.toBeInTheDocument();
  });

  it("shows who is signed in, and a Sign out button that posts to /auth/sign-out, in the sidebar", async () => {
    vi.stubGlobal("fetch", sessionApi({ isSignInOn: true, name: "Alpha" }));

    renderLayout();

    const sidebar = screen.getByRole("complementary");
    expect(await within(sidebar).findByText("Signed in as Alpha.")).toBeInTheDocument();
    const signOut = within(sidebar).getByRole("button", { name: "Sign out" });
    expect(signOut).toHaveAttribute("type", "submit");
    const form = signOut.closest("form");
    expect(form).toHaveAttribute("method", "post");
    expect(form).toHaveAttribute("action", "/auth/sign-out");
  });

  it("shows no signed-in name or Sign out button when sign-in is off", async () => {
    const fetch = sessionApi({ isSignInOn: false });
    vi.stubGlobal("fetch", fetch);

    renderLayout();
    // Nothing on screen changes once the session is read, so wait for the
    // request, then flush.
    await waitFor(() => expect(fetch).toHaveBeenCalled());
    await act(async () => {});

    const sidebar = within(screen.getByRole("complementary"));
    expect(sidebar.queryByText(/Signed in as/)).not.toBeInTheDocument();
    expect(sidebar.queryByRole("button", { name: "Sign out" })).not.toBeInTheDocument();
  });

  it("keeps saving after the browser goes offline when no connection is required", async () => {
    const fetch = connectionApi(false);
    vi.stubGlobal("fetch", fetch);
    function SavePage() {
      const save = useMutation({ mutationFn: async () => "Saved" });
      return <button onClick={() => save.mutate()}>{save.data ?? "Save"}</button>;
    }
    renderLayout(<SavePage />);
    // Let the app learn its data is local before the network drops. Nothing
    // on screen changes while online, so wait for the request, then flush.
    await waitFor(() => expect(fetch).toHaveBeenCalled());
    await act(async () => {});

    act(() => {
      window.dispatchEvent(new Event("offline"));
    });
    await userEvent.setup().click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByRole("button", { name: "Saved" })).toBeInTheDocument();
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
