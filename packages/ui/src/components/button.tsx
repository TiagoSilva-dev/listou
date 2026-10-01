import { cva, type VariantProps } from "class-variance-authority";
import type { ButtonHTMLAttributes } from "react";
import { cn } from "../cn";

export const buttonVariants = cva(
  [
    "inline-flex select-none items-center justify-center gap-2 whitespace-nowrap font-semibold",
    "transition-all duration-(--duration-fast) ease-out-soft",
    "focus-visible:shadow-focus focus-visible:outline-none",
    "active:scale-[0.98] disabled:pointer-events-none disabled:opacity-50",
  ],
  {
    variants: {
      variant: {
        primary: "bg-primary text-primary-ink shadow-soft hover:bg-primary-hover",
        secondary: "bg-surface text-ink shadow-hairline hover:bg-canvas-deep",
        soft: "bg-primary-soft text-primary hover:bg-primary-soft/70",
        ghost: "text-ink-soft hover:bg-canvas-deep hover:text-ink",
        dark: "bg-ink text-white hover:bg-ink-soft",
        danger: "bg-danger-soft text-danger hover:bg-danger-soft/70",
      },
      size: {
        sm: "h-9 rounded-chip px-3.5 text-sm",
        md: "h-11 rounded-control px-5 text-sm",
        lg: "h-14 rounded-control px-7 text-base",
        icon: "size-10 rounded-full",
      },
      block: { true: "w-full" },
    },
    defaultVariants: { variant: "primary", size: "md" },
  },
);

export interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> {
  loading?: boolean;
}

export function Button({
  className,
  variant,
  size,
  block,
  loading,
  disabled,
  children,
  type = "button",
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(buttonVariants({ variant, size, block }), className)}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...props}
    >
      {loading ? (
        <span
          aria-hidden
          className="size-4 animate-spin rounded-full border-2 border-current border-r-transparent"
        />
      ) : null}
      {children}
    </button>
  );
}
