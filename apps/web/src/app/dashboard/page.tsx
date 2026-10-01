import type { Metadata } from "next";
import Link from "next/link";
import { Plus } from "lucide-react";
import { z } from "zod";
import { Event, eventTypeMeta, formatEventDate } from "@listou/types";
import { Badge, buttonVariants, EmptyState, Glyph } from "@listou/ui";
import { CoverArt } from "@/components/cover-art";
import { serverApi } from "@/lib/server-api";

export const metadata: Metadata = { title: "Minhas listas" };

export default async function DashboardPage() {
  const { events } = await serverApi("/events", z.object({ events: z.array(Event) }), {
    auth: true,
  });

  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-end justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="font-display text-display-sm font-semibold tracking-tight">
            Minhas listas
          </h1>
          <p className="text-ink-muted">Seus momentos especiais, num só lugar.</p>
        </div>
        {events.length > 0 ? (
          <Link href="/criar" className={buttonVariants({ size: "md" })}>
            <Plus className="size-4" /> Criar lista
          </Link>
        ) : null}
      </div>

      {events.length === 0 ? (
        <EmptyState
          icon="gift"
          title="Vamos criar sua primeira lista?"
          description="Escolha o momento, monte a lista e compartilhe com quem você ama. Leva poucos minutos."
          action={
            <Link href="/criar" className={buttonVariants({ size: "lg" })}>
              Criar minha primeira lista
            </Link>
          }
        />
      ) : (
        <ul className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {events.map((e, i) => {
            const meta = eventTypeMeta(e.type);
            const date = formatEventDate(e.eventDate);
            return (
              <li key={e.id} className="animate-fade-up" style={{ animationDelay: `${i * 60}ms` }}>
                <Link
                  href={`/dashboard/${e.id}`}
                  className="rounded-card bg-surface shadow-soft duration-(--duration-base) ease-out-soft hover:shadow-lift group flex h-full flex-col overflow-hidden transition-all hover:-translate-y-1"
                >
                  <CoverArt theme={e.theme} emoji={meta.emoji} className="h-36 p-4">
                    <Badge tone="glass">
                      <Glyph name={meta.emoji} className="mr-1 size-3.5" /> {meta.label}
                    </Badge>
                  </CoverArt>
                  <div className="flex flex-1 flex-col gap-3 p-5">
                    <h2 className="font-display text-balance text-xl font-semibold leading-snug">
                      {e.title}
                    </h2>
                    <div className="text-ink-muted mt-auto flex items-center justify-between text-sm">
                      <span>{date ?? "Sem data"}</span>
                      <Badge tone={e.status === "PUBLISHED" ? "success" : "neutral"}>
                        {e.status === "PUBLISHED"
                          ? "No ar"
                          : e.status === "DRAFT"
                            ? "Rascunho"
                            : "Arquivada"}
                      </Badge>
                    </div>
                  </div>
                </Link>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
