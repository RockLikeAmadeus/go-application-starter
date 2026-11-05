import { act, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { connectionApi } from "../../test/connectionApi";
import { renderWithClient } from "../../test/render";
import { OfflineNotice } from "./OfflineNotice";

/** Starts the browser with or without a connection; returns a way to change it, as the browser would. */
function connection(isOnline: boolean) {
  const onLine = vi.spyOn(navigator, "onLine", "get").mockReturnValue(isOnline);
  return (isNowOnline: boolean) => {
    onLine.mockReturnValue(isNowOnline);
    act(() => {
      window.dispatchEvent(new Event(isNowOnline ? "online" : "offline"));
    });
  };
}

describe("OfflineNotice", () => {
  beforeEach(() => {
    // The server: its data is reached over the network.
    vi.stubGlobal("fetch", connectionApi(true));
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("says the app is offline when the browser starts without a connection", () => {
    connection(false);

    renderWithClient(<OfflineNotice />);

    expect(screen.getByRole("status")).toHaveTextContent("You're offline");
  });

  it("says the app is offline as soon as the connection drops", () => {
    const change = connection(true);
    renderWithClient(<OfflineNotice />);

    change(false);

    expect(screen.getByRole("status")).toHaveTextContent("You're offline");
  });

  it("clears the notice as soon as the connection comes back", () => {
    const change = connection(false);
    renderWithClient(<OfflineNotice />);

    change(true);

    expect(screen.getByRole("status")).toBeEmptyDOMElement();
  });

  it("stays hidden when no connection is required, even when the browser is offline", async () => {
    vi.stubGlobal("fetch", connectionApi(false));
    connection(false);

    renderWithClient(<OfflineNotice />);

    await waitFor(() => expect(screen.getByRole("status")).toBeEmptyDOMElement());
  });
});
