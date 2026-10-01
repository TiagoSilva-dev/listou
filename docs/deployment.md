# Deploy

## Alvos

- **Web:** Vercel (projeto apontando para `apps/web`, install `pnpm install`, build
  `pnpm --filter @listou/web build`). Env: `API_URL`, `NEXT_PUBLIC_SITE_URL`.
  Alternativa: imagem `infra/docker/web.Dockerfile` (Next standalone).
- **API:** container `infra/docker/api.Dockerfile` (distroless, nonroot) em Fly.io / Render /
  Cloud Run / ECS. Probes: liveness `GET /health`, readiness `GET /ready`.
- **Banco:** PostgreSQL 16 gerenciado (Neon, Supabase, RDS). Extensão `pg_trgm`.
- **Storage:** compatível com S3 (Cloudflare R2) — quando houver upload.

## Migrations

O binário `migrate` vai na mesma imagem da API. Rodar `/app/migrate up` como release step
antes de trocar o tráfego. Migrations devem ser compatíveis com a versão anterior do código
(expand → migrate → contract).

## Variáveis da API

`APP_ENV=production`, `DATABASE_URL`, `PUBLIC_WEB_URL`, `COOKIE_SECURE=true`,
`COOKIE_DOMAIN` (se web e API compartilharem domínio pai), `SESSION_TTL`, `RESERVATION_TTL`,
`FEATURE_FLAGS`, `LOG_LEVEL`, `SENTRY_DSN` (quando habilitado).

## CI/CD

`.github/workflows/ci.yml`: gofmt, go vet, migrations up/reset/up, `go test -race`, build;
prettier, eslint, tsc, vitest, `next build`; build das imagens Docker. Deploy contínuo a ser
ligado quando os ambientes existirem.

## Observabilidade

Logs JSON (`request_id`, `user_id`, `endpoint`, `status`, `duration_ms`). `X-Request-Id`
propagado do web para a API. Error tracking preparado para Sentry (DSN via ambiente).
