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
