# Listou

Listas de presentes bonitas para casamento, chá de bebê, casa nova e todos os momentos que importam.

**Stack:** Next.js 16 (App Router) + Tailwind v4 · Go 1.24 (modular monolith, REST) · PostgreSQL 16 ·
pnpm workspaces · Docker · GitHub Actions.

## Começando

Pré-requisitos: Go 1.24+, Node 22+, pnpm 10+, Docker (ou um Postgres 16 local).

```bash
cp .env.example .env
make install
make up        # Postgres em Docker
make db-reset  # migrations + seed de desenvolvimento
make dev       # API em http://localhost:8080, web em http://localhost:3000
```

Ou tudo em containers: `make dev-docker`.

Login de desenvolvimento (seed): `tiago@listou.dev` / `listou123`.
Lista pública de exemplo: http://localhost:3000/l/tiago-e-julia

## Documentação

- [CLAUDE.md](CLAUDE.md) — regras e convenções do projeto
- [docs/product.md](docs/product.md) — visão de produto e métricas
- [docs/architecture.md](docs/architecture.md) — arquitetura
- [docs/database.md](docs/database.md) — modelo de dados e ERD
- [docs/api.md](docs/api.md) e [docs/openapi.yaml](docs/openapi.yaml) — API
- [docs/affiliate-providers.md](docs/affiliate-providers.md) — afiliados
- [docs/decisions.md](docs/decisions.md) — ADRs
- [docs/security.md](docs/security.md) · [docs/deployment.md](docs/deployment.md) · [docs/roadmap.md](docs/roadmap.md)
