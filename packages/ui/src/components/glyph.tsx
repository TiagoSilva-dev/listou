import type { ReactNode, SVGProps } from "react";
import { cn } from "../cn";

/**
 * Listou's own illustration set. Every glyph is drawn on a 24px grid with a
 * rounded 1.6 stroke and one soft tinted shape, and inherits `currentColor`,
 * so the tone always comes from a token (text-primary, text-blush...).
 *
 * Content stored in the database still carries legacy emoji strings; GLYPH_BY_EMOJI
 * maps each one to a glyph so nothing renders as an emoji anymore.
 */
const T = { fill: "currentColor", fillOpacity: 0.16, stroke: "none" } as const;

const PATHS: Record<string, ReactNode> = {
  gift: (
    <>
      <rect x="3.5" y="9" width="17" height="11.5" rx="2.2" {...T} />
      <rect x="3.5" y="9" width="17" height="11.5" rx="2.2" />
      <path d="M2.5 9h19v-2.2a1.3 1.3 0 0 0-1.3-1.3H3.8a1.3 1.3 0 0 0-1.3 1.3Z" />
      <path d="M12 5.5V20.5" />
      <path d="M12 5.5c-1.2-2.6-4.6-2.9-4.6-.9 0 1.5 2.6 1 4.6.9Zm0 0c1.2-2.6 4.6-2.9 4.6-.9 0 1.5-2.6 1-4.6.9Z" />
    </>
  ),
  bed: (
    <>
      <rect x="3" y="11" width="18" height="6" rx="1.6" {...T} />
      <path d="M3 19.5V6" />
      <path d="M21 19.5V13a2 2 0 0 0-2-2h-8v6" />
      <path d="M3 17h18" />
      <circle cx="7" cy="13.8" r="1.7" />
    </>
  ),
  bath: (
    <>
      <path d="M3 12h18v2.5a4.5 4.5 0 0 1-4.5 4.5h-9A4.5 4.5 0 0 1 3 14.5Z" {...T} />
      <path d="M3 12h18v2.5a4.5 4.5 0 0 1-4.5 4.5h-9A4.5 4.5 0 0 1 3 14.5Z" />
      <path d="M6 12V6.5a2.5 2.5 0 0 1 4.6-1.3" />
      <path d="M7 19l-1 2M17 19l1 2" />
      <path d="M14.5 8.5v.01M17 7v.01M16.5 9.8v.01" />
    </>
  ),
  pan: (
    <>
      <circle cx="9.5" cy="13.5" r="6.5" {...T} />
      <circle cx="9.5" cy="13.5" r="6.5" />
      <path d="M16 13.5h5.5" />
      <path d="M7 12.2c.6-1.1 1.6-1.7 2.8-1.8" />
    </>
  ),
  basket: (
    <>
      <path d="M4 10h16l-1.6 9a1.6 1.6 0 0 1-1.6 1.3H7.2A1.6 1.6 0 0 1 5.6 19Z" {...T} />
      <path d="M4 10h16l-1.6 9a1.6 1.6 0 0 1-1.6 1.3H7.2A1.6 1.6 0 0 1 5.6 19Z" />
      <path d="M8 10l4-6 4 6" />
      <path d="M9.5 14v3.5M12 14v3.5M14.5 14v3.5" />
    </>
  ),
  suitcase: (
    <>
      <rect x="4" y="7.5" width="16" height="12.5" rx="2.4" {...T} />
      <rect x="4" y="7.5" width="16" height="12.5" rx="2.4" />
      <path d="M9 7.5V5.8A1.8 1.8 0 0 1 10.8 4h2.4A1.8 1.8 0 0 1 15 5.8v1.7" />
      <path d="M9 11v5M15 11v5" />
    </>
  ),
  heart: (
    <>
      <path d="M12 20.2S3.5 15 3.5 8.9A4.4 4.4 0 0 1 12 7.2a4.4 4.4 0 0 1 8.5 1.7C20.5 15 12 20.2 12 20.2Z" {...T} />
      <path d="M12 20.2S3.5 15 3.5 8.9A4.4 4.4 0 0 1 12 7.2a4.4 4.4 0 0 1 8.5 1.7C20.5 15 12 20.2 12 20.2Z" />
    </>
  ),
  party: (
    <>
      <path d="M4 20 8.2 8.6a1 1 0 0 1 1.6-.4l6 6a1 1 0 0 1-.4 1.6Z" {...T} />
      <path d="M4 20 8.2 8.6a1 1 0 0 1 1.6-.4l6 6a1 1 0 0 1-.4 1.6Z" />
      <path d="M13 4.5c.6 1.2.6 2.3 0 3.5M17 8c1.2-.5 2.3-.4 3.5.2M16.5 3.5v.01M20.5 11.5v.01M19 5v.01" />
    </>
  ),
  sun: (
    <>
      <path d="M6.5 17a5.5 5.5 0 0 1 11 0Z" {...T} />
      <path d="M6.5 17a5.5 5.5 0 0 1 11 0" />
      <path d="M2.5 17h19M5 20.5h14" />
      <path d="M12 6.5V4M5.6 9.1 4 7.5M18.4 9.1 20 7.5" />
    </>
  ),
  coffee: (
    <>
      <path d="M5 9h11v5.5a4.5 4.5 0 0 1-4.5 4.5h-2A4.5 4.5 0 0 1 5 14.5Z" {...T} />
      <path d="M5 9h11v5.5a4.5 4.5 0 0 1-4.5 4.5h-2A4.5 4.5 0 0 1 5 14.5Z" />
      <path d="M16 10.5h1.2a2.3 2.3 0 0 1 0 4.6H15.6" />
      <path d="M8.5 3.5c-.8 1-.8 2 0 3M12 3.5c-.8 1-.8 2 0 3" />
    </>
  ),
  bear: (
    <>
      <circle cx="12" cy="13" r="7" {...T} />
      <circle cx="12" cy="13" r="7" />
      <circle cx="6.5" cy="7" r="2.2" />
      <circle cx="17.5" cy="7" r="2.2" />
      <path d="M10 12v.01M14 12v.01" />
      <path d="M10.7 15.2c.8.6 1.8.6 2.6 0" />
    </>
  ),
  house: (
    <>
      <path d="M5 11v8a1.5 1.5 0 0 0 1.5 1.5h11A1.5 1.5 0 0 0 19 19v-8" {...T} />
      <path d="M3 11.5 12 4l9 7.5" />
      <path d="M5 10v9a1.5 1.5 0 0 0 1.5 1.5h11A1.5 1.5 0 0 0 19 19v-9" />
      <path d="M10 20.5v-5a2 2 0 0 1 4 0v5" />
    </>
  ),
  plate: (
    <>
      <circle cx="12" cy="12" r="8.5" {...T} />
      <circle cx="12" cy="12" r="8.5" />
      <circle cx="12" cy="12" r="4.6" />
    </>
  ),
  sofa: (
    <>
      <path d="M4 12.5V10a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v2.5" />
      <rect x="2.5" y="12" width="19" height="6" rx="2.2" {...T} />
      <rect x="2.5" y="12" width="19" height="6" rx="2.2" />
      <path d="M6 18v2M18 18v2" />
    </>
  ),
  stroller: (
    <>
      <path d="M4 5h2.2l1.6 8h10.2A6 6 0 0 0 12 7.5V13" {...T} />
      <path d="M3.5 4.5h3L8 13h10.5A6.5 6.5 0 0 0 12 6.5V13" />
      <circle cx="9" cy="18" r="2" />
      <circle cx="17" cy="18" r="2" />
    </>
  ),
  shirt: (
    <>
      <path d="m8.5 4-5 3 2.2 4 2-1V20h9.6v-10l2 1 2.2-4-5-3a3.5 3.5 0 0 1-7 0Z" {...T} />
      <path d="m8.5 4-5 3 2.2 4 2-1V20h9.6v-10l2 1 2.2-4-5-3a3.5 3.5 0 0 1-7 0Z" />
    </>
  ),
  bag: (
    <>
      <path d="M5 8h14l1 12H4Z" {...T} />
      <path d="M5 8h14l1 12H4Z" />
      <path d="M9 11V7a3 3 0 0 1 6 0v4" />
    </>
  ),
  plane: (
    <>
      <path d="M10.5 13.5 4 17l-1-1.5 5-4-3.5-5.5L6 5l6.5 4.5 5-3.5a2 2 0 0 1 2.5 3l-4.5 4.5L18 20l-1.5.8-3.2-5.8Z" {...T} />
      <path d="M10.5 13.5 4 17l-1-1.5 5-4-3.5-5.5L6 5l6.5 4.5 5-3.5a2 2 0 0 1 2.5 3l-4.5 4.5L18 20l-1.5.8-3.2-5.8Z" />
    </>
  ),
  baby: (
    <>
      <circle cx="12" cy="9" r="5" {...T} />
      <circle cx="12" cy="9" r="5" />
      <path d="M12 14v3.5" />
      <rect x="10" y="17.5" width="4" height="3.5" rx="1.75" />
      <path d="M10.2 8.5v.01M13.8 8.5v.01M10.8 10.8c.7.5 1.7.5 2.4 0" />
    </>
  ),
  ring: (
    <>
      <circle cx="12" cy="15" r="5.5" {...T} />
      <circle cx="12" cy="15" r="5.5" />
      <path d="m9.5 4 1-1.5h3l1 1.5-2.5 3.5Z" />
    </>
  ),
  cake: (
    <>
      <rect x="4" y="12" width="16" height="8" rx="2" {...T} />
      <rect x="4" y="12" width="16" height="8" rx="2" />
      <path d="M4 15.5c2 1.5 3 1.5 4 0s2-1.5 4 0 3 1.5 4 0 2-1.5 4 0" />
      <path d="M12 12V8.5" />
      <path d="M12 3.5c1.2 1 1.4 2 0 3-1.4-1-1.2-2 0-3Z" />
    </>
  ),
  cap: (
    <>
      <path d="m2.5 9.5 9.5-4.5 9.5 4.5-9.5 4.5Z" {...T} />
      <path d="m2.5 9.5 9.5-4.5 9.5 4.5-9.5 4.5Z" />
      <path d="M6.5 11.8v4c0 1.4 2.5 2.7 5.5 2.7s5.5-1.3 5.5-2.7v-4" />
      <path d="M21.5 9.5V15" />
    </>
  ),
  tree: (
    <>
      <path d="M12 3 6 10.5h3L5 16h14l-4-5.5h3Z" {...T} />
      <path d="M12 3 6 10.5h3L5 16h14l-4-5.5h3Z" />
      <path d="M12 16v4.5" />
    </>
  ),
  star: (
    <>
      <path d="m12 3.5 2.6 5.4 5.9.8-4.3 4.1 1 5.9L12 16.9l-5.2 2.8 1-5.9-4.3-4.1 5.9-.8Z" {...T} />
      <path d="m12 3.5 2.6 5.4 5.9.8-4.3 4.1 1 5.9L12 16.9l-5.2 2.8 1-5.9-4.3-4.1 5.9-.8Z" />
    </>
  ),
  bulb: (
    <>
      <path d="M12 3.5a6 6 0 0 0-3.6 10.8c.7.6 1.1 1.3 1.1 2.2h5c0-.9.4-1.6 1.1-2.2A6 6 0 0 0 12 3.5Z" {...T} />
      <path d="M12 3.5a6 6 0 0 0-3.6 10.8c.7.6 1.1 1.3 1.1 2.2h5c0-.9.4-1.6 1.1-2.2A6 6 0 0 0 12 3.5Z" />
      <path d="M9.5 19h5M10.5 21.2h3" />
    </>
  ),
  link: (
    <>
      <path d="M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1 1" />
      <path d="M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1-1" />
    </>
  ),
  search: (
    <>
      <circle cx="10.5" cy="10.5" r="6.5" {...T} />
      <circle cx="10.5" cy="10.5" r="6.5" />
      <path d="m15.5 15.5 5 5" />
    </>
  ),
  folder: (
    <>
      <path d="M3 7.5A2 2 0 0 1 5 5.5h4l2 2.5h8a2 2 0 0 1 2 2V18a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z" {...T} />
      <path d="M3 7.5A2 2 0 0 1 5 5.5h4l2 2.5h8a2 2 0 0 1 2 2V18a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z" />
    </>
  ),
  tv: (
    <>
      <rect x="3" y="6" width="18" height="12" rx="2.4" {...T} />
      <rect x="3" y="6" width="18" height="12" rx="2.4" />
      <path d="M8.5 21h7M12 18v3" />
    </>
  ),
  fridge: (
    <>
      <rect x="6" y="3" width="12" height="18" rx="2.5" {...T} />
      <rect x="6" y="3" width="12" height="18" rx="2.5" />
      <path d="M6 10.5h12M9 6.5v1.5M9 13v2.5" />
    </>
  ),
  flag: (
    <>
      <path d="M5 21V4" />
      <path d="M5 5h12l-2.5 3.5L17 12H5Z" {...T} />
      <path d="M5 5h12l-2.5 3.5L17 12H5" />
    </>
  ),
};

