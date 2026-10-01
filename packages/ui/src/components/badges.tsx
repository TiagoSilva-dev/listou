import type { HTMLAttributes } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "../cn";

export const badgeVariants = cva(
  "inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-semibold leading-none",
  {
    variants: {
      tone: {
        neutral: "bg-canvas-deep text-ink-soft",
        primary: "bg-primary-soft text-primary",
        success: "bg-success-soft text-success",
        warning: "bg-warning-soft text-warning",
        blush: "bg-blush-soft text-[#a4553d]",
        glass: "bg-white/85 text-ink shadow-hairline backdrop-blur",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

export interface BadgeProps
  extends HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

export function Badge({ className, tone, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ tone }), className)} {...props} />;
}

const merchantDot: Record<string, string> = {
  AMAZON: "bg-merchant-amazon",
  MERCADO_LIVRE: "bg-merchant-mercadolivre",
  SHOPEE: "bg-merchant-shopee",
};

export function MarketplaceBadge({
  code,
  name,
  className,
}: {
  code: string;
  name: string;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "text-ink-soft inline-flex items-center gap-1.5 text-xs font-medium",
        className,
      )}
    >
      <span aria-hidden className={cn("bg-ink-muted size-1.5 rounded-full", merchantDot[code])} />
      {name}
    </span>
  );
}
