# Reconta

Aplicação web para controle financeiro pessoal — gerenciamento de contas a pagar/receber, categorização de transações e notificações de vencimento.

**Site:** [reconta.app](https://reconta.app)

## Stack

- **Frontend:** Vue.js 3 + Vite + TypeScript (`web/`)
- **Backend:** Go 1.26+ (`api/`)
- **Runtime JS:** Bun
- **Banco de dados:** SQLite
- **Servidor:** VPS Linux com Nginx + systemd

## Estrutura

```
reconta/
├── web/          # Frontend Vue.js 3 + Vite
├── api/          # Backend Go
├── files/        # Configurações Nginx e systemd para a VPS
├── scripts/      # Scripts de deploy
└── .github/      # Workflows de CI/CD
```

## Desenvolvimento

Pré-requisitos: [Bun](https://bun.sh) e [Go 1.26+](https://go.dev)

```sh
# Instalar dependências JS
bun install

# Iniciar frontend (Vite dev server)
bun run dev

# Iniciar backend Go
bun run api:dev
```

O frontend roda em `http://localhost:5173` por padrão.  
A API Go roda em `http://localhost:3020` por padrão.

### Rodando em `https://reconta.local` (domínio local)

Para desenvolver com um domínio e HTTPS parecidos com produção, use o Makefile na raiz do projeto.

Pré-requisito único: `openssl` (já vem instalado na maioria das distros/macOS).

```sh
make dev
```

Esse comando:

1. **Gera um certificado TLS autoassinado** para `reconta.local` em `certs/` (via `make certs`), caso ainda não exista. O certificado é local e ignorado pelo Git.
2. **Verifica** se `reconta.local` resolve para `127.0.0.1` em `/etc/hosts` (via `make hosts-check`) e, se não resolver, imprime o comando para adicionar:
   ```sh
   sudo sh -c 'echo "127.0.0.1 reconta.local" >> /etc/hosts'
   ```
3. Sobe a API Go e o Vite (via `bun run dev`). O Vite detecta o certificado em `certs/` automaticamente e passa a servir em HTTPS.

Depois disso, acesse **https://reconta.local:5173**. Como o certificado é autoassinado, o navegador vai exibir um aviso de segurança na primeira visita — é esperado, basta prosseguir/confiar manualmente (não há CA local instalada).

Alvos individuais do Makefile, se quiser rodar as etapas separadamente:

```sh
make certs        # gera o certificado, se não existir
make hosts-check  # confere a entrada em /etc/hosts
```

### Simulando produção com Podman (`make up`)

Além do dev server do Vite, o Makefile também sobe API + Nginx/Web em containers Podman, num único pod, reaproveitando o mesmo certificado de `certs/`. Serve **apenas na porta 443** (sem HTTP nem porta alternativa):

```sh
make up      # builda as imagens e sobe o pod em https://reconta.local
make logs    # segue os logs de API e Web
make down    # para e remove o pod (mantém o volume com o banco)
```

Como a porta 443 é privilegiada, em algumas distros o Podman rootless pode recusar o bind. Se acontecer, ajuste:

```sh
sudo sysctl net.ipv4.ip_unprivileged_port_start=443
```

## Build

```sh
# Build do frontend
bun run build

# Build da API Go
bun run api:build   # gera api/bin/server
```

## Deploy

O deploy é automático: push para `main` dispara o GitHub Actions que envia os arquivos para a VPS via SCP e executa `scripts/deploy.sh` remotamente.

O script de deploy:

1. Preserva `.env` e o banco SQLite
2. Instala dependências e gera o build de produção
3. Para o serviço, substitui os arquivos e reinicia

Segredos necessários no repositório: `SSH_HOST`, `SSH_USER`, `SSH_PASS`, `SSH_PORT`, `PROJECT_PATH`.

## Login via Google (variáveis de ambiente)

O login via Google é opcional: se `GOOGLE_CLIENT_ID` estiver vazio no `.env`, essa opção fica desabilitada e a aplicação funciona normalmente só com login por email/senha.

As variáveis usadas (`api/.env`, veja `api/.env.example`):

```sh
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:3020/api/auth/google/callback
```

### Como gerar as credenciais no Google Cloud Console

1. **Crie (ou selecione) um projeto**
   - Acesse [console.cloud.google.com](https://console.cloud.google.com/)
   - No topo, clique no seletor de projetos → **Novo Projeto**
   - Dê um nome (ex.: `reconta`) e clique em **Criar**

2. **Configure a tela de consentimento OAuth**
   - No menu lateral: **APIs e Serviços** → **Tela de consentimento OAuth**
   - Escolha o tipo **Externo** (a menos que você use Google Workspace) e clique em **Criar**
   - Preencha os campos obrigatórios:
     - Nome do app: `Reconta`
     - Email de suporte do usuário: seu email
     - Email de contato do desenvolvedor: seu email
   - Em **Escopos**, adicione `openid`, `.../auth/userinfo.email` e `.../auth/userinfo.profile` (são os escopos que a API usa — veja `api/internal/auth/google.go`)
   - Em **Usuários de teste** (se o app estiver em modo "Teste"), adicione os emails do Google que farão login durante o desenvolvimento
   - Salve e continue até finalizar

3. **Crie as credenciais OAuth 2.0**
   - Menu lateral: **APIs e Serviços** → **Credenciais**
   - Clique em **Criar Credenciais** → **ID do cliente OAuth**
   - Tipo de aplicativo: **Aplicativo da Web**
   - Nome: `Reconta Web` (ou outro nome interno, sem impacto funcional)
   - Em **URIs de redirecionamento autorizados**, adicione as URLs de callback correspondentes a cada ambiente:
     - Desenvolvimento: `http://localhost:3020/api/auth/google/callback`
     - Produção: `https://reconta.app/api/auth/google/callback`
   - Clique em **Criar**

4. **Copie o Client ID e o Client Secret**
   - Um modal exibirá o **ID do cliente** e a **Chave secreta do cliente**
   - Copie esses valores para `api/.env`:
     ```sh
     GOOGLE_CLIENT_ID=seu-client-id.apps.googleusercontent.com
     GOOGLE_CLIENT_SECRET=sua-chave-secreta
     GOOGLE_REDIRECT_URL=http://localhost:3020/api/auth/google/callback
     ```
   - Se perder a chave depois, você pode gerar uma nova em **Credenciais** → clique no ID do cliente criado → **Adicionar chave secreta**

5. **Publique o app (opcional, para produção)**
   - Enquanto a tela de consentimento estiver em modo **Teste**, só os usuários de teste cadastrados no passo 2 conseguem logar
   - Para permitir qualquer conta Google, volte em **Tela de consentimento OAuth** e clique em **Publicar aplicativo**
   - Se os escopos usados forem apenas `openid`/`email`/`profile` (não sensíveis), a publicação costuma ser aprovada automaticamente, sem revisão manual do Google

### Configuração em produção (VPS)

O `api/.env` de produção precisa das mesmas três variáveis, mas com `GOOGLE_REDIRECT_URL` apontando para o domínio real (`https://reconta.app/api/auth/google/callback`) — essa URL **precisa** estar cadastrada nos "URIs de redirecionamento autorizados" do passo 3, senão o Google recusa o callback com erro `redirect_uri_mismatch`.

Lembre-se: o script de deploy (`scripts/deploy.sh`) preserva o `.env` existente na VPS entre deploys, então essas variáveis só precisam ser configuradas manualmente uma vez no servidor.

## Planos e assinaturas (Mercado Pago)

O site tem três planos, exibidos em `/planos`: **Gratuito**, **Essencial** e **Profissional**. Os dois pagos aceitam assinatura **mensal ou anual**, com pagamento via **PIX, boleto, cartão de débito ou cartão de crédito**, processado pelo [Mercado Pago](https://www.mercadopago.com.br) (Checkout API / Checkout Transparente — o usuário não sai do site).

### Como funciona

- **Página `/planos`**: pública, com toggle mensal/anual. Preços, descrições e benefícios dos planos vêm do banco (tabela `plans`) e são editáveis no painel de admin (aba **Planos**, permissão `manage_plans`). O plano gratuito nunca tem preço.
- **Checkout**: modal responsivo na própria página. PIX gera QR Code + copia-e-cola (expira em 30 min); boleto gera o link do PDF (vence em 3 dias e exige CPF/CNPJ + endereço); cartões (débito e crédito) são tokenizados no navegador pelo SDK JS do Mercado Pago — os dados do cartão **nunca passam pelo nosso backend** — com detecção automática de bandeira pelos 6 primeiros dígitos.
- **Confirmação**: via webhook (`POST /api/billing/webhook`, com validação de assinatura `x-signature` quando `MP_WEBHOOK_SECRET` está definido) e também por polling do modal (`GET /api/billing/payments/{id}`), que reconsulta o Mercado Pago — útil em desenvolvimento, onde o webhook não alcança a máquina local.
- **Renovação**: as assinaturas não renovam sozinhas (PIX e boleto não permitem cobrança automática). Em vez disso, o usuário recebe **lembretes no site (sino/SSE) e por e-mail 7, 3 e 1 dia antes do vencimento**, com link para renovar em `/planos`. Renovar antes do fim do ciclo soma o novo período ao atual — nenhum dia é perdido.
- **Cobranças/lembretes via systemd**: a varredura roda pela **mesma unit e timer já existentes** (`reconta-notifications.timer` → `reconta-notifications.service`, de hora em hora). A unit ganhou um segundo `ExecStart` que chama `POST /api/internal/billing/scan` (protegido por `INTERNAL_API_TOKEN`), que envia os lembretes e expira assinaturas vencidas (o usuário volta ao plano Gratuito). Após alterar a unit na VPS: `systemctl daemon-reload`.
- **Cancelamento** (em Configurações → Plano e assinatura), a qualquer momento, em dois modos:
  - **Usar até o fim do ciclo**: sem reembolso; a assinatura fica marcada para não renovar e expira sozinha na data.
  - **Cancelar agora com reembolso parcial**: o acesso termina na hora e o backend solicita ao Mercado Pago um reembolso proporcional ao tempo não usado do ciclo (calculado sobre o último pagamento aprovado).

### Tabelas

| Tabela                  | Conteúdo                                                                 |
| ----------------------- | ------------------------------------------------------------------------ |
| `plans`                 | Código, nome, preços mensal/anual, benefícios (JSON) e destaque          |
| `subscriptions`         | Assinatura por usuário: plano, ciclo, status, fim do período, lembretes  |
| `subscription_payments` | Cada cobrança: id do pagamento no MP, status, QR PIX, link do boleto     |

### Variáveis de ambiente

No `api/.env` (backend):

```sh
MP_ACCESS_TOKEN=      # Access Token (credencial privada) do Mercado Pago
MP_WEBHOOK_SECRET=    # Assinatura secreta do webhook (opcional, recomendado em produção)
```

No `web/.env` / `web/.env.production` (frontend, usado só para tokenizar cartões):

```sh
VITE_MP_PUBLIC_KEY=   # Public Key do Mercado Pago
```

Sem `MP_ACCESS_TOKEN`, o site funciona normalmente e apenas o checkout fica desabilitado (HTTP 503).

### Guia: obtendo as chaves no site do Mercado Pago

1. **Crie/acesse sua conta** em [mercadopago.com.br](https://www.mercadopago.com.br) e entre no painel de desenvolvedores: [mercadopago.com.br/developers](https://www.mercadopago.com.br/developers/pt) → **Suas integrações**.
2. **Crie uma aplicação**: botão **Criar aplicação** → nome `Reconta` → em "Qual tipo de solução?" escolha **Pagamentos on-line** → **CheckoutAPI (Transparente)** → confirme.
3. **Credenciais de teste** (para desenvolver): dentro da aplicação, menu **Credenciais de teste**. Copie:
   - **Public Key** → `VITE_MP_PUBLIC_KEY` no `web/.env`
   - **Access Token** → `MP_ACCESS_TOKEN` no `api/.env`
   - Com credenciais de teste, use os [cartões de teste](https://www.mercadopago.com.br/developers/pt/docs/checkout-api/additional-content/your-integrations/test/cards) do Mercado Pago (ex.: aprovar com titular `APRO`, recusar com `OTHE`).
4. **Credenciais de produção**: menu **Credenciais de produção**. O Mercado Pago pede um breve formulário (indústria/site) na primeira vez. Copie a **Public Key** para `web/.env.production` e o **Access Token** para o `api/.env` da VPS (o deploy preserva o `.env` entre publicações).
5. **Configure o webhook**: na aplicação, menu **Webhooks** → **Configurar notificações**:
   - URL de produção: `https://reconta.app/api/billing/webhook`
   - Evento: marque **Pagamentos** (`payment`)
   - Salve e copie a **assinatura secreta** exibida → `MP_WEBHOOK_SECRET` no `api/.env` da VPS.
6. **Reinicie o serviço** na VPS após editar o `.env`: `sudo systemctl restart reconta`.

> Dica: em desenvolvimento não é preciso webhook — o modal de checkout faz polling e reconsulta o status direto na API do Mercado Pago.

## Estatísticas de visitas (painel de admin)

O painel `/admin` tem uma aba **Estatísticas** com visitas únicas, visitas totais, novos vs. recorrentes, série diária (gráfico com range de datas selecionável), páginas mais visitadas, referrers, navegador/SO/dispositivo, localização por IP e uma tabela de visitas recentes (com IP, país/cidade, navegador, SO e referrer). Também mostra um indicador de "ativos agora" (visitantes únicos nos últimos 5 minutos).

### Como funciona

- Como o site é uma SPA (o Nginx serve `index.html` para qualquer rota), não existe log de navegação por página no servidor — cada troca de rota é reportada por um beacon (`POST /api/track`, disparado em `router.afterEach` no front-end). O beacon nunca bloqueia nem falha visivelmente a navegação.
- **Visitante único**: cookie próprio `rc_vid` (HttpOnly, 1 ano) gerado pelo backend na primeira visita. `rc_sid` é um cookie de sessão do navegador (sem expiração fixa), usado para agrupar visitas da mesma sessão.
- **IP real**: o handler lê, nessa ordem, `CF-Connecting-IP` → `X-Real-IP` → o IP da conexão TCP. Ver seção de Nginx abaixo para o IP do Cloudflare chegar corretamente.
- **Geolocalização**: via [GeoLite2 City](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) da MaxMind, lido de um arquivo `.mmdb` local (`GEOIP_DB_PATH`). Se a variável não estiver definida ou o arquivo não existir, a geolocalização fica desabilitada e os campos de país/cidade ficam vazios — o resto do rastreamento continua funcionando normalmente.
- **Navegador/SO/dispositivo**: parseados do `User-Agent` no backend (`github.com/mileusna/useragent`). Bots conhecidos são marcados (`is_bot`) e excluídos de todas as estatísticas.
- A aba **Estatísticas** é visível a qualquer usuário com a permissão `admin_panel` (mesmo critério de acesso ao restante do painel); os dados de IP/localização são de uso interno/administrativo.

### Tabelas

| Tabela        | Conteúdo                                                                                     |
| ------------- | ---------------------------------------------------------------------------------------------- |
| `page_visits` | Uma linha por navegação: visitante/sessão, path, referrer, IP, geo, user agent, browser/SO/dispositivo, timestamp |

### Variáveis de ambiente

No `api/.env` (backend):

```sh
GEOIP_DB_PATH=   # Caminho absoluto do GeoLite2-City.mmdb (opcional — sem ele, geolocalização fica desabilitada)
```

### Configuração em produção (VPS)

**1. IP real do Cloudflare no Nginx** — o site roda atrás do Cloudflare, então por padrão o Nginx (e o backend) só enxergam o IP da borda do Cloudflare, não o do visitante. `files/cloudflare-realip.conf` configura o [`ngx_http_realip_module`](https://nginx.org/en/docs/http/ngx_http_realip_module.html) para confiar no cabeçalho `CF-Connecting-IP` **somente** quando a requisição vem de uma faixa de IP publicada pelo Cloudflare (evita spoofing por quem não vier da borda deles). `files/reconta.conf` inclui esse arquivo no server block principal.

   > ⚠️ O `scripts/deploy.sh` **não** copia nem recarrega a configuração do Nginx — só o app (API + frontend). Mudanças em `files/reconta.conf` ou `files/cloudflare-realip.conf` precisam ser aplicadas manualmente na VPS. Na VPS de produção, o site fica em `/etc/nginx/sites.d/40-reconta.app.conf` (carregado via `include /etc/nginx/sites.d/*.conf` no `nginx.conf` principal — `/etc/nginx/conf.d/` **não** é incluído, apesar de existir um arquivo antigo lá):
   > ```sh
   > scp files/reconta.conf root@<vps>:/etc/nginx/sites.d/40-reconta.app.conf
   > scp files/cloudflare-realip.conf root@<vps>:/etc/nginx/sites.d/cloudflare-realip.conf
   > ssh root@<vps> "nginx -t && systemctl reload nginx"
   > ```
   > As faixas de IP do Cloudflare mudam raramente, mas convém revisar de tempos em tempos contra [cloudflare.com/ips-v4](https://www.cloudflare.com/ips-v4/) e [/ips-v6](https://www.cloudflare.com/ips-v6/).

**2. GeoLite2 via `geoipupdate`** — a ferramenta oficial da MaxMind. Em distros sem pacote pronto (ex.: Rocky/RHEL sem repo próprio), instalar a partir do [release oficial no GitHub](https://github.com/maxmind/geoipupdate/releases):

   ```sh
   curl -sL -o /tmp/geoipupdate.rpm https://github.com/maxmind/geoipupdate/releases/latest/download/geoipupdate_8.0.0_linux_amd64.rpm
   sudo dnf install -y /tmp/geoipupdate.rpm   # ou dpkg -i / apt install geoipupdate em Debian/Ubuntu
   sudo nano /etc/GeoIP.conf
   ```
   ```
   AccountID   SEU_ACCOUNT_ID
   LicenseKey  SUA_LICENSE_KEY
   EditionIDs  GeoLite2-City
   ```
   ```sh
   sudo geoipupdate -v   # baixa o banco pela primeira vez (grava em /usr/share/GeoIP por padrão)
   ```

   O binário **não vem com timer systemd próprio** — `files/geoipupdate.service` e `files/geoipupdate.timer` cobrem isso (roda 1x/dia, com atraso aleatório de até 1h para não bater exatamente na hora cheia):

   ```sh
   scp files/geoipupdate.service files/geoipupdate.timer root@<vps>:/etc/systemd/system/
   ssh root@<vps> "systemctl daemon-reload && systemctl enable --now geoipupdate.timer"
   ```

   Por padrão o `geoipupdate` grava em `/usr/share/GeoIP/GeoLite2-City.mmdb` (verifique `DatabaseDirectory` em `/etc/GeoIP.conf` se for diferente). Defina esse caminho em `GEOIP_DB_PATH` no `.env` de produção da API. O backend recarrega o arquivo do disco a cada hora sozinho, então atualizações do `geoipupdate` são pegas sem reiniciar o serviço.

### Privacidade

Os dados de IP, geolocalização e User-Agent são coletados para fins administrativos internos (métricas de uso e segurança), não são compartilhados com terceiros e ficam restritos a quem tem a permissão `admin_panel`. Antes de usar em produção com tráfego real, vale revisar com um jurídico se a política de privacidade do site precisa mencionar essa coleta (LGPD, interesse legítimo).

## Endpoints da API

Todas as rotas usam o prefixo `/api` (sem versionamento). Rotas marcadas como **protegidas** exigem sessão autenticada (cookie `session_token`), validada pelo middleware `auth.Handler.RequireUser()`.

### Health check

| Método | Rota          | Auth |
| ------ | ------------- | ---- |
| GET    | `/api/health` | Não  |

### Usuários

| Método | Rota         | Auth |
| ------ | ------------ | ---- |
| POST   | `/api/users` | Não  |
| GET    | `/api/users` | Não  |

### Autenticação

| Método | Rota                        | Auth |
| ------ | --------------------------- | ---- |
| POST   | `/api/auth/login`           | Não  |
| POST   | `/api/auth/logout`          | Não  |
| GET    | `/api/auth/me`              | Não  |
| GET    | `/api/auth/google/login`    | Não  |
| GET    | `/api/auth/google/callback` | Não  |

### Contas (`accounts`)

| Método | Rota                 | Auth |
| ------ | -------------------- | ---- |
| GET    | `/api/accounts`      | Sim  |
| POST   | `/api/accounts`      | Sim  |
| PUT    | `/api/accounts/{id}` | Sim  |
| DELETE | `/api/accounts/{id}` | Sim  |

### Categorias (`categories`)

| Método | Rota                   | Auth |
| ------ | ---------------------- | ---- |
| GET    | `/api/categories`      | Sim  |
| POST   | `/api/categories`      | Sim  |
| PUT    | `/api/categories/{id}` | Sim  |
| DELETE | `/api/categories/{id}` | Sim  |

### Tags

| Método | Rota             | Auth |
| ------ | ---------------- | ---- |
| GET    | `/api/tags`      | Sim  |
| POST   | `/api/tags`      | Sim  |
| PUT    | `/api/tags/{id}` | Sim  |
| DELETE | `/api/tags/{id}` | Sim  |

### Transações (`transactions`)

| Método | Rota                                | Auth |
| ------ | ----------------------------------- | ---- |
| GET    | `/api/transactions`                 | Sim  |
| POST   | `/api/transactions`                 | Sim  |
| PATCH  | `/api/transactions`                 | Sim  |
| DELETE | `/api/transactions`                 | Sim  |
| POST   | `/api/transactions/auto-categorize` | Sim  |
| GET    | `/api/transactions/opening-balance` | Sim  |
| POST   | `/api/transactions/opening-balance` | Sim  |
| GET    | `/api/transactions/{id}`            | Sim  |
| PUT    | `/api/transactions/{id}`            | Sim  |
| DELETE | `/api/transactions/{id}`            | Sim  |

### Planos e assinaturas (`billing`)

| Método | Rota                                | Auth                        |
| ------ | ----------------------------------- | --------------------------- |
| GET    | `/api/plans`                        | Não                         |
| GET    | `/api/billing/subscription`         | Sim                         |
| POST   | `/api/billing/subscribe`            | Sim                         |
| GET    | `/api/billing/payments/{id}`        | Sim                         |
| POST   | `/api/billing/subscription/cancel`  | Sim                         |
| POST   | `/api/billing/webhook`              | Assinatura do Mercado Pago  |
| POST   | `/api/internal/billing/scan`        | Header `X-Internal-Token`   |
| GET    | `/api/admin/plans`                  | Permissão `manage_plans`    |
| PUT    | `/api/admin/plans/{id}`             | Permissão `manage_plans`    |

### Estatísticas de visitas (`analytics`)

| Método | Rota                             | Auth                       |
| ------ | --------------------------------- | --------------------------- |
| POST   | `/api/track`                      | Não                         |
| GET    | `/api/admin/analytics/overview`   | Permissão `admin_panel`     |
| GET    | `/api/admin/analytics/pages`      | Permissão `admin_panel`     |
| GET    | `/api/admin/analytics/referrers`  | Permissão `admin_panel`     |
| GET    | `/api/admin/analytics/locations`  | Permissão `admin_panel`     |
| GET    | `/api/admin/analytics/devices`    | Permissão `admin_panel`     |
| GET    | `/api/admin/analytics/visitors`   | Permissão `admin_panel`     |
| GET    | `/api/admin/analytics/active`     | Permissão `admin_panel`     |

## Testes

```sh
# Frontend
bun test

# Backend Go
bun run api:test
```
