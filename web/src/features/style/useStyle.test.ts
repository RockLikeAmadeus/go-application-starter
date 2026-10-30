import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { readStoredStyle, writeStoredStyle } from "./style";
import { useStyle } from "./useStyle";

describe("useStyle", () => {
  afterEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute("data-style");
  });

  it("starts from the stored choice and applies it to the page", () => {
    writeStoredStyle("terminal", localStorage);

    const { result } = renderHook(() => useStyle());

    expect(result.current[0]).toBe("terminal");
    expect(document.documentElement).toHaveAttribute("data-style", "terminal");
  });

  it("applies and saves a chosen style", () => {
    const { result } = renderHook(() => useStyle());

    act(() => result.current[1]("soft"));

    expect(result.current[0]).toBe("soft");
    expect(document.documentElement).toHaveAttribute("data-style", "soft");
    expect(readStoredStyle(localStorage)).toBe("soft");
  });
});
