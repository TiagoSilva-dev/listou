# Segurança e privacidade

## Implementado

- **Sessões:** cookie `listou_session` `HttpOnly`, `SameSite=Lax`, `Secure` em produção; token de
  256 bits; banco guarda apenas SHA-256; logout revoga.
- **Senhas:** argon2id com salt aleatório; mínimo 8 caracteres; comparação em tempo constante;
  mensagens de login genéricas (não revelam se o e-mail existe).
- **CSRF:** `SameSite=Lax` + middleware `SameOrigin` (rejeita `Origin` não permitido em métodos
  mutáveis).
- **Rate limiting:** por IP e rota sensível (login, registro, reservas, analytics).
- **Validação de entrada:** JSON com limite de 1 MB e campos desconhecidos rejeitados;
  validação de domínio em cada service; Zod no frontend.
- **SQL injection:** apenas queries parametrizadas (pgx).
- **Output encoding:** React escapa por padrão; API devolve só JSON.
- **Headers:** `nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy`, CSP restritiva na API, HSTS
  em produção, `Permissions-Policy` no web.
- **Autorização:** todo acesso a evento/lista/item resolve o dono a partir do banco — IDs do
  frontend nunca são confiados; rotas públicas são read-only e filtram campos.
- **Reservas:** token de gestão aleatório entregue uma única vez ao convidado (hash no banco).
- **Erros:** envelope padrão com `requestId`; nunca stack trace; panics recuperados.
- **Logs:** sem senhas, tokens, corpos ou query strings.
- **Audit log** para publicar/excluir evento e cancelamentos feitos pelo dono.
- **Secrets:** só via ambiente; `.env` no `.gitignore`; `.env.example` sem valores.

## Privacidade (LGPD)

- Analytics first-party, sem Google Analytics para métricas de negócio.
- Visitante identificado por UUID aleatório em cookie first-party (`listou_vid`), sem
  fingerprinting, sem IP armazenado; referrer guardado apenas como host.
- Nome de quem reservou não é exibido publicamente salvo opção do criador.
- Contato do convidado é opcional e só visível ao criador (fora do modo surpresa).
- Listas `PRIVATE` não são acessíveis publicamente; `UNLISTED` e `PRIVATE` recebem `noindex`.

## Pendências conhecidas

- Upload de imagens (validação de tipo/tamanho + URLs assinadas S3/R2) — milestone futuro.
- Banner de consentimento quando houver cookies não essenciais.
- Verificação de e-mail e recuperação de senha.
