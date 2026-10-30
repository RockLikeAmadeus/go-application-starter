import { render, screen } from "@testing-library/react";
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
});
