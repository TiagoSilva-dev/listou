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

## ADR-0010: AI List Builder sobre a interface `recommendations.Provider`

**Status:** aceito (M10).

**Contexto:** o produto precisa sugerir desejos para completar a lista, mas ainda não há LLM nem Jev contratados, e a aplicação não pode depender deles.

**Decisão:** `recommendations.Builder` ordena e filtra o que um `Provider` devolve (remove o que a lista já tem por comparação de tokens, prioriza por importância e pelo texto livre do usuário, limita a 24 desejos). Hoje o provider é por regras; um provider LLM/Jev entra pela mesma interface. O módulo `listbuilder` expõe sugerir e aplicar atrás da flag `AI_LIST_BUILDER` (404 `FEATURE_DISABLED` quando desligada). Sugestões são apenas desejos: nunca produtos, preços ou links; o produto continua vindo do `MatchingService`.

**Consequências:** trocar o motor não muda API nem UI. O texto livre só afeta a ordenação até existir um provider generativo; qualquer provider futuro deve tratar o prompt como dado não confiável e passar pelos mesmos limites.

## ADR-0011: Produtos por link colado (`LINK`) enquanto não há API oficial

**Status:** aceito (M11).

**Contexto:** Mercado Livre não tem API de afiliados documentada que possamos usar e a Amazon exige conta com vendas qualificadas para liberar a Creators API. Scraping por n8n ou similar violaria termos das lojas, quebra a cada mudança de HTML e contraria a regra de nunca inventar dados de parceiros.

**Decisão:** provider `LINK` (`internal/affiliate/link`) atrás de `affiliate.Provider`. O usuário cola a URL; a API baixa a página (SSRF-safe) e lê só os metadados públicos de pré-visualização (JSON-LD `Product`, Open Graph, `<title>`): título, imagem, marca, descrição. Preço, disponibilidade e rating **nunca** são lidos; o item entra como desejo. Se a página não pode ser lida (bloqueio de bot, sem metadados, título igual ao nome da loja), o usuário digita o nome. `BuildAffiliateURL` devolve a URL colada sem alterar; só parâmetros de rastreamento de analytics (`utm_*`, `fbclid`, `gclid`…) são removidos. O clique continua passando por `/go/{offerId}`. Sem n8n nem novo serviço.

**Consequências:** nenhuma monetização por afiliado nesse caminho até haver o formato de link aprovado de cada programa (`docs/integrations/<loja>.md`). Adapters oficiais entram pela mesma interface. O fetch de URL de usuário é superfície de SSRF, mitigada em `docs/integrations/link.md`.

## ADR-0012: Catálogo curado com links de afiliado nossos (`CURATED`)

**Status:** aceito (M11).

**Contexto:** o "colar link" não monetiza (ADR-0011). Os programas permitem gerar links de afiliado no painel, mas ainda não há API oficial utilizável para busca ou preço.

**Decisão:** provider `CURATED` (`internal/affiliate/curated`) lê um catálogo (inicialmente `catalog.json` embutido; hoje tabelas, ver ADR-0013): produtos reais com título, imagem e um link de afiliado por loja gerado por nós. `BuildAffiliateURL` devolve o link sem alterá-lo. Sem preço nem disponibilidade. O catálogo é validado ao carregar. Um provider por merchant (índice único), então o Mercado Livre passa a `CURATED` e Amazon/Shopee seguem no `MOCK` até terem links.

**Consequências:** quem escolhe produtos pela busca sai por um link monetizado; quem cola link próprio não. Crescer o catálogo é um PR em `catalog.json`; se virar centenas de itens, migrar para tabela com admin. Imagens e divulgação de afiliado dependem dos termos de cada programa (`docs/integrations/curated.md`).

## ADR-0013: Painel admin e catálogo curado em tabelas

**Status:** aceito (M11).

**Contexto:** o ADR-0012 previa migrar o `catalog.json` para tabela com admin quando o catálogo crescesse. Editar JSON exige PR e redeploy a cada produto, e quem cura o catálogo não é necessariamente quem programa.

**Decisão:** `curated_products`/`curated_offers` (migration 00009, com os 3 produtos iniciais) substituem o JSON; `curated.Provider` lê de uma `Source` (Postgres). Módulo `internal/admin` expõe `/api/v1/admin/curated-products` atrás de `RequireAdminFunc` (papel `ADMIN`, concedido só por `migrate admin <email>`), com validação em `curated.Input`, auditoria em `audit_logs` e a tela `/admin/produtos` (404 para quem não é admin). O leitor de link existente (`/products/link-preview`) preenche título e imagem.

**Consequências:** o catálogo muda sem deploy. Cada busca consulta o Postgres (volume pequeno; sem cache). Sem edição pública: o papel ADMIN não pode ser dado pela interface. Preços continuam fora.
