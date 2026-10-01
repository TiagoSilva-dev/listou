# CLAUDE.md — Listou

Listou é uma plataforma consumer-first de listas de presentes (universal gift registry) com
monetização por afiliados. Para o usuário é "uma maneira bonita e inteligente de criar, organizar
e compartilhar listas para momentos especiais". A monetização fica por trás da experiência.

Loop central: **CREATE → DISCOVER → SHARE → GIFT**.

## Regras obrigatórias

- **ANTES DE ALTERAR ARQUITETURA IMPORTANTE:** ler `/docs/architecture.md` e `/docs/decisions.md`.
- **ANTES DE IMPLEMENTAR MARKETPLACE:** ler `/docs/affiliate-providers.md`. Nunca inventar endpoints,
  credenciais, formatos de URL, preços, ratings ou regras comerciais de parceiros.
- Lógica de Amazon / Mercado Livre / Shopee **só** dentro de `apps/api/internal/affiliate/<provider>`.
  O domínio fala apenas com a interface `affiliate.Provider`.
- IA sugere **desejos** (ListItem), nunca produtos/preços/disponibilidade apresentados como reais.
- Nenhuma alteração manual de schema: tudo via `apps/api/migrations` (goose, com `Up` e `Down`).
- Nunca commitar secrets. Configuração nova vai em `.env.example` sem valores reais.
- Não adicionar Redis, Kafka, GraphQL, Kubernetes ou microsserviços sem ADR aprovando.

## Estrutura

```
apps/api            Go modular monolith (REST /api/v1)
  cmd/api           entrypoint HTTP
  cmd/migrate       up | down | status | reset | seed
  internal/app      composição dos módulos (router, middlewares)
  internal/platform config, logger, httpx (erros, middlewares), database, flags, health
  internal/<módulo> auth, events, lists, catalog, affiliate, reservations, analytics, ...
  migrations        SQL versionado (goose) embutido no binário
apps/web            Next.js (App Router) — frontend + BFF (rewrites para a API)
packages/ui         design tokens (tokens.css) + componentes base
packages/types      schemas Zod, tipos da API, formatadores compartilhados
packages/config     tsconfig base
docs/               arquitetura, banco, API, ADRs, segurança, deploy
infra/docker        Dockerfiles
```

Cada módulo Go segue: `model.go` (tipos/regras puras) → `repository.go` (SQL, pgx) →
`service.go` (casos de uso, transações) → `handler.go` (HTTP fino: decode, chama service, encode).
Handlers nunca contêm regra de negócio. Módulos só se conhecem via interfaces pequenas definidas
no consumidor.

## Comandos

```
make up            # Postgres via Docker
make dev           # migra e roda API (:8080) + web (:3000)
make dev-docker    # stack inteira em Docker
make migrate | migrate-down | seed | db-reset
make test          # Go + Vitest
make e2e           # Playwright (stack rodando)
make lint          # go vet, gofmt, eslint, tsc
make format
make build
```

### Testes individuais

```
cd apps/api && go test -race ./internal/lists -run TestName     # um teste Go
cd apps/api && TEST_DATABASE_URL=postgres://... go test ./internal/app   # integração (pula sem a var; use um banco descartável!)
pnpm --filter @listou/web exec vitest run path/to/file.test.ts          # um teste Vitest
pnpm --filter @listou/web exec playwright test e2e/registry.spec.ts     # um e2e (stack rodando)
```

Login de desenvolvimento (após `make db-reset`): `tiago@listou.dev` / `listou123`; lista pública
de exemplo em `/l/joao-e-maria`.

## Arquitetura (visão geral)

- **Web é BFF**: o navegador só fala com a origem web; `/api/v1/*` e `/go/*` são reescritos para a API
  (ADR-0005), então não há CORS e o cookie de sessão é first-party. Server Components chamam a API
  direto encaminhando o cookie. Não chame a API de outra origem no cliente.
- **Sessão**: token opaco em cookie HttpOnly (só o SHA-256 vai ao banco), não JWT (ADR-0003).
  CSRF mitigado por `SameSite=Lax` + checagem de `Origin`.
- **Status de item é derivado** (`lists.DeriveStatus`) das quantidades desired/purchased/reserved +
  `archived_at`; nunca persista status (ADR-0007).
- **Feature flags**: `FEATURE_FLAGS` no `.env` (`internal/platform/flags`). O AI List Builder
  (`listbuilder`, sobre `recommendations.Provider`) fica atrás de `AI_LIST_BUILDER`.
- Rate limiting é em memória (ADR-0004); não assuma múltiplas instâncias da API.
- Marketplace hoje é `MockProvider` (catálogo fictício); adapters reais só com doc oficial em
  `docs/integrations/`.

## Convenções

Go

- `context.Context` em toda chamada de I/O; erros embrulhados com `fmt.Errorf("...: %w", err)`.
- Erros de domínio são `*httpx.Error` com `code` estável (`LIST_NOT_FOUND`); nunca stack trace no cliente.
- Logs estruturados (`slog`) via `logger.From(ctx)` — sem senhas, tokens ou PII.
- IDs UUIDv7 gerados na aplicação (`uuid.NewV7`). Dinheiro sempre em centavos (`int64`) + `currency`.
- Transações explícitas com `database.WithTx`. Concorrência de reservas: `SELECT ... FOR UPDATE` +
  CHECK constraint no banco.
- Interfaces apenas onde há mais de uma implementação real ou fronteira de módulo.

TypeScript

- `strict`, sem `any`. Validação com Zod (`@listou/types`). Formulários com React Hook Form.
- Server Components por padrão; `"use client"` só onde há interação.
- Nenhum valor visual arbitrário: cores, raios, sombras, tipografia e motion vêm de
  `packages/ui/src/styles/tokens.css`.
- UX: em toda tela, "o usuário entende o próximo passo em menos de 3 segundos?".

## Definition of Done

Código implementado · lint passa · testes passam · build passa · tipos passam · migrations
up/down funcionam · erros tratados · loading e empty states · mobile verificado · documentação
relevante atualizada · nenhuma secret commitada.
