"use client";

import { useState } from "react";
import { Controller, useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Link2, Pencil, Plus, Trash2 } from "lucide-react";
import { z } from "zod";
import {
  CuratedProduct,
  LinkPreview,
  type CuratedMerchant,
} from "@listou/types";
import {
  Badge,
  Button,
  EmptyState,
  Field,
  Input,
  MarketplaceBadge,
  Select,
  Sheet,
  Switch,
  useToast,
} from "@listou/ui";
import { api, apiVoid, applyFieldErrors, ApiError } from "@/lib/api";
import { ItemVisual } from "./item-visual";

const CATEGORY_SUGGESTIONS = [
  "Cozinha",
  "Mesa posta",
  "Quarto",
  "Banheiro",
  "Sala",
  "Limpeza",
  "Organização",
  "Bebê",
  "Viagem",
];

const httpUrl = (message: string) =>
  z
    .string()
    .trim()
    .refine((v) => /^https?:\/\/\S+$/.test(v), message);

const schema = z.object({
  title: z.string().trim().min(3, "Dê um nome ao produto").max(200),
  brand: z.string().trim().max(80),
  category: z.string().trim().max(60),
  keywords: z.string().trim().max(500),
  imageUrl: z
    .string()
    .trim()
    .refine((v) => v === "" || /^https?:\/\/\S+$/.test(v), "Endereço de imagem inválido"),
  active: z.boolean(),
  offers: z
    .array(z.object({ merchant: z.string().min(1, "Escolha a loja"), url: httpUrl("Link inválido") }))
    .min(1, "Informe ao menos um link de afiliado"),
});
type FormValues = z.infer<typeof schema>;

function emptyValues(merchants: CuratedMerchant[]): FormValues {
  return {
    title: "",
    brand: "",
    category: "",
    keywords: "",
    imageUrl: "",
    active: true,
    offers: [{ merchant: merchants[0]?.code ?? "", url: "" }],
  };
}

