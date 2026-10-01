import Link from "next/link";
import { buttonVariants } from "@listou/ui";
import { Logo } from "./logo";

export function SiteHeader() {
  return (
    <header className="bg-canvas/80 sticky top-0 z-30 border-b border-transparent backdrop-blur-md">
      <div className="container-page flex h-16 items-center justify-between">
        <Logo />
        <nav className="flex items-center gap-1">
          <Link href="/entrar" className={buttonVariants({ variant: "ghost", size: "sm" })}>
            Entrar
          </Link>
          <Link href="/criar" className={buttonVariants({ variant: "dark", size: "sm" })}>
            Criar lista
          </Link>
        </nav>
      </div>
    </header>
  );
}
