import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { EventType } from "@listou/types";
import { CreateWizard } from "@/components/create-wizard";
import { getCurrentUser } from "@/lib/server-api";

export const metadata: Metadata = { title: "Criar lista", robots: { index: false } };

export default async function CreatePage({
  searchParams,
}: {
  searchParams: Promise<{ tipo?: string }>;
}) {
  const { tipo } = await searchParams;
  const parsed = EventType.safeParse(tipo);
  const initialType = parsed.success ? parsed.data : null;
  if (!(await getCurrentUser())) {
    const back = `/criar${initialType ? `?tipo=${initialType}` : ""}`;
    redirect(`/criar-conta?next=${encodeURIComponent(back)}`);
  }
  return <CreateWizard initialType={initialType} />;
}
