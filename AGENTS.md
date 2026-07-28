# AGENTS.md — reconta

## Visão geral

Reconta é uma aplicação de finanças pessoais: contas, transações, categorias, contas fixas (fixedbill), faturas/cobranças (billing via Mercado Pago), relatórios (PDF/Excel), extratos (statement), compartilhamento (share) e notificações.

Monorepo com backend em **Go** (`api/`) e frontend em **Vue 3 + TypeScript** (`web/`), runtime/gerenciador de pacotes **Bun**, deploy local via **Podman** (não Docker).

> Este repositório já tem `CLAUDE.md`. Mantenha este AGENTS.md como a fonte cross-tool (Cursor, Codex, Copilot, etc.) e deixe peculiaridades específicas do Claude Code no CLAUDE.md — evite duplicar conteúdo entre os dois.

## Estrutura real do projeto

```
.
├── api/                    # backend Go — entrypoint único em api/main.go (sem cmd/)
│   ├── internal/
│   │   ├── account/ auth/ billing/ category/ db/ email/
│   │   ├── fixedbill/ health/ notification/ report/
│   │   └── seed/ share/ statement/ tag/ transaction/ user/
│   ├── data/                # SQLite local (reconta.db, -shm, -wal) — NUNCA editar/commitar dados reais
│   ├── insomnia.yaml         # coleção de requests — fonte de verdade da API, atualizar ao criar endpoints
│   └── README.md
├── web/                    # frontend Vue 3
│   └── src/
│       ├── api/ components/ composables/ layouts/
│       ├── router.ts         # rotas declaradas manualmente (NÃO é file-based routing)
│       ├── config.ts styles/ types/ utils/ views/
├── certs/                  # certificados TLS locais autoassinados (reconta.local) — sensível, não regenerar sem necessidade
├── files/                  # unidades systemd --user + config nginx para produção/local
├── podman/                 # Containerfile.api, Containerfile.web, nginx.local.conf
├── scripts/deploy.sh
├── Makefile                 # orquestra dev, build, deploy (ver seção abaixo)
├── package.json / bun.lock   # raiz do workspace (orquestra api+web em dev)
└── skills-lock.json          # lockfile de Skills do Claude usadas no projeto
```

## Stack e decisões que a LLM não deve tentar "corrigir"

- **Backend Go 1.26.2**, módulo `github.com/re-conta/reconta/api`.
- **SQLite via `modernc.org/sqlite`** (driver puro Go, sem CGO). Não sugerir trocar por `mattn/go-sqlite3` — quebraria builds sem CGO e os `Containerfile.api`.
- Geração de relatórios: `github.com/go-pdf/fpdf`, `github.com/ledongthuc/pdf` (PDF) e `github.com/xuri/excelize/v2` (Excel) — usar essas libs para qualquer feature nova de exportação, não introduzir alternativas.
- **Billing/pagamentos via Mercado Pago** (`github.com/mercadopago/sdk-go`). Código de billing e webhooks é sensível: nunca logar tokens/segredos, tratar webhooks como idempotentes.
- Auth com `golang.org/x/oauth2` + `golang.org/x/crypto`.
- Pacotes em `internal/` são organizados **por feature**, não por camada (não existe pasta `handlers/`, `services/` na raiz). Uma feature nova ganha seu próprio pacote em `internal/<feature>/`, seguindo o padrão dos existentes.
- Endpoint interno de notificações (`/api/internal/notifications/scan`) usa autenticação própria via header `X-Internal-Token` — diferente da auth de usuário. Não misturar os dois mecanismos.

