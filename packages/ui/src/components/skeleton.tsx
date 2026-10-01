import { cn } from "../cn";

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      aria-hidden
      className={cn(
        "animate-shimmer rounded-control bg-[linear-gradient(90deg,var(--color-canvas-deep)_0%,var(--color-surface)_50%,var(--color-canvas-deep)_100%)] bg-[length:200%_100%]",
        className,
      )}
    />
  );
}
