import type { PublicItem } from "@listou/types";

export interface PriceBand {
  id: string;
  label: string;
  min: number;
  max: number;
}

export const PRICE_BANDS: PriceBand[] = [
  { id: "all", label: "Qualquer valor", min: 0, max: Infinity },
  { id: "ate-100", label: "Até R$ 100", min: 0, max: 10_000 },
  { id: "100-300", label: "R$ 100 a R$ 300", min: 10_000, max: 30_000 },
  { id: "300-1000", label: "R$ 300 a R$ 1.000", min: 30_000, max: 100_000 },
  { id: "1000+", label: "Acima de R$ 1.000", min: 100_000, max: Infinity },
];

/** Cheapest in-stock offer, falling back to the owner's reference price. */
export function displayPriceCents(item: PublicItem): number | null {
  const offer = item.offers.find((o) => o.priceCents != null && o.availability !== "OUT_OF_STOCK");
  return offer?.priceCents ?? item.priceReferenceCents ?? null;
}

export function matchesFilters(
  item: PublicItem,
  opts: { query: string; band: PriceBand; categoryId: string },
): boolean {
  if (opts.categoryId !== "all") {
    if (opts.categoryId === "none" ? item.categoryId !== null : item.categoryId !== opts.categoryId)
      return false;
  }
  if (opts.query) {
    const hay = `${item.title} ${item.product?.canonicalTitle ?? ""} ${item.product?.brand ?? ""}`
      .normalize("NFD")
      .replace(/[̀-ͯ]/g, "")
      .toLowerCase();
    const needle = opts.query.normalize("NFD").replace(/[̀-ͯ]/g, "").toLowerCase();
    if (!hay.includes(needle)) return false;
  }
  if (opts.band.id !== "all") {
    const price = displayPriceCents(item);
    if (price === null || price < opts.band.min || price >= opts.band.max) return false;
  }
  return true;
}

export function isGone(item: PublicItem): boolean {
  return item.availableQuantity <= 0;
}
