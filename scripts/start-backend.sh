#!/usr/bin/env bash
# Inicia wsapi en el host usando PM2.
# Idempotente: no duplica el proceso si ya está registrado en PM2.
# La base de datos debe estar corriendo en el host; este script no levanta Docker.
# Uso:
#   ./scripts/start-backend.sh
# Puede ejecutarse desde cualquier directorio.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST_BIN="$REPO_ROOT/dist/wsapi"
BACKEND_DIR="$REPO_ROOT/backend"
APP_NAME="wsapi"

if [ ! -f "$DIST_BIN" ]; then
  echo "ERROR: No se encontró $DIST_BIN." >&2
  echo "       Ejecuta scripts/build-backend.sh primero." >&2
  exit 1
fi

if ! command -v pm2 &>/dev/null; then
  echo "PM2 no encontrado. Instalando globalmente con npm..."
  npm install -g pm2
fi

if pm2 describe "$APP_NAME" &>/dev/null 2>&1; then
  echo "wsapi ya está registrado en PM2. No se crea duplicado."
  echo "Para reiniciar usa: pm2 restart $APP_NAME"
  pm2 status "$APP_NAME"
  exit 0
fi

echo "Iniciando wsapi con PM2..."
# --cwd apunta a backend/ para que godotenv.Load() encuentre backend/.env
pm2 start "$DIST_BIN" --name "$APP_NAME" --cwd "$BACKEND_DIR"
echo "wsapi activo. Comandos útiles:"
echo "  pm2 logs $APP_NAME      — ver logs en tiempo real"
echo "  pm2 status              — estado del proceso"
echo "  pm2 stop $APP_NAME      — detener"
echo "  pm2 restart $APP_NAME   — reiniciar"
