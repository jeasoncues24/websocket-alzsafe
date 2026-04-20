# Producción WSAPI

Guía mínima para desplegar WSAPI con Docker Compose en producción.

## Requisitos

- Docker y Docker Compose instalados.
- Puertos libres para `APP_PORT`, `FRONTEND_PORT` y `3306`.

## Variables

1. Copia `.env.example` a `.env` y completa:
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASS`
- `MARIADB_ROOT_PASSWORD`
- `APP_ENV`
- `APP_PORT`
- `FRONTEND_PORT`
- `NEXT_PUBLIC_API_URL`
- `NEXT_INTERNAL_API_URL`

2. Si vas a usar frontend local fuera de Compose, copia `frontend/.env.example` a `frontend/.env.local`.

## Verificar puertos

Ejecuta:

```bash
bash docker/check-ports.sh
```

Si alguno está ocupado, cambia manualmente `APP_PORT`, `FRONTEND_PORT` o el puerto de MariaDB y vuelve a correr el script.

## Build y arranque

```bash
docker compose up -d --build
```

## Validación

- Backend: `http://127.0.0.1:${APP_PORT}`
- Frontend: `http://127.0.0.1:${FRONTEND_PORT}`
- MariaDB: `127.0.0.1:3306`

## Operación

- Logs: `docker compose logs -f backend frontend mariadb`
- Estado: `docker compose ps`
- Detener: `docker compose down`
