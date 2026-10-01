import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { AuthShell } from "@/components/auth-shell";
import { LoginForm } from "@/components/auth-form";
import { safeNext } from "@/lib/nav";
import { getCurrentUser } from "@/lib/server-api";

export const metadata: Metadata = { title: "Entrar", robots: { index: false } };

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ next?: string }>;
}) {
  const next = safeNext((await searchParams).next);
  if (await getCurrentUser()) redirect(next);
  return (
    <AuthShell
      title="Que bom ter você de volta"
      subtitle="Entre para continuar montando sua lista."
    >
      <LoginForm next={next} />
    </AuthShell>
  );
}
