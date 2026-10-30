import { act, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { OfflineNotice } from "./OfflineNotice";

/** Starts the browser with or without a connection; returns a way to change it, as the browser would. */
function connection(online: boolean) {
  const onLine = vi.spyOn(navigator, "onLine", "get").mockReturnValue(online);
  return (nowOnline: boolean) => {
    onLine.mockReturnValue(nowOnline);
    act(() => {
      window.dispatchEvent(new Event(nowOnline ? "online" : "offline"));
    });
  };
}

describe("OfflineNotice", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("says the app is offline when the browser starts without a connection", () => {
    connection(false);

    render(<OfflineNotice />);

    expect(screen.getByRole("status")).toHaveTextContent("You're offline");
  });

  it("says the app is offline as soon as the connection drops", () => {
    const change = connection(true);
    render(<OfflineNotice />);

    change(false);

    expect(screen.getByRole("status")).toHaveTextContent("You're offline");
  });

  it("clears the notice as soon as the connection comes back", () => {
    const change = connection(false);
    render(<OfflineNotice />);

    change(true);

    expect(screen.getByRole("status")).toBeEmptyDOMElement();
  });
});
