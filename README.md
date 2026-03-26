# ReConta — Controle Financeiro Pessoal

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./public/images/banner.svg" />
    <img src="./public/images/banner-light.svg" alt="ReConta" />
  </picture>
</p>

> **reconta.app** · Gerencie suas finanças, analise extratos bancários e acompanhe sua poupança.

[![Deploy](https://github.com/sistematico/reconta/actions/workflows/deploy.yml/badge.svg)](https://github.com/sistematico/reconta/actions/workflows/deploy.yml)

## Stack

- **Next.js 16** (TypeScript · App Router)
- **Drizzle ORM + SQLite** (banco de dados local em `reconta.db`)
- **Tailwind CSS v4** (tema dark)
- **Recharts** (gráficos)
- **Radix UI** (componentes acessíveis)
- **Biome** (linting e formatação)
- **pnpm**

---

## Instalação

```bash
# Clone e entre no diretório
cd reconta

# Instale as dependências
pnpm install

# Crie o banco de dados (SQLite)
pnpm push
```

## Desenvolvimento

```bash
pnpm dev
```

Acesse [http://localhost:3000](http://localhost:3000).

O banco é criado automaticamente em `reconta.db` na raiz do projeto. As categorias e conta padrão são inseridas na primeira execução via `src/instrumentation.ts`.

## Produção

```bash
pnpm build
pnpm start
```

---

## Páginas

| Página | Rota | Descrição |
|--------|------|-----------|
| **Dashboard** | `/` | Visão geral: KPIs do mês, gráfico dos últimos 6 meses, gastos por categoria, contas pendentes e últimos lançamentos |
| **Lançamentos** | `/transacoes` | Livro-caixa completo com filtros por tipo/mês, busca por descrição, totalizadores e CRUD |
| **Contas Fixas** | `/contas` | Alertas de cobranças recorrentes (condomínio, luz, internet etc.) com controle de pagamento por mês |
| **Relatórios** | `/relatorios` | Comparativo mês atual vs anterior, gráfico de poupança, análise por categoria |
| **Importar Extrato** | `/importar` | Upload de extrato bancário em PDF via drag & drop com parsing automático de transações |
| **Categorias** | `/categorias` | CRUD de categorias com cores customizáveis (receita, despesa ou ambos) |
| **Contas Bancárias** | `/contas-bancarias` | Gerenciamento de contas (corrente, poupança, crédito, investimentos) com saldo total |

---

## Funcionalidades

- **Importação de extrato PDF** — tenta detectar automaticamente o formato de extratos brasileiros (Itaú, Bradesco, BB, Nubank etc.). O PDF precisa conter texto selecionável (não apenas imagem).
- **Alertas de contas fixas** — contas vencidas aparecem em vermelho; contas com vencimento em até 3 dias aparecem em amarelo.
- **Taxa de poupança** — calculada automaticamente como `(receitas - despesas) / receitas × 100`.
- **Comparativo mensal** — variação percentual em relação ao mês anterior para receitas, despesas e saldo.
- **Navegação por mês/ano** — todas as views permitem navegar entre meses.
- **Seed automático** — categorias padrão e conta inicial são criados automaticamente na primeira execução.

---

## Scripts disponíveis

```bash
pnpm dev          # Servidor de desenvolvimento
pnpm build        # Build de produção
pnpm start        # Servidor de produção
pnpm push         # Aplica o schema ao banco SQLite
pnpm generate     # Gera arquivos de migração (drizzle-kit)
pnpm studio       # Abre o Drizzle Studio (GUI para o banco)
pnpm lint         # Verifica problemas com Biome
pnpm format       # Formata o código com Biome
pnpm check        # Checa o código por erros de sintaxe
```

---

## Estrutura do projeto

```
src/
├── app/
│   ├── api/                  # Rotas da API REST
│   │   ├── accounts/
│   │   ├── bills/
│   │   ├── categories/
│   │   ├── dashboard/
│   │   ├── import/
│   │   └── transactions/
│   ├── categorias/
│   ├── contas/
│   ├── contas-bancarias/
│   ├── importar/
│   ├── relatorios/
│   ├── transacoes/
│   ├── globals.css
│   ├── layout.tsx
│   └── page.tsx              # Dashboard
├── components/
│   ├── ui/                   # Componentes base (Button, Card, Dialog…)
│   ├── layout/               # Sidebar e Header
│   ├── dashboard/            # Gráficos e cards do dashboard
│   ├── transactions/         # Lista e formulário de lançamentos
│   ├── bills/                # Lista e formulário de contas fixas
│   ├── reports/              # Gráficos de relatórios
│   ├── import/               # Upload de PDF
│   ├── categories/           # CRUD de categorias
│   └── accounts/             # CRUD de contas bancárias
├── hooks/
│   └── use-accounts.ts
├── lib/
│   ├── db/
│   │   ├── index.ts          # Conexão Drizzle + SQLite
│   │   ├── schema.ts         # Tabelas e tipos
│   │   └── seed.ts           # Dados iniciais
│   ├── pdf-parser.ts         # Parsing de extratos PDF
│   └── utils.ts              # Helpers (formatação, datas)
└── instrumentation.ts        # Seed executado na inicialização
```