function ProductForm({
  product,
  merchants,
  onSaved,
}: {
  product: CuratedProduct | null;
  merchants: CuratedMerchant[];
  onSaved: (p: CuratedProduct) => void;
}) {
  const toast = useToast();
  const [reading, setReading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const {
    register,
    control,
    handleSubmit,
    setError,
    setValue,
    getValues,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: product
      ? {
          title: product.title,
          brand: product.brand,
          category: product.category,
          keywords: product.keywords,
          imageUrl: product.imageUrl,
          active: product.active,
          offers: product.offers,
        }
      : emptyValues(merchants),
  });
  const { fields, append, remove } = useFieldArray({ control, name: "offers" });

  /** Fill title and image from the page behind the first affiliate link. */
  async function readLink() {
    const url = getValues("offers.0.url").trim();
    if (!/^https?:\/\/\S+$/.test(url)) {
      setError("offers.0.url", { message: "Cole o link de afiliado primeiro" });
      return;
    }
    setReading(true);
    try {
      const { preview } = await api("/products/link-preview", z.object({ preview: LinkPreview }), {
        method: "POST",
        body: { url },
      });
      if (preview.readable && !getValues("title").trim()) setValue("title", preview.title);
      if (preview.imageUrl && !getValues("imageUrl").trim()) setValue("imageUrl", preview.imageUrl);
      toast(
        preview.readable
          ? "Título e imagem preenchidos. Confira os campos."
          : "A loja não deixou ler a página. Preencha à mão.",
        preview.readable ? "success" : "neutral",
      );
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível ler o link agora.", "error");
    } finally {
      setReading(false);
    }
  }

  const onSubmit = handleSubmit(async (v) => {
    setFormError(null);
    try {
      const body = { ...v, offers: v.offers.map((o) => ({ merchant: o.merchant, url: o.url })) };
      const res = product
        ? await api(`/admin/curated-products/${product.id}`, z.object({ product: CuratedProduct }), {
            method: "PUT",
            body,
          })
        : await api("/admin/curated-products", z.object({ product: CuratedProduct }), {
            method: "POST",
            body,
          });
      onSaved(res.product);
    } catch (err) {
      setFormError(applyFieldErrors(err, setError as never));
    }
  });

  return (
    <form noValidate className="flex flex-col gap-5" onSubmit={onSubmit}>
      <div className="flex flex-col gap-3">
        <p className="text-ink text-sm font-semibold">Links de afiliado</p>
        {fields.map((f, i) => (
          <div key={f.id} className="flex flex-col gap-2">
            <div className="grid grid-cols-[minmax(0,9rem)_minmax(0,1fr)] gap-2">
              <Select aria-label="Loja" {...register(`offers.${i}.merchant`)}>
                {merchants.map((m) => (
                  <option key={m.code} value={m.code}>
                    {m.name}
                  </option>
                ))}
              </Select>
              <Input
                aria-label="Link de afiliado"
                inputMode="url"
                autoComplete="off"
                placeholder="https://meli.la/…"
                aria-invalid={!!errors.offers?.[i]?.url}
                {...register(`offers.${i}.url`)}
              />
            </div>
            {errors.offers?.[i]?.url || errors.offers?.[i]?.merchant ? (
              <p role="alert" className="text-danger text-sm">
                {errors.offers?.[i]?.url?.message ?? errors.offers?.[i]?.merchant?.message}
              </p>
            ) : null}
            <div className="flex gap-2">
              {i === 0 ? (
                <Button type="button" variant="soft" size="sm" loading={reading} onClick={readLink}>
                  <Link2 className="size-4" /> Ler link e preencher
                </Button>
              ) : null}
              {fields.length > 1 ? (
                <Button type="button" variant="ghost" size="sm" onClick={() => remove(i)}>
                  Remover loja
                </Button>
              ) : null}
            </div>
          </div>
        ))}
        {typeof errors.offers?.message === "string" ? (
          <p role="alert" className="text-danger text-sm">
            {errors.offers.message}
          </p>
        ) : null}
        {errors.offers?.root?.message ? (
          <p role="alert" className="text-danger text-sm">
            {errors.offers.root.message}
          </p>
        ) : null}
        {fields.length < merchants.length ? (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="self-start"
            onClick={() => {
              const used = new Set(getValues("offers").map((o) => o.merchant));
              append({ merchant: merchants.find((m) => !used.has(m.code))?.code ?? "", url: "" });
            }}
          >
            <Plus className="size-4" /> Adicionar outra loja
          </Button>
        ) : null}
      </div>

      <Field label="Nome do produto" htmlFor="cp-title" error={errors.title?.message}>
        <Input id="cp-title" aria-invalid={!!errors.title} {...register("title")} />
      </Field>
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Marca" htmlFor="cp-brand" optional error={errors.brand?.message}>
          <Input id="cp-brand" {...register("brand")} />
        </Field>
        <Field label="Categoria" htmlFor="cp-category" optional error={errors.category?.message}>
          <Input id="cp-category" list="cp-categories" {...register("category")} />
          <datalist id="cp-categories">
            {CATEGORY_SUGGESTIONS.map((c) => (
              <option key={c} value={c} />
            ))}
          </datalist>
        </Field>
      </div>
      <Field
        label="Imagem (endereço)"
        htmlFor="cp-image"
        optional
        hint="Use uma imagem que você tem direito de usar."
        error={errors.imageUrl?.message}
      >
        <Input id="cp-image" inputMode="url" {...register("imageUrl")} />
      </Field>
      <Field
        label="Palavras de busca"
        htmlFor="cp-keywords"
        optional
        hint="O que as pessoas digitariam: air fryer, fritadeira, sem óleo…"
        error={errors.keywords?.message}
      >
        <Input id="cp-keywords" {...register("keywords")} />
      </Field>
      <Controller
        control={control}
        name="active"
        render={({ field }) => (
          <Switch
            id="cp-active"
            checked={field.value}
            onCheckedChange={field.onChange}
            label="Aparece na busca"
            description="Desativado, some da busca mas os itens já adicionados a listas continuam funcionando."
          />
        )}
      />
      {formError ? (
        <p role="alert" className="text-danger text-sm">
          {formError}
        </p>
      ) : null}
      <Button type="submit" size="lg" loading={isSubmitting}>
        {product ? "Salvar alterações" : "Adicionar ao catálogo"}
      </Button>
    </form>
  );
}

