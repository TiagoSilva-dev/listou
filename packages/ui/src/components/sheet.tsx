"use client";

import * as DialogPrimitive from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "../cn";

export interface SheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  /** Visually hide the title (still announced to screen readers). */
  hideTitle?: boolean;
  children: ReactNode;
  className?: string;
}

/**
 * Responsive overlay: a bottom sheet on mobile, a centered dialog from `sm` up.
 * Used for BottomSheet, Dialog and ShareDialog patterns.
 */
export function Sheet({
  open,
  onOpenChange,
  title,
  description,
  hideTitle,
  children,
  className,
}: SheetProps) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="animate-fade-in bg-ink/40 fixed inset-0 z-40 backdrop-blur-[2px]" />
        <DialogPrimitive.Content
          className={cn(
            "animate-sheet-up rounded-t-sheet bg-surface shadow-sheet fixed inset-x-0 bottom-0 z-50 max-h-[92dvh] overflow-y-auto p-6 pb-[max(1.5rem,env(safe-area-inset-bottom))]",
            "sm:animate-pop sm:rounded-sheet sm:shadow-lift sm:inset-auto sm:left-1/2 sm:top-1/2 sm:w-full sm:max-w-lg sm:-translate-x-1/2 sm:-translate-y-1/2",
            "focus:outline-none",
            className,
          )}
        >
          <div
            aria-hidden
            className="bg-line mx-auto -mt-2 mb-4 h-1.5 w-10 rounded-full sm:hidden"
          />
          <div
            className={cn("mb-5 flex items-start justify-between gap-4", hideTitle && "sr-only")}
          >
            <div className="flex flex-col gap-1">
              <DialogPrimitive.Title className="font-display text-ink text-2xl font-semibold">
                {title}
              </DialogPrimitive.Title>
              {description ? (
                <DialogPrimitive.Description className="text-ink-muted text-sm">
                  {description}
                </DialogPrimitive.Description>
              ) : null}
            </div>
          </div>
          {!description ? (
            <DialogPrimitive.Description className="sr-only">{title}</DialogPrimitive.Description>
          ) : null}
          <DialogPrimitive.Close
            className="text-ink-muted hover:bg-canvas-deep hover:text-ink focus-visible:shadow-focus absolute right-4 top-4 grid size-9 place-items-center rounded-full transition-colors focus-visible:outline-none"
            aria-label="Fechar"
          >
            <X className="size-5" />
          </DialogPrimitive.Close>
          {children}
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
