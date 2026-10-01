import { cn } from "@listou/ui";

/** Product image, or an emoji tile when the item has no picture. */
export function ItemVisual({
  imageUrl,
  emoji,
  alt,
  className,
}: {
  imageUrl?: string | null;
  emoji?: string | null;
  alt: string;
  className?: string;
}) {
  if (imageUrl) {
    return (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={imageUrl}
        alt={alt}
        loading="lazy"
        className={cn("size-full object-cover", className)}
      />
    );
  }
  return (
    <span
      aria-hidden
      className={cn(
        "from-primary-soft/70 via-canvas-deep to-blush-soft grid size-full place-items-center bg-gradient-to-br text-5xl",
        className,
      )}
    >
      {emoji ?? "🎁"}
    </span>
  );
}
