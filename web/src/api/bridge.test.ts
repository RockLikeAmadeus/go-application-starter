import { describe, expect, it, vi } from "vitest";
import { bridgeFetch, type BridgeCall } from "./bridge";

describe("bridgeFetch", () => {
  it("sends the request over the bridge and returns the reply as a Response", async () => {
    const call = vi.fn<BridgeCall>(async () => ({
      status: 201,
      contentType: "application/json",
      body: '{"name":"Widget"}',
    }));
    const request = new Request("http://localhost/api/v1/items?limit=2", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: '{"name":"Widget"}',
    });

    const response = await bridgeFetch(call)(request);

    expect(call).toHaveBeenCalledWith("POST", "/api/v1/items?limit=2", '{"name":"Widget"}');
    expect(response.status).toBe(201);
    expect(response.headers.get("Content-Type")).toBe("application/json");
    expect(await response.text()).toBe('{"name":"Widget"}');
  });

  // Deleting an item answers 204, and a Response for 204 can't have a body,
  // not even an empty one.
  it("returns a Response with no body for 204", async () => {
    const call = vi.fn<BridgeCall>(async () => ({ status: 204, contentType: "", body: "" }));

    const response = await bridgeFetch(call)(
      new Request("http://localhost/api/v1/items/item-1", { method: "DELETE" }),
    );

    expect(response.status).toBe(204);
    expect(response.body).toBeNull();
  });
});
