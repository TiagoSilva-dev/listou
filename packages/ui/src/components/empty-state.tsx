import type { ReactNode } from "react";
import { cn } from "../cn";

export interface EmptyStateProps {
  emoji?: string;
  title: string;
  description?: string;
  action?: ReactNode;
  className?: string;
}

export function EmptyState({
  emoji = "🎁",
  title,
  description,
  action,
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        "rounded-card border-line-strong bg-surface/60 flex flex-col items-center gap-4 border border-dashed px-6 py-14 text-center",
        className,
      )}
    >
      <span
        aria-hidden
        className="bg-primary-soft grid size-16 place-items-center rounded-full text-3xl"
      >
        {emoji}
      </span>
      <div className="flex max-w-sm flex-col gap-1.5">
        <h3 className="font-display text-ink text-xl font-semibold">{title}</h3>
        {description ? (
          <p className="text-ink-muted text-sm leading-relaxed">{description}</p>
        ) : null}
      </div>
      {action}
    </div>
  );
}
