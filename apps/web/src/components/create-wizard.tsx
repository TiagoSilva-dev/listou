"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { ArrowLeft, ArrowRight, Check, Sparkles, PenLine } from "lucide-react";
import { z } from "zod";
import { EVENT_TYPES, Event, EventType, eventTypeMeta } from "@listou/types";
import { Badge, Button, cn, Field, Input, Textarea } from "@listou/ui";
import { api, applyFieldErrors } from "@/lib/api";
import { Logo } from "./logo";

type Template = "EMPTY" | "SUGGESTED";

const titleHints: Partial<Record<EventType, string>> = {
  WEDDING: "Casamento de Tiago e Julia",
  HOUSEWARMING: "Chá de casa nova do Tiago e Julia",
  BABY_SHOWER: "Chá de bebê da Helena",
  BIRTHDAY: "Aniversário de 30 anos da Ana",
  GRADUATION: "Formatura da Mariana",
  TRAVEL: "Nossa viagem para a Itália",
  CHRISTMAS: "Natal da família Souza",
  WISHLIST: "Minha wishlist",
  CUSTOM: "Meu momento especial",
};

export function CreateWizard({ initialType }: { initialType: EventType | null }) {
  const router = useRouter();
  const [step, setStep] = useState(initialType ? 2 : 1);
  const [type, setType] = useState<EventType | null>(initialType);
  const [title, setTitle] = useState("");
  const [hostNames, setHostNames] = useState("");
  const [eventDate, setEventDate] = useState("");
  const [description, setDescription] = useState("");
  const [template, setTemplate] = useState<Template>("SUGGESTED");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [titleError, setTitleError] = useState<string | undefined>();

  const meta = type ? eventTypeMeta(type) : null;

  function next() {
    if (step === 2 && title.trim().length === 0) {
      setTitleError("Dê um nome ao momento");
      return;
    }
    setTitleError(undefined);
    setStep((s) => Math.min(3, s + 1));
  }

  async function create() {
    if (!type) return;
    setBusy(true);
    setError(null);
    try {
      const res = await api("/events", z.object({ event: Event }), {
        method: "POST",
        body: {
          type,
          title: title.trim(),
          hostNames: hostNames.trim() || undefined,
          eventDate: eventDate || undefined,
          description: description.trim() || undefined,
          template,
        },
      });
      router.push(`/dashboard/${res.event.id}?novo=1`);
    } catch (err) {
      setError(applyFieldErrors(err, (_name, e) => setError(e.message)));
      setBusy(false);
    }
  }

  return (
    <div className="flex min-h-dvh flex-col">
      <header className="container-page flex h-16 items-center justify-between">
        <Logo />
        <ol className="flex items-center gap-2" aria-label="Progresso">
          {[1, 2, 3].map((n) => (
            <li
              key={n}
              aria-current={n === step ? "step" : undefined}
              className={cn(
                "duration-(--duration-slow) ease-out-soft h-1.5 rounded-full transition-all",
                n === step
                  ? "bg-primary w-8"
                  : n < step
                    ? "bg-primary/50 w-4"
                    : "bg-line-strong w-4",
              )}
            />
          ))}
        </ol>
      </header>

      <main className="container-page flex flex-1 flex-col items-center justify-center py-8">
        <div key={step} className="animate-fade-up flex w-full max-w-2xl flex-col gap-8">
          {step === 1 ? (
            <>
              <Heading
                title="Qual é o momento?"
                subtitle="Escolha a ocasião e preparamos tudo para você."
              />
              <ul className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                {EVENT_TYPES.map((t) => (
                  <li key={t.type}>
                    <button
                      type="button"
                      onClick={() => {
                        setType(t.type);
                        setStep(2);
                      }}
                      className={cn(
                        "rounded-card bg-surface shadow-hairline duration-(--duration-base) ease-out-soft group flex h-full w-full flex-col items-start gap-5 p-5 text-left transition-all",
                        "hover:shadow-lift focus-visible:shadow-focus hover:-translate-y-0.5 focus-visible:outline-none",
                        type === t.type && "shadow-[0_0_0_2px_var(--color-primary)]",
                      )}
                    >
                      <span
                        aria-hidden
                        className="duration-(--duration-base) ease-spring text-4xl transition-transform group-hover:scale-110"
                      >
                        {t.emoji}
                      </span>
                      <span className="font-semibold">{t.label}</span>
                    </button>
                  </li>
                ))}
              </ul>
            </>
          ) : null}

          {step === 2 && meta ? (
            <>
              <Heading
                emoji={meta.emoji}
                title="Conte um pouco sobre o evento"
                subtitle="Só o essencial. Você pode mudar tudo depois."
              />
              <div className="flex flex-col gap-5">
                <Field label="Nome do evento" htmlFor="title" error={titleError}>
                  <Input
                    id="title"
                    autoFocus
                    value={title}
                    maxLength={120}
                    placeholder={titleHints[meta.type] ?? ""}
                    onChange={(e) => setTitle(e.target.value)}
                    aria-invalid={!!titleError}
                  />
                </Field>
                <Field
                  label="Quem são os anfitriões?"
                  htmlFor="hostNames"
                  optional
                  hint="Aparece em destaque na página da lista, ex.: Tiago & Julia."
                >
                  <Input
                    id="hostNames"
                    value={hostNames}
                    maxLength={120}
                    onChange={(e) => setHostNames(e.target.value)}
                  />
                </Field>
                <Field label="Data" htmlFor="date" optional>
                  <Input
                    id="date"
                    type="date"
                    value={eventDate}
                    onChange={(e) => setEventDate(e.target.value)}
                  />
                </Field>
                <Field label="Uma mensagem para os convidados" htmlFor="description" optional>
                  <Textarea
                    id="description"
                    value={description}
                    maxLength={2000}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="Conte por que esse momento é especial…"
                  />
                </Field>
              </div>
            </>
          ) : null}

          {step === 3 && meta ? (
            <>
              <Heading
                emoji="✨"
                title="Vamos montar sua lista"
                subtitle="Como você prefere começar?"
              />
              <div className="flex flex-col gap-3">
                <ChoiceCard
                  selected={template === "SUGGESTED"}
                  onSelect={() => setTemplate("SUGGESTED")}
                  icon={<Sparkles className="size-5" />}
                  title="Começar com sugestões"
                  text={
                    meta.suggestedCategories.length
                      ? `Categorias e itens prontos para ${meta.label.toLowerCase()}: ${meta.suggestedCategories.map((c) => c.name).join(", ")}. Você edita o que quiser.`
                      : "Uma lista pronta para você adicionar seus desejos."
                  }
                  badge="Recomendado"
                />
                <ChoiceCard
                  selected={template === "EMPTY"}
                  onSelect={() => setTemplate("EMPTY")}
                  icon={<PenLine className="size-5" />}
                  title="Montar sozinho"
                  text="Comece com a lista vazia e adicione item por item, de qualquer loja ou à mão."
                />
                <div className="rounded-card border-line-strong flex items-center gap-4 border border-dashed p-5 opacity-70">
                  <span className="bg-canvas-deep grid size-11 place-items-center rounded-full">
                    🤖
                  </span>
                  <div className="flex flex-1 flex-col gap-0.5">
                    <p className="font-semibold">Criar com IA</p>
                    <p className="text-ink-muted text-sm">
                      Descreva sua casa e o orçamento. Em breve.
                    </p>
                  </div>
                  <Badge>Em breve</Badge>
                </div>
              </div>
              {error ? (
                <p
                  role="alert"
                  className="rounded-control bg-danger-soft text-danger px-4 py-3 text-sm"
                >
                  {error}
                </p>
              ) : null}
            </>
          ) : null}

          <div className="flex items-center justify-between gap-3">
            {step > (initialType ? 2 : 1) ? (
              <Button variant="ghost" onClick={() => setStep((s) => s - 1)}>
                <ArrowLeft className="size-4" /> Voltar
              </Button>
            ) : (
              <span />
            )}
            {step === 2 ? (
              <Button size="lg" onClick={next}>
                Continuar <ArrowRight className="size-4" />
              </Button>
            ) : null}
            {step === 3 ? (
              <Button size="lg" onClick={create} loading={busy}>
                <Check className="size-4" /> Criar minha lista
              </Button>
            ) : null}
          </div>
        </div>
      </main>
    </div>
  );
}

