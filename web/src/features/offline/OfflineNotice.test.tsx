import { act, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { OfflineNotice } from "./OfflineNotice";

describe("OfflineNotice", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("says the app is offline when the browser starts without a connection", () => {
    vi.spyOn(navigator, "onLine", "get").mockReturnValue(false);

    render(<OfflineNotice />);

    expect(screen.getByRole("status")).toHaveTextContent("You're offline");
  });

  it("says the app is offline as soon as the connection drops", () => {
    const onLine = vi.spyOn(navigator, "onLine", "get").mockReturnValue(true);
    render(<OfflineNotice />);

    onLine.mockReturnValue(false);
    act(() => {
      window.dispatchEvent(new Event("offline"));
    });

    expect(screen.getByRole("status")).toHaveTextContent("You're offline");
  });
});
