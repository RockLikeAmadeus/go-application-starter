import { describe, expect, it } from "vitest";
import {
  applyStyle,
  defaultStyle,
  readStoredStyle,
  resolveStyle,
  styles,
  writeStoredStyle,
} from "./style";

describe("styles", () => {
  it("offers Crisp, Terminal and Soft, with the first as the default", () => {
    expect(styles.map((style) => style.name)).toEqual(["Crisp", "Terminal", "Soft"]);
    expect(defaultStyle).toBe(styles[0]?.id);
  });
});

describe("resolveStyle", () => {
  it("returns the stored style when it is known, and otherwise the default", () => {
    expect(resolveStyle("terminal")).toBe("terminal");
    expect(resolveStyle("soft")).toBe("soft");
    expect(resolveStyle("retro")).toBe(defaultStyle);
    expect(resolveStyle(null)).toBe(defaultStyle);
  });
});

describe("readStoredStyle", () => {
  it("returns the saved style, or null when storage throws or holds an unknown value", () => {
    expect(readStoredStyle({ getItem: () => "terminal" })).toBe("terminal");
    expect(readStoredStyle({ getItem: () => "retro" })).toBeNull();
    // Browsers throw like this when site data is blocked.
    const blocked = {
      getItem: () => {
        throw new DOMException("The operation is insecure.", "SecurityError");
      },
    };
    expect(readStoredStyle(blocked)).toBeNull();
  });
});

describe("writeStoredStyle", () => {
  it("saves a choice that readStoredStyle returns, and does nothing when storage is unavailable", () => {
    const saved = new Map<string, string>();
    const storage = {
      getItem: (key: string) => saved.get(key) ?? null,
      setItem: (key: string, value: string) => void saved.set(key, value),
    };

    writeStoredStyle("soft", storage);
    expect(readStoredStyle(storage)).toBe("soft");

    // Browsers throw like this when site data is blocked.
    const blocked = {
      setItem: () => {
        throw new DOMException("The operation is insecure.", "SecurityError");
      },
    };
    expect(() => writeStoredStyle("soft", blocked)).not.toThrow();
  });
});

describe("applyStyle", () => {
  it("sets the root element's data-style to the style", () => {
    const root = document.createElement("html");

    applyStyle("terminal", root);
    expect(root).toHaveAttribute("data-style", "terminal");

    applyStyle("soft", root);
    expect(root).toHaveAttribute("data-style", "soft");
  });
});
