import type { Metadata } from "next";
import { Badge, Card } from "@listou/ui";
import { Logo } from "@/components/logo";

export const metadata: Metadata = {
  title: "Loja de demonstração",
  robots: { index: false, follow: false },
};

const NAMES: Record<string, string> = {
  amazon: "Amazon",
  mercado_livre: "Mercado Livre",
  shopee: "Shopee",
};

/**
 * Destination of the MOCK affiliate provider. It stands in for a marketplace
 * product page so the whole /go redirect flow can be exercised without any
 * real store, credentials or invented URLs.
 */
export default async function DemoStorePage({
  params,
  searchParams,
}: {
  params: Promise<{ merchant: string; id: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const { merchant, id } = await params;
  const query = await searchParams;
  const name = NAMES[merchant] ?? merchant;
  const attribution = Object.entries(query).filter(([k]) => k === "demo_tag" || k === "click");

  return (
    <main className="container-page flex min-h-dvh max-w-xl flex-col justify-center gap-6 py-12">
      <Logo />
      <Card className="flex flex-col gap-5 p-8">
        <Badge tone="warning" className="w-fit">
          Loja de demonstração
        </Badge>
        <h1 className="font-display text-3xl font-semibold">Você chegou à loja “{name}”</h1>
        <p className="text-ink-soft leading-relaxed">
          Esta página representa o produto{" "}
          <code className="bg-canvas-deep rounded px-1.5 py-0.5 text-sm">{id}</code> numa loja. Ela
          existe apenas para testar o redirecionamento de afiliados: nenhuma loja real, preço ou
          compra está envolvida.
        </p>
        {attribution.length > 0 ? (
          <dl className="rounded-control bg-canvas grid gap-1 p-4 text-sm">
            {attribution.map(([k, v]) => (
              <div key={k} className="flex justify-between gap-4">
                <dt className="text-ink-muted">{k}</dt>
                <dd className="break-all font-mono">{Array.isArray(v) ? v.join(",") : v}</dd>
              </div>
            ))}
          </dl>
        ) : null}
        <p className="text-ink-muted text-sm">
          Pode fechar esta aba para voltar à lista de presentes.
        </p>
      </Card>
    </main>
  );
}
