import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { readStoredTheme, writeStoredTheme } from "./theme";
import { useTheme } from "./useTheme";

/** Stubs the system colour-scheme preference, which jsdom does not provide. */
function systemPrefers(scheme: "dark" | "light") {
  vi.stubGlobal("matchMedia", (query: string) => ({
    matches: query === `(prefers-color-scheme: ${scheme})`,
  }));
}

describe("useTheme", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
    document.documentElement.removeAttribute("style");
    document.documentElement.removeAttribute("class");
  });

  it("starts from the stored choice and applies it to the page", () => {
    systemPrefers("dark");
    writeStoredTheme("light", localStorage);

    const { result } = renderHook(() => useTheme());

    expect(result.current[0]).toBe("light");
    expect(document.documentElement.style.colorScheme).toBe("light");
  });

  it("follows the system preference when there is no stored choice", () => {
    systemPrefers("light");

    const { result } = renderHook(() => useTheme());

    expect(result.current[0]).toBe("light");
    expect(document.documentElement.style.colorScheme).toBe("light");
  });

  it("applies and saves a chosen theme", () => {
    systemPrefers("dark");
    const { result } = renderHook(() => useTheme());

    act(() => result.current[1]("light"));

    expect(result.current[0]).toBe("light");
    expect(document.documentElement.style.colorScheme).toBe("light");
    expect(readStoredTheme(localStorage)).toBe("light");
  });
});
