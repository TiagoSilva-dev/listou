"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CalendarDays, MapPin, Search } from "lucide-react";
import {
  eventTypeMeta,
  formatEventDate,
  formatPrice,
  PublicList,
  type PublicItem,
} from "@listou/types";
import {
  Avatar,
  Badge,
  buttonVariants,
  Card,
  CategoryTabs,
  cn,
  EmptyState,
  Input,
  MarketplaceBadge,
  Progress,
  Select, Glyph } from "@listou/ui";
import { api, apiVoid } from "@/lib/api";
import { displayPriceCents, isGone, matchesFilters, PRICE_BANDS } from "@/lib/public-list";
import { useStoredReservations } from "@/lib/reservations-store";
import { plural } from "@/lib/labels";
import { CoverArt } from "./cover-art";
import { GiftDrawer } from "./gift-drawer";
import { ItemVisual } from "./item-visual";
import { ShareDialog } from "./share-dialog";
import { Share2 } from "lucide-react";
import { Button } from "@listou/ui";

export function PublicRegistry({ initial, siteUrl }: { initial: PublicList; siteUrl: string }) {
  const [data, setData] = useState(initial);
  const [query, setQuery] = useState("");
  const [bandId, setBandId] = useState("all");
  const [category, setCategory] = useState("all");
  const [selected, setSelected] = useState<string | null>(null);
  const [shareOpen, setShareOpen] = useState(false);

  const { event } = data;
  const meta = eventTypeMeta(event.type);
  const mine = useStoredReservations(event.slug);
  const band = PRICE_BANDS.find((b) => b.id === bandId) ?? PRICE_BANDS[0]!;

  useEffect(() => {
    if (!data.preview) {
      void apiVoid("/analytics/track", {
        method: "POST",
        body: { name: "LIST_VIEWED", slug: event.slug },
      }).catch(() => undefined);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [event.slug]);

  const refetch = useCallback(async () => {
    try {
      setData(await api(`/public/lists/${event.slug}`, PublicList));
    } catch {
      // keep showing the last known state
    }
  }, [event.slug]);

  const tabs = useMemo(() => {
    const counts = new Map<string, number>();
    for (const it of data.items)
      counts.set(it.categoryId ?? "none", (counts.get(it.categoryId ?? "none") ?? 0) + 1);
    const out = [
      { id: "all", label: "Tudo", emoji: null as string | null, count: data.items.length },
    ];
    for (const c of data.categories)
      if (counts.get(c.id))
        out.push({ id: c.id, label: c.name, emoji: c.emoji, count: counts.get(c.id) ?? 0 });
    if (counts.has("none") && out.length > 1)
      out.push({ id: "none", label: "Outros", emoji: null, count: counts.get("none") ?? 0 });
    return out;
  }, [data]);

  const emojiByCategory = useMemo(
    () => new Map(data.categories.map((c) => [c.id, c.emoji])),
    [data.categories],
  );
  const visible = data.items.filter((it) =>
    matchesFilters(it, { query: query.trim(), band, categoryId: category }),
  );
  const selectedItem = data.items.find((i) => i.id === selected) ?? null;
  const chosen = data.progress.reservedUnits + data.progress.purchasedUnits;
  const date = formatEventDate(event.eventDate);
  const filtering = query.trim() !== "" || bandId !== "all" || category !== "all";

  return (
    <div className="bg-canvas min-h-dvh">
      {data.preview ? (
        <div role="status" className="bg-ink px-4 py-2.5 text-center text-sm text-white">
          Pré-visualização: só você vê esta página. Publique a lista no painel para compartilhar.{" "}
          <Link href="/dashboard" className="font-semibold underline underline-offset-2">
            Voltar ao painel
          </Link>
        </div>
      ) : null}

      <header className="relative">
        {event.coverImageUrl ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={event.coverImageUrl}
            alt=""
            fetchPriority="high"
            className="h-56 w-full object-cover sm:h-80"
          />
        ) : (
          <CoverArt theme={event.theme} emoji={meta.emoji} className="h-56 sm:h-80" />
        )}
        <div className="absolute inset-x-0 top-0 flex items-center justify-between p-4 sm:p-6">
          <Link
            href="/"
            className="font-display shadow-hairline rounded-full bg-white/85 px-3.5 py-1.5 text-sm font-semibold backdrop-blur"
          >
            Listou
          </Link>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setShareOpen(true)}
            className="bg-white/85 backdrop-blur"
          >
            <Share2 className="size-4" /> Compartilhar
          </Button>
        </div>
      </header>

      <section className="container-page -mt-14 flex max-w-3xl flex-col items-center gap-4 text-center sm:-mt-16">
        <Avatar
          name={event.hostNames ?? event.title}
          src={event.avatarUrl}
          size="xl"
          className="shadow-lift"
        />
        <Badge tone="primary">
          <Glyph name={meta.emoji} className="text-primary mr-1.5 inline size-[0.9em] align-[-0.1em]" /> {meta.headline}
        </Badge>
        <h1 className="font-display text-display-sm sm:text-display text-balance font-semibold tracking-tight">
          {event.hostNames ?? event.title}
        </h1>
        {event.hostNames && event.hostNames !== event.title ? (
          <p className="text-ink-soft text-lg">{event.title}</p>
        ) : null}
        <div className="text-ink-muted flex flex-wrap items-center justify-center gap-x-5 gap-y-1 text-sm">
          {date ? (
            <span className="inline-flex items-center gap-1.5">
              <CalendarDays className="size-4" /> {date}
            </span>
          ) : null}
          {event.location ? (
            <span className="inline-flex items-center gap-1.5">
              <MapPin className="size-4" /> {event.location}
            </span>
          ) : null}
        </div>
        {event.description ? (
          <p className="text-ink-soft max-w-xl text-pretty leading-relaxed">{event.description}</p>
        ) : null}

        {data.progress.totalUnits > 0 ? (
          <div className="mt-2 flex w-full max-w-md flex-col gap-2">
            <div className="text-ink-soft flex justify-between text-sm font-semibold">
              <span>
                {chosen} de {plural(data.progress.totalUnits, "presente", "presentes")} escolhidos
              </span>
              <span className="tabular-nums">
                {Math.round((chosen / data.progress.totalUnits) * 100)}%
              </span>
            </div>
            <Progress value={chosen} max={data.progress.totalUnits} label="Presentes escolhidos" />
          </div>
        ) : null}
      </section>

      <div className="border-line/70 bg-canvas/90 sticky top-0 z-20 mt-10 border-y backdrop-blur-md">
        <div className="container-page flex flex-col gap-3 py-3">
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Search
                aria-hidden
                className="text-ink-muted pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2"
              />
              <Input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Buscar na lista"
                aria-label="Buscar na lista"
                className="h-11 pl-11"
              />
            </div>
            <Select
              value={bandId}
              onChange={(e) => setBandId(e.target.value)}
              aria-label="Faixa de preço"
              className="h-11 w-40 shrink-0 sm:w-52"
            >
              {PRICE_BANDS.map((b) => (
                <option key={b.id} value={b.id}>
                  {b.label}
                </option>
              ))}
            </Select>
          </div>
          {tabs.length > 2 ? (
            <CategoryTabs tabs={tabs} value={category} onChange={setCategory} />
          ) : null}
        </div>
      </div>

      <main className="container-page py-8">
        {data.items.length === 0 ? (
          <EmptyState
            icon="gift"
            title="A lista ainda está sendo montada"
            description="Volte em breve para ver os presentes escolhidos com carinho."
          />
        ) : visible.length === 0 ? (
          <EmptyState
            icon="search"
            title="Nenhum presente encontrado"
            description="Tente outra busca ou outra faixa de preço."
            action={
              filtering ? (
                <Button
                  variant="secondary"
                  onClick={() => {
                    setQuery("");
                    setBandId("all");
                    setCategory("all");
                  }}
                >
                  Limpar filtros
                </Button>
              ) : undefined
            }
          />
        ) : (
          <ul className="grid grid-cols-2 gap-3 sm:gap-5 md:grid-cols-3 lg:grid-cols-4">
            {visible.map((it, i) => (
              <li
                key={it.id}
                className="animate-fade-up"
                style={{ animationDelay: `${Math.min(i, 8) * 40}ms` }}
              >
                <PublicItemCard
                  item={it}
                  categoryEmoji={it.categoryId ? emojiByCategory.get(it.categoryId) : null}
                  mine={mine.some((r) => r.itemId === it.id)}
                  onOpen={() => setSelected(it.id)}
                />
              </li>
            ))}
          </ul>
        )}
      </main>

      <section className="container-page pb-16">
        <CoverArt
          theme="lavender"
          className="rounded-hero flex flex-col items-center gap-4 px-6 py-12 text-center"
        >
          <Glyph name="gift" className="text-primary size-6" />
          <h2 className="font-display text-2xl font-semibold sm:text-3xl">Gostou da ideia?</h2>
          <p className="text-ink-soft max-w-md">
            Crie sua lista gratuitamente e compartilhe com quem você ama.
          </p>
          <Link href="/criar" className={buttonVariants({ variant: "dark", size: "lg" })}>
            Criar minha lista
          </Link>
        </CoverArt>
        <p className="text-ink-muted mt-8 text-center text-xs leading-relaxed">
          Ao comprar por links desta página, as lojas podem pagar uma comissão ao Listou, sem custo
          extra para você.
        </p>
      </section>

      <GiftDrawer
        slug={event.slug}
        item={selectedItem}
        allowReservations={data.list.allowReservations}
        mine={mine}
        onClose={() => setSelected(null)}
        onChanged={() => void refetch()}
      />
      <ShareDialog
        open={shareOpen}
        onOpenChange={setShareOpen}
        url={`${siteUrl}/l/${event.slug}`}
        title={event.title}
        slug={event.slug}
      />
    </div>
  );
}

