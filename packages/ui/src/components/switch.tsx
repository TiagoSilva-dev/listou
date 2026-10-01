"use client";

import { cn } from "../cn";

export interface SwitchProps {
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  label: string;
  description?: string;
  id?: string;
}

export function Switch({ checked, onCheckedChange, label, description, id }: SwitchProps) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div className="flex flex-col gap-0.5">
        <label htmlFor={id} className="text-ink text-sm font-semibold">
          {label}
        </label>
        {description ? (
          <p className="text-ink-muted text-sm leading-relaxed">{description}</p>
        ) : null}
      </div>
      <button
        id={id}
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onCheckedChange(!checked)}
        className={cn(
          "duration-(--duration-base) focus-visible:shadow-focus relative mt-0.5 h-7 w-12 shrink-0 rounded-full transition-colors focus-visible:outline-none",
          checked ? "bg-primary" : "bg-line-strong",
        )}
      >
        <span
          aria-hidden
          className={cn(
            "shadow-soft duration-(--duration-base) ease-spring absolute left-0.5 top-0.5 size-6 rounded-full bg-white transition-transform",
            checked && "translate-x-5",
          )}
        />
      </button>
    </div>
  );
}
