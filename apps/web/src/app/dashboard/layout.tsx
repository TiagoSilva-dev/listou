import { redirect } from "next/navigation";
import { AppHeader } from "@/components/app-header";
import { getCurrentUser } from "@/lib/server-api";

export const metadata = { robots: { index: false, follow: false } };

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser();
  if (!user) redirect("/entrar?next=/dashboard");
  return (
    <>
      <AppHeader userName={user.name} isAdmin={user.role === "ADMIN"} />
      <main className="container-page py-8 sm:py-12">{children}</main>
    </>
  );
}
