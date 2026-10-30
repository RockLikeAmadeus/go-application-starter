import { describe, expect, it } from "vitest";
import { defaultStyle, styles } from "./style";

describe("styles", () => {
  it("offers Crisp, Terminal and Soft, with the first as the default", () => {
    expect(styles.map((style) => style.name)).toEqual(["Crisp", "Terminal", "Soft"]);
    expect(defaultStyle).toBe(styles[0]?.id);
  });
});
