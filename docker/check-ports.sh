#!/usr/bin/env bash
set -euo pipefail

backend_port="${1:-${APP_PORT:-8080}}"
frontend_port="${2:-${FRONTEND_PORT:-3000}}"
mariadb_port="${3:-3306}"

port_in_use() {
  local port="$1"

  if command -v ss >/dev/null 2>&1; then
    ss -ltnH "sport = :${port}" 2>/dev/null | grep -q .
    return
  fi

  if command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
    return
  fi

  echo "No se encontró ss ni lsof para verificar puertos." >&2
  exit 2
}

next_free_port() {
  local port="$1"
  local tries=0

  while port_in_use "$port"; do
    port=$((port + 1))
    tries=$((tries + 1))
    if [ "$tries" -gt 50 ]; then
      break
    fi
  done

  printf '%s' "$port"
}

check_port() {
  local label="$1"
  local port="$2"
  local env_name="$3"

  if port_in_use "$port"; then
    suggestion="$(next_free_port "$port")"
    echo "${label} ocupado: ${port}"
    echo "Sugerencia para ${env_name}: ${suggestion}"
    return 1
  fi

  echo "${label} libre: ${port}"
}

status=0

check_port "Backend" "$backend_port" "APP_PORT" || status=1
check_port "Frontend" "$frontend_port" "FRONTEND_PORT" || status=1
check_port "MariaDB" "$mariadb_port" "MariaDB host port" || status=1

if [ "$status" -ne 0 ]; then
  echo "Actualiza manualmente .env y vuelve a ejecutar el despliegue."
fi

exit "$status"
