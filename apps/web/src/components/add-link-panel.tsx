"use client";

import { useState } from "react";
import { Link2 } from "lucide-react";
import { z } from "zod";
import { LinkPreview, ListItem } from "@listou/types";
import { Button, Field, Input, MarketplaceBadge, useToast } from "@listou/ui";
import { api, ApiError } from "@/lib/api";
import { ItemVisual } from "./item-visual";

/** Paste a product link: we read its public title and image and add it as a wish. */
export function AddLinkPanel({
  listId,
  onAdded,
}: {
  listId: string;
  onAdded: (item: ListItem) => void;
}) {
  const toast = useToast();
  const [url, setUrl] = useState("");
  const [preview, setPreview] = useState<LinkPreview | null>(null);
  const [title, setTitle] = useState("");
  const [checking, setChecking] = useState(false);
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function check() {
    const value = url.trim();
    if (!value) return;
    setChecking(true);
    setError(null);
    setPreview(null);
    try {
      const res = await api("/products/link-preview", z.object({ preview: LinkPreview }), {
        method: "POST",
        body: { url: value },
      });
      setPreview(res.preview);
      setTitle(res.preview.title);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Não foi possível ler o link agora.");
    } finally {
      setChecking(false);
    }
  }

  async function add() {
    if (!preview) return;
    setAdding(true);
    try {
      const imported = await api(
        "/products/import-link",
        z.object({ product: z.object({ id: z.string() }) }),
        { method: "POST", body: { url: preview.url, title: title.trim() } },
      );
      const { item } = await api(`/lists/${listId}/items`, z.object({ item: ListItem }), {
        method: "POST",
        body: { productId: imported.product.id, desiredQuantity: 1 },
      });
      onAdded(item);
      toast(`“${item.title}” foi adicionado à lista`, "success");
      setUrl("");
      setPreview(null);
      setTitle("");
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível adicionar.", "error");
    } finally {
      setAdding(false);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <form
        className="flex flex-col gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          void check();
        }}
      >
        <Field
          label="Link do produto"
          htmlFor="link-url"
          hint="Cole o endereço da página do produto em qualquer loja."
          error={error ?? undefined}
        >
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Link2
                aria-hidden
                className="text-ink-muted pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2"
              />
              <Input
                id="link-url"
                autoFocus
                inputMode="url"
                autoComplete="off"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://loja.com.br/produto"
                className="pl-11"
                aria-invalid={!!error}
              />
            </div>
            <Button type="submit" variant="soft" loading={checking} disabled={!url.trim()}>
              Ler link
            </Button>
          </div>
        </Field>
      </form>

      {preview ? (
        <div className="rounded-card bg-canvas flex flex-col gap-4 p-4">
          <div className="flex items-center gap-4">
            <div className="rounded-control bg-surface size-20 shrink-0 overflow-hidden">
              <ItemVisual imageUrl={preview.imageUrl} emoji="link" alt={title || "Produto"} />
            </div>
            <div className="flex min-w-0 flex-1 flex-col gap-1">
              <MarketplaceBadge code="LINK" name={preview.storeName} />
              <p className="text-ink-muted text-xs">
                Preço e disponibilidade não são lidos do link. Confira na loja.
              </p>
            </div>
          </div>
          <Field
            label="Nome do item"
            htmlFor="link-title"
            hint={
              preview.readable
                ? "Você pode ajustar o nome antes de adicionar."
                : "Essa loja não deixou a gente ler a página. Digite o nome do item."
            }
          >
            <Input
              id="link-title"
              value={title}
              maxLength={200}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Ex.: Máquina de café"
            />
          </Field>
          <Button onClick={add} loading={adding} disabled={!title.trim()}>
            Adicionar à lista
          </Button>
        </div>
      ) : null}
    </div>
  );
}
