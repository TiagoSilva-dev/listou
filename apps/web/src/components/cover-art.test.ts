import { describe, expect, it } from "vitest";
import { themeOf } from "./cover-art";

describe("themeOf", () => {
  it("falls back to blush for unknown themes", () => {
    expect(themeOf("lavender")).toBe("lavender");
    expect(themeOf("neon")).toBe("blush");
    expect(themeOf(null)).toBe("blush");
  });
});
