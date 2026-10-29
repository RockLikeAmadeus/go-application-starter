import { describe, expect, it } from "vitest";
import { resolveTheme } from "./theme";

describe("resolveTheme", () => {
  it("is dark with no stored choice and no system preference", () => {
    expect(resolveTheme(null, null)).toBe("dark");
  });
});
