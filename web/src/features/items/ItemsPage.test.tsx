import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Item } from "../../api/client";
import { renderWithClient } from "../../test/render";
import { ItemsPage } from "./ItemsPage";

/** A tiny in-memory fake of the items API, installed as global fetch. */
function fakeApi(initial: Item[] = []) {
  const items = [...initial];
  const json = (body: unknown, status = 200) =>
    new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });

  return vi.fn(async (input: Request) => {
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

  it("adds an item", async () => {
    const user = userEvent.setup();
    renderWithClient(<ItemsPage />);

    await user.type(await screen.findByLabelText("New item name"), "Widget");
    await user.click(screen.getByRole("button", { name: "Add item" }));

    expect(await screen.findByRole("list", { name: "Items" })).toHaveTextContent("Widget");
    expect(screen.getByLabelText("New item name")).toHaveValue("");
  });

  it("shows validation errors from the server", async () => {
    const user = userEvent.setup();
    renderWithClient(<ItemsPage />);

    await screen.findByText("No items yet.");
    await user.click(screen.getByRole("button", { name: "Add item" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("name must not be empty");
  });
});
