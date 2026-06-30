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

## Testes

```sh
# Frontend
bun test

# Backend Go
bun run api:test
```
