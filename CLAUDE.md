# Reconta — Guia para Claude

## Estrutura do monorepo

```
reconta/
├── web/                    # Frontend Vue.js 3 + Vite + TypeScript
│   ├── src/
│   │   ├── App.vue
│   │   ├── main.ts
│   │   ├── components/
│   │   ├── assets/
│   │   └── style.css
│   ├── public/             # Imagens, ícones e SVGs estáticos
│   ├── index.html
│   ├── vite.config.ts
│   └── package.json
├── api/                    # Backend Go
│   └── go.mod              # module github.com/lucasbrum/reconta/api
├── files/                  # Arquivos de configuração do servidor (VPS)
│   ├── reconta.service     # Systemd service da aplicação
│   └── reconta.conf        # Configuração Nginx
├── scripts/
│   └── deploy.sh           # Script de deploy executado na VPS
├── .github/workflows/
│   └── deploy.yml          # CI/CD: push para main → SCP + SSH na VPS
└── package.json            # Raiz do monorepo (Bun workspaces)
```

---

## Stack

| Camada     | Tecnologia                        |
|------------|-----------------------------------|
| Frontend   | Vue.js 3 + Vite + TypeScript      |
| Backend    | Go (`api/`)                       |
| Runtime JS | Bun                               |
| Servidor   | VPS Linux (usuário `nginx`)       |
| Proxy      | Nginx → `localhost:3020`          |
| Processo   | systemd (`reconta.service`)       |
| Deploy     | GitHub Actions → SCP + SSH        |
| Domínio    | reconta.app (HTTPS via Let's Encrypt) |

---

## Bun: caminho absoluto na VPS

Na VPS, o Bun está instalado em `/home/nginx/.bun/bin/` e **não é adicionado ao PATH automaticamente**. Em scripts do `package.json` executados na VPS, usar `bun` ou `bunx` sem caminho absoluto falha silenciosamente.

**Regra:** Sempre que um script do `package.json` chamar `bunx`, use o caminho absoluto na VPS:

```json
"build": "vue-tsc -b && /home/nginx/.bun/bin/bunx --bun vite build"
```

O `scripts/deploy.sh` contorna isso para comandos de alto nível com:
```bash
PATH=$PATH:/home/nginx/.bun/bin
```
Mas isso **não propaga** para os scripts internos do `package.json` — por isso o caminho absoluto é necessário lá dentro.

Em ambiente local (dev), `bunx` funciona normalmente pois Bun está no PATH do desenvolvedor.

---

## web/ (Vue + Vite)

```sh
# Desenvolvimento local
bun run dev

# Build de produção
bun run build

# Preview do build
bun run preview
```

Scripts ficam em `web/package.json`. Na raiz, todos delegam com `--cwd web`.

---

## api/ (Go)

Module: `github.com/lucasbrum/reconta/api`  
Go: 1.26+

Estrutura esperada:
```
api/
├── main.go     # ponto de entrada
├── internal/   # lógica de negócio (não exportada)
└── go.mod
```

```sh
# Na raiz (via package.json)
bun run api:dev      # go run .
bun run api:build    # go build -o bin/server .
bun run api:test     # go test ./...

# Diretamente em api/
go run .
go test ./...
```

---

## Deploy

1. Push para `main` dispara o workflow `.github/workflows/deploy.yml`
2. GitHub Actions envia os arquivos via SCP para a VPS
3. Em seguida, executa `bash scripts/deploy.sh` via SSH
4. O deploy script:
   - Copia o diretório atual para `/tmp/reconta` (preserva `.env` e banco SQLite)
   - Roda `bun install` e `bun run build`
   - Para o serviço, substitui `/var/www/reconta`, reinicia o serviço e o Nginx

Segredos necessários no GitHub: `SSH_HOST`, `SSH_USER`, `SSH_PASS`, `SSH_PORT`, `PROJECT_PATH`.
