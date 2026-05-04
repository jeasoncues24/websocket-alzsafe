#!/usr/bin/env bash
# Compila el binario wsapi y lo deja en ./dist/wsapi.
# Uso:
#   ./scripts/build-backend.sh           # build normal (usa caché Docker)
#   ./scripts/build-backend.sh --no-cache # build limpio sin caché
# Debe ejecutarse desde cualquier directorio; usa rutas relativas al repo.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$REPO_ROOT/backend/.env"
DIST_DIR="$REPO_ROOT/dist"
DIST_BIN="$DIST_DIR/wsapi"
NO_CACHE=""

for arg in "$@"; do
  case "$arg" in
    --no-cache) NO_CACHE="--no-cache" ;;
  esac
done

# Limpia el binario anterior antes de validar para no dejar artefactos viejos
# si la validación falla en ejecuciones siguientes.
rm -f "$DIST_BIN"

if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: Falta backend/.env." >&2
  echo "       Copia backend/.env.copy como backend/.env y configura las variables antes de compilar." >&2
  exit 1
fi

mkdir -p "$DIST_DIR"

cd "$REPO_ROOT"
echo "Compilando wsapi... (contexto: ./backend)"
# shellcheck disable=SC2086
docker compose build $NO_CACHE backend-build
docker compose run --rm backend-build

echo "Binario disponible en $DIST_BIN"
