"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { LogOut, Plus } from "lucide-react";
import { Avatar, Button, buttonVariants } from "@listou/ui";
import { apiVoid } from "@/lib/api";
import { Logo } from "./logo";

export function AppHeader({ userName, isAdmin }: { userName: string; isAdmin?: boolean }) {
  const router = useRouter();
  async function logout() {
    await apiVoid("/auth/logout", { method: "POST" });
    router.replace("/");
    router.refresh();
  }
  return (
    <header className="border-line/70 bg-canvas/85 sticky top-0 z-30 border-b backdrop-blur-md">
      <div className="container-page flex h-16 items-center justify-between gap-3">
        <div className="flex items-center gap-6">
          <Logo href="/dashboard" />
          <Link
            href="/dashboard"
            className="text-ink-soft hover:text-ink hidden text-sm font-semibold sm:block"
          >
            Minhas listas
          </Link>
          {isAdmin ? (
            <Link
              href="/admin/produtos"
              className="text-ink-soft hover:text-ink hidden text-sm font-semibold sm:block"
            >
              Produtos curados
            </Link>
          ) : null}
        </div>
        <div className="flex items-center gap-2">
          <Link href="/criar" className={buttonVariants({ variant: "soft", size: "sm" })}>
            <Plus className="size-4" /> Nova lista
          </Link>
          <Avatar name={userName} size="sm" className="ring-2" />
          <Button variant="ghost" size="icon" onClick={logout} aria-label="Sair">
            <LogOut className="size-4" />
          </Button>
        </div>
      </div>
    </header>
  );
}
