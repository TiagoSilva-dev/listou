"use client";

import { useEffect, useState } from "react";
import { Check, ExternalLink, Gift, Minus, Plus, ShieldCheck } from "lucide-react";
import { z } from "zod";
import { formatPrice, Reservation, type PublicItem } from "@listou/types";
import {
  Badge,
  Button,
  buttonVariants,
  Field,
  Input,
  MarketplaceBadge,
  Price,
  Sheet,
  Textarea,
  useToast,
} from "@listou/ui";
import { api, ApiError, apiVoid } from "@/lib/api";
import {
  addReservation,
  removeReservation,
  updateReservation,
  type StoredReservation,
} from "@/lib/reservations-store";
import { ItemVisual } from "./item-visual";

type Mode = "choose" | "reserve" | "purchased" | "done";

export interface GiftDrawerProps {
  slug: string;
  item: PublicItem | null;
  allowReservations: boolean;
  mine: StoredReservation[];
  onClose: () => void;
  /** Called after any server-side change so the page can refetch availability. */
  onChanged: () => void;
}

export function GiftDrawer({
  slug,
  item,
  allowReservations,
  mine,
  onClose,
  onChanged,
}: GiftDrawerProps) {
  return (
    <Sheet
      open={item !== null}
      onOpenChange={(o) => !o && onClose()}
      title={item?.title ?? "Presentear"}
      hideTitle
      className="sm:max-w-xl"
    >
      {item ? (
        <DrawerBody
          key={item.id}
          slug={slug}
          item={item}
          allowReservations={allowReservations}
          mine={mine}
          onClose={onClose}
          onChanged={onChanged}
        />
      ) : null}
    </Sheet>
  );
}

