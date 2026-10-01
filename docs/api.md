# API

Base: `/api/v1` (JSON, camelCase). Especificação: [`openapi.yaml`](openapi.yaml).

## Convenções

- Autenticação por cookie de sessão `listou_session` (definido em `/auth/login` e `/auth/register`).
- Erros:
  ```json
  { "error": { "code": "LIST_NOT_FOUND", "message": "…", "requestId": "…", "fields": {} } }
  ```
  `fields` aparece em `VALIDATION_FAILED` (422). Nunca há stack trace.
- `X-Request-Id` é aceito (8–64 chars `[A-Za-z0-9_-]`) ou gerado, e devolvido em toda resposta.
- Dinheiro em centavos (`priceCents`) + `currency`. Datas ISO-8601; `eventDate` é `YYYY-MM-DD`.

## Operacional

| Método | Rota            | Descrição               |
| ------ | --------------- | ----------------------- |
| GET    | `/health`       | liveness                |
| GET    | `/ready`        | readiness (checa banco) |
| GET    | `/api/v1/flags` | feature flags ativas    |
