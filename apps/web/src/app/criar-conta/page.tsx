import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { AuthShell } from "@/components/auth-shell";
import { RegisterForm } from "@/components/auth-form";
import { safeNext } from "@/lib/nav";
import { getCurrentUser } from "@/lib/server-api";

export const metadata: Metadata = { title: "Criar conta", robots: { index: false } };

export default async function RegisterPage({
  searchParams,
}: {
  searchParams: Promise<{ next?: string }>;
}) {
  const next = safeNext((await searchParams).next, "/criar");
  if (await getCurrentUser()) redirect(next);
  return (
    <AuthShell
      title="Vamos começar?"
      subtitle="Crie sua conta grátis e monte sua lista em poucos minutos."
    >
      <RegisterForm next={next} />
    </AuthShell>
  );
}
