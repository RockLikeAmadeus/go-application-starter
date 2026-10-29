import { describe, expect, it } from "vitest";
import { readStoredTheme, resolveTheme } from "./theme";

describe("resolveTheme", () => {
  it("is dark with no stored choice and no system preference", () => {
    expect(resolveTheme(null, null)).toBe("dark");
  });

  it("follows the system preference with no stored choice", () => {
    expect(resolveTheme(null, "light")).toBe("light");
    expect(resolveTheme(null, "dark")).toBe("dark");
  });

  it("prefers a stored choice over the system preference", () => {
    expect(resolveTheme("light", "dark")).toBe("light");
    expect(resolveTheme("dark", "light")).toBe("dark");
  });
});

describe("readStoredTheme", () => {
  it("returns the value saved in storage", () => {
    const storage = { getItem: () => "light" };
    expect(readStoredTheme(storage)).toBe("light");
  });

  it("returns null when storage is unavailable", () => {
    // Browsers throw like this when site data is blocked.
    const storage = {
      getItem: () => {
        throw new DOMException("The operation is insecure.", "SecurityError");
      },
    };
    expect(readStoredTheme(storage)).toBeNull();
  });
});
