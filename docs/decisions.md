# Decisões de arquitetura (ADRs)

Formato: contexto → decisão → consequências. Novas ADRs são adicionadas no fim; decisões
substituídas ficam marcadas como _Superseded_.

## ADR-0001 — Monorepo pnpm + modular monolith em Go

**Contexto:** time pequeno, produto em validação, necessidade de evoluir rápido sem perder fronteiras.
**Decisão:** monorepo (`apps/web`, `apps/api`, `packages/*`) com pnpm workspaces; backend como
modular monolith Go com módulos em `internal/`. Sem Turborepo por enquanto (scripts `pnpm -r`
bastam).
**Consequências:** um deploy de API; fronteiras de módulo garantidas por convenção e interfaces;
extração para serviço possível depois.

## ADR-0002 — stdlib `net/http` + pgx + goose

**Contexto:** Go 1.22+ trouxe roteamento com método e path params na stdlib.
**Decisão:** `net/http.ServeMux` (sem framework), `pgx/v5` com SQL explícito (sem ORM),
`goose` para migrations SQL embutidas no binário (`migrate up|down|status|reset|seed`).
**Consequências:** poucas dependências, SQL revisável, migrations rodam no mesmo binário do deploy.

## ADR-0003 — Sessões opacas no servidor em vez de JWT

**Contexto:** precisamos de logout real, revogação e cookies seguros; não há múltiplos serviços
validando tokens.
**Decisão:** token aleatório de 256 bits em cookie `HttpOnly; SameSite=Lax; Secure` (produção);
banco guarda só SHA-256. Senhas com argon2id. Identidades em `auth_identities` para Google/Apple.
**Consequências:** uma query por request autenticada (indexada); revogação imediata.

## ADR-0004 — Sem Redis no MVP

**Contexto:** o prompt pede Redis só onde houver ganho real.
**Decisão:** rate limiting em memória (uma instância). Expiração de reservas por varredura
periódica + filtragem por `expires_at` na leitura.
**Consequências:** quando houver mais de uma instância da API, o rate limiter migra para Redis
(é o único ponto que muda).

## ADR-0005 — Next.js como BFF via rewrites

**Contexto:** cookies de sessão devem ser first-party; a URL do marketplace não deve vazar para
componentes.
**Decisão:** o navegador chama apenas a origem web; `/api/v1/*` e `/go/*` são reescritos para a API.
Server Components chamam a API diretamente encaminhando o cookie.
**Consequências:** sem CORS; CSRF mitigado por `SameSite=Lax` + checagem de `Origin` na API.

## ADR-0006 — Afiliados por adapters; mock antes de real

**Contexto:** não podemos inventar APIs/URLs de parceiros.
**Decisão:** interface `affiliate.Provider`; `MockProvider` com catálogo fictício explicitamente
marcado como demonstração. Adapter real só após documentação oficial registrada em
`docs/integrations/`.
**Consequências:** o domínio não conhece marketplaces; trocar/adicionar parceiros não afeta
`lists`, `events` ou `reservations`.

## ADR-0007 — Status do item derivado, não armazenado

**Contexto:** `status` armazenado diverge facilmente das quantidades.
**Decisão:** armazenar `desired/purchased/reserved_quantity` + `archived_at`; derivar status em
`lists.DeriveStatus`. CHECK de capacidade no banco.
**Consequências:** uma fonte de verdade; filtros por status usam expressões sobre quantidades.

## ADR-0008 — Design tokens em CSS (Tailwind v4 `@theme`)

**Contexto:** precisamos de um design system consistente e premium sem números mágicos.
**Decisão:** tokens em `packages/ui/src/styles/tokens.css` (cores, raio, sombra, tipografia,
motion); componentes base próprios no estilo shadcn/ui (Radix para Dialog). Tema claro apenas no
MVP. Capas de evento geradas por gradientes temáticos quando não há foto.
**Consequências:** utilitários como `bg-canvas`, `rounded-card`, `shadow-soft` em todo o app.

## ADR-0009 — Membros no nível do evento

**Contexto:** o modelo conceitual cita `GiftListMember`, mas permissões são do evento inteiro.
**Decisão:** tabela `event_members` (CO_OWNER, VIEWER). Dono é `events.owner_id`.
**Consequências:** múltiplas listas por evento herdam as permissões.
