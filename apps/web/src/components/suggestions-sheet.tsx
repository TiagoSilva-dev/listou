"use client";

import { useState } from "react";
import { Sparkles } from "lucide-react";
import { ApplySuggestionsResult, Suggestions, type SuggestedCategory } from "@listou/types";
import { Badge, Button, EmptyState, Input, Sheet, Skeleton, useToast } from "@listou/ui";
import { api, ApiError } from "@/lib/api";

const PRIORITY = { ESSENTIAL: "HIGH", RECOMMENDED: "MEDIUM", OPTIONAL: "LOW" } as const;
const IMPORTANCE_LABEL = {
  ESSENTIAL: "Essencial",
  RECOMMENDED: "Recomendado",
  OPTIONAL: "Opcional",
};

const keyOf = (c: number, d: number) => `${c}:${d}`;

export function SuggestionsSheet({
  open,
  onOpenChange,
  eventId,
  onApplied,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  eventId: string;
  onApplied: () => void;
}) {
  const toast = useToast();
  const [prompt, setPrompt] = useState("");
  const [loading, setLoading] = useState(false);
  const [applying, setApplying] = useState(false);
  const [categories, setCategories] = useState<SuggestedCategory[] | null>(null);
  const [picked, setPicked] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string | null>(null);

  async function suggest() {
    setLoading(true);
    setError(null);
    try {
      const res = await api(`/events/${eventId}/suggestions`, Suggestions, {
        method: "POST",
        body: { prompt },
      });
      setCategories(res.categories);
      // Essentials start selected; the rest is the owner's call.
      const initial = new Set<string>();
      res.categories.forEach((c, ci) =>
        c.desires.forEach((d, di) => d.importance === "ESSENTIAL" && initial.add(keyOf(ci, di))),
      );
      setPicked(initial);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Não foi possível sugerir agora.");
    } finally {
      setLoading(false);
    }
  }

  function toggle(key: string) {
    setPicked((prev) => {
      const next = new Set(prev);
      if (!next.delete(key)) next.add(key);
      return next;
    });
  }

  async function apply() {
    if (!categories) return;
    const body = {
      categories: categories
        .map((c, ci) => ({
          name: c.name,
          emoji: c.emoji,
          desires: c.desires
            .filter((_, di) => picked.has(keyOf(ci, di)))
            .map((d) => ({
              title: d.title,
              emoji: d.emoji,
              quantity: d.quantity,
              priority: PRIORITY[d.importance],
            })),
        }))
        .filter((c) => c.desires.length > 0),
    };
    setApplying(true);
    try {
      const res = await api(`/events/${eventId}/suggestions/apply`, ApplySuggestionsResult, {
        method: "POST",
        body,
      });
      toast(
        `${res.addedItems} ${res.addedItems === 1 ? "item adicionado" : "itens adicionados"} ✨`,
        "success",
      );
      setCategories(null);
      setPicked(new Set());
      onOpenChange(false);
      onApplied();
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível adicionar.", "error");
    } finally {
      setApplying(false);
    }
  }

  return (
    <Sheet
      open={open}
      onOpenChange={onOpenChange}
      title="Ideias para a sua lista"
      description="Sugestões de desejos que ainda não estão na lista. Você escolhe o que entra."
    >
      <form
        className="flex gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          void suggest();
        }}
      >
        <Input
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          maxLength={300}
          placeholder="Ex.: banheiro, apartamento pequeno (opcional)"
          aria-label="Conte o que você procura"
        />
        <Button type="submit" loading={loading}>
          <Sparkles className="size-4" /> {categories ? "Atualizar" : "Sugerir"}
        </Button>
      </form>

      {error ? (
        <p role="alert" className="text-danger mt-4 text-sm">
          {error}
        </p>
      ) : null}

      {loading ? (
        <div className="mt-6 flex flex-col gap-3">
          <Skeleton className="h-6 w-1/3" />
          <Skeleton className="h-12" />
          <Skeleton className="h-12" />
        </div>
      ) : categories && categories.length === 0 ? (
        <div className="mt-6">
          <EmptyState
            emoji="🎉"
            title="Sua lista já está completa"
            description="Não temos mais ideias para este momento. Que tal buscar produtos específicos?"
          />
        </div>
      ) : categories ? (
        <div className="mt-6 flex flex-col gap-6">
          {categories.map((c, ci) => (
            <section key={c.name} aria-label={c.name} className="flex flex-col gap-2">
              <div>
                <h3 className="font-display text-lg font-semibold">
                  <span aria-hidden>{c.emoji}</span> {c.name}
                </h3>
                <p className="text-ink-muted text-sm">{c.reason}</p>
              </div>
              <ul className="flex flex-col gap-1.5">
                {c.desires.map((d, di) => {
                  const key = keyOf(ci, di);
                  return (
                    <li key={key}>
                      <label className="border-line hover:bg-primary-soft/40 flex cursor-pointer items-center gap-3 rounded-2xl border p-3">
                        <input
                          type="checkbox"
                          className="accent-primary size-5"
                          checked={picked.has(key)}
                          onChange={() => toggle(key)}
                        />
                        <span aria-hidden className="text-xl">
                          {d.emoji}
                        </span>
                        <span className="flex-1 text-sm font-medium">
                          {d.title}
                          {d.quantity > 1 ? (
                            <span className="text-ink-muted font-normal"> · {d.quantity} un.</span>
                          ) : null}
                        </span>
                        <Badge tone={d.importance === "ESSENTIAL" ? "primary" : "neutral"}>
                          {IMPORTANCE_LABEL[d.importance]}
                        </Badge>
                      </label>
                    </li>
                  );
                })}
              </ul>
            </section>
          ))}
          <div className="bg-surface border-line sticky -bottom-6 -mx-6 -mb-6 border-t p-4">
            <Button
              block
              size="lg"
              disabled={picked.size === 0}
              loading={applying}
              onClick={() => void apply()}
            >
              Adicionar {picked.size} {picked.size === 1 ? "item" : "itens"} à lista
            </Button>
          </div>
        </div>
      ) : null}
    </Sheet>
  );
}
