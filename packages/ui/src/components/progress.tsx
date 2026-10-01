import { cn } from "../cn";

export interface ProgressProps {
  value: number;
  max: number;
  label?: string;
  className?: string;
  tone?: "primary" | "light";
}

export function Progress({ value, max, label, className, tone = "primary" }: ProgressProps) {
  const pct = max > 0 ? Math.min(100, Math.round((value / max) * 100)) : 0;
  return (
    <div
      role="progressbar"
      aria-valuemin={0}
      aria-valuemax={max}
      aria-valuenow={value}
      aria-label={label}
      className={cn(
        "h-2 w-full overflow-hidden rounded-full",
        tone === "light" ? "bg-white/30" : "bg-primary-soft",
        className,
      )}
    >
      <div
        className={cn(
          "duration-(--duration-slow) ease-out-soft h-full rounded-full transition-[width]",
          tone === "light" ? "bg-white" : "from-primary to-blush bg-gradient-to-r",
        )}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}
