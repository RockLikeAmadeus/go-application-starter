import { describe, expect, it } from "vitest";
import { applyTheme, readStoredTheme, resolveTheme, writeStoredTheme } from "./theme";

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

  it("returns null when storage holds an unknown value", () => {
    const storage = { getItem: () => "purple" };
    expect(readStoredTheme(storage)).toBeNull();
  });
});

describe("writeStoredTheme", () => {
  it("saves a choice that readStoredTheme returns", () => {
    const saved = new Map<string, string>();
    const storage = {
      getItem: (key: string) => saved.get(key) ?? null,
      setItem: (key: string, value: string) => void saved.set(key, value),
    };

    writeStoredTheme("light", storage);
    expect(readStoredTheme(storage)).toBe("light");

    writeStoredTheme("dark", storage);
    expect(readStoredTheme(storage)).toBe("dark");
  });

  it("does nothing when storage is unavailable", () => {
    // Browsers throw like this when site data is blocked.
    const storage = {
      setItem: () => {
        throw new DOMException("The operation is insecure.", "SecurityError");
      },
    };
    expect(() => writeStoredTheme("light", storage)).not.toThrow();
  });
});

describe("applyTheme", () => {
  it("sets the root element's color-scheme to the theme", () => {
    const root = document.createElement("html");

    applyTheme("light", root);
    expect(root.style.colorScheme).toBe("light");

    applyTheme("dark", root);
    expect(root.style.colorScheme).toBe("dark");
  });

  it("marks the root element with the dark class only for the dark theme", () => {
    const root = document.createElement("html");

    applyTheme("dark", root);
    expect(root).toHaveClass("dark");

    applyTheme("light", root);
    expect(root).not.toHaveClass("dark");
  });
});
