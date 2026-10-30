import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { routes } from "./router";

function renderAt(path: string) {
  const router = createMemoryRouter(routes, { initialEntries: [path] });
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("routes", () => {
  beforeEach(() => {
    vi.stubGlobal("matchMedia", () => ({ matches: false }));
    vi.stubGlobal("fetch", () => new Promise(() => {}));
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  // sample:start
  it("opens on the Items page", async () => {
    renderAt("/");

    expect(await screen.findByRole("heading", { name: "Items" })).toBeInTheDocument();
  });

  // sample:end
  it("shows Page not found for an unknown path", async () => {
    renderAt("/nowhere");

    expect(await screen.findByText("Page not found.")).toBeInTheDocument();
  });
});
