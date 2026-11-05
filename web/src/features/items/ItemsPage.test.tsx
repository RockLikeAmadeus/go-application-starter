import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Item } from "../../api/client";
import { renderWithClient } from "../../test/render";
import { ItemsPage } from "./ItemsPage";

/**
 * A tiny in-memory fake of the items API, installed as global fetch. Its plan
 * has item export, which answers a fixed CSV.
 */
function fakeApi(initial: Item[] = []) {
  const items = [...initial];
  const json = (body: unknown, status = 200) =>
    new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });

  return vi.fn(async (input: Request) => {
    const path = new URL(input.url).pathname;
    if (path === "/api/v1/entitlements") {
      return json({
        plan: "pro",
        limits: {},
        features: {
          "item-export": {
            name: "Item export",
            isOn: true,
            whenOff: "show-locked",
            availableOn: [],
          },
        },
      });
    }
    if (path === "/api/v1/items/export") {
      return new Response("name,createdAt\nWidget,2026-11-05T09:00:00Z\n", {
        headers: { "Content-Type": "text/csv" },
      });
    }
    if (input.method === "GET") return json({ items });
    if (input.method === "POST") {
      const { name } = (await input.json()) as { name: string };
      if (!name.trim()) {
        return json(
          { title: "Unprocessable Entity", status: 422, detail: "name must not be empty" },
          422,
        );
      }
      const now = new Date().toISOString();
      const item = {
        id: String(items.length + 1),
        name: name.trim(),
        createdAt: now,
        updatedAt: now,
      };
      items.push(item);
      return json(item, 201);
    }
    return json({ title: "Not Implemented", status: 501 }, 501);
  });
}

describe("ItemsPage", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", fakeApi());
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("shows the empty state", async () => {
    renderWithClient(<ItemsPage />);
    expect(await screen.findByText("No items yet.")).toBeInTheDocument();
  });

  it("lists items in a table, one row each with its rename and delete actions", async () => {
    const now = new Date().toISOString();
    vi.stubGlobal("fetch", fakeApi([{ id: "1", name: "Widget", createdAt: now, updatedAt: now }]));
    renderWithClient(<ItemsPage />);

    const table = await screen.findByRole("table", { name: "Items" });
    const [header, row] = within(table).getAllByRole("row");
    expect(within(header!).getByRole("columnheader", { name: "Name" })).toBeInTheDocument();
    const cells = within(row!);
    expect(cells.getByRole("cell", { name: "Widget" })).toBeInTheDocument();
    expect(cells.getByRole("button", { name: "Rename Widget" })).toBeInTheDocument();
    expect(cells.getByRole("button", { name: "Delete Widget" })).toBeInTheDocument();
  });

  it("adds an item", async () => {
    const user = userEvent.setup();
    renderWithClient(<ItemsPage />);

    await user.type(await screen.findByLabelText("New item name"), "Widget");
    await user.click(screen.getByRole("button", { name: "Add item" }));

    expect(await screen.findByRole("table", { name: "Items" })).toHaveTextContent("Widget");
    expect(screen.getByLabelText("New item name")).toHaveValue("");
  });

  it("shows validation errors from the server", async () => {
    const user = userEvent.setup();
    renderWithClient(<ItemsPage />);

    await screen.findByText("No items yet.");
    await user.click(screen.getByRole("button", { name: "Add item" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("name must not be empty");
  });

  it("stays loading until the plan's features arrive, so Export is in its final state on first paint", async () => {
    const api = fakeApi();
    let release!: () => void;
    const entitlementsHeld = new Promise<void>((resolve) => (release = resolve));
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: Request) => {
        if (new URL(input.url).pathname === "/api/v1/entitlements") await entitlementsHeld;
        return api(input);
      }),
    );
    const { client } = renderWithClient(<ItemsPage />);

    await waitFor(() => expect(client.getQueryState(["items"])?.status).toBe("success"));
    expect(screen.getByText("Loading…")).toBeInTheDocument();
    expect(screen.queryByText("No items yet.")).not.toBeInTheDocument();

    release();
    expect(await screen.findByText("No items yet.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Export" })).toBeEnabled();
  });

  it("downloads the items as a CSV file when Export is pressed", async () => {
    const user = userEvent.setup();
    // jsdom has no object URLs, and a real click would navigate.
    const createObjectURL = vi.fn<(file: Blob) => string>(() => "blob:items");
    Object.assign(URL, { createObjectURL, revokeObjectURL: vi.fn() });
    const click = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});
    renderWithClient(<ItemsPage />);

    const exportButton = await screen.findByRole("button", { name: "Export" });
    await waitFor(() => expect(exportButton).toBeEnabled());
    await user.click(exportButton);

    await waitFor(() => expect(click).toHaveBeenCalledOnce());
    const link = click.mock.contexts[0] as HTMLAnchorElement;
    expect(link.download).toBe("items.csv");
    expect(link.href).toBe("blob:items");
    expect(await createObjectURL.mock.calls[0]![0].text()).toBe(
      "name,createdAt\nWidget,2026-11-05T09:00:00Z\n",
    );
    click.mockRestore();
    Reflect.deleteProperty(URL, "createObjectURL");
    Reflect.deleteProperty(URL, "revokeObjectURL");
  });
});
