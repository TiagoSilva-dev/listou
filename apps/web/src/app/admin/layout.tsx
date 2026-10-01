import { notFound, redirect } from "next/navigation";
import { AppHeader } from "@/components/app-header";
import { getCurrentUser } from "@/lib/server-api";

export const metadata = { title: "Admin", robots: { index: false, follow: false } };

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser();
  if (!user) redirect("/entrar?next=/admin/produtos");
  if (user.role !== "ADMIN") notFound();
  return (
    <>
      <AppHeader userName={user.name} isAdmin />
      <main className="container-page py-8 sm:py-12">{children}</main>
    </>
  );
}
