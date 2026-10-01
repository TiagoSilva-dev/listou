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

## Autenticação

| Método | Rota                    | Descrição                                                       |
| ------ | ----------------------- | --------------------------------------------------------------- |
| POST   | `/api/v1/auth/register` | cria conta e sessão (rate limit por IP)                         |
| POST   | `/api/v1/auth/login`    | login; resposta idêntica para e-mail inexistente e senha errada |
| POST   | `/api/v1/auth/logout`   | revoga a sessão                                                 |
| GET    | `/api/v1/me`            | usuário atual                                                   |

## Criador (requer sessão)

| Método           | Rota                                                       | Notas                                                                        |
| ---------------- | ---------------------------------------------------------- | ---------------------------------------------------------------------------- |
| POST/GET         | `/events`                                                  | `template: SUGGESTED` cria categorias e itens sugeridos por regras           |
| GET/PATCH/DELETE | `/events/{id}`                                             | PATCH publica (`status`), muda `visibility`, `surpriseMode`, `slug`, `theme` |
| GET              | `/events/{id}/list`                                        | em modo surpresa, quantidades por item ficam ocultas para o dono             |
| GET              | `/events/{id}/dashboard`                                   | métricas agregadas e atividade recente                                       |
| POST/GET         | `/lists/{id}/items`                                        | item manual ou com `productId`                                               |
| PATCH/DELETE     | `/items/{id}`                                              | DELETE arquiva se já houve reserva                                           |
| POST             | `/lists/{id}/categories` · PATCH/DELETE `/categories/{id}` |                                                                              |
| GET              | `/products/search?q=&sort=`                                | providers habilitados; `sort`: `relevance`, `price_asc`, `price_desc`        |
| POST             | `/products/import`                                         | persiste produto + ofertas do provider                                       |
| GET              | `/products/{id}` · `/products/{id}/offers`                 | ofertas trazem `goUrl`, nunca a URL da loja                                  |
| DELETE           | `/manage/reservations/{id}`                                | dono cancela reserva (audit log)                                             |

## Público (sem login)

| Método | Rota                                               | Notas                                                      |
| ------ | -------------------------------------------------- | ---------------------------------------------------------- |
| GET    | `/public/lists/{slug}`                             | `404 LIST_NOT_FOUND` para inexistente, rascunho ou privada |
| POST   | `/public/lists/{slug}/items/{itemId}/reservations` | devolve `manageToken` **uma vez**                          |
| DELETE | `/reservations/{id}`                               | header `X-Reservation-Token`                               |
| POST   | `/reservations/{id}/confirm`                       | reserva → compra                                           |
| POST   | `/analytics/track`                                 | lista branca de eventos de navegador                       |
| GET    | `/go/{offerId}?item=&utm_*`                        | `302` para a URL do provider; registra clique              |

## Códigos de erro frequentes

`UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION_FAILED`, `EMAIL_TAKEN`, `INVALID_CREDENTIALS`, `EVENT_NOT_FOUND`,
`LIST_NOT_FOUND`, `ITEM_NOT_FOUND`, `SLUG_TAKEN`, `ITEM_NOT_AVAILABLE`, `RESERVATIONS_DISABLED`,
`RESERVATION_NOT_FOUND`, `RESERVATION_NOT_ACTIVE`, `ITEM_QUANTITY_CONFLICT`, `OFFER_NOT_FOUND`,
`CSRF_REJECTED`, `RATE_LIMITED`.

Recursos de outro usuário respondem `404` (não `403`) para não revelar existência.
