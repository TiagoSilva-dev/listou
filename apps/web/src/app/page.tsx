import Link from "next/link";
import { ArrowRight, Gift, HeartHandshake, Lock, Share2, Store } from "lucide-react";
import { EVENT_TYPES } from "@listou/types";
import { Badge, buttonVariants, Progress, Glyph } from "@listou/ui";
import { CoverArt } from "@/components/cover-art";
import { SiteFooter } from "@/components/site-footer";
import { SiteHeader } from "@/components/site-header";

const BulbIcon = ({ className }: { className?: string }) => (
  <Glyph name="bulb" className={className} />
);

const steps = [
  {
    icon: BulbIcon,
    title: "Crie",
    text: "Escolha o momento e monte sua lista em minutos — sozinho ou com sugestões.",
  },
  {
    icon: Store,
    title: "Descubra",
    text: "Encontre produtos em várias lojas ou adicione qualquer desejo à mão.",
  },
  {
    icon: Share2,
    title: "Compartilhe",
    text: "Um link lindo para WhatsApp, Instagram, e-mail ou QR Code no convite.",
  },
  {
    icon: Gift,
    title: "Presenteie",
    text: "Convidados reservam ou compram sem cadastro. Nada de presente repetido.",
  },
];

const promises = [
  {
    icon: HeartHandshake,
    title: "Sem cadastro para convidados",
    text: "Quem recebe o link só escolhe e presenteia.",
  },
  {
    icon: Lock,
    title: "Presente surpresa",
    text: "Acompanhe o progresso sem descobrir quem deu o quê.",
  },
  {
    icon: Store,
    title: "Qualquer loja",
    text: "Amazon, Mercado Livre, Shopee ou um link qualquer.",
  },
];

