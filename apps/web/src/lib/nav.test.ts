import { describe, expect, it } from "vitest";
import { safeNext } from "./nav";

describe("safeNext", () => {
  it("keeps relative paths and rejects open redirects", () => {
    expect(safeNext("/criar?tipo=WEDDING")).toBe("/criar?tipo=WEDDING");
    expect(safeNext("//evil.com")).toBe("/dashboard");
    expect(safeNext("https://evil.com")).toBe("/dashboard");
    expect(safeNext("/\\evil.com")).toBe("/dashboard");
    expect(safeNext(undefined)).toBe("/dashboard");
    expect(safeNext(["/a", "/b"])).toBe("/a");
  });
});
