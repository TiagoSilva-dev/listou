"use client";

import Link from "next/link";
import { useCallback, useMemo, useState } from "react";
import {
  Eye,
  Gift,
  MousePointerClick,
  Pencil,
  Plus,
  Rocket,
  Settings,
  Share2,
  ShoppingBag,
} from "lucide-react";
import { z } from "zod";
import {
  Dashboard,
  Event,
  eventTypeMeta,
  formatEventDate,
  formatPrice,
  OwnerList,
  type ListItem,
} from "@listou/types";
import {
  Badge,
  Button,
  buttonVariants,
  Card,
  CategoryTabs,
  EmptyState,
  MarketplaceBadge,
  Progress,
  useToast,
} from "@listou/ui";
import { api, ApiError } from "@/lib/api";
import { plural, STATUS_LABEL, STATUS_TONE } from "@/lib/labels";
import { AddItemSheet } from "./add-item-sheet";
import { CoverArt } from "./cover-art";
import { EditItemSheet } from "./edit-item-sheet";
import { ItemVisual } from "./item-visual";
import { SettingsSheet } from "./settings-sheet";
import { ShareDialog } from "./share-dialog";

interface Props {
  initialDashboard: Dashboard;
  initialList: OwnerList;
  siteUrl: string;
  justCreated: boolean;
}

const ACTIVITY_LABEL = {
  RESERVED: "reservou",
  PURCHASED: "marcou como comprado",
  CANCELLED: "cancelou a reserva de",
  CLICKED: "abriu a loja de",
} as const;

