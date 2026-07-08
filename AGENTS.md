# Reconta — Guia para Agentes de IA

## O que é este projeto

Reconta é uma aplicação web para controle financeiro pessoal — gerenciamento de contas a pagar/receber, categorização de transações e notificações de vencimento.

## Estrutura do monorepo

```
reconta/
├── web/                    # Frontend Vue.js 3 + Vite + TypeScript
│   ├── src/
│   │   ├── App.vue         # Componente raiz
│   │   ├── main.ts         # Entry point
│   │   ├── components/     # Componentes Vue
│   │   ├── assets/         # Imagens e recursos locais
│   │   └── style.css
│   ├── public/             # Arquivos estáticos servidos diretamente
│   │   ├── favicon.svg
│   │   ├── icons.svg
│   │   └── images/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
├── api/                    # Backend Go
│   └── go.mod              # module github.com/lucasbrum/reconta/api
├── files/                  # Configurações do servidor (não editar sem contexto)
│   ├── reconta.service     # Systemd: processo principal da aplicação
│   └── reconta.conf        # Nginx: proxy reverso + SSL
├── scripts/
│   └── deploy.sh           # Deploy na VPS
├── .github/workflows/
│   └── deploy.yml          # CI/CD automático no push para main
├── package.json            # Raiz do workspace Bun
├── CLAUDE.md               # Instruções para Claude Code
├── AGENTS.md               # Este arquivo
└── README.md
```

## Stack técnica

- **Frontend:** Vue.js 3 + Vite + TypeScript (em `web/`)
- **Backend:** Go 1.26+ (em `api/`)
- **Runtime JS:** Bun (não use npm, yarn, pnpm, node ou ts-node)
- **Banco de dados:** SQLite (arquivo local na VPS, preservado entre deploys)
- **Proxy:** Nginx → `localhost:3020`
- **Processo:** systemd (`reconta.service`)
- **Domínio:** reconta.app com HTTPS via Let's Encrypt

## Regras críticas

### Bun na VPS — caminho absoluto obrigatório

Na VPS (usuário `nginx`), o Bun está em `/home/nginx/.bun/bin/` e **não fica no PATH** por padrão. Scripts do `package.json` que chamam `bunx` ou `bun` sem caminho absoluto **falham silenciosamente** na VPS.

Sempre use caminho absoluto em scripts do `package.json` que rodam na VPS:

```json
"/home/nginx/.bun/bin/bunx --bun vite build"
```

Em desenvolvimento local, `bunx` funciona normalmente.

### Não use

- `npm`, `yarn`, `pnpm` — use `bun`
- `node`, `ts-node` — use `bun`
- `express`, `fastify` — a API é em Go
- `dotenv` — Bun carrega `.env` automaticamente
- `vite` diretamente — use `bunx --bun vite` (ou caminho absoluto na VPS)

## Comandos principais

```sh
# Frontend (raiz ou web/)
bun run dev          # Vite dev server
bun run build        # Build de produção
bun run preview      # Preview do build

# Backend Go (na raiz)
bun run api:dev      # go run ./cmd/server
bun run api:build    # go build -o bin/server ./cmd/server
bun run api:test     # go test ./...
```

## Deploy

Push para `main` → GitHub Actions → SCP dos arquivos → SSH executa `scripts/deploy.sh` na VPS.

O deploy preserva `.env` e o banco SQLite durante a troca de versão.

## Arquivos sensíveis

- `web/.env` / `api/.env` — nunca commitar, preservados pelo deploy
- `web/.env.production` / `api/.env.production` — gerados pelo deploy a partir dos `.env`
- Banco SQLite — em `/var/www/reconta/`, preservado entre deploys
