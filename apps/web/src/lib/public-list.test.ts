import { describe, expect, it } from "vitest";
import type { PublicItem } from "@listou/types";
import { displayPriceCents, matchesFilters, PRICE_BANDS } from "./public-list";

const base: PublicItem = {
  id: "1",
  categoryId: "c1",
  title: "Jogo de panelas",
  description: null,
  imageUrl: null,
  emoji: null,
  externalUrl: null,
  priceReferenceCents: 50_000,
  currency: "BRL",
  priority: "MEDIUM",
  desiredQuantity: 1,
  purchasedQuantity: 0,
  reservedQuantity: 0,
  availableQuantity: 1,
  status: "AVAILABLE",
  product: null,
  offers: [],
};

describe("public list filters", () => {
  it("prefers the cheapest in-stock offer over the reference price", () => {
    const item: PublicItem = {
      ...base,
      offers: [
        {
          id: "a",
          merchant: { code: "AMAZON", name: "Amazon" },
          title: "x",
          priceCents: 30_000,
          originalPriceCents: null,
          currency: "BRL",
          availability: "OUT_OF_STOCK",
          imageUrl: null,
          goUrl: "/go/a",
          lastSyncedAt: null,
          demo: true,
        },
        {
          id: "b",
          merchant: { code: "SHOPEE", name: "Shopee" },
          title: "x",
          priceCents: 40_000,
          originalPriceCents: null,
          currency: "BRL",
          availability: "IN_STOCK",
          imageUrl: null,
          goUrl: "/go/b",
          lastSyncedAt: null,
          demo: true,
        },
      ],
    };
    expect(displayPriceCents(item)).toBe(40_000);
    expect(displayPriceCents(base)).toBe(50_000);
  });

  it("filters by accent-insensitive text, category and price band", () => {
    const all = PRICE_BANDS[0]!;
    const band = PRICE_BANDS.find((b) => b.id === "300-1000")!;
    expect(matchesFilters(base, { query: "PANELA", band: all, categoryId: "all" })).toBe(true);
    expect(matchesFilters(base, { query: "liquidificador", band: all, categoryId: "all" })).toBe(
      false,
    );
    expect(matchesFilters(base, { query: "", band, categoryId: "all" })).toBe(true);
    expect(
      matchesFilters(base, {
        query: "",
        band: PRICE_BANDS.find((b) => b.id === "ate-100")!,
        categoryId: "all",
      }),
    ).toBe(false);
    expect(matchesFilters(base, { query: "", band: all, categoryId: "c2" })).toBe(false);
  });
});