export function EventWorkspace({ initialDashboard, initialList, siteUrl, justCreated }: Props) {
  const toast = useToast();
  const [dash, setDash] = useState(initialDashboard);
  const [list, setList] = useState(initialList);
  const [category, setCategory] = useState("all");
  const [addOpen, setAddOpen] = useState(false);
  const [editing, setEditing] = useState<ListItem | null>(null);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [shareOpen, setShareOpen] = useState(false);
  const [publishing, setPublishing] = useState(false);

  const event = dash.event;
  const meta = eventTypeMeta(event.type);
  const published = event.status === "PUBLISHED";
  const publicUrl = `${siteUrl}/l/${event.slug}`;

  const refresh = useCallback(async () => {
    const [d, l] = await Promise.all([
      api(`/events/${event.id}/dashboard`, Dashboard),
      api(`/events/${event.id}/list`, OwnerList),
    ]);
    setDash(d);
    setList(l);
  }, [event.id]);

  async function setStatus(status: "PUBLISHED" | "DRAFT") {
    setPublishing(true);
    try {
      const res = await api(`/events/${event.id}`, z.object({ event: Event }), {
        method: "PATCH",
        body: { status, visibility: event.visibility },
      });
      setDash((d) => ({ ...d, event: res.event }));
      if (status === "PUBLISHED") {
        toast("Sua lista está no ar! 🎉", "success");
        setShareOpen(true);
      } else toast("Lista voltou para rascunho");
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível alterar.", "error");
    } finally {
      setPublishing(false);
    }
  }

  const tabs = useMemo(() => {
    const counts = new Map<string, number>();
    for (const it of list.items)
      counts.set(it.categoryId ?? "none", (counts.get(it.categoryId ?? "none") ?? 0) + 1);
    const base = [{ id: "all", label: "Tudo", count: list.items.length }];
    for (const c of list.categories)
      base.push({ id: c.id, label: c.name, count: counts.get(c.id) ?? 0 });
    if (counts.has("none") && list.categories.length > 0)
      base.push({ id: "none", label: "Outros", count: counts.get("none") ?? 0 });
    return base;
  }, [list]);

  const emojiByCategory = useMemo(
    () => new Map(list.categories.map((c) => [c.id, c.emoji])),
    [list.categories],
  );
  const visible = list.items.filter(
    (it) =>
      category === "all" ||
      (category === "none" ? it.categoryId === null : it.categoryId === category),
  );
  const { stats } = dash;
  const done = stats.reservedUnits + stats.purchasedUnits;

  return (
    <div className="flex flex-col gap-8">
      <CoverArt theme={event.theme} emoji={meta.emoji} className="rounded-hero p-6 sm:p-10">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex flex-col gap-3">
            <div className="flex flex-wrap items-center gap-2">
              <Badge tone="glass">
                {meta.emoji} {meta.label}
              </Badge>
              <Badge tone={published ? "success" : "glass"}>
                {published ? "No ar" : "Rascunho"}
              </Badge>
            </div>
            <h1 className="font-display text-display-sm sm:text-display text-balance font-semibold tracking-tight">
              {event.title}
            </h1>
            <p className="text-ink-soft">
              {formatEventDate(event.eventDate) ?? "Sem data definida"}
              {stats.daysRemaining != null && stats.daysRemaining >= 0 ? (
                <span className="ml-2 font-semibold">
                  ·{" "}
                  {stats.daysRemaining === 0
                    ? "é hoje! 🎉"
                    : `faltam ${plural(stats.daysRemaining, "dia", "dias")}`}
                </span>
              ) : null}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button variant="dark" onClick={() => setShareOpen(true)}>
              <Share2 className="size-4" /> Compartilhar
            </Button>
            <Link
              href={`/l/${event.slug}`}
              target="_blank"
              className={buttonVariants({ variant: "secondary" })}
            >
              <Eye className="size-4" /> {published ? "Ver página" : "Pré-visualizar"}
            </Link>
            <Button
              variant="secondary"
              size="icon"
              onClick={() => setSettingsOpen(true)}
              aria-label="Configurações"
            >
              <Settings className="size-4" />
            </Button>
          </div>
        </div>
      </CoverArt>

      {!published ? (
        <Card className="bg-primary-soft flex flex-col gap-4 p-5 shadow-none sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-start gap-4">
            <span
              aria-hidden
              className="bg-surface grid size-11 shrink-0 place-items-center rounded-full text-xl"
            >
              {justCreated ? "🎉" : "🚀"}
            </span>
            <div>
              <p className="font-semibold">
                {justCreated ? "Sua lista foi criada!" : "Sua lista ainda é um rascunho"}
              </p>
              <p className="text-ink-soft text-sm">
                {list.items.length === 0
                  ? "Adicione alguns itens e publique para compartilhar com os convidados."
                  : "Quando estiver pronta, publique para gerar o link público."}
              </p>
            </div>
          </div>
          <Button
            onClick={() => setStatus("PUBLISHED")}
            loading={publishing}
            disabled={list.items.length === 0}
          >
            <Rocket className="size-4" /> Publicar lista
          </Button>
        </Card>
      ) : null}

      <section aria-label="Resumo" className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <Stat
          icon={<Gift className="size-4" />}
          label="Itens"
          value={stats.itemsCount}
          hint={plural(stats.totalUnits, "presente desejado", "presentes desejados")}
        />
        <Stat
          icon={<ShoppingBag className="size-4" />}
          label="Reservados"
          value={stats.reservedUnits}
          hint={dash.surpriseMode ? "modo surpresa" : `de ${stats.totalUnits}`}
        />
        <Stat
          icon={<Gift className="size-4" />}
          label="Comprados"
          value={stats.purchasedUnits}
          hint={dash.surpriseMode ? "modo surpresa" : `de ${stats.totalUnits}`}
        />
        <Stat
          icon={<Eye className="size-4" />}
          label="Visualizações"
          value={stats.views}
          hint={`${plural(stats.outboundClicks, "clique", "cliques")} nas lojas`}
        />
      </section>

      {stats.totalUnits > 0 ? (
        <Card className="flex flex-col gap-3 p-5">
          <div className="flex items-baseline justify-between">
            <p className="font-semibold">
              {done} de {stats.totalUnits} presentes escolhidos
            </p>
            <p className="text-ink-muted text-sm tabular-nums">
              {Math.round((done / stats.totalUnits) * 100)}%
            </p>
          </div>
          <Progress value={done} max={stats.totalUnits} label="Presentes escolhidos" />
          {dash.surpriseMode ? (
            <p className="text-ink-muted text-sm">
              🎁 Modo surpresa ligado: você vê o total, mas não quem deu o quê.
            </p>
          ) : null}
        </Card>
      ) : null}

      <section className="flex flex-col gap-5" aria-label="Itens da lista">
        <div className="flex items-center justify-between gap-3">
          <h2 className="font-display text-2xl font-semibold">Itens da lista</h2>
          <Button onClick={() => setAddOpen(true)}>
            <Plus className="size-4" /> Adicionar item
          </Button>
        </div>

        {list.items.length > 0 && tabs.length > 2 ? (
          <CategoryTabs tabs={tabs} value={category} onChange={setCategory} />
        ) : null}

        {list.items.length === 0 ? (
          <EmptyState
            emoji="🛍️"
            title="Sua lista está vazia"
            description="Busque produtos de várias lojas ou adicione qualquer desejo à mão."
            action={
              <Button size="lg" onClick={() => setAddOpen(true)}>
                <Plus className="size-4" /> Adicionar o primeiro item
              </Button>
            }
          />
        ) : visible.length === 0 ? (
          <EmptyState
            emoji="🗂️"
            title="Nada nesta categoria ainda"
            description="Adicione um item e escolha esta categoria."
          />
        ) : (
          <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {visible.map((it) => (
              <li key={it.id}>
                <OwnerItemCard
                  item={it}
                  categoryEmoji={it.categoryId ? emojiByCategory.get(it.categoryId) : null}
                  hideStatus={dash.surpriseMode}
                  onEdit={() => setEditing(it)}
                />
              </li>
            ))}
          </ul>
        )}
      </section>

      {dash.recentActivity.length > 0 ? (
        <section className="flex flex-col gap-4" aria-label="Atividade recente">
          <h2 className="font-display text-2xl font-semibold">Atividade recente</h2>
          <Card className="divide-line divide-y">
            {dash.recentActivity.map((a, i) => (
              <div key={i} className="flex items-center gap-3 p-4 text-sm">
                <span
                  aria-hidden
                  className="bg-primary-soft grid size-9 place-items-center rounded-full"
                >
                  {a.kind === "CLICKED" ? (
                    <MousePointerClick className="text-primary size-4" />
                  ) : (
                    <Gift className="text-primary size-4" />
                  )}
                </span>
                <p className="text-ink-soft flex-1">
                  {a.guestName ? (
                    <strong className="text-ink">{a.guestName} </strong>
                  ) : (
                    <strong className="text-ink">Alguém </strong>
                  )}
                  {ACTIVITY_LABEL[a.kind]}
                  {a.itemTitle ? (
                    <strong className="text-ink"> {a.itemTitle}</strong>
                  ) : (
                    " um presente"
                  )}
                </p>
                <time className="text-ink-muted text-xs" dateTime={a.at}>
                  {new Date(a.at).toLocaleDateString("pt-BR", { day: "numeric", month: "short" })}
                </time>
              </div>
            ))}
          </Card>
        </section>
      ) : null}

      {published ? (
        <p className="text-ink-muted text-center text-sm">
          <button
            type="button"
            className="text-primary font-semibold hover:underline"
            onClick={() => setStatus("DRAFT")}
          >
            Voltar para rascunho
          </button>{" "}
          (a página pública deixa de funcionar até você publicar de novo)
        </p>
      ) : null}

      <AddItemSheet
        open={addOpen}
        onOpenChange={setAddOpen}
        listId={list.list.id}
        categories={list.categories}
        onAdded={() => void refresh()}
      />
      <EditItemSheet
        item={editing}
        categories={list.categories}
        onClose={() => setEditing(null)}
        onSaved={() => void refresh()}
        onDeleted={() => void refresh()}
      />
      <SettingsSheet
        key={event.updatedAt}
        open={settingsOpen}
        onOpenChange={setSettingsOpen}
        event={event}
        onSaved={(e) => setDash((d) => ({ ...d, event: e, surpriseMode: e.surpriseMode }))}
      />
      <ShareDialog
        open={shareOpen}
        onOpenChange={setShareOpen}
        url={publicUrl}
        title={event.title}
        slug={event.slug}
      />
    </div>
  );
}