export default function HomePage() {
  return (
    <>
      <SiteHeader />
      <main>
        <section className="container-page grid items-center gap-12 pb-20 pt-10 md:grid-cols-[1.1fr_0.9fr] md:pb-28 md:pt-16">
          <div className="animate-fade-up flex flex-col gap-7">
            <Badge tone="primary" className="w-fit">
              <Glyph name="gift" className="size-3.5" /> Grátis para criar e compartilhar
            </Badge>
            <h1 className="font-display text-display md:text-display-lg text-balance font-semibold tracking-tight">
              A lista de presentes que o seu momento <em className="text-primary">merece</em>.
            </h1>
            <p className="text-ink-soft max-w-lg text-pretty text-lg leading-relaxed">
              Crie, compartilhe e receba exatamente o que você precisa — de qualquer loja, sem
              complicação para quem vai presentear.
            </p>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Link href="/criar" className={buttonVariants({ size: "lg" })}>
                Criar minha lista grátis <ArrowRight className="size-4" />
              </Link>
              <Link
                href="/l/joao-e-maria"
                className={buttonVariants({ variant: "secondary", size: "lg" })}
              >
                Ver um exemplo
              </Link>
            </div>
          </div>
          <HeroPreview />
        </section>

        <section className="bg-surface py-20">
          <div className="container-page flex flex-col gap-10">
            <div className="flex max-w-xl flex-col gap-3">
              <h2 className="font-display text-display-sm font-semibold tracking-tight">
                Para cada momento da vida
              </h2>
              <p className="text-ink-muted">
                Comece por um modelo pensado para a ocasião — você ajusta tudo depois.
              </p>
            </div>
            <ul className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
              {EVENT_TYPES.filter((t) => t.type !== "CUSTOM").map((t) => (
                <li key={t.type}>
                  <Link
                    href={`/criar?tipo=${t.type}`}
                    className="rounded-card bg-canvas duration-(--duration-base) ease-out-soft hover:shadow-lift group flex h-full flex-col gap-6 p-5 transition-all hover:-translate-y-0.5"
                  >
                    <span
                      className="duration-(--duration-base) ease-spring text-3xl transition-transform group-hover:scale-110"
                      aria-hidden
                    >
                      <Glyph name={t.emoji} className="text-primary" />
                    </span>
                    <span className="font-semibold">{t.label}</span>
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </section>

        <section className="container-page flex flex-col gap-12 py-24">
          <h2 className="font-display text-display-sm max-w-2xl text-balance font-semibold tracking-tight">
            Criar, descobrir, compartilhar e presentear. Simples assim.
          </h2>
          <ol className="grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
            {steps.map((s, i) => (
              <li key={s.title} className="flex flex-col gap-3">
                <span className="flex items-center gap-3">
                  <span className="bg-primary-soft text-primary grid size-11 place-items-center rounded-full">
                    <s.icon className="size-5" />
                  </span>
                  <span className="text-ink-muted text-xs font-semibold uppercase tracking-widest">
                    Passo {i + 1}
                  </span>
                </span>
                <h3 className="font-display text-xl font-semibold">{s.title}</h3>
                <p className="text-ink-muted leading-relaxed">{s.text}</p>
              </li>
            ))}
          </ol>
        </section>

        <section className="container-page pb-24">
          <CoverArt theme="night" className="rounded-hero px-6 py-14 text-white sm:px-14 sm:py-20">
            <div className="grid gap-10 md:grid-cols-[1fr_1fr] md:items-center">
              <div className="flex flex-col gap-5">
                <h2 className="font-display text-display-sm md:text-display text-balance font-semibold tracking-tight">
                  Seus convidados vão achar tudo muito fácil.
                </h2>
                <p className="max-w-md leading-relaxed text-white/75">
                  Uma página rápida, bonita no celular, que mostra o que ainda falta e leva direto
                  para a loja.
                </p>
                <Link
                  href="/criar"
                  className={buttonVariants({
                    variant: "secondary",
                    size: "lg",
                    className: "w-fit",
                  })}
                >
                  Começar agora
                </Link>
              </div>
              <ul className="flex flex-col gap-3">
                {promises.map((p) => (
                  <li
                    key={p.title}
                    className="rounded-card flex gap-4 bg-white/10 p-5 backdrop-blur"
                  >
                    <p.icon className="mt-0.5 size-5 shrink-0 text-[#f5c3b3]" />
                    <div>
                      <p className="font-semibold">{p.title}</p>
                      <p className="text-sm text-white/70">{p.text}</p>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </CoverArt>
        </section>
      </main>
      <SiteFooter />
    </>
  );
}

function HeroPreview() {
  const items = [
    {
      emoji: "pan",
      title: "Air Fryer 5L",
      price: "R$ 379,00",
      tag: "Disponível",
      tone: "success" as const,
    },
    {
      emoji: "bed",
      title: "Jogo de cama queen",
      price: "R$ 249,90",
      tag: "Reservado",
      tone: "warning" as const,
    },
    {
      emoji: "coffee",
      title: "Cafeteira espresso",
      price: "R$ 899,00",
      tag: "Disponível",
      tone: "success" as const,
    },
  ];
  return (
    <div className="animate-fade-up relative mx-auto w-full max-w-sm [animation-delay:120ms] md:max-w-none">
      <div
        aria-hidden
        className="from-primary-soft via-blush-soft to-sun-soft absolute -inset-6 -z-10 rounded-[3rem] bg-gradient-to-br blur-2xl"
      />
      <div className="rounded-hero bg-surface shadow-lift overflow-hidden">
        <CoverArt theme="blush" emoji="house" className="flex h-44 items-end p-5">
          <Badge tone="glass"><Glyph name="house" className="size-3.5" /> Chá de casa nova</Badge>
        </CoverArt>
        <div className="flex flex-col gap-5 p-5">
          <div>
            <p className="font-display text-2xl font-semibold">João &amp; Maria</p>
            <p className="text-ink-muted text-sm">12 de dezembro · São Paulo</p>
          </div>
          <div className="flex flex-col gap-2">
            <div className="text-ink-soft flex justify-between text-xs font-semibold">
              <span>8 de 24 presentes</span>
              <span>33%</span>
            </div>
            <Progress value={8} max={24} label="Progresso da lista" />
          </div>
          <ul className="flex flex-col gap-2.5">
            {items.map((it) => (
              <li
                key={it.title}
                className="rounded-control bg-canvas flex items-center gap-3 p-2.5"
              >
                <span
                  className="rounded-chip bg-surface grid size-12 place-items-center text-2xl"
                  aria-hidden
                >
                  <Glyph name={it.emoji} className="text-primary" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-semibold">{it.title}</p>
                  <p className="text-ink-muted text-xs tabular-nums">{it.price}</p>
                </div>
                <Badge tone={it.tone}>{it.tag}</Badge>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}
