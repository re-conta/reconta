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

## Testes

```sh
# Frontend
bun test

# Backend Go
bun run api:test
```