### Frontend
- **Vue 3.5 + vue-router 5**, roteador declarado manualmente em `src/router.ts` — **não é file-based routing**, não sugerir `unplugin-vue-router`.
- **Tailwind CSS v4** via `@tailwindcss/vite` (config CSS-first) — não criar `tailwind.config.js` no estilo v3.
- Gráficos com `chart.js` + `vue-chartjs`; ícones com `lucide-vue-next`.
- **Lint/format com oxlint + oxfmt** (não ESLint/Prettier). Comandos: `bun run lint`, `bun run fix`, `bun run format`, `bun run check`.
- Type-check com `vue-tsc`; o script `build` já roda `vue-tsc -b` antes do `vite build` — não pular essa etapa.

### Runtime — o que não usar
- `npm`, `yarn`, `pnpm` — usar `bun`.
- `node`, `ts-node` — usar `bun`.
- `express`, `fastify` ou qualquer framework HTTP JS — a API é em Go (`net/http` puro, sem framework).

## Comandos

### Backend
```bash
cd api
go build ./...
go test ./... -v
go vet ./...
gofmt -l .          # deve retornar vazio
```

### Frontend
```bash
cd web
bun install
bun run dev          # Vite dev server
bun run build          # vue-tsc -b && vite build
bun run lint            # oxlint
bun run fix              # oxlint --fix
bun run format            # oxfmt
bun run check              # oxfmt --check
```

### Ambiente local completo (via Makefile, na raiz)
```bash
make dev              # gera certs + checa /etc/hosts + roda `bun run dev` (API + Vite juntos)
make up                # build das imagens Podman e sobe pod (simula produção em https://reconta.local)
make down                # para o pod (mantém volume do banco)
make logs                  # segue logs do pod
make status                  # status do pod/containers
make clean                    # remove pod, imagens e volume (apaga banco local!)
make notify                     # dispara manualmente o scan de notificações
make notify-install/uninstall     # instala/remove timer systemd --user
```
`make up` usa **Podman**, não Docker — não sugerir `docker-compose` como alternativa. O domínio local é `reconta.local` (HTTPS obrigatório, porta 443), exige entrada em `/etc/hosts` checada pelo próprio `make dev`/`make up`.

## Testes

- Toda mudança em `internal/<feature>/` precisa de teste Go cobrindo caminho feliz e erro.
- Rodar `go test ./...` (backend) antes de qualquer PR que toque `api/`.
- Não há suíte de testes de frontend configurada em `web/package.json` no momento — se adicionar uma, registrar o comando aqui.

## Segurança

- `api/data/*.db*` contém dados reais em dev local — nunca commitar, nunca ler/alterar diretamente fora de migrations/seeds.
- `certs/*.pem` são certificados autoassinados locais — não gerar novos sem necessidade nem versionar chaves de produção.
- Segredos de ambiente ficam em `api/.env` e `web/.env` (com fallback para os respectivos `.env.example` conforme o `Makefile`) — nunca commitar `.env`/`.env.production` reais. O `main.go` carrega `api/.env` manualmente (não é auto-load do Bun).
- Webhooks do Mercado Pago e o token interno de notificações (`X-Internal-Token`) exigem validação estrita — não relaxar checagem "para debugar mais rápido".

## Regras de commit / PR

- Mensagem: `tipo(escopo): descrição curta` (ex: `feat(billing): trata webhook de assinatura cancelada`).
- Prefixo no título do PR: `[api]` ou `[web]` quando a mudança for isolada a um lado.
- Rodar `go test ./...` e `bun run check`/`bun run lint` antes de abrir PR.
- Ao adicionar/alterar endpoint, atualizar `api/insomnia.yaml`.

## Limites — o que o agente NÃO deve fazer

- Não trocar `modernc.org/sqlite` por um driver com CGO.
- Não trocar oxlint/oxfmt por ESLint/Prettier.
- Não introduzir Docker/docker-compose no lugar de Podman.
- Não criar `tailwind.config.js` estilo v3 — a config é CSS-first (v4).
- Não editar `api/data/*.db*` nem `certs/*.pem` diretamente.
- Não alterar a lógica de auth interna (`X-Internal-Token`) sem revisão humana explícita.
