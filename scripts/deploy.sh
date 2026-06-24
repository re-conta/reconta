#!/usr/bin/env bash
set -euo pipefail

NAME="reconta"
TMPDIR="/tmp/$NAME"
WORKDIR="/var/www/$NAME"
SERVICE="${NAME}.service"
PATH=$PATH:/home/nginx/.local/share/pnpm

echo "📦 Preparando ambiente de deploy..."

[ -e "$TMPDIR" ] && rm -rf "$TMPDIR"
[ -e "$WORKDIR" ] && cp -af "$WORKDIR" "$TMPDIR"
cd "$TMPDIR" || exit 1

# Mantém .env e o banco de dados (incluindo arquivos WAL/SHM do SQLite) intactos
git clean -fxd -e .env -e 'drizzle/reconta.db*'
cp .env .env.production

echo "📥 Instalando dependências..."
pnpm install

echo "🏗️ Buildando aplicação..."
if ! pnpm run build; then
  echo "❌ Falha no build. Abortando deploy."
  exit 1
fi
echo "✅ Build concluído com sucesso!"

echo "⏸️ Parando serviço para aplicar migrações com segurança..."
sudo /usr/bin/systemctl stop "$SERVICE"

# Garantir que o banco de dados exista no servidor
mkdir -p "$WORKDIR/drizzle"
[ -f "$WORKDIR/drizzle/reconta.db" ] || touch "$WORKDIR/drizzle/reconta.db"

# Copia o banco de dados mais atual (já com o serviço parado, sem escritas concorrentes)
mkdir -p "$TMPDIR/drizzle"
cp -af "$WORKDIR"/drizzle/reconta.db* "$TMPDIR/drizzle/" 2>/dev/null || true

echo "🗃️ Aplicando migrações do banco de dados..."
if ! pnpm drizzle-kit migrate; then
  echo "❌ Falha ao aplicar migrações. Abortando deploy e reiniciando serviço antigo."
  sudo /usr/bin/systemctl start "$SERVICE"
  exit 1
fi
echo "✅ Migrações aplicadas com sucesso!"

[ -e "$WORKDIR" ] && rm -rf "$WORKDIR"
cp -af "$TMPDIR" "$WORKDIR"
sudo /usr/bin/systemctl start "$SERVICE"
echo "🚀 Serviço reiniciado!"

# Instala/atualiza os units do cron de notificações
sudo cp "$WORKDIR/files/reconta-cron.service" /etc/systemd/system/
sudo cp "$WORKDIR/files/reconta-cron.timer" /etc/systemd/system/
sudo /usr/bin/systemctl daemon-reload
sudo /usr/bin/systemctl enable --now reconta-cron.timer
echo "⏰ Timer de notificações atualizado!"
