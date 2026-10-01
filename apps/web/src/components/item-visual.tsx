import { cn, Glyph } from "@listou/ui";

/** Product image, or a tile with a Listou glyph when the item has no picture. */
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
        "from-primary-soft/70 via-canvas-deep to-blush-soft grid size-full place-items-center bg-gradient-to-br text-primary/80 text-5xl",
        className,
      )}
    >
      <Glyph name={emoji} />
    </span>
  );
}