export type GlyphName = keyof typeof PATHS;

const GLYPH_BY_EMOJI: Record<string, GlyphName> = {
  "🎁": "gift", "📦": "gift", "🟫": "gift",
  "🛏": "bed", "💤": "bed", "🌙": "bed",
  "🛁": "bath", "🧻": "bath", "🚿": "bath",
  "🍳": "pan", "🍲": "pan", "🍟": "pan", "🍝": "pan", "🍚": "pan",
  "🧺": "basket", "🧹": "basket",
  "🧳": "suitcase", "🗺": "suitcase",
  "💜": "heart", "❤": "heart",
  "🎉": "party", "🥂": "party",
  "🌅": "sun",
  "☕": "coffee", "🥤": "coffee",
  "🧸": "bear", "🧶": "bear", "🧣": "bear",
  "🏠": "house",
  "🍽": "plate", "🍴": "plate",
  "🛋": "sofa",
  "🚼": "stroller", "🚗": "stroller",
  "👕": "shirt",
  "🛍": "bag", "👜": "bag",
  "✈": "plane",
  "👶": "baby", "🧷": "baby",
  "💍": "ring",
  "🎂": "cake", "🧁": "cake",
  "🎓": "cap",
  "🎄": "tree",
  "⭐": "star",
  "💡": "bulb", "✨": "bulb", "🤖": "bulb",
  "🔗": "link",
  "🔎": "search", "🔍": "search",
  "🗂": "folder",
  "📺": "tv", "📻": "tv", "📟": "tv",
  "🧊": "fridge",
  "🚀": "flag",
};

/** Resolves a glyph key or a legacy emoji string to a glyph name (never throws). */
export function glyphName(value?: string | null): GlyphName {
  if (!value) return "gift";
  if (value in PATHS) return value;
  const bare = value.replace(/️/g, "");
  return GLYPH_BY_EMOJI[bare] ?? "gift";
}

export interface GlyphProps extends Omit<SVGProps<SVGSVGElement>, "name"> {
  /** A glyph key ("gift") or a legacy emoji ("🎁"). */
  name?: string | null;
}

export function Glyph({ name, className, ...rest }: GlyphProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
      focusable="false"
      className={cn("size-[1em] shrink-0", className)}
      {...rest}
    >
      {PATHS[glyphName(name)]}
    </svg>
  );
}
