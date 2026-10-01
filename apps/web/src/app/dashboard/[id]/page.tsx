import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { Dashboard, OwnerList } from "@listou/types";
import { z } from "zod";
import { EventWorkspace } from "@/components/event-workspace";
import { ApiError } from "@/lib/api";
import { siteUrl } from "@/lib/env";
import { serverApi } from "@/lib/server-api";

export const metadata: Metadata = { title: "Painel da lista" };

export default async function EventPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ novo?: string }>;
}) {
  const { id } = await params;
  const { novo } = await searchParams;
  const loaded = await Promise.all([
    serverApi(`/events/${id}/dashboard`, Dashboard, { auth: true }),
    serverApi(`/events/${id}/list`, OwnerList, { auth: true }),
    // The flag only gates the UI; the API enforces it on its own.
    serverApi("/flags", z.object({ flags: z.record(z.string(), z.boolean()) }), {
      revalidate: 60,
    }).catch(() => ({ flags: {} as Record<string, boolean> })),
  ]).catch((err: unknown) => {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  });
  if (!loaded) notFound();
  const [dashboard, list, flags] = loaded;
  return (
    <EventWorkspace
      initialDashboard={dashboard}
      initialList={list}
      siteUrl={siteUrl()}
      justCreated={novo === "1"}
      aiBuilder={flags.flags.AI_LIST_BUILDER === true}
    />
  );
}
