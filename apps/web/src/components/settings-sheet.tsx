"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { z } from "zod";
import { Event } from "@listou/types";
import { Button, cn, Field, Input, Sheet, Switch, Textarea, useToast } from "@listou/ui";
import { api, ApiError, apiVoid } from "@/lib/api";
import { THEMES, type ThemeName } from "./cover-art";

const THEME_LABEL: Record<ThemeName, string> = {
  blush: "Blush",
  lavender: "Lavanda",
  sage: "Sálvia",
  sun: "Sol",
  night: "Noite",
};

export function SettingsSheet({
  open,
  onOpenChange,
  event,
  onSaved,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  event: Event;
  onSaved: (event: Event) => void;
}) {
  const router = useRouter();
  const toast = useToast();
  const [title, setTitle] = useState(event.title);
  const [hostNames, setHostNames] = useState(event.hostNames ?? "");
  const [eventDate, setEventDate] = useState(event.eventDate ?? "");
  const [location, setLocation] = useState(event.location ?? "");
  const [description, setDescription] = useState(event.description ?? "");
  const [slug, setSlug] = useState(event.slug);
  const [theme, setTheme] = useState(event.theme);
  const [unlisted, setUnlisted] = useState(event.visibility !== "PUBLIC");
  const [surprise, setSurprise] = useState(event.surpriseMode);
  const [busy, setBusy] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  async function save() {
    setBusy(true);
    setErrors({});
    try {
      const res = await api(`/events/${event.id}`, z.object({ event: Event }), {
        method: "PATCH",
        body: {
          title,
          hostNames,
          eventDate,
          location,
          description,
          slug,
          theme,
          visibility: unlisted ? "UNLISTED" : "PUBLIC",
          surpriseMode: surprise,
        },
      });
      onSaved(res.event);
      toast("Configurações salvas", "success");
      onOpenChange(false);
      if (res.event.slug !== event.slug) router.refresh();
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === "SLUG_TAKEN") setErrors({ slug: err.message });
        else if (err.fields) setErrors(err.fields);
        else toast(err.message, "error");
      } else toast("Não foi possível salvar.", "error");
    } finally {
      setBusy(false);
    }
  }

  async function remove() {
    if (!window.confirm("Excluir esta lista? O link público deixará de funcionar.")) return;
    try {
      await apiVoid(`/events/${event.id}`, { method: "DELETE" });
      router.replace("/dashboard");
      router.refresh();
    } catch (err) {
      toast(err instanceof ApiError ? err.message : "Não foi possível excluir.", "error");
    }
  }

  return (
    <Sheet
      open={open}
      onOpenChange={onOpenChange}
      title="Configurações da lista"
      className="sm:max-w-xl"
    >
      <div className="flex flex-col gap-6">
        <Field label="Nome do evento" htmlFor="s-title" error={errors.title}>
          <Input id="s-title" value={title} onChange={(e) => setTitle(e.target.value)} />
        </Field>
        <Field label="Anfitriões" htmlFor="s-host" optional error={errors.hostNames}>
          <Input id="s-host" value={hostNames} onChange={(e) => setHostNames(e.target.value)} />
        </Field>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Data" htmlFor="s-date" optional error={errors.eventDate}>
            <Input
              id="s-date"
              type="date"
              value={eventDate}
              onChange={(e) => setEventDate(e.target.value)}
            />
          </Field>
          <Field label="Local" htmlFor="s-loc" optional error={errors.location}>
            <Input id="s-loc" value={location} onChange={(e) => setLocation(e.target.value)} />
          </Field>
        </div>
        <Field
          label="Mensagem para os convidados"
          htmlFor="s-desc"
          optional
          error={errors.description}
        >
          <Textarea
            id="s-desc"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </Field>
        <Field
          label="Endereço da lista"
          htmlFor="s-slug"
          hint={`listou.com.br/l/${slug || "seu-endereco"}`}
          error={errors.slug}
        >
          <Input
            id="s-slug"
            value={slug}
            onChange={(e) => setSlug(e.target.value.toLowerCase())}
            aria-invalid={!!errors.slug}
          />
        </Field>

        <fieldset className="flex flex-col gap-3">
          <legend className="mb-1 text-sm font-semibold">Visual da capa</legend>
          <div className="flex flex-wrap gap-3">
            {(Object.keys(THEMES) as ThemeName[]).map((name) => (
              <button
                key={name}
                type="button"
                onClick={() => setTheme(name)}
                aria-pressed={theme === name}
                aria-label={THEME_LABEL[name]}
                className={cn(
                  "ring-offset-surface focus-visible:shadow-focus size-12 rounded-full ring-offset-2 transition-all focus-visible:outline-none",
                  theme === name ? "ring-primary ring-2" : "ring-line ring-1",
                )}
                style={{
                  background: `linear-gradient(135deg, ${THEMES[name].a}, ${THEMES[name].b})`,
                }}
              />
            ))}
          </div>
        </fieldset>

        <div className="rounded-card bg-canvas flex flex-col gap-5 p-5">
          <Switch
            id="s-unlisted"
            checked={unlisted}
            onCheckedChange={setUnlisted}
            label="Só quem tem o link"
            description="Sua lista não aparece em buscas (Google). Desative para permitir indexação."
          />
          <Switch
            id="s-surprise"
            checked={surprise}
            onCheckedChange={setSurprise}
            label="Presente surpresa"
            description="Você vê só o total de presentes. Não mostramos quem deu o quê. Os convidados continuam vendo a disponibilidade."
          />
        </div>

        <div className="flex flex-col gap-3 sm:flex-row-reverse">
          <Button size="lg" onClick={save} loading={busy} className="sm:flex-1">
            Salvar
          </Button>
          <Button size="lg" variant="danger" onClick={remove}>
            Excluir lista
          </Button>
        </div>
      </div>
    </Sheet>
  );
}