function PublicItemCard({
  item,
  categoryEmoji,
  mine,
  onOpen,
}: {
  item: PublicItem;
  categoryEmoji?: string | null;
  mine: boolean;
  onOpen: () => void;
}) {
  const gone = isGone(item);
  const price = displayPriceCents(item);
  const hasOffers = item.offers.length > 0;
  return (
    <Card
      className={cn(
        "duration-(--duration-base) ease-out-soft group flex h-full flex-col overflow-hidden transition-all",
        !gone && "hover:shadow-lift hover:-translate-y-1",
      )}
    >
      <button
        type="button"
        onClick={onOpen}
        className="bg-canvas-deep focus-visible:shadow-focus relative aspect-square overflow-hidden text-left focus-visible:outline-none"
        aria-label={`Ver ${item.title}`}
      >
        <ItemVisual
          imageUrl={item.imageUrl ?? item.product?.imageUrl}
          emoji={item.emoji ?? categoryEmoji}
          alt={item.title}
          className={cn(
            "duration-(--duration-slow) ease-out-soft transition-all group-hover:scale-105",
            gone && "opacity-40 grayscale",
          )}
        />
        <span className="absolute left-2.5 top-2.5 flex flex-col items-start gap-1.5">
          {gone ? (
            <Badge tone="neutral">Já escolhido</Badge>
          ) : item.status === "PARTIALLY_RESERVED" ? (
            <Badge tone="warning">{item.availableQuantity} restantes</Badge>
          ) : item.desiredQuantity > 1 ? (
            <Badge tone="glass">{item.desiredQuantity}× desejado</Badge>
          ) : null}
          {mine ? <Badge tone="primary">Sua escolha</Badge> : null}
        </span>
      </button>
      <div className="flex flex-1 flex-col gap-2 p-3.5 sm:p-4">
        <h3 className="line-clamp-2 text-sm font-semibold leading-snug sm:text-base">
          {item.title}
        </h3>
        {price != null ? (
          <p className="text-sm tabular-nums sm:text-base">
            {!hasOffers ? <span className="text-ink-muted mr-1 text-xs">~</span> : null}
            <span className="font-semibold">{formatPrice(price)}</span>
          </p>
        ) : null}
        {hasOffers ? (
          <div className="flex flex-wrap gap-x-2.5 gap-y-0.5">
            {item.offers.slice(0, 3).map((o) => (
              <MarketplaceBadge key={o.id} code={o.merchant.code} name={o.merchant.name} />
            ))}
          </div>
        ) : null}
        <Button
          size="sm"
          variant={gone ? "secondary" : "primary"}
          className="mt-auto"
          block
          onClick={onOpen}
          disabled={gone && !mine}
        >
          {mine ? "Ver minha escolha" : gone ? "Indisponível" : "Presentear"}
        </Button>
      </div>
    </Card>
  );
}
