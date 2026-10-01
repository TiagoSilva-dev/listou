# Roadmap

| #   | Milestone             | Escopo                                                                                                                       | Status |
| --- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------ |
| 0   | Foundation            | monorepo, docs, CLAUDE.md, API base (health/ready, request id, logs, erros), migrations, design system, Docker, Makefile, CI | ✅     |
| 1   | Authentication        | registro/login/logout/me, sessões, rate limit, páginas entrar/criar conta                                                    | ✅     |
| 2   | Events                | CRUD de eventos, onboarding de criação em 3 passos, slug                                                                     | ✅     |
| 3   | Gift Lists            | lista principal, categorias, itens manuais, quantidades                                                                      | ✅     |
| 4   | Products              | catálogo, MockAffiliateProvider, busca, associar produto, ofertas, seed                                                      | ✅     |
| 5   | Public Registry       | `/l/{slug}` SSR, filtros, busca, drawer de produto, SEO                                                                      | ✅     |
| 6   | Reservations          | reservar/comprar/cancelar, expiração, presente surpresa                                                                      | ✅     |
| 7   | Affiliate Redirect    | `/go/{offerId}`, registro de clique                                                                                          | ✅     |
| 8   | Sharing               | copiar link, WhatsApp, e-mail, QR Code, Web Share, OG image                                                                  | ✅     |
| 9   | Analytics             | eventos de produto, dashboard do criador                                                                                     | 🟡     |
| 10  | AI Builder            | `RecommendationProvider`, construtor com IA (flag `AI_LIST_BUILDER`)                                                         | ✅     |
| 11  | Real Marketplace      | adicionar por link colado (`LINK`) ✅; 1º provider de API real após leitura da documentação oficial                         | 🟡     |
| 12  | Growth & Optimization | templates, comparação de ofertas, Jev matching se provar ganho                                                               | —      |

Primeiro release utilizável = milestones 1–9.

## Notas do primeiro release

- M9 (🟡): eventos e dashboard do criador prontos; falta o funil agregado (Activation, Share Rate, CTR) como consultas/painel admin.
- M8: copiar link, WhatsApp, e-mail, Instagram (copia o link), Web Share, QR Code e OG image dinâmica. Falta SHARE_OPENED via parâmetro de origem.
- Pendências conhecidas: upload de imagens (hoje só URL), verificação de e-mail, recuperação de senha, admin básico, `show_reserver_names` (coluna existe; sem UI), notificações (reserva expirando), job de expiração roda a cada minuto dentro da API.
