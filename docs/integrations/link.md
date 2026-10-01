# Integração: link colado (`LINK`)

Não é integração com marketplace: nenhuma API, credencial ou programa de afiliados é usado. A
decisão está na ADR-0011.

## Fluxo

1. `POST /products/link-preview {url}`: valida, baixa a página e devolve título, imagem e nome da
   loja (`og:site_name` ou host). `readable=false` quando não há título útil.
2. `POST /products/import-link {url,title?}`: grava `products` + um `product_offers` do merchant
   `LINK` (preço `NULL`, disponibilidade `UNKNOWN`, `metadata.storeName`). Dedupe pela URL canônica.
3. O item da lista é criado como qualquer produto (`productId`). O presenteador sai por `/go/{offerId}`.

## O que é lido da página

Ordem de prioridade do título: JSON-LD `Product.name` → `og:title` → `twitter:title` → `<title>`.
Imagem: JSON-LD → `og:image` → `twitter:image` (só `http(s)`, resolvida contra a URL final).
Também marca e descrição. **Preço, disponibilidade e rating não são lidos.** Título igual ao nome/host
da loja é descartado (página inicial, captcha, erro).

## Proteções (SSRF e abuso)

- Só `http`/`https`, portas 80/443, sem credenciais na URL, até 2048 caracteres.
- IP de destino é verificado **na conexão** (dialer `Control`), então vale para DNS que aponta para
  rede interna, rebinding e redirecionamentos: bloqueia loopback, privados, link-local (inclui
  `169.254.169.254`), CGNAT, multicast, reservados.
- Sem proxy de ambiente; até 5 redirecionamentos, cada um revalidado; timeout total de 8 s;
  corpo limitado a 1 MiB; só `text/html`; status 2xx.
- Rate limit de 20 requisições/min por IP nos dois endpoints (em memória, ADR-0004).
- Imagem remota é só URL guardada; o navegador a carrega direto (sem proxy de imagens).

## Limitações conhecidas

- Várias lojas (Amazon, Shopee, às vezes Mercado Livre) bloqueiam leitura automática ou devolvem
  páginas sem metadados do produto; nesse caso o usuário digita o nome.
- A URL colada é usada como veio (sem tag de afiliado) até existir o formato aprovado por programa.
- Respeite `robots.txt`/termos de cada loja ao evoluir isso; não há cache nem crawling em lote.
