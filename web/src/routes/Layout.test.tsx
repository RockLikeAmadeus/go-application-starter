import { render, screen, within } from "@testing-library/react";
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
  });

  it("shows the Dark mode switch in the header", () => {
    renderLayout();
    expect(
      within(screen.getByRole("banner")).getByRole("switch", { name: "Dark mode" }),
    ).toBeInTheDocument();
  });
});
