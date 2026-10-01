import { describe, expect, it } from "vitest";
import { daysUntil, formatPrice, parsePriceToCents, slugify } from "./format";

describe("format", () => {
  it("formats BRL", () => {
    expect(formatPrice(39990)?.replace(/\s/g, " ")).toBe("R$ 399,90");
    expect(formatPrice(null)).toBeNull();
  });

  it.each([
    ["1.299,90", 129990],
    ["1299.90", 129990],
    ["1.299", 129900],
    ["R$ 50", 5000],
    ["abc", null],
  ])("parses %s", (input, cents) => {
    expect(parsePriceToCents(input)).toBe(cents);
  });

  it("slugifies portuguese names", () => {
    expect(slugify("Chá de casa nova do Tiago & Júlia!")).toBe("cha-de-casa-nova-do-tiago-e-julia");
  });

  it("counts days until a date", () => {
    expect(daysUntil("2026-12-12", new Date(2026, 11, 2))).toBe(10);
  });
});
