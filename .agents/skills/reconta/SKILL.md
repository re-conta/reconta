---
name: reconta
description: Aplicativo de controle financeiro pessoal com Next.js, Drizzle ORM e SQLite. Use para tarefas relacionadas ao desenvolvimento, manutenção ou análise do ReConta.
---

# ReConta — Controle Financeiro Pessoal

## Visão Geral

ReConta é um aplicativo de controle financeiro pessoal que permite gerenciar finanças, analisar extratos bancários e acompanhar poupança. Desenvolvido com Next.js 16, Drizzle ORM, SQLite e Tailwind CSS v4.

## Stack Tecnológica

- **Frontend**: Next.js 16 (App Router), React 19, TypeScript
- **Banco de dados**: Drizzle ORM + SQLite (`reconta.db`)
- **Estilização**: Tailwind CSS v4
- **UI Components**: Radix UI, Recharts
- **Utilitários**: Biome (linting), date-fns, jspdf, xlsx
- **Autenticação**: better-auth
- **Email**: nodemailer + email-templates

## Estrutura do Projeto

```
src/
├── app/
│   ├── api/                  # Rotas REST (accounts, bills, categories, dashboard, import, transactions)
│   ├── (rotas principais)/   # Páginas: categorias, contas, contas-bancarias, importar, relatorios, transacoes
│   ├── globals.css
│   ├── layout.tsx
│   └── page.tsx              # Dashboard
├── components/
│   ├── ui/                   # Componentes base (Button, Card, Dialog, etc.)
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

## Páginas Principais

| Rota | Descrição |
|------|-----------|
| `/` | Dashboard: KPIs do mês, gráfico dos últimos 6 meses, gastos por categoria, contas pendentes |
| `/transacoes` | Livro-caixa completo com filtros, busca e CRUD |
| `/contas` | Contas fixas com alertas de cobranças (condomínio, luz, internet) |
| `/relatorios` | Comparativo mês atual vs anterior, gráfico de poupança |
| `/importar` | Upload de extrato PDF com parsing automático |
| `/categorias` | CRUD de categorias com cores customizáveis |
| `/contas-bancarias` | Gerenciamento de contas (corrente, poupança, crédito, investimentos) |

## Scripts Disponíveis

```bash
pnpm dev          # Servidor de desenvolvimento (porta 3000)
pnpm build        # Build de produção
pnpm start        # Servidor de produção (porta 3020)
pnpm push         # Aplica schema ao banco SQLite
pnpm generate     # Gera migrações Drizzle
pnpm studio       # Drizzle Studio (GUI)
pnpm cron         # Job de cron (verificar contas vencidas)
pnpm lint         # Verifica problemas com Biome
pnpm format       # Formata código com Biome
pnpm check        # Checa sintaxe TypeScript
```

**Nota**: Este projeto usa `pnpm` como gerenciador de pacotes (não npm).

## Funcionalidades Principais

### Importação de Extrato PDF
- Detecta automaticamente formatos brasileiros (Itaú, Bradesco, BB, Nubank)
- Requer texto selecionável no PDF (não apenas imagem)

### Contas Fixas
- Alertas visuais: vermelho (vencidas), amarelo (vencimento em até 3 dias)
- Controle de pagamento por mês

### Cálculos
- Taxa de poupança: `(receitas - despesas) / receitas × 100`
- Comparativo mensal: variação percentual vs mês anterior

### Navegação
- Todas as views suportam navegação por mês/ano

## Banco de Dados

- SQLite local em `reconta.db` na raiz do projeto
- Schema definido em `src/lib/db/schema.ts`
- Seed executado via `src/instrumentation.ts` na inicialização
- Categorias e conta padrão inseridas automaticamente

## Regras de Estilo

- **Biome**: Linting e formatação obrigatórios
- **TypeScript**: Sempre usar tipos explícitos
- **React**: Sempre incluir todas as variáveis do `useEffect` na dependency array

## Notas Importantes

- O projeto usa Next.js 16 com mudanças breaking (consultar docs em `node_modules/next/dist/docs/`)
- Tailwind CSS v4 é usado (versão diferente de v3)
- O banco de dados é criado automaticamente na primeira execução
- Páginas estão em `src/app/` com rotas tradicionais (não App Router)
- API routes estão em `src/app/api/`