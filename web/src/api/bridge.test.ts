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
});
