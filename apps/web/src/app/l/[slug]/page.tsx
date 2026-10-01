import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { eventTypeMeta, formatEventDate, PublicList } from "@listou/types";
import { PublicRegistry } from "@/components/public-registry";
import { ApiError } from "@/lib/api";
import { siteUrl } from "@/lib/env";
import { serverApi } from "@/lib/server-api";

type Params = { params: Promise<{ slug: string }> };

/** Live lists are fetched anonymously (cacheable); drafts fall back to an owner preview. */
async function load(slug: string): Promise<PublicList | null> {
  try {
    return await serverApi(`/public/lists/${encodeURIComponent(slug)}`, PublicList, {
      revalidate: 5,
    });
  } catch (err) {
    if (!(err instanceof ApiError) || err.status !== 404) throw err;
  }
  try {
    return await serverApi(`/public/lists/${encodeURIComponent(slug)}`, PublicList, { auth: true });
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}

export async function generateMetadata({ params }: Params): Promise<Metadata> {
  const { slug } = await params;
  const data = await load(slug);
  if (!data) return { title: "Lista não encontrada", robots: { index: false } };
  const { event } = data;
  const meta = eventTypeMeta(event.type);
  const date = formatEventDate(event.eventDate);
  const title = `${meta.headline}: ${event.hostNames ?? event.title}`;
  const description =
    event.description ??
    `Veja a lista de presentes${date ? ` · ${date}` : ""}. Escolha um presente sem precisar criar conta.`;
  const indexable = event.visibility === "PUBLIC" && !data.preview;
  return {
    title,
    description,
    alternates: { canonical: `/l/${event.slug}` },
    robots: indexable ? { index: true, follow: true } : { index: false, follow: false },
    openGraph: {
      title,
      description,
      url: `${siteUrl()}/l/${event.slug}`,
      type: "website",
      locale: "pt_BR",
    },
    twitter: { card: "summary_large_image", title, description },
  };
}

export default async function PublicListPage({ params }: Params) {
  const { slug } = await params;
  const data = await load(slug);
  if (!data) notFound();
  return <PublicRegistry initial={data} siteUrl={siteUrl()} />;
}
