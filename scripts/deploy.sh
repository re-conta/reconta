#!/usr/bin/env bash

NAME="reconta"
TMPDIR="/tmp/$NAME"
WORKDIR="/var/www/$NAME"
SERVICE="${NAME}.service"
PATH=$PATH:/home/nginx/.local/share/pnpm

echo "📦 Preparando ambiente de deploy..."

[ -e $TMPDIR ] && rm -rf $TMPDIR
[ -e $WORKDIR ] && cp -af $WORKDIR $TMPDIR
cd $TMPDIR || exit 1

git clean -fxd -e .env -e drizzle/reconta.db
cp .env .env.production

if [ ! -f drizzle/meta/_journal.json ]; then
  echo "❌ drizzle/meta/_journal.json não encontrado. Verifique se os arquivos de migração foram commitados."
  exit 1
fi

echo "📥 Instalando dependências..."
pnpm install

echo "🗃️ Preparando journal de migrações (transição push → migrate)..."
pnpm tsx scripts/seed-journal.ts

echo "🗃️ Aplicando migrações do banco de dados..."
if ! pnpm drizzle-kit migrate; then
  echo "❌ Falha ao aplicar migrações. Abortando deploy."
  exit 1
fi
echo "✅ Migrações aplicadas com sucesso!"

if ! pnpm run seed; then
  echo "❌ Falha ao executar seed. Abortando deploy."
  exit 1
fi
echo "✅ Seed concluído com sucesso!"

if pnpm run build; then
  echo "✅ Build concluído com sucesso!"
  sudo /usr/bin/systemctl stop $SERVICE
  [ -e $WORKDIR ] && rm -rf $WORKDIR
  [ -e $TMPDIR ] && cp -af $TMPDIR $WORKDIR
  sudo /usr/bin/systemctl start $SERVICE
  echo "🚀 Serviço reiniciado!"
fi