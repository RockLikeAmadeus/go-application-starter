import { onlineManager, useMutation } from "@tanstack/react-query";
import { act, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { connectionApi } from "../../test/connectionApi";
import { renderHookWithClient } from "../../test/render";
import { useIsConnectionRequired } from "./useIsConnectionRequired";
import { useLocalDataOnline } from "./useLocalDataOnline";

describe("useLocalDataOnline", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    onlineManager.setOnline(true);
  });

  it("lets mutations run after the browser goes offline when no connection is required", async () => {
    vi.stubGlobal("fetch", connectionApi(false));
    const { result } = renderHookWithClient(() => {
      useLocalDataOnline();
      return {
        isConnectionRequired: useIsConnectionRequired(),
        save: useMutation({ mutationFn: async () => "saved" }),
      };
    });
    await waitFor(() => expect(result.current.isConnectionRequired).toBe(false));

    act(() => {
      window.dispatchEvent(new Event("offline"));
    });
    act(() => {
      result.current.save.mutate();
    });

    await waitFor(() => expect(result.current.save.data).toBe("saved"));
  });
});