function DrawerBody({
  slug,
  item,
  allowReservations,
  mine,
  onClose,
  onChanged,
}: Omit<GiftDrawerProps, "item"> & { item: PublicItem }) {
  const toast = useToast();
  const myReservations = mine.filter((r) => r.itemId === item.id);
  const [mode, setMode] = useState<Mode>("choose");
  const [name, setName] = useState("");
  const [contact, setContact] = useState("");
  const [message, setMessage] = useState("");
  const [qty, setQty] = useState(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [clickedOffer, setClickedOffer] = useState(false);
  const [finished, setFinished] = useState<"RESERVATION" | "PURCHASE" | null>(null);

  const available = item.availableQuantity;
  const gone = available <= 0;
  const offers = item.offers;
  const hero = item.imageUrl ?? item.product?.imageUrl;

  useEffect(() => {
    if (offers.length > 0) {
      void apiVoid("/analytics/track", {
        method: "POST",
        body: { name: "OFFER_VIEWED", slug, itemId: item.id },
      }).catch(() => undefined);
    }
  }, [offers.length, slug, item.id]);

  async function submit(kind: "RESERVATION" | "PURCHASE") {
    if (name.trim().length === 0) {
      setError("Como devemos te chamar?");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const { reservation } = await api(
        `/public/lists/${slug}/items/${item.id}/reservations`,
        z.object({ reservation: Reservation }),
        {
          method: "POST",
          body: {
            kind,
            quantity: qty,
            guestName: name.trim(),
            guestContact: contact.trim() || undefined,
            message: message.trim() || undefined,
          },
        },
      );
      if (reservation.manageToken) {
        addReservation(slug, {
          id: reservation.id,
          token: reservation.manageToken,
          itemId: item.id,
          itemTitle: item.title,
          kind,
          quantity: qty,
          createdAt: reservation.createdAt,
        });
      }
      setFinished(kind);
      setMode("done");
      onChanged();
    } catch (err) {
      if (err instanceof ApiError && err.code === "ITEM_NOT_AVAILABLE") onChanged();
      setError(
        err instanceof ApiError
          ? err.fields
            ? (Object.values(err.fields)[0] ?? err.message)
            : err.message
          : "Não foi possível concluir. Tente novamente.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function cancel(r: StoredReservation) {
    setBusy(true);
    try {
      await apiVoid(`/reservations/${r.id}`, {
        method: "DELETE",
        headers: { "X-Reservation-Token": r.token },
      });
      removeReservation(slug, r.id);
      toast("Reserva cancelada. O item voltou a ficar disponível.");
      onChanged();
      onClose();
    } catch (err) {
      if (
        err instanceof ApiError &&
        (err.code === "RESERVATION_NOT_FOUND" || err.code === "RESERVATION_NOT_ACTIVE")
      ) {
        removeReservation(slug, r.id);
        onChanged();
      }
      toast(err instanceof ApiError ? err.message : "Não foi possível cancelar.", "error");
    } finally {
      setBusy(false);
    }
  }

  async function confirm(r: StoredReservation) {
    setBusy(true);
    try {
      await apiVoid(`/reservations/${r.id}/confirm`, {
        method: "POST",
        headers: { "X-Reservation-Token": r.token },
      });
      updateReservation(slug, r.id, { kind: "PURCHASE" });
      toast("Obrigado! Marcamos como comprado.", "success");
      onChanged();
      onClose();
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível confirmar.", "error");
    } finally {
      setBusy(false);
    }
  }

  if (mode === "done" && finished) {
    return (
      <div className="animate-pop flex flex-col items-center gap-5 py-6 text-center">
        <span className="bg-success-soft text-success grid size-16 place-items-center rounded-full">
          <Check className="size-8" />
        </span>
        <div className="flex flex-col gap-2">
          <h2 className="font-display text-2xl font-semibold">
            {finished === "RESERVATION" ? "Presente reservado!" : "Obrigado pelo presente!"}
          </h2>
          <p className="text-ink-muted max-w-sm">
            {finished === "RESERVATION"
              ? "Guardamos este item para você por alguns dias para que ninguém repita. Você pode cancelar ou confirmar a compra por aqui quando quiser."
              : "Marcamos este item como comprado para que os outros convidados não repitam."}
          </p>
        </div>
        <Button size="lg" onClick={onClose}>
          Voltar para a lista
        </Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex gap-4">
        <div className="rounded-card bg-canvas-deep size-28 shrink-0 overflow-hidden sm:size-32">
          <ItemVisual imageUrl={hero} emoji={item.emoji ?? item.product?.emoji} alt={item.title} />
        </div>
        <div className="flex min-w-0 flex-col justify-center gap-2 pr-8">
          <h2 className="font-display text-balance text-xl font-semibold leading-snug sm:text-2xl">
            {item.title}
          </h2>
          {item.product?.brand ? (
            <p className="text-ink-muted text-sm">{item.product.brand}</p>
          ) : null}
          <div className="flex flex-wrap items-center gap-2">
            {item.desiredQuantity > 1 ? (
              <Badge tone="primary">
                {available} de {item.desiredQuantity} disponíveis
              </Badge>
            ) : null}
            {item.product?.demo ? <Badge>demonstração</Badge> : null}
          </div>
        </div>
      </div>
      {item.description ? (
        <p className="text-ink-soft leading-relaxed">{item.description}</p>
      ) : null}

      {myReservations.map((r) => (
        <div key={r.id} className="rounded-card bg-primary-soft flex flex-col gap-3 p-4">
          <p className="text-primary text-sm font-semibold">
            {r.kind === "PURCHASE"
              ? "Você marcou este presente como comprado"
              : `Você reservou ${r.quantity} ${r.quantity === 1 ? "unidade" : "unidades"} deste presente`}
          </p>
          <div className="flex flex-wrap gap-2">
            {r.kind === "RESERVATION" ? (
              <Button size="sm" onClick={() => confirm(r)} loading={busy}>
                <Check className="size-4" /> Já comprei
              </Button>
            ) : null}
            <Button size="sm" variant="secondary" onClick={() => cancel(r)} disabled={busy}>
              Cancelar {r.kind === "PURCHASE" ? "compra" : "reserva"}
            </Button>
          </div>
        </div>
      ))}

      {gone && myReservations.length === 0 ? (
        <div className="rounded-card bg-canvas-deep p-5 text-center">
          <p className="font-semibold">Este presente já foi escolhido</p>
          <p className="text-ink-muted text-sm">Que tal olhar outro item da lista?</p>
        </div>
      ) : null}

      {!gone && mode === "choose" ? (
        <div className="flex flex-col gap-5">
          {offers.length > 0 ? (
            <section className="flex flex-col gap-3" aria-label="Onde comprar">
              <h3 className="text-ink-muted text-sm font-semibold">Onde comprar</h3>
              <ul className="flex flex-col gap-2">
                {offers.map((o, i) => (
                  <li key={o.id} className="rounded-control bg-canvas flex items-center gap-3 p-3">
                    <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                      <MarketplaceBadge
                        code={o.merchant.code}
                        name={o.merchant.name}
                        className="text-ink text-sm font-semibold"
                      />
                      <span className="flex flex-wrap items-center gap-2">
                        <Price
                          cents={o.priceCents}
                          originalCents={o.originalPriceCents}
                          currency={o.currency}
                          className="text-sm"
                        />
                        {i === 0 && offers.length > 1 ? (
                          <Badge tone="success">Menor preço</Badge>
                        ) : null}
                      </span>
                    </div>
                    <a
                      href={`${o.goUrl}?item=${item.id}&utm_source=list&utm_medium=registry&utm_campaign=${encodeURIComponent(slug)}`}
                      target="_blank"
                      rel="sponsored noopener noreferrer"
                      onClick={() => setClickedOffer(true)}
                      className={buttonVariants({
                        size: "sm",
                        variant: i === 0 ? "primary" : "secondary",
                      })}
                    >
                      Comprar <ExternalLink className="size-3.5" />
                    </a>
                  </li>
                ))}
              </ul>
              <p className="text-ink-muted flex items-start gap-2 text-xs leading-relaxed">
                <ShieldCheck className="mt-0.5 size-3.5 shrink-0" />
                Você será levado à loja. Se comprar por aqui, a loja pode pagar uma pequena comissão
                ao Listou — sem custo extra para você.
              </p>
            </section>
          ) : item.externalUrl ? (
            <a
              href={item.externalUrl}
              target="_blank"
              rel="noopener noreferrer nofollow"
              className={buttonVariants({ variant: "secondary", size: "lg", block: true })}
            >
              Ver produto na loja <ExternalLink className="size-4" />
            </a>
          ) : (
            <p className="rounded-control bg-canvas text-ink-muted p-4 text-sm">
              {item.priceReferenceCents != null
                ? `Valor de referência: ${formatPrice(item.priceReferenceCents)}. `
                : ""}
              Escolha onde comprar à sua maneira e depois avise aqui para ninguém repetir.
            </p>
          )}

          {clickedOffer ? (
            <div className="animate-fade-up rounded-card bg-sun-soft flex flex-col gap-3 p-4">
              <p className="text-sm font-semibold">Concluiu a compra na loja?</p>
              <p className="text-ink-soft text-sm">
                Avise por aqui para os outros convidados não comprarem o mesmo presente.
              </p>
              <Button size="sm" className="w-fit" onClick={() => setMode("purchased")}>
                <Gift className="size-4" /> Marcar como comprado
              </Button>
            </div>
          ) : null}

          <div className="border-line flex flex-col gap-2 border-t pt-5">
            {allowReservations ? (
              <Button variant="soft" size="lg" onClick={() => setMode("reserve")}>
                Reservar este presente
              </Button>
            ) : null}
            {!clickedOffer && (offers.length > 0 || item.externalUrl || !allowReservations) ? (
              <Button variant="ghost" onClick={() => setMode("purchased")}>
                Já comprei em outro lugar
              </Button>
            ) : null}
          </div>
        </div>
      ) : null}

      {!gone && (mode === "reserve" || mode === "purchased") ? (
        <form
          className="animate-fade-up flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            void submit(mode === "reserve" ? "RESERVATION" : "PURCHASE");
          }}
        >
          <h3 className="font-display text-xl font-semibold">
            {mode === "reserve" ? "Reservar presente" : "Marcar como comprado"}
          </h3>
          <Field
            label="Seu nome"
            htmlFor="g-name"
            error={error && name.trim() === "" ? error : undefined}
          >
            <Input
              id="g-name"
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={80}
              autoComplete="name"
            />
          </Field>
          <Field
            label="E-mail ou telefone"
            htmlFor="g-contact"
            optional
            hint="Só o criador da lista pode ver, caso precise falar com você."
          >
            <Input
              id="g-contact"
              value={contact}
              onChange={(e) => setContact(e.target.value)}
              maxLength={120}
              autoComplete="email"
            />
          </Field>
          <Field label="Recado carinhoso" htmlFor="g-msg" optional>
            <Textarea
              id="g-msg"
              className="min-h-20"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              maxLength={500}
            />
          </Field>
          {available > 1 ? (
            <div className="rounded-control bg-canvas flex items-center justify-between p-3">
              <span className="text-sm font-semibold">Quantas unidades?</span>
              <div className="flex items-center gap-3">
                <Button
                  type="button"
                  variant="secondary"
                  size="icon"
                  className="size-9"
                  onClick={() => setQty((q) => Math.max(1, q - 1))}
                  aria-label="Menos"
                >
                  <Minus className="size-4" />
                </Button>
                <span className="w-6 text-center font-semibold tabular-nums" aria-live="polite">
                  {qty}
                </span>
                <Button
                  type="button"
                  variant="secondary"
                  size="icon"
                  className="size-9"
                  onClick={() => setQty((q) => Math.min(available, q + 1))}
                  aria-label="Mais"
                >
                  <Plus className="size-4" />
                </Button>
              </div>
            </div>
          ) : null}
          {mode === "reserve" ? (
            <p className="text-ink-muted text-xs">
              A reserva fica guardada por alguns dias. Depois disso o item volta a ficar disponível.
            </p>
          ) : null}
          {error && name.trim() !== "" ? (
            <p
              role="alert"
              className="rounded-control bg-danger-soft text-danger px-4 py-3 text-sm"
            >
              {error}
            </p>
          ) : null}
          <div className="flex flex-col gap-2 sm:flex-row-reverse">
            <Button type="submit" size="lg" loading={busy} className="sm:flex-1">
              {mode === "reserve" ? "Confirmar reserva" : "Confirmar"}
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="lg"
              onClick={() => {
                setMode("choose");
                setError(null);
              }}
            >
              Voltar
            </Button>
          </div>
        </form>
      ) : null}
    </div>
  );
}
