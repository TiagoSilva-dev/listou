import type { ReactNode } from "react";
import { CoverArt } from "./cover-art";
import { Logo } from "./logo";

export function AuthShell({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle: string;
  children: ReactNode;
}) {
  return (
    <main className="grid min-h-dvh lg:grid-cols-[1fr_1.05fr]">
      <section className="flex flex-col px-5 py-6 sm:px-10">
        <Logo />
        <div className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center gap-8 py-10">
          <div className="animate-fade-up flex flex-col gap-2">
            <h1 className="font-display text-display-sm font-semibold tracking-tight">{title}</h1>
            <p className="text-ink-muted">{subtitle}</p>
          </div>
          {children}
        </div>
      </section>
      <CoverArt theme="lavender" emoji="gift" className="hidden items-end p-12 lg:flex">
        <figure className="flex max-w-md flex-col gap-4">
          <blockquote className="font-display text-ink text-balance text-3xl font-semibold leading-tight">
            “Em dois minutos nossa lista estava no ar. Os convidados amaram a facilidade.”
          </blockquote>
          <figcaption className="text-ink-soft text-sm">
            Exemplo de depoimento · Casal fictício de demonstração
          </figcaption>
        </figure>
      </CoverArt>
    </main>
  );
}