function Heading({ title, subtitle, emoji }: { title: string; subtitle: string; emoji?: string }) {
  return (
    <div className="flex flex-col gap-2">
      {emoji ? (
        <span aria-hidden className="text-4xl">
          {emoji}
        </span>
      ) : null}
      <h1 className="font-display text-display-sm sm:text-display text-balance font-semibold tracking-tight">
        {title}
      </h1>
      <p className="text-ink-muted text-lg">{subtitle}</p>
    </div>
  );
}

function ChoiceCard({
  selected,
  onSelect,
  icon,
  title,
  text,
  badge,
}: {
  selected: boolean;
  onSelect: () => void;
  icon: React.ReactNode;
  title: string;
  text: string;
  badge?: string;
}) {
  return (
    <button
      type="button"
      role="radio"
      aria-checked={selected}
      onClick={onSelect}
      className={cn(
        "rounded-card bg-surface shadow-hairline duration-(--duration-base) ease-out-soft hover:shadow-soft focus-visible:shadow-focus flex items-start gap-4 p-5 text-left transition-all focus-visible:outline-none",
        selected && "shadow-[0_0_0_2px_var(--color-primary)]",
      )}
    >
      <span
        className={cn(
          "grid size-11 shrink-0 place-items-center rounded-full",
          selected ? "bg-primary text-primary-ink" : "bg-primary-soft text-primary",
        )}
      >
        {icon}
      </span>
      <span className="flex flex-1 flex-col gap-1">
        <span className="flex items-center gap-2 font-semibold">
          {title}
          {badge ? <Badge tone="primary">{badge}</Badge> : null}
        </span>
        <span className="text-ink-muted text-sm leading-relaxed">{text}</span>
      </span>
    </button>
  );
}
