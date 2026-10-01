"use client";

import { cn } from "../cn";
import { Glyph } from "./glyph";

export interface CategoryTab {
  id: string;
  label: string;
  emoji?: string | null;
  count?: number;
}

export interface CategoryTabsProps {
  tabs: CategoryTab[];
  value: string;
  onChange: (id: string) => void;
  className?: string;
}

export function CategoryTabs({ tabs, value, onChange, className }: CategoryTabsProps) {
  return (
    <div
      role="tablist"
      className={cn("-mx-4 flex gap-2 overflow-x-auto px-4 pb-1 [scrollbar-width:none]", className)}
    >
      {tabs.map((tab) => {
        const active = tab.id === value;
        return (
          <button
            key={tab.id}
            role="tab"
            type="button"
            aria-selected={active}
            onClick={() => onChange(tab.id)}
            className={cn(
              "duration-(--duration-fast) flex h-10 shrink-0 items-center gap-1.5 rounded-full px-4 text-sm font-semibold transition-all",
              active
                ? "bg-ink shadow-soft text-white"
                : "bg-surface text-ink-soft shadow-hairline hover:text-ink",
            )}
          >
            {tab.emoji ? <Glyph name={tab.emoji} className="size-4" /> : null}
            {tab.label}
            {tab.count != null ? (
              <span
                className={cn("text-xs tabular-nums", active ? "text-white/70" : "text-ink-muted")}
              >
                {tab.count}
              </span>
            ) : null}
          </button>
        );
      })}
    </div>
  );
}
