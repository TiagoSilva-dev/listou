import Link from "next/link";
import { cn } from "@listou/ui";

export function Logo({ className, href = "/" }: { className?: string; href?: string }) {
  return (
    <Link
      href={href}
      className={cn("group inline-flex items-center gap-2", className)}
      aria-label="Listou — início"
    >
      <span
        aria-hidden
        className="from-primary to-blush shadow-soft duration-(--duration-base) ease-spring grid size-8 place-items-center rounded-[0.7rem] bg-gradient-to-br text-white transition-transform group-hover:-rotate-6"
      >
        <svg
          viewBox="0 0 24 24"
          className="size-4.5"
          fill="none"
          stroke="currentColor"
          strokeWidth="2.2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M20 12v9H4v-9" />
          <path d="M2 7h20v5H2z" />
          <path d="M12 21V7" />
          <path d="M12 7c-1.5-3-5-4-5.5-1.5S9.5 7 12 7Zm0 0c1.5-3 5-4 5.5-1.5S14.5 7 12 7Z" />
        </svg>
      </span>
      <span className="font-display text-ink text-xl font-semibold tracking-tight">Listou</span>
    </Link>
  );
}
