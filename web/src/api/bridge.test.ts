import { afterEach, describe, expect, it, vi } from "vitest";
import { apiFetch, bridgeFetch, type BridgeCall, type BridgeWindow } from "./bridge";

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

describe("apiFetch", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("sends requests over the bridge when Android's Wails bridge is there when it's created", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>();
    vi.stubGlobal("fetch", fetch);
    const call = vi.fn<BridgeCall>(async () => ({
      status: 200,
      contentType: "application/json",
      body: '{"isRequired":false}',
    }));
    const win: BridgeWindow = { wails: { invokeAsync: () => {} } };

    const send = apiFetch(win, async () => call);
    // Importing the Wails runtime replaces window.wails, after the check.
    delete win.wails;
    const response = await send(new Request("http://localhost/api/v1/connection"));

    expect(call).toHaveBeenCalledWith("GET", "/api/v1/connection", "");
    expect(fetch).not.toHaveBeenCalled();
    expect(await response.json()).toEqual({ isRequired: false });
  });

  it("uses plain fetch, looked up per request, when there's no Wails bridge", async () => {
    const loadCall = vi.fn(async (): Promise<BridgeCall> => async () => ({
      status: 200,
      contentType: "application/json",
      body: '{"bridged":true}',
    }));
    const send = apiFetch({}, loadCall);
    // Tests stub fetch after the API client's module has loaded.
    const fetch = vi.fn<typeof globalThis.fetch>(async () => Response.json({ isRequired: true }));
    vi.stubGlobal("fetch", fetch);
    const request = new Request("http://localhost/api/v1/connection");

    const response = await send(request);

    expect(fetch).toHaveBeenCalledWith(request);
    expect(loadCall).not.toHaveBeenCalled();
    expect(await response.json()).toEqual({ isRequired: true });
  });
});