function Stat({
  icon,
  label,
  value,
  hint,
}: {
  icon: React.ReactNode;
  label: string;
  value: number;
  hint: string;
}) {
  return (
    <Card className="flex flex-col gap-1 p-5">
      <span className="text-ink-muted flex items-center gap-2 text-sm font-semibold">
        {icon}
        {label}
      </span>
      <span className="font-display text-4xl font-semibold tabular-nums">{value}</span>
      <span className="text-ink-muted text-xs">{hint}</span>
    </Card>
  );
}

function OwnerItemCard({
  item,
  categoryEmoji,
  hideStatus,
  onEdit,
}: {
  item: ListItem;
  categoryEmoji?: string | null;
  hideStatus: boolean;
  onEdit: () => void;
}) {
  const best = item.offers[0];
  return (
    <Card className="duration-(--duration-base) ease-out-soft hover:shadow-lift group flex h-full flex-col overflow-hidden transition-all hover:-translate-y-0.5">
      <div className="bg-canvas-deep relative aspect-[4/3] overflow-hidden">
        <ItemVisual
          imageUrl={item.imageUrl ?? item.product?.imageUrl}
          emoji={item.emoji ?? categoryEmoji}
          alt={item.title}
        />
        {item.desiredQuantity > 1 ? (
          <Badge tone="glass" className="absolute left-3 top-3">
            {item.desiredQuantity}× desejado
          </Badge>
        ) : null}
        <Button
          variant="secondary"
          size="icon"
          onClick={onEdit}
          aria-label={`Editar ${item.title}`}
          className="absolute right-3 top-3 size-9 opacity-0 transition-opacity focus-visible:opacity-100 group-hover:opacity-100 max-sm:opacity-100"
        >
          <Pencil className="size-4" />
        </Button>
      </div>
      <div className="flex flex-1 flex-col gap-3 p-4">
        <h3 className="line-clamp-2 font-semibold leading-snug">{item.title}</h3>
        <div className="mt-auto flex flex-col gap-2">
          <div className="flex items-center justify-between gap-2">
            <span className="text-sm tabular-nums">
              {best?.priceCents != null ? (
                formatPrice(best.priceCents, best.currency)
              ) : item.priceReferenceCents != null ? (
                <>~ {formatPrice(item.priceReferenceCents)}</>
              ) : (
                <span className="text-ink-muted">Sem valor</span>
              )}
            </span>
            {hideStatus ? null : (
              <Badge tone={STATUS_TONE[item.status]}>{STATUS_LABEL[item.status]}</Badge>
            )}
          </div>
          {item.offers.length > 0 ? (
            <div className="flex flex-wrap gap-x-3 gap-y-1">
              {item.offers.map((o) => (
                <MarketplaceBadge key={o.id} code={o.merchant.code} name={o.merchant.name} />
              ))}
            </div>
          ) : (
            <p className="text-ink-muted text-xs">
              {item.externalUrl ? "Com link de referência" : "Item livre (sem loja)"}
            </p>
          )}
        </div>
      </div>
    </Card>
  );
}
