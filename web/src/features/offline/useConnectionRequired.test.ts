import { onlineManager } from "@tanstack/react-query";
import { waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { connectionApi } from "../../test/connectionApi";
import { renderHookWithClient } from "../../test/render";
import { useConnectionRequired } from "./useConnectionRequired";

describe("useConnectionRequired", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    onlineManager.setOnline(true);
  });

  it("assumes a connection is required until the server answers", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => new Promise<Response>(() => {})),
    );

    const { result } = renderHookWithClient(() => useConnectionRequired());

    expect(result.current).toBe(true);
  });

  it("is false when the server says no connection is required", async () => {
    vi.stubGlobal("fetch", connectionApi(false));

    const { result } = renderHookWithClient(() => useConnectionRequired());

    await waitFor(() => expect(result.current).toBe(false));
  });

  it("asks the server even when the browser reports offline", async () => {
    vi.stubGlobal("fetch", connectionApi(false));
    onlineManager.setOnline(false);

    const { result } = renderHookWithClient(() => useConnectionRequired());

    await waitFor(() => expect(result.current).toBe(false));
  });

  it("stays true when the request fails", async () => {
    const fetch = vi.fn(async () => {
      throw new TypeError("Failed to fetch");
    });
    vi.stubGlobal("fetch", fetch);

    const { result } = renderHookWithClient(() => useConnectionRequired());

    await waitFor(() => expect(fetch).toHaveBeenCalled());
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(result.current).toBe(true);
  });
});
