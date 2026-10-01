import { cn } from "@listou/ui";

/**
 * Generated cover art used when an event has no photo. Each theme is a set of
 * soft layered gradients plus grain, so every list looks finished on day one.
 */
export const THEMES = {
  blush: { base: "#f6ded5", a: "#f0b49f", b: "#d9c7ef", c: "#fff4ec" },
  lavender: { base: "#e4dcf3", a: "#b9a3e3", b: "#f3c6d3", c: "#f7f2ff" },
  sage: { base: "#dce9df", a: "#a9c9b3", b: "#f1dfb9", c: "#f6faf4" },
  sun: { base: "#f8e7c4", a: "#f1bf73", b: "#f2a493", c: "#fff8ea" },
  night: { base: "#2f2741", a: "#6b4fa0", b: "#e8947a", c: "#463a5e" },
} as const;

export type ThemeName = keyof typeof THEMES;

export function themeOf(name: string | null | undefined): ThemeName {
  return name && name in THEMES ? (name as ThemeName) : "blush";
}

export function CoverArt({
  theme,
  emoji,
  className,
  children,
}: {
  theme: string | null | undefined;
  emoji?: string;
  className?: string;
  children?: React.ReactNode;
}) {
  const t = THEMES[themeOf(theme)];
  return (
    <div
      className={cn("relative isolate overflow-hidden", className)}
      style={{
        backgroundColor: t.base,
        backgroundImage: [
          `radial-gradient(60% 80% at 15% 20%, ${t.a} 0%, transparent 60%)`,
          `radial-gradient(50% 70% at 85% 30%, ${t.b} 0%, transparent 65%)`,
          `radial-gradient(70% 60% at 60% 100%, ${t.c} 0%, transparent 70%)`,
        ].join(","),
      }}
    >
      <div aria-hidden className="grain absolute inset-0 -z-10 mix-blend-multiply" />
      {emoji ? (
        <span
          aria-hidden
          className="pointer-events-none absolute -bottom-6 -right-4 -z-10 rotate-[-12deg] select-none text-[9rem] opacity-30 blur-[1px] sm:text-[12rem]"
        >
          {emoji}
        </span>
      ) : null}
      {children}
    </div>
  );
}
