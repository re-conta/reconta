#!/usr/bin/env bash

NAME="reconta"
TMPDIR="/tmp/$NAME"
WORKDIR="/var/www/$NAME"
SERVICE="${NAME}.service"
PATH=$PATH:/home/nginx/.local/share/pnpm

echo "📦 Preparando ambiente de deploy..."

[ -e $TMPDIR ] && rm -rf $TMPDIR
[ -e $WORKDIR ] && cp -af $WORKDIR $TMPDIR
# Garantir que o banco de dados exista no servidor
if [ ! -f "$WORKDIR/reconta.db" ]; then
  echo "💾 Criando banco de dados inicial..."
  touch "$WORKDIR/reconta.db"
fi
cd $TMPDIR || exit 1

git clean -fxd -e .env -e drizzle/reconta.db -e reconta.db
cp .env .env.production

echo "📥 Instalando dependências..."
pnpm install

echo "🗃️ Aplicando migrações do banco de dados..."
if ! pnpm drizzle-kit migrate; then
  echo "❌ Falha ao aplicar migrações. Abortando deploy."
  exit 1
fi
echo "✅ Migrações aplicadas com sucesso!"

# if ! pnpm run seed; then
#   echo "❌ Falha ao executar seed. Abortando deploy."
#   exit 1
# fi
#echo "✅ Seed concluído com sucesso!"

if pnpm run build; then
  echo "✅ Build concluído com sucesso!"
  sudo /usr/bin/systemctl stop $SERVICE
  [ -e $WORKDIR ] && rm -rf $WORKDIR
  [ -e $TMPDIR ] && cp -af $TMPDIR $WORKDIR
  # Garantir que o banco de dados persista
  if [ -f "$TMPDIR/reconta.db" ]; then
    cp -f "$TMPDIR/reconta.db" "$WORKDIR/reconta.db"
  fi
  sudo /usr/bin/systemctl start $SERVICE
  echo "🚀 Serviço reiniciado!"

  # Instala/atualiza os units do cron de notificações
  sudo cp $WORKDIR/files/reconta-cron.service /etc/systemd/system/
  sudo cp $WORKDIR/files/reconta-cron.timer /etc/systemd/system/
  sudo /usr/bin/systemctl daemon-reload
  sudo /usr/bin/systemctl enable --now reconta-cron.timer
  echo "⏰ Timer de notificações atualizado!"
fi