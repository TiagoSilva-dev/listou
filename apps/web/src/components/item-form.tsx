"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { formatPrice, parsePriceToCents, type Category, type ListItem } from "@listou/types";
import { Button, Field, Input, Select, Textarea } from "@listou/ui";

const schema = z.object({
  title: z.string().trim().min(1, "Dê um nome ao item").max(160),
  description: z.string().trim().max(1000),
  desiredQuantity: z
    .string()
    .regex(/^\d+$/, "Informe um número")
    .refine((v) => Number(v) >= 1 && Number(v) <= 999, "Entre 1 e 999"),
  price: z
    .string()
    .refine((v) => v.trim() === "" || parsePriceToCents(v) !== null, "Valor inválido"),
  externalUrl: z
    .string()
    .trim()
    .refine((v) => v === "" || /^https?:\/\/\S+$/.test(v), "Link inválido"),
  categoryId: z.string(),
  notes: z.string().trim().max(1000),
});
type FormValues = z.infer<typeof schema>;

export interface ItemPayload {
  title: string;
  description: string;
  desiredQuantity: number;
  priceReferenceCents?: number;
  externalUrl: string;
  categoryId: string;
  notes: string;
}

export function ItemForm({
  categories,
  item,
  submitLabel,
  onSubmit,
  onDelete,
  showLink = true,
}: {
  categories: Category[];
  item?: ListItem;
  submitLabel: string;
  onSubmit: (payload: ItemPayload) => Promise<void>;
  onDelete?: () => Promise<void>;
  showLink?: boolean;
}) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: item?.title ?? "",
      description: item?.description ?? "",
      desiredQuantity: String(item?.desiredQuantity ?? 1),
      price:
        item?.priceReferenceCents != null
          ? (formatPrice(item.priceReferenceCents)?.replace(/[^\d,.]/g, "") ?? "")
          : "",
      externalUrl: item?.externalUrl ?? "",
      categoryId: item?.categoryId ?? "",
      notes: item?.notes ?? "",
    },
  });

  return (
    <form
      noValidate
      className="flex flex-col gap-5"
      onSubmit={handleSubmit(async (v) => {
        const cents = v.price.trim() === "" ? undefined : (parsePriceToCents(v.price) ?? undefined);
        await onSubmit({
          title: v.title,
          description: v.description,
          desiredQuantity: Number(v.desiredQuantity),
          priceReferenceCents: cents,
          externalUrl: v.externalUrl,
          categoryId: v.categoryId,
          notes: v.notes,
        });
      })}
    >
      <Field label="Nome do item" htmlFor="item-title" error={errors.title?.message}>
        <Input
          id="item-title"
          autoFocus
          placeholder="Ex.: Máquina de café"
          aria-invalid={!!errors.title}
          {...register("title")}
        />
      </Field>
      <div className="grid grid-cols-2 gap-4">
        <Field label="Quantidade" htmlFor="item-qty" error={errors.desiredQuantity?.message}>
          <Input
            id="item-qty"
            inputMode="numeric"
            aria-invalid={!!errors.desiredQuantity}
            {...register("desiredQuantity")}
          />
        </Field>
        <Field label="Valor estimado" htmlFor="item-price" optional error={errors.price?.message}>
          <Input
            id="item-price"
            inputMode="decimal"
            placeholder="R$ 0,00"
            aria-invalid={!!errors.price}
            {...register("price")}
          />
        </Field>
      </div>
      {categories.length > 0 ? (
        <Field label="Categoria" htmlFor="item-category" optional>
          <Select id="item-category" {...register("categoryId")}>
            <option value="">Sem categoria</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </Select>
        </Field>
      ) : null}
      <Field label="Descrição" htmlFor="item-desc" optional>
        <Textarea
          id="item-desc"
          className="min-h-20"
          placeholder="Cor, tamanho, modelo preferido…"
          {...register("description")}
        />
      </Field>
      {showLink ? (
        <Field
          label="Link de referência"
          htmlFor="item-link"
          optional
          hint="Uma loja de sua preferência, se quiser."
          error={errors.externalUrl?.message}
        >
          <Input
            id="item-link"
            type="url"
            inputMode="url"
            placeholder="https://"
            aria-invalid={!!errors.externalUrl}
            {...register("externalUrl")}
          />
        </Field>
      ) : null}
      {item ? (
        <Field label="Observação privada" htmlFor="item-notes" optional hint="Só você vê.">
          <Textarea id="item-notes" className="min-h-16" {...register("notes")} />
        </Field>
      ) : null}
      <div className="flex flex-col gap-2 sm:flex-row-reverse">
        <Button type="submit" size="lg" loading={isSubmitting} className="sm:flex-1">
          {submitLabel}
        </Button>
        {onDelete ? (
          <Button type="button" variant="danger" size="lg" onClick={onDelete}>
            Remover
          </Button>
        ) : null}
      </div>
    </form>
  );
}