export function CuratedAdmin({
  initialProducts,
  merchants,
}: {
  initialProducts: CuratedProduct[];
  merchants: CuratedMerchant[];
}) {
  const toast = useToast();
  const [products, setProducts] = useState(initialProducts);
  const [editing, setEditing] = useState<CuratedProduct | "new" | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const nameOf = (code: string) => merchants.find((m) => m.code === code)?.name ?? code;

  function saved(p: CuratedProduct) {
    setProducts((cur) =>
      cur.some((x) => x.id === p.id) ? cur.map((x) => (x.id === p.id ? p : x)) : [p, ...cur],
    );
    setEditing(null);
    toast(`“${p.title}” salvo`, "success");
  }

  async function toggle(p: CuratedProduct, active: boolean) {
    setBusyId(p.id);
    try {
      const { product } = await api(
        `/admin/curated-products/${p.id}`,
        z.object({ product: CuratedProduct }),
        { method: "PUT", body: { ...p, active } },
      );
      setProducts((cur) => cur.map((x) => (x.id === product.id ? product : x)));
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível atualizar.", "error");
    } finally {
      setBusyId(null);
    }
  }

  async function remove(p: CuratedProduct) {
    if (!window.confirm(`Remover “${p.title}” do catálogo? Itens já adicionados a listas continuam funcionando.`))
      return;
    setBusyId(p.id);
    try {
      await apiVoid(`/admin/curated-products/${p.id}`, { method: "DELETE" });
      setProducts((cur) => cur.filter((x) => x.id !== p.id));
      toast("Produto removido", "success");
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível remover.", "error");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="font-display text-ink text-3xl font-semibold">Produtos curados</h1>
          <p className="text-ink-muted max-w-xl text-sm">
            Quem escolhe um desses produtos na busca sai da lista pelo seu link de afiliado, exatamente
            como cadastrado.
          </p>
        </div>
        <Button onClick={() => setEditing("new")} disabled={merchants.length === 0}>
          <Plus className="size-4" /> Novo produto
        </Button>
      </div>

      {merchants.length === 0 ? (
        <p role="alert" className="text-danger text-sm">
          Nenhuma loja usa o catálogo curado. Configure um provider CURATED para uma loja.
        </p>
      ) : null}

      {products.length === 0 ? (
        <EmptyState
          icon="gift"
          title="Nenhum produto curado ainda"
          description="Cole um link de afiliado e a gente preenche título e imagem."
          action={<Button onClick={() => setEditing("new")}>Adicionar o primeiro</Button>}
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {products.map((p) => (
            <li
              key={p.id}
              className="rounded-card bg-surface shadow-hairline flex flex-col gap-4 p-4 sm:flex-row sm:items-center"
            >
              <div className="rounded-control bg-canvas size-20 shrink-0 overflow-hidden">
                <ItemVisual imageUrl={p.imageUrl || null} emoji="gift" alt={p.title} />
              </div>
              <div className="flex min-w-0 flex-1 flex-col gap-1.5">
                <p className="text-ink line-clamp-2 font-semibold">{p.title}</p>
                <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                  {p.category ? <Badge tone="neutral">{p.category}</Badge> : null}
                  {p.offers.map((o) => (
                    <MarketplaceBadge key={o.merchant} code={o.merchant} name={nameOf(o.merchant)} />
                  ))}
                  {!p.active ? <Badge tone="warning">Fora da busca</Badge> : null}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="soft"
                  size="sm"
                  disabled={busyId === p.id}
                  onClick={() => toggle(p, !p.active)}
                >
                  {p.active ? "Desativar" : "Ativar"}
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Editar ${p.title}`}
                  onClick={() => setEditing(p)}
                >
                  <Pencil className="size-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Remover ${p.title}`}
                  disabled={busyId === p.id}
                  onClick={() => remove(p)}
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
            </li>
          ))}
        </ul>
      )}

      <Sheet
        open={editing !== null}
        onOpenChange={(o) => !o && setEditing(null)}
        title={editing === "new" ? "Novo produto" : "Editar produto"}
        description="O link de afiliado é usado exatamente como você colar."
      >
        {editing ? (
          <ProductForm
            key={editing === "new" ? "new" : editing.id}
            product={editing === "new" ? null : editing}
            merchants={merchants}
            onSaved={saved}
          />
        ) : null}
      </Sheet>
    </div>
  );
}
