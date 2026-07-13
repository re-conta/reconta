#!/usr/bin/env bash

set -euo pipefail

NAME="reconta"
TMPDIR="/tmp/$NAME"
WORKDIR="/var/www/$NAME"
SERVICE="${NAME}.service"
PATH=$PATH:/home/nginx/.bun/bin

echo "📦 Preparando ambiente de deploy..."

[ -e "$TMPDIR" ] && rm -rf "$TMPDIR"
[ -e "$WORKDIR" ] && cp -af "$WORKDIR" "$TMPDIR"
cd "$TMPDIR" || exit 1

# Mantém .env e o banco de dados (incluindo arquivos WAL/SHM do SQLite) intactos
git clean -fxd -e web/.env -e api/.env -e api/data
[ -e web/.env ] && cp web/.env web/.env.production
[ -e api/.env ] && cp api/.env api/.env.production

echo "📥 Instalando dependências..."
bun install

echo "🏗️ Buildando frontend (Vue.js)..."
if ! bun run web:build; then
  echo "❌ Falha no build do frontend. Abortando deploy."
  exit 1
fi

echo "🏗️ Buildando backend (Go)..."
if ! bun run api:build; then
  echo "❌ Falha no build da API. Abortando deploy."
  exit 1
fi
echo "✅ Build concluído com sucesso!"

echo "⏸️ Parando serviço para aplicar migrações com segurança..."
sudo /usr/bin/systemctl stop "$SERVICE"

[ -e "$WORKDIR" ] && rm -rf "$WORKDIR"
cp -af "$TMPDIR" "$WORKDIR"

sudo /usr/bin/systemctl start "$SERVICE"
sudo /usr/bin/systemctl restart nginx.service

echo "⏰ Instalando timer de notificações de contas fixas..."
sudo cp -f "$WORKDIR/files/reconta-notifications.service" /etc/systemd/system/reconta-notifications.service
sudo cp -f "$WORKDIR/files/reconta-notifications.timer" /etc/systemd/system/reconta-notifications.timer
sudo /usr/bin/systemctl daemon-reload
sudo /usr/bin/systemctl enable --now reconta-notifications.timer

echo "🚀 Serviço reiniciado!"
