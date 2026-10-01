# Produto

## Visão

Plataforma brasileira de **Universal Gift Registry + Shopping Discovery + Affiliate Commerce + AI
Recommendations**. Para o usuário: uma maneira bonita e inteligente de criar, organizar e
compartilhar listas para momentos especiais. Loop: **CREATE → DISCOVER → SHARE → GIFT**.

## Ocasiões

Chá de bebê, casamento, casa nova, aniversário, formatura, Natal, viagem, mudança, lista pessoal,
wishlist e evento personalizado (`EventType`).

## Loop do MVP

Cria conta → cria lista → adiciona itens → personaliza → compartilha → convidado acessa sem
cadastro → escolhe item → reserva ou compra → redirect `/go/{offerId}` → clique registrado →
provider de afiliado atribui → criador acompanha reservados/comprados.

## Princípios de experiência

- Criar lista é um onboarding de 3 passos, não um formulário de ERP.
- A página pública (`/l/{slug}`) é a página mais importante: rápida, mobile-first, sem login.
- "O usuário entende o próximo passo em menos de 3 segundos?" — se não, simplificar.
- Emoção, celebração, confiança e facilidade; nunca complexidade técnica.
- Transparência: divulgamos que lojas podem pagar comissão; nada de dark patterns.
- Recomendação orgânica nunca é distorcida por comissão sem transparência.

## Regras de negócio-chave

- `ListItem` existe sem produto de marketplace (ex.: "Dinheiro para lua de mel").
- Quantidades: `desired = purchased + reserved + available`; status derivado
  (`AVAILABLE`, `PARTIALLY_RESERVED`, `RESERVED`, `PURCHASED`, `ARCHIVED`).
- Reservas expiram (TTL configurável) e devolvem disponibilidade.
- Nome de quem reservou não é público, salvo opção do criador.
- **Presente surpresa:** criador vê apenas contagem agregada; convidados continuam vendo
  disponibilidade para evitar duplicidade.
- Compra em grupo: domínio será desenhado sem movimentação financeira até especificação legal.

## Métricas

| Métrica                | Definição                                                           |
| ---------------------- | ------------------------------------------------------------------- |
| Activation             | % usuários com uma lista com ≥ 3 itens                              |
| Share Rate             | % listas com `SHARE_CREATED`                                        |
| Visitor Conversion     | visitantes únicos → `OUTBOUND_CLICKED`                              |
| Reservation Conversion | visitantes únicos → `ITEM_RESERVED`                                 |
| Affiliate CTR          | `OFFER_VIEWED` → `OUTBOUND_CLICKED`                                 |
| Virality               | `USER_REGISTERED` com origem em lista compartilhada                 |
| GMV Outbound           | soma estimada de `price_cents` nos cliques (≠ receita de afiliados) |

**North Star:** número de interações de presente bem-sucedidas, aproximado por reservas +
cliques outbound qualificados.

## Monetização

Começa por **AFFILIATE**. Arquitetura não bloqueia PREMIUM, SPONSORED (sempre rotulado),
SUBSCRIPTION e TRANSACTION_FEE.
