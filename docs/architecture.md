# Arquitetura

> Leia este documento e `decisions.md` antes de qualquer mudança arquitetural.

## Visão geral

```
Browser ──► Next.js (apps/web, Vercel)
              │  Server Components buscam a API com o cookie do usuário
              │  rewrites: /api/v1/* e /go/* ─────────────┐
              ▼                                           ▼
           Go API (apps/api, container) ───────► PostgreSQL (gerenciado)
              │
              └─► affiliate.Provider (Mock hoje; Amazon/ML/Shopee após docs oficiais)
```

- **Modular monolith** em Go. Cada pasta em `internal/` é um módulo com fronteira clara; a
  extração futura para serviço é possível porque módulos só se falam por interfaces pequenas.
- **Next.js como BFF**: o navegador só conversa com a origem web. Rewrites encaminham `/api/v1` e
  `/go` para a API, mantendo o cookie de sessão first-party (`SameSite=Lax`, `HttpOnly`).
- **REST versionado** em `/api/v1`, envelope de erro padronizado, OpenAPI em `docs/openapi.yaml`.

## Módulos da API

| Módulo            | Responsabilidade                                                                                                                       |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `platform/*`      | config, logger (slog JSON), httpx (erros, middlewares, rate limit, CSRF), database (pool, tx), flags, health                           |
| `auth`            | registro, login, sessões opacas (hash SHA-256 no banco), middleware de autenticação; preparado para Google/Apple via `auth_identities` |
| `events`          | Event + GiftList principal + categorias; slug; publicação; autorização owner/co-owner                                                  |
| `lists`           | ListItem: CRUD, quantidades, status derivado, associação a produto                                                                     |
| `catalog`         | busca unificada via providers, persistência de Product/ProductOffer, `ProductMatchingService`                                          |
| `affiliate`       | interface `Provider`, registro de providers, `MockProvider`, redirect `/go/{offerId}`                                                  |
| `reservations`    | reservas transacionais, tokens de gestão para convidados, expiração                                                                    |
| `analytics`       | eventos de produto internos e cliques outbound (privacy-first)                                                                         |
| `publiclists`     | leitura read-only otimizada da página pública                                                                                          |
| `dashboard`       | métricas agregadas do criador (respeita presente surpresa)                                                                             |
| `recommendations` | `RecommendationProvider` (regras hoje; LLM/Jev depois)                                                                                 |
| `decision`        | `DecisionEngine` (Rule/Mock hoje; Jev opcional depois)                                                                                 |
| `access`          | resolução de papéis (evento/lista/item/categoria); recurso alheio responde 404                                                         |
| `platform/audit`  | trilha de auditoria (ex.: exclusão de evento)                                                                                          |

Camadas por módulo: `model` (regras puras, testáveis) → `repository` (SQL) → `service`
(casos de uso + transações) → `handler` (HTTP fino).

## Fluxo outbound `/go/{offerId}`

1. Busca oferta + merchant + provider habilitado.
2. `Provider.BuildAffiliateURL(offer, ctx)` gera a URL permitida (o mock devolve a `product_url`
   com parâmetros de rastreio fictícios claramente marcados).
3. Registra `click_events` (oferta, item, evento, visitor id aleatório first-party, host do
   referrer, UTM, classe de dispositivo) — sem IP, sem fingerprint.
4. `302` para o marketplace. A URL do marketplace nunca é enviada ao frontend.

## Presente surpresa

Com `surprise_mode` ativo, o dono não recebe estado por item nem nomes de convidados: o dashboard
agrega apenas totais, e `GET /events/{id}/list` passa por `redactIfSurprise`. Convidados continuam
vendo disponibilidade na página pública. Há teste de integração cobrindo o vazamento.

## Sitemap público

`GET /api/v1/public/sitemap` lista slugs publicados; o Next gera `sitemap.xml` a partir dele.

## Concorrência de reservas

`reservations.Service.Reserve` abre transação, faz `SELECT ... FOR UPDATE` no item, valida
disponibilidade, cria a reserva e incrementa `reserved_quantity`. A constraint
`purchased_quantity + reserved_quantity <= desired_quantity` garante integridade mesmo que uma
regra de aplicação falhe.

## Front-end

- App Router, Server Components por padrão; páginas públicas renderizadas no servidor com
  `revalidate` curto e metadata (OG, canonical, robots conforme visibilidade).
- Design system em `packages/ui`: tokens em CSS (`@theme` Tailwind v4) + componentes.
- Formulários: React Hook Form + Zod (`packages/types`).

## Escala ("simple now, scalable later")

- Índices orientados às queries reais; cursor pagination onde listas crescem.
- Cliques/analytics em tabelas append-only, prontas para particionamento por mês.
- Rate limit em memória hoje; move para Redis quando houver mais de uma instância (ADR-0004).
- Página pública é cacheável em CDN; invalidação por revalidate.
