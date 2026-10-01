# Integração: catálogo curado (`CURATED`)

Catálogo mantido por nós no painel **/admin/produtos** (tabelas `curated_products` e `curated_offers`).
Cada oferta tem o **link de afiliado que nós geramos no painel do programa** (ex.: `https://meli.la/…`).
Nenhuma API de loja é chamada, então **não há preço nem disponibilidade** (`NULL` / `UNKNOWN`).
A saída é sempre `/go/{offerId}` → o link exatamente como cadastrado (nada é acrescentado).

## Como adicionar um produto

1. Gere o link de afiliado no painel da loja.
2. Entre como administrador, abra **Produtos curados** (`/admin/produtos`) e clique em **Novo produto**.
3. Escolha a loja, cole o link e use **Ler link e preencher** para trazer título e imagem da página
   (o mesmo leitor do "colar link"; se a loja bloquear, preencha à mão). Ajuste nome, marca,
   categoria e palavras de busca.
4. Salve. O produto aparece na busca ("Buscar") imediatamente, sem redeploy.

- **Desativar** tira o produto da busca, mas itens que já estão em listas continuam funcionando
  (o `GetProduct` resolve também os inativos).
- **Remover** apaga do catálogo; ofertas já gravadas em listas mantêm o link.
- O `external_id` (`CUR-<uuid>`, ou o id antigo dos itens migrados) é estável e nunca muda na edição.
- Uma oferta por loja; só lojas cujo provider habilitado é `CURATED` aparecem no formulário.

## Quem é administrador

O papel `ADMIN` só é concedido por linha de comando, nunca pela interface:

```
cd apps/api && DATABASE_URL=… go run ./cmd/migrate admin voce@exemplo.com
```

`make seed` já torna o usuário de desenvolvimento (`tiago@listou.dev`) administrador.
Em Docker: `docker compose run --rm migrate admin voce@exemplo.com`.

## API (somente ADMIN; 403 `FORBIDDEN` para os demais)

| Método e rota                                | Descrição                                             |
| -------------------------------------------- | ----------------------------------------------------- |
| `GET /api/v1/admin/curated-products`         | `{ products, merchants }` (inclui inativos)           |
| `POST /api/v1/admin/curated-products`        | cria; corpo `{ title, brand, category, keywords, imageUrl, active, offers[{merchant,url}] }` |
| `PUT /api/v1/admin/curated-products/{id}`    | substitui os campos e as ofertas                      |
| `DELETE /api/v1/admin/curated-products/{id}` | remove (204)                                          |

Toda alteração grava `audit_logs` (`CURATED_PRODUCT_CREATED|UPDATED|DELETED`).

## Lojas

Só `MERCADO_LIVRE` usa `CURATED` hoje (migration 00008). `AMAZON` e `SHOPEE` seguem no `MOCK`
até haver links de afiliado para elas; para trocar, atualize `affiliate_providers.adapter` numa
nova migration (só um provider habilitado por merchant).

## Pontos de atenção

- **Imagens:** `imageUrl` aponta para o CDN da loja (hotlink). Os termos de cada programa sobre uso
  de imagens não foram verificados; confirme antes de ampliar o catálogo ou prefira imagens próprias.
- **Divulgação de afiliado:** os programas costumam exigir aviso de que há link de afiliado. Confira
  as regras de cada um e mostre o aviso no produto.
- Itens adicionados antes da troca de provider podem manter ofertas de demonstração do `MOCK`.
- Editar o link de um produto atualiza a oferta quando o produto for importado de novo; itens já
  materializados em listas só pegam o novo link na próxima importação.
