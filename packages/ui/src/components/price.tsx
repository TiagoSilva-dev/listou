import { formatPrice } from "@listou/types";
import { cn } from "../cn";

export interface PriceProps {
  cents: number | null;
  originalCents?: number | null;
  currency?: string;
  /** Prefix such as "a partir de" for reference prices. */
  prefix?: string;
  className?: string;
}

export function Price({ cents, originalCents, currency = "BRL", prefix, className }: PriceProps) {
  const value = formatPrice(cents, currency);
  if (!value) return null;
  const original =
    originalCents && cents && originalCents > cents ? formatPrice(originalCents, currency) : null;
  return (
    <span className={cn("inline-flex items-baseline gap-1.5 tabular-nums", className)}>
      {prefix ? <span className="text-ink-muted text-xs font-normal">{prefix}</span> : null}
      <span className="text-ink font-semibold">{value}</span>
      {original ? <s className="text-ink-muted text-xs">{original}</s> : null}
    </span>
  );
}
