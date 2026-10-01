import { cn } from "../cn";

export interface AvatarProps {
  name: string;
  src?: string | null;
  size?: "sm" | "md" | "lg" | "xl";
  className?: string;
}

const sizes = {
  sm: "size-8 text-xs",
  md: "size-11 text-sm",
  lg: "size-16 text-lg",
  xl: "size-24 text-2xl",
};

function initials(name: string): string {
  const parts = name.replace(/&| e /gi, " ").split(/\s+/).filter(Boolean);
  return parts
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? "")
    .join("");
}

export function Avatar({ name, src, size = "md", className }: AvatarProps) {
  const base = cn(
    "relative inline-flex shrink-0 items-center justify-center overflow-hidden rounded-full ring-4 ring-surface",
    sizes[size],
    className,
  );
  if (src) {
    return (
      <span className={base}>
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={src} alt={name} className="size-full object-cover" />
      </span>
    );
  }
  return (
    <span
      className={cn(
        base,
        "from-primary-soft to-blush-soft font-display text-primary bg-gradient-to-br font-semibold",
      )}
      aria-label={name}
    >
      {initials(name)}
    </span>
  );
}
