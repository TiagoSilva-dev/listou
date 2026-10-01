"use client";

import { z } from "zod";
import { ListItem, type Category } from "@listou/types";
import { Sheet, useToast } from "@listou/ui";
import { api, ApiError, apiVoid } from "@/lib/api";
import { ItemForm, type ItemPayload } from "./item-form";

export function EditItemSheet({
  item,
  categories,
  onClose,
  onSaved,
  onDeleted,
}: {
  item: ListItem | null;
  categories: Category[];
  onClose: () => void;
  onSaved: (item: ListItem) => void;
  onDeleted: (id: string) => void;
}) {
  const toast = useToast();

  async function save(p: ItemPayload) {
    if (!item) return;
    try {
      const { item: saved } = await api(`/items/${item.id}`, z.object({ item: ListItem }), {
        method: "PATCH",
        body: {
          title: p.title,
          description: p.description,
          desiredQuantity: p.desiredQuantity,
          priceReferenceCents: p.priceReferenceCents,
          externalUrl: p.externalUrl,
          categoryId: p.categoryId,
          notes: p.notes,
        },
      });
      onSaved(saved);
      toast("Item atualizado", "success");
      onClose();
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível salvar.", "error");
    }
  }

  async function remove() {
    if (!item || !window.confirm(`Remover “${item.title}” da lista?`)) return;
    try {
      await apiVoid(`/items/${item.id}`, { method: "DELETE" });
      onDeleted(item.id);
      toast("Item removido");
      onClose();
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível remover.", "error");
    }
  }

  return (
    <Sheet open={item !== null} onOpenChange={(o) => !o && onClose()} title="Editar item">
      {item ? (
        <ItemForm
          key={item.id}
          item={item}
          categories={categories}
          submitLabel="Salvar alterações"
          onSubmit={save}
          onDelete={remove}
        />
      ) : null}
    </Sheet>
  );
}
