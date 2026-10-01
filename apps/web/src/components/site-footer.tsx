import { Logo } from "./logo";

export function SiteFooter() {
  return (
    <footer className="border-line border-t">
      <div className="container-page text-ink-muted flex flex-col gap-6 py-10 text-sm sm:flex-row sm:items-start sm:justify-between">
        <div className="flex max-w-sm flex-col gap-3">
          <Logo />
          <p className="leading-relaxed">
            Listas de presentes bonitas para os momentos que importam. Quando um convidado compra
            pelo nosso link, a loja pode nos pagar uma pequena comissão — sem custo extra para
            ninguém.
          </p>
        </div>
        <p>© {new Date().getFullYear()} Listou · Feito no Brasil</p>
      </div>
    </footer>
  );
}
