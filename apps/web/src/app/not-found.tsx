import Link from "next/link";
import { buttonVariants, EmptyState } from "@listou/ui";
import { SiteHeader } from "@/components/site-header";

export default function NotFound() {
  return (
    <>
      <SiteHeader />
      <main className="container-page py-20">
        <EmptyState
          icon="search"
          title="Não encontramos esta página"
          description="O link pode estar incompleto ou a lista pode ter sido arquivada pelo criador."
          action={
            <Link href="/" className={buttonVariants({ variant: "secondary" })}>
              Voltar ao início
            </Link>
          }
        />
      </main>
    </>
  );
}
