# Affiliate providers

> **Leia antes de implementar qualquer marketplace.** Não inventar endpoints, credenciais,
> formatos de URL, preços, ratings ou regras comerciais.

## Interface

```go
type Provider interface {
    Code() string                                                     // "MOCK", "AMAZON_ASSOCIATES", ...
    SearchProducts(ctx, SearchQuery) ([]SearchResult, error)
    GetProduct(ctx, externalID string) (ProductData, error)
    GetOffers(ctx, externalID string) ([]OfferData, error)
    BuildAffiliateURL(ctx, Offer, ClickContext) (string, error)
}
```

`TrackOutboundClick` do prompt é responsabilidade do módulo `analytics` (comum a todos os
providers); providers que exigirem notificação própria implementam a interface opcional
`ClickTracker`.

## Regras

1. Lógica específica de marketplace vive **apenas** em `internal/affiliate/<provider>`.
2. Credenciais vêm do ambiente; `affiliate_providers.config` guarda só dados não secretos.
3. Preço, disponibilidade e rating só são exibidos se vierem da fonte autorizada, com
   `last_synced_at`. Rating só se os termos permitirem.
4. Sem scraping onde não for permitido.
5. Comissão nunca altera a ordenação orgânica sem rotulagem.

## Processo para um provider real

1. Consultar documentação oficial atual do programa.
2. Registrar em `docs/integrations/<provider>.md`: programa, link da doc, autenticação, limites de
   taxa, formato de link aprovado, regras de exibição de preço, cache permitido, termos.
3. Implementar a interface + testes com fixtures gravadas da doc.
4. Habilitar via `affiliate_providers.enabled` e flag `MULTI_MARKETPLACE`.

## Status

| Provider                              | Status                                                            |
| ------------------------------------- | ----------------------------------------------------------------- |
| `MOCK`                                | ✅ catálogo de demonstração (dados fictícios, rotulados como tal) |
| `CURATED`                             | ✅ catálogo manual com nossos links de afiliado (`integrations/curated.md`) |
| `LINK`                                | ✅ link colado pelo usuário, sem API de loja (`integrations/link.md`) |
| Amazon (Associates / Creators API)    | ⛔ aguardando documentação oficial e credenciais                  |
| Mercado Livre (programa de afiliados) | ⛔ aguardando documentação oficial e credenciais                  |
| Shopee (Affiliate Program)            | ⛔ aguardando documentação oficial e credenciais                  |
