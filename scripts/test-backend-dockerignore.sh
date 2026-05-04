#!/usr/bin/env bash
# Verifica reglas mínimas de backend/.dockerignore para el build context del backend.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DOCKERIGNORE="$REPO_ROOT/backend/.dockerignore"

fail() {
  echo "FAIL: $1" >&2
  exit 1
}

require_line() {
  local value="$1"
  grep -Fxq "$value" "$DOCKERIGNORE" || fail "Falta la regla exacta: $value"
}

require_comment() {
  local value="$1"
  grep -Fq "$value" "$DOCKERIGNORE" || fail "Falta el comentario de sección: $value"
}

forbidden_line() {
  local value="$1"
  if grep -Fxq "$value" "$DOCKERIGNORE"; then
    fail "No se debe excluir un archivo/ruta requerida para compilar: $value"
  fi
}

[ -f "$DOCKERIGNORE" ] || fail "No existe $DOCKERIGNORE"

require_comment "# Secretos / entorno"
require_comment "# Documentación"
require_comment "# Artefactos de test y cobertura"
require_comment "# Metadata de herramientas y temporales"

require_line ".env"
require_line ".env.*"
require_line ".env.copy"
require_line "docs/"
require_line "*.md"
require_line "*.test"
require_line "coverage.txt"
require_line "*.log"
require_line ".gitignore"
require_line "*.swp"
require_line "*.swo"

forbidden_line "go.mod"
forbidden_line "go.sum"
forbidden_line "main.go"
forbidden_line "internal/"
forbidden_line "cmd/"

echo "OK: backend/.dockerignore cumple reglas esperadas"
