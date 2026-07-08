# Reconta API

API REST em Go que dá suporte ao Reconta, uma aplicação de controle financeiro pessoal (contas, categorias, tags, transações, importação de extratos bancários em PDF e geração de relatórios/backup).

## Sumário

- [Visão geral](#visão-geral)
- [Autenticação e sessão](#autenticação-e-sessão)
- [Convenções gerais](#convenções-gerais)
- [Endpoints](#endpoints)
  - [Health check](#health-check)
  - [Autenticação](#autenticação)
  - [Login com Google (OAuth2)](#login-com-google-oauth2)
  - [Usuários](#usuários)
  - [Contas (accounts)](#contas-accounts)
  - [Categorias](#categorias)
  - [Tags](#tags)
  - [Transações](#transações)
  - [Importação de extratos (PDF)](#importação-de-extratos-pdf)
  - [Relatórios e backup](#relatórios-e-backup)
  - [Contas Recorrentes](#contas-fixas)
  - [Notificações](#notificações)
- [Modelos de dados](#modelos-de-dados)
- [Variáveis de ambiente](#variáveis-de-ambiente)
- [Executando localmente](#executando-localmente)
- [Timer systemd de notificações (produção)](#timer-systemd-de-notificações-produção)

---

## Visão geral

A API é servida por um único binário Go (`main.go`) que registra rotas em um `http.ServeMux` nativo (Go 1.22+, com suporte a métodos e path params). Cada domínio de negócio vive em seu próprio pacote dentro de `internal/`:

| Pacote                  | Responsabilidade                                                                  |
| ----------------------- | --------------------------------------------------------------------------------- |
| `internal/auth`         | Login por e-mail/senha, sessão via cookie, login com Google OAuth2                |
| `internal/user`         | CRUD de usuários, papéis (roles), perfil e senha                                  |
| `internal/account`      | Contas bancárias/carteiras do usuário                                             |
| `internal/category`     | Categorias de transação (com padrões de auto-categorização)                       |
| `internal/tag`          | Etiquetas livres associáveis a transações                                         |
| `internal/transaction`  | Lançamentos financeiros (receitas/despesas), filtros, saldo de abertura           |
| `internal/statement`    | Extração e parsing de extratos bancários em PDF para importação                   |
| `internal/report`       | Exportação de relatórios (XLSX/ODS/PDF/JSON) e importação de backup JSON          |
| `internal/seed`         | Popula categorias/conta padrão para novos usuários                                |
| `internal/db`           | Conexão e migrações do banco SQLite                                               |
| `internal/fixedbill`    | Contas fixas (despesas recorrentes): ciclo de vida e pagamentos                   |
| `internal/notification` | Notificações de contas fixas (site em tempo real via SSE + e-mail) e preferências |
| `internal/email`        | Envio de e-mail via SMTP (`net/smtp`)                                             |

O banco de dados é SQLite (caminho configurável via `DB_PATH`).

## Autenticação e sessão

A autenticação é feita por **cookie de sessão HTTP-only** (`session_token`), gerado no login e validado a cada requisição autenticada.

1. **Login por e-mail/senha**: `POST /api/auth/login` valida a senha com bcrypt e cria a sessão.
2. **Login com Google**: fluxo OAuth2 (`/api/auth/google/login` → Google → `/api/auth/google/callback`), disponível apenas se `GOOGLE_CLIENT_ID` estiver configurado.
3. Em ambos os casos, o servidor grava o cookie `session_token` (HttpOnly, SameSite=Lax, Secure quando `ENV=production`), válido por **7 dias**.
4. Rotas protegidas usam o middleware `auth.Handler.RequireUser`, que resolve o usuário a partir do cookie e injeta o `userID` no handler; se a sessão for inválida/ausente, retornam `401 Unauthorized`.
5. Rotas administrativas usam `requireRole`, retornando `403 Forbidden` se o papel do usuário não for suficiente.

Papéis (`role`) existentes: `user`, `admin`, `super_admin`.

Como o cookie é a única forma de autenticação, o frontend deve enviar as requisições com `credentials: "include"` (o CORS do servidor já habilita `Access-Control-Allow-Credentials: true` refletindo a origem da requisição).

## Convenções gerais

- **Base path**: todas as rotas ficam sob `/api`.
- **Formato**: request e response bodies em JSON (`Content-Type: application/json`), exceto uploads (`multipart/form-data`) e downloads de relatório (binário/octet-stream conforme o formato).
- **Erros**: respostas de erro seguem o formato `{"error": "mensagem"}` com o status HTTP apropriado (`400`, `401`, `403`, `404`, `409`, `422`, `500`).
- **Autenticação**: coluna "Auth" na tabela de endpoints indica o nível de proteção:
  - `Público` — sem autenticação.
  - `Sessão` — requer cookie de sessão válido (qualquer papel).
  - `admin+` — requer papel `admin` ou `super_admin`.
  - `super_admin` — requer papel `super_admin`.
- **Escopo por usuário**: contas, categorias, tags e transações são sempre filtradas pelo `userID` da sessão — um usuário nunca acessa dados de outro.
- **IDs em rota**: `{id}` é sempre um inteiro (`int64`); IDs inválidos retornam `400`.

---

## Endpoints

### Health check

| Método | Rota          | Auth    | Descrição                                                     |
| ------ | ------------- | ------- | ------------------------------------------------------------- |
| GET    | `/api/health` | Público | Verifica se o servidor está no ar. Retorna `{"status":"ok"}`. |

### Autenticação

| Método | Rota               | Auth    | Body / Parâmetros                         | Descrição                                                                   |
| ------ | ------------------ | ------- | ----------------------------------------- | --------------------------------------------------------------------------- |
| POST   | `/api/auth/login`  | Público | `{ "email": string, "password": string }` | Autentica por e-mail/senha, cria sessão e grava o cookie. Retorna o `User`. |
| POST   | `/api/auth/logout` | Sessão  | —                                         | Remove a sessão atual e limpa o cookie. Retorna `204 No Content`.           |
| GET    | `/api/auth/me`     | Sessão  | —                                         | Retorna o usuário autenticado (`User`).                                     |

Respostas de erro do login: `400` (corpo inválido), `401` (e-mail/senha incorretos).

### Login com Google (OAuth2)

Disponível apenas quando a variável `GOOGLE_CLIENT_ID` está configurada.

| Método | Rota                        | Auth    | Parâmetros             | Descrição                                                                                                                                                                                        |
| ------ | --------------------------- | ------- | ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| GET    | `/api/auth/google/login`    | Público | —                      | Gera o `state` OAuth2, grava cookie temporário e redireciona (`302`) para a tela de consentimento do Google.                                                                                     |
| GET    | `/api/auth/google/callback` | Público | Query: `state`, `code` | Callback do Google. Troca o `code` por token, busca o perfil, cria/vincula o usuário e a sessão, e redireciona para `APP_URL`. Em caso de erro, redireciona para `APP_URL/login?error=<motivo>`. |

Comportamento de vinculação de conta:

- Se já existir um usuário com o mesmo `google_id`, reutiliza-o (atualiza `avatarUrl`).
- Se existir um usuário com o mesmo e-mail (cadastrado via senha), vincula o `google_id` a ele.
- Caso contrário, cria um novo usuário e executa o callback de seed de dados padrão (contas/categorias).

### Usuários

| Método | Rota                     | Auth        | Body / Parâmetros                                                  | Descrição                                                                                                     |
| ------ | ------------------------ | ----------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------- |
| POST   | `/api/users`             | Público     | `{ "name": string, "email": string, "password": string (min. 8) }` | Cria um novo usuário (cadastro). Popula categorias/conta padrão. Retorna `201` com o `User`.                  |
| GET    | `/api/users`             | admin+      | —                                                                  | Lista todos os usuários.                                                                                      |
| PATCH  | `/api/users/{id}/role`   | super_admin | Path: `id` · Body: `{ "role": "user"\|"admin"\|"super_admin" }`    | Altera o papel de um usuário. `404` se não existir, `403` se a alteração não for permitida.                   |
| PATCH  | `/api/users/me`          | Sessão      | `{ "name": string, "email": string }`                              | Atualiza nome/e-mail do usuário autenticado. `409` se o e-mail já estiver em uso.                             |
| PATCH  | `/api/users/me/password` | Sessão      | `{ "currentPassword": string, "newPassword": string (min. 8) }`    | Atualiza a senha. Usuários criados via Google (sem senha ainda) não precisam informar a atual. Retorna `204`. |

Validações: `name` obrigatório, `email` deve conter formato válido (`x@y.z`), `password`/`newPassword` com mínimo de 8 caracteres. Erros: `422` (validação), `409` (e-mail já cadastrado).

### Contas (accounts)

Todas as rotas exigem sessão (`auth.RequireUser`) e operam apenas sobre as contas do usuário autenticado.

| Método | Rota                 | Auth   | Body / Parâmetros                                       | Descrição                                                          |
| ------ | -------------------- | ------ | ------------------------------------------------------- | ------------------------------------------------------------------ |
| GET    | `/api/accounts`      | Sessão | —                                                       | Lista as contas do usuário.                                        |
| POST   | `/api/accounts`      | Sessão | `{ "name": string, "type": string, "balance": number }` | Cria uma conta. `type` padrão: `"checking"`.                       |
| PUT    | `/api/accounts/{id}` | Sessão | Path: `id` · Body igual ao `POST`                       | Atualiza uma conta. `404` se não existir/não pertencer ao usuário. |
| DELETE | `/api/accounts/{id}` | Sessão | Path: `id`                                              | Remove uma conta. Retorna `{"success": true}`.                     |

### Categorias

| Método | Rota                   | Auth   | Body / Parâmetros                                                                                              | Descrição                                                                                                                                                |
| ------ | ---------------------- | ------ | -------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| GET    | `/api/categories`      | Sessão | —                                                                                                              | Lista as categorias do usuário.                                                                                                                          |
| POST   | `/api/categories`      | Sessão | `{ "name": string, "color": string, "icon": string, "type": "income"\|"expense"\|"both", "patterns": string }` | Cria categoria. Padrões: `color="#6366f1"`, `icon="circle"`, `type="both"`. `patterns` é uma lista de regex (uma por linha) usada na auto-categorização. |
| PUT    | `/api/categories/{id}` | Sessão | Path: `id` · Body igual ao `POST`                                                                              | Atualiza categoria. `404` se não pertencer ao usuário.                                                                                                   |
| DELETE | `/api/categories/{id}` | Sessão | Path: `id`                                                                                                     | Remove categoria. Retorna `{"success": true}`.                                                                                                           |

### Tags

| Método | Rota             | Auth   | Body / Parâmetros                     | Descrição                                        |
| ------ | ---------------- | ------ | ------------------------------------- | ------------------------------------------------ |
| GET    | `/api/tags`      | Sessão | —                                     | Lista as tags do usuário.                        |
| POST   | `/api/tags`      | Sessão | `{ "name": string, "color": string }` | Cria tag. `color` padrão: `"#6366f1"`.           |
| PUT    | `/api/tags/{id}` | Sessão | Path: `id` · Body igual ao `POST`     | Atualiza tag. `404` se não pertencer ao usuário. |
| DELETE | `/api/tags/{id}` | Sessão | Path: `id`                            | Remove tag. Retorna `{"success": true}`.         |

### Transações

Todas as rotas exigem sessão. `{id}` sempre restrito ao usuário autenticado.

| Método | Rota                                | Auth   | Body / Query / Parâmetros                                                          | Descrição                                                                                                                           |
| ------ | ----------------------------------- | ------ | ---------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| GET    | `/api/transactions`                 | Sessão | Query: `type`, `search`, `month`, `year`, `categoryId`, `tagId`, `page`, `limit`   | Lista transações paginadas, com filtros. Retorna `{ data, totals, pagination }`, incluindo `tags` associadas.                       |
| GET    | `/api/transactions/periods`         | Sessão | —                                                                                  | Lista os períodos (mês/ano) que possuem transações lançadas.                                                                        |
| POST   | `/api/transactions`                 | Sessão | `{ date, description, amount, type, categoryId?, accountId?, notes?, tagIds? }`    | Cria uma transação. `date`, `description`, `amount` (≠0) e `type` são obrigatórios. `amount` é sempre armazenado em valor absoluto. |
| GET    | `/api/transactions/{id}`            | Sessão | Path: `id`                                                                         | Busca uma transação por ID, com tags. `404` se não existir.                                                                         |
| PUT    | `/api/transactions/{id}`            | Sessão | Path: `id` · Body igual ao `POST`                                                  | Atualiza uma transação. Se `tagIds` for enviado (mesmo vazio), substitui as tags associadas.                                        |
| DELETE | `/api/transactions/{id}`            | Sessão | Path: `id`                                                                         | Remove uma transação. Retorna `{"success": true}`.                                                                                  |
| PATCH  | `/api/transactions`                 | Sessão | `{ "ids": number[], "fields": { "type"?, "categoryId"?, "accountId"?, "date"? } }` | Atualização em massa. `categoryId`/`accountId` aceitam `null` ou `"_none"` para desvincular. Retorna `{"updated": n}`.              |
| DELETE | `/api/transactions`                 | Sessão | `{ "scope": "month"\|"year"\|"all", "month"?: number, "year"?: number }`           | Remoção em massa por escopo. Retorna `{"deleted": n}`.                                                                              |
| POST   | `/api/transactions/auto-categorize` | Sessão | —                                                                                  | Aplica os padrões (regex) das categorias às transações sem categoria. Retorna `{"updated": n, "checked": n}`.                       |
| GET    | `/api/transactions/opening-balance` | Sessão | Query: `month`, `year` (obrigatórios)                                              | Retorna o saldo de abertura do período; se não houver um definido, retorna a soma dos saldos das contas.                            |
| POST   | `/api/transactions/opening-balance` | Sessão | `{ "month": number, "year": number, "amount": number }`                            | Define/atualiza (upsert) o saldo de abertura de um mês/ano.                                                                         |

### Importação de extratos (PDF)

Todas as rotas exigem sessão.

| Método | Rota                               | Auth   | Body / Parâmetros                                                                           | Descrição                                                                                                                                                                                                                               |
| ------ | ---------------------------------- | ------ | ------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| GET    | `/api/transactions/import/banks`   | Sessão | —                                                                                           | Lista os bancos suportados para detecção automática (`bb`, `sicredi`, `bradesco`, `nubank`, `mercadopago`, `itau`, `generic`).                                                                                                          |
| POST   | `/api/transactions/import/preview` | Sessão | `multipart/form-data`: campo `file` (PDF, máx. 20MB), campo opcional `bank` (força o banco) | Extrai o texto do PDF, detecta o banco, faz o parsing dos lançamentos, marca duplicados e sugere categoria via padrões. Retorna `{ bank, bankLabel, transactions[] }`. `422` se o PDF não puder ser lido/nenhum lançamento reconhecido. |
| POST   | `/api/transactions/import/confirm` | Sessão | `{ "bank": string, "accountId"?: number, "transactions": ParsedRow[] }`                     | Confirma a importação, criando as transações válidas (`date`, `description`, `amount`≠0 e `type` em `income`/`expense`). Retorna `{ imported, total }`.                                                                                 |

### Relatórios e backup

Todas as rotas exigem sessão.

| Método | Rota                  | Auth   | Body / Parâmetros                                                                                                                                                    | Descrição                                                                                                                                                                                                     |
| ------ | --------------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| POST   | `/api/reports/export` | Sessão | `{ "format": "xlsx"\|"ods"\|"pdf"\|"json", "scope": "month"\|"year"\|"range"\|"all", "month"?, "year"?, "dateFrom"?, "dateTo"?, "charts"?: [{ title, pngBase64 }] }` | Gera e retorna o arquivo do relatório (download binário) com transações, totais e (opcionalmente) gráficos embutidos. `format="json"` gera um arquivo de backup completo.                                     |
| POST   | `/api/reports/import` | Sessão | `multipart/form-data`: campo `file` (JSON de backup, máx. 10MB)                                                                                                      | Restaura transações a partir de um backup JSON gerado por `export` (`format=json`), recriando categorias/contas/tags por nome quando necessário e pulando duplicadas. Retorna `{ imported, skipped, total }`. |

Escopos de exportação (`scope`):

- `month` — exige `month` (1–12) e `year`.
- `year` — exige `year`.
- `range` — exige `dateFrom` e `dateTo` (formato `YYYY-MM-DD`).
- `all` — todo o histórico, sem filtro de data.

### Contas fixas

Todas as rotas exigem sessão e operam apenas sobre as contas fixas do usuário autenticado.

| Método | Rota                               | Auth   | Body / Parâmetros                                                                    | Descrição                                                                                                                                                                                                                                              |
| ------ | ---------------------------------- | ------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| GET    | `/api/fixed-bills`                 | Sessão | —                                                                                    | Lista as contas fixas do usuário (ativas, congeladas e encerradas).                                                                                                                                                                                    |
| POST   | `/api/fixed-bills`                 | Sessão | `{ name, amount, categoryId?, accountId?, periodicity, dueDate, notes? }`            | Cria uma conta fixa. `periodicity`: `weekly`, `biweekly`, `monthly`, `bimonthly`, `quarterly`, `semiannual`, `annual`, `biennial`.                                                                                                                     |
| PUT    | `/api/fixed-bills/{id}`            | Sessão | Path: `id` · Body igual ao `POST`                                                    | Atualiza uma conta fixa.                                                                                                                                                                                                                               |
| DELETE | `/api/fixed-bills/{id}`            | Sessão | Path: `id`                                                                           | Remove uma conta fixa (e seu histórico de pagamentos).                                                                                                                                                                                                 |
| POST   | `/api/fixed-bills/{id}/freeze`     | Sessão | Path: `id`                                                                           | Congela a conta (para de gerar lembretes e não pode ser paga até reativar).                                                                                                                                                                            |
| POST   | `/api/fixed-bills/{id}/reactivate` | Sessão | Path: `id`                                                                           | Reativa uma conta congelada ou encerrada.                                                                                                                                                                                                              |
| POST   | `/api/fixed-bills/{id}/close`      | Sessão | Path: `id`                                                                           | Encerra definitivamente a conta.                                                                                                                                                                                                                       |
| POST   | `/api/fixed-bills/{id}/pay`        | Sessão | Body opcional: `{ bank?, paymentMethod?, paidAt?, amountPaid?, accountId?, notes? }` | Registra o pagamento do ciclo atual: cria uma transação de despesa (visível em `/transacoes`, gráficos e exports), grava o pagamento e avança `dueDate` conforme a periodicidade. Sem body, usa valor/data padrão. `422` se a conta não estiver ativa. |
| GET    | `/api/fixed-bills/{id}/payments`   | Sessão | Path: `id`                                                                           | Histórico de pagamentos da conta.                                                                                                                                                                                                                      |

### Notificações

Lembretes de contas fixas vencendo/vencidas, entregues em tempo real no site (SSE) e por e-mail conforme as preferências do usuário.

| Método | Rota                               | Auth    | Body / Parâmetros                                  | Descrição                                                                                                                                                                |
| ------ | ---------------------------------- | ------- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| GET    | `/api/notifications`               | Sessão  | —                                                  | Lista as últimas notificações do usuário (vazio se notificações no site estiverem desativadas).                                                                          |
| GET    | `/api/notifications/unread-count`  | Sessão  | —                                                  | Retorna `{ "count": n }` de notificações não lidas.                                                                                                                      |
| GET    | `/api/notifications/stream`        | Sessão  | —                                                  | Conexão SSE (`text/event-stream`) que emite cada notificação nova (`data: <Notification>`) assim que é criada.                                                           |
| POST   | `/api/notifications/{id}/read`     | Sessão  | Path: `id`                                         | Marca uma notificação como lida.                                                                                                                                         |
| POST   | `/api/notifications/read-all`      | Sessão  | —                                                  | Marca todas as notificações do usuário como lidas.                                                                                                                       |
| GET    | `/api/notification-settings`       | Sessão  | —                                                  | Retorna as preferências do usuário (cria com padrão na primeira vez).                                                                                                    |
| PUT    | `/api/notification-settings`       | Sessão  | `{ siteEnabled, emailEnabled, offsets: number[] }` | Atualiza as preferências. `offsets` são minutos de antecedência antes do vencimento (ex.: `[2880,1440,120,60]`).                                                         |
| POST   | `/api/internal/notifications/scan` | Interno | Header `X-Internal-Token: <INTERNAL_API_TOKEN>`    | Varre todas as contas fixas ativas, gera notificações (dedupe automático) e dispara e-mails. Chamada pelo timer systemd, nunca pelo frontend. `401` sem o token correto. |

---

## Modelos de dados

### `User`

```jsonc
{
  "id": 1,
  "name": "Fulano",
  "email": "fulano@email.com",
  "role": "user", // "user" | "admin" | "super_admin"
  "avatarUrl": "",
  "hasPassword": true,
  "createdAt": "2026-01-01T00:00:00Z",
}
```

### `Account`

```jsonc
{
  "id": 1,
  "name": "Conta Corrente",
  "type": "checking",
  "balance": 1500.5,
  "createdAt": "2026-01-01T00:00:00Z",
}
```

### `Category`

```jsonc
{
  "id": 1,
  "name": "Mercado",
  "color": "#6366f1",
  "icon": "circle",
  "type": "both", // "income" | "expense" | "both"
  "patterns": "supermercado\nmercado.*",
}
```

### `Tag`

```jsonc
{
  "id": 1,
  "name": "Viagem",
  "color": "#6366f1",
}
```

### `Transaction`

```jsonc
{
  "id": 1,
  "date": "2026-01-15",
  "description": "Supermercado XPTO",
  "amount": 250.3, // sempre valor absoluto; sinal definido por "type"
  "type": "expense", // "income" | "expense"
  "categoryId": 3,
  "categoryName": "Mercado", // presente em algumas listagens
  "categoryColor": "#6366f1",
  "accountId": 1,
  "notes": null,
  "importedFrom": null, // "pdf" | "backup" | null
  "bank": null,
  "pixBeneficiary": null,
  "createdAt": "2026-01-15T12:00:00Z",
  "tags": [{ "id": 1, "name": "Viagem", "color": "#6366f1" }],
}
```

### `ListResult` (listagem paginada de transações)

```jsonc
{
  "data": [/* Transaction[] */],
  "totals": { "income": 5000, "expense": 3200, "balance": 1800, "count": 42 },
  "pagination": { "page": 1, "limit": 20, "total": 42 },
}
```

### `FixedBill`

```jsonc
{
  "id": 1,
  "name": "Energia elétrica",
  "amount": 150.5,
  "categoryId": 3,
  "categoryName": "Utilidades",
  "categoryColor": "#6366f1",
  "accountId": 1,
  "accountName": "Conta Corrente",
  "periodicity": "monthly",
  "dueDate": "2026-02-10",
  "status": "active", // "active" | "frozen" | "closed"
  "notes": null,
  "createdAt": "2026-01-01T00:00:00Z",
  "updatedAt": "2026-01-10T00:00:00Z",
}
```

### `Notification`

```jsonc
{
  "id": 1,
  "fixedBillId": 1,
  "fixedBillName": "Energia elétrica",
  "kind": "bill_due_soon", // "bill_due_soon" | "bill_overdue"
  "title": "Conta vencendo: Energia elétrica",
  "message": "Energia elétrica vence em 2 dia(s) (2026-02-10).",
  "dueDate": "2026-02-10",
  "readAt": null,
  "createdAt": "2026-02-08T13:00:00Z",
}
```

## Variáveis de ambiente

| Variável               | Padrão                  | Descrição                                                                                                                                                                                                       |
| ---------------------- | ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PORT`                 | `3020`                  | Porta HTTP do servidor.                                                                                                                                                                                         |
| `DB_PATH`              | `./data/reconta.db`     | Caminho do arquivo SQLite.                                                                                                                                                                                      |
| `ENV`                  | `development`           | Quando `production`, cookies de sessão são marcados como `Secure`.                                                                                                                                              |
| `APP_URL`              | `http://localhost:5173` | URL do frontend, usada nos redirecionamentos do fluxo OAuth2 do Google.                                                                                                                                         |
| `GOOGLE_CLIENT_ID`     | —                       | Client ID do OAuth2 do Google. Se ausente, o login com Google fica desabilitado.                                                                                                                                |
| `GOOGLE_CLIENT_SECRET` | —                       | Client Secret do OAuth2 do Google.                                                                                                                                                                              |
| `GOOGLE_REDIRECT_URL`  | —                       | URL de callback registrada no Google Cloud Console (`/api/auth/google/callback`).                                                                                                                               |
| `INTERNAL_API_TOKEN`   | —                       | Token compartilhado exigido pelo header `X-Internal-Token` na rota `/api/internal/notifications/scan`. Sem ele, a rota fica desabilitada (sempre `401`). Gere um valor aleatório (ex.: `openssl rand -hex 32`). |
| `SMTP_HOST`            | —                       | Host do servidor SMTP para envio de e-mails de lembrete. Se ausente, o envio vira no-op (apenas loga).                                                                                                          |
| `SMTP_PORT`            | `587`                   | Porta do servidor SMTP.                                                                                                                                                                                         |
| `SMTP_USER`            | —                       | Usuário para autenticação SMTP (`PlainAuth`). Também usado como remetente se `SMTP_FROM` não for definido.                                                                                                      |
| `SMTP_PASS`            | —                       | Senha/token do usuário SMTP.                                                                                                                                                                                    |
| `SMTP_FROM`            | valor de `SMTP_USER`    | Endereço de remetente dos e-mails.                                                                                                                                                                              |

Variáveis podem ser definidas em um arquivo `.env` na raiz de `api/` (carregado apenas em desenvolvimento; em produção o systemd injeta o `EnvironmentFile`).

## Executando localmente

```sh
# Na raiz do monorepo
bun run api:dev      # equivalente a: cd api && go run .

# Rodar os testes
bun run api:test     # equivalente a: cd api && go test ./...

# Build de produção
bun run api:build    # equivalente a: cd api && go build -o bin/server .
```

O servidor sobe em `http://localhost:3020` por padrão (ajustável via `PORT`), com CORS liberado para a origem do Vite em desenvolvimento (`http://localhost:5173`). Em produção, o Nginx faz proxy same-origin em `/api/`, dispensando CORS.

## Timer systemd de notificações (produção)

Os lembretes de contas fixas dependem de um timer systemd que chama `POST /api/internal/notifications/scan` a cada hora. Os arquivos de unidade ficam em `files/reconta-notifications.service` e `files/reconta-notifications.timer`.

Instalação na VPS (uma vez, após o primeiro deploy):

```sh
sudo cp files/reconta-notifications.service files/reconta-notifications.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now reconta-notifications.timer
```

Requisitos:

- `INTERNAL_API_TOKEN` definido em `api/.env` (o mesmo valor é lido pelo serviço via `EnvironmentFile`).
- Para lembretes por e-mail, configurar também `SMTP_HOST`/`SMTP_PORT`/`SMTP_USER`/`SMTP_PASS`/`SMTP_FROM`.

Verificação:

```sh
sudo systemctl list-timers reconta-notifications.timer
sudo systemctl start reconta-notifications.service   # dispara uma varredura manualmente
journalctl -u reconta-notifications.service -f
```
