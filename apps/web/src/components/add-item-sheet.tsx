"use client";

import { useEffect, useState } from "react";
import { Plus, Search } from "lucide-react";
import { z } from "zod";
import { formatPrice, ListItem, SearchResult, type Category, type OwnerList } from "@listou/types";
import {
  Badge,
  Button,
  cn,
  Input,
  MarketplaceBadge,
  Select,
  Sheet,
  Skeleton,
  useToast,
} from "@listou/ui";
import { api, ApiError } from "@/lib/api";
import { ItemForm, type ItemPayload } from "./item-form";
import { ItemVisual } from "./item-visual";

type Tab = "search" | "manual";

export function AddItemSheet({
  open,
  onOpenChange,
  listId,
  categories,
  onAdded,
  initialQuery = "",
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  listId: string;
  categories: Category[];
  onAdded: (item: ListItem) => void;
  initialQuery?: string;
}) {
  const toast = useToast();
  const [tab, setTab] = useState<Tab>("search");
  const [query, setQuery] = useState(initialQuery);
  const [sort, setSort] = useState("relevance");
  const [adding, setAdding] = useState<string | null>(null);
  // The last completed search; "loading" is derived from it not matching the current inputs.
  const [searched, setSearched] = useState<{
    key: string;
    results: SearchResult[];
    error: string | null;
  } | null>(null);

  const q = query.trim();
  const searchKey = `${q}|${sort}`;
  const active = q.length >= 2;
  const loading = open && active && searched?.key !== searchKey;
  const results = active ? (searched?.results ?? null) : null;
  const searchError = active && searched?.key === searchKey ? searched.error : null;

  useEffect(() => {
    if (!open || !active) return;
    let cancelled = false;
    const handle = window.setTimeout(async () => {
      try {
        const res = await api(
          `/products/search?q=${encodeURIComponent(q)}&sort=${sort}`,
          z.object({ results: z.array(SearchResult) }),
        );
        if (!cancelled) setSearched({ key: searchKey, results: res.results, error: null });
      } catch (err) {
        if (!cancelled)
          setSearched({
            key: searchKey,
            results: [],
            error: err instanceof ApiError ? err.message : "Não foi possível buscar agora.",
          });
      }
    }, 300);
    return () => {
      cancelled = true;
      window.clearTimeout(handle);
    };
  }, [q, sort, open, active, searchKey]);

  async function addProduct(r: SearchResult) {
    setAdding(r.externalId);
    try {
      const imported = await api(
        "/products/import",
        z.object({ product: z.object({ id: z.string() }) }),
        {
          method: "POST",
          body: { providerCode: r.providerCode, externalId: r.externalId },
        },
      );
      const { item } = await api(`/lists/${listId}/items`, z.object({ item: ListItem }), {
        method: "POST",
        body: { productId: imported.product.id, desiredQuantity: 1 },
      });
      onAdded(item);
      toast(`“${item.title}” foi adicionado à lista`, "success");
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível adicionar.", "error");
    } finally {
      setAdding(null);
    }
  }

  async function addManual(p: ItemPayload) {
    try {
      const { item } = await api(`/lists/${listId}/items`, z.object({ item: ListItem }), {
        method: "POST",
        body: {
          title: p.title,
          description: p.description || undefined,
          desiredQuantity: p.desiredQuantity,
          priceReferenceCents: p.priceReferenceCents,
          externalUrl: p.externalUrl || undefined,
          categoryId: p.categoryId || undefined,
        },
      });
      onAdded(item);
      toast(`“${item.title}” foi adicionado à lista`, "success");
      onOpenChange(false);
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível adicionar.", "error");
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange} title="Adicionar item" className="sm:max-w-2xl">
      <div className="flex flex-col gap-5">
        <div role="tablist" className="rounded-control bg-canvas-deep grid grid-cols-2 gap-1 p-1">
          {(
            [
              ["search", "Buscar produtos"],
              ["manual", "Adicionar à mão"],
            ] as const
          ).map(([id, label]) => (
            <button
              key={id}
              role="tab"
              type="button"
              aria-selected={tab === id}
              onClick={() => setTab(id)}
              className={cn(
                "rounded-chip duration-(--duration-fast) h-10 text-sm font-semibold transition-all",
                tab === id ? "bg-surface text-ink shadow-soft" : "text-ink-muted hover:text-ink",
              )}
            >
              {label}
            </button>
          ))}
        </div>

        {tab === "search" ? (
          <div className="flex flex-col gap-4">
            <div className="flex gap-2">
              <div className="relative flex-1">
                <Search
                  aria-hidden
                  className="text-ink-muted pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2"
                />
                <Input
                  autoFocus
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Buscar: air fryer, jogo de cama…"
                  className="pl-11"
                  aria-label="Buscar produtos"
                />
              </div>
              <Select
                value={sort}
                onChange={(e) => setSort(e.target.value)}
                className="w-40 shrink-0"
                aria-label="Ordenar"
              >
                <option value="relevance">Relevância</option>
                <option value="price_asc">Menor preço</option>
                <option value="price_desc">Maior preço</option>
              </Select>
            </div>
            <p className="text-ink-muted text-xs">
              Catálogo de demonstração: produtos e preços fictícios. Nenhum dado de loja real é
              exibido ainda.
            </p>

            {loading && searched === null ? (
              <div className="flex flex-col gap-3">
                {[0, 1, 2].map((i) => (
                  <Skeleton key={i} className="rounded-card h-24" />
                ))}
              </div>
            ) : searchError ? (
              <p
                role="alert"
                className="rounded-control bg-danger-soft text-danger px-4 py-3 text-sm"
              >
                {searchError}
              </p>
            ) : results === null ? (
              <p className="text-ink-muted py-10 text-center text-sm">
                Digite o que você quer ganhar e mostramos opções.
              </p>
            ) : results.length === 0 ? (
              <div className="flex flex-col items-center gap-3 py-10 text-center">
                <p className="font-semibold">Nenhum produto encontrado para “{query}”</p>
                <Button variant="soft" onClick={() => setTab("manual")}>
                  Adicionar “{query}” à mão
                </Button>
              </div>
            ) : (
              <ul className={cn("flex flex-col gap-3 transition-opacity", loading && "opacity-60")}>
                {results.map((r) => (
                  <li
                    key={`${r.providerCode}-${r.externalId}`}
                    className="rounded-card bg-canvas flex items-center gap-4 p-3"
                  >
                    <div className="rounded-control bg-surface size-20 shrink-0 overflow-hidden">
                      <ItemVisual imageUrl={r.imageUrl} emoji={r.emoji} alt={r.title} />
                    </div>
                    <div className="flex min-w-0 flex-1 flex-col gap-1">
                      <p className="line-clamp-2 text-sm font-semibold leading-snug">{r.title}</p>
                      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                        {r.offers.map((o) => (
                          <MarketplaceBadge
                            key={o.merchant.code}
                            code={o.merchant.code}
                            name={o.merchant.name}
                          />
                        ))}
                      </div>
                      <p className="text-sm tabular-nums">
                        {r.lowestPriceCents != null ? (
                          <>
                            <span className="text-ink-muted text-xs">a partir de </span>
                            <span className="font-semibold">{formatPrice(r.lowestPriceCents)}</span>
                          </>
                        ) : null}
                        {r.demo ? <Badge className="ml-2">demonstração</Badge> : null}
                      </p>
                    </div>
                    <Button
                      size="sm"
                      onClick={() => addProduct(r)}
                      loading={adding === r.externalId}
                      disabled={adding !== null}
                    >
                      <Plus className="size-4" /> Adicionar
                    </Button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        ) : (
          <ItemForm categories={categories} submitLabel="Adicionar à lista" onSubmit={addManual} />
        )}
      </div>
    </Sheet>
  );
}

export type { OwnerList };
