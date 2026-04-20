# Producción WSAPI

Guía mínima para desplegar WSAPI con Docker Compose en producción.

## Requisitos

- Docker y Docker Compose instalados.
- MariaDB/MySQL accesible desde el contenedor backend.
- Puertos libres para `APP_PORT` y `FRONTEND_PORT`.

Si la base de datos corre en el mismo host, usa `DB_HOST=host.docker.internal` en `.env`.

## Variables

1. Copia `.env.example` a `.env` y completa:

- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASS`
- `APP_PORT`
- `FRONTEND_PORT`

2. Si vas a usar frontend local fuera de Compose, copia `frontend/.env.example` a `frontend/.env.local`.

## Verificar puertos

Ejecuta:

```bash
bash docker/check-ports.sh
```

Si alguno está ocupado, cambia manualmente `APP_PORT` o `FRONTEND_PORT` en `.env` y vuelve a correr el script.

## Build y arranque

```bash
docker compose build
docker compose up -d
```

## Validación

- Backend: `http://localhost:${APP_PORT}`
- Frontend: `http://localhost:${FRONTEND_PORT}`

## Operación

- Logs: `docker compose logs -f backend frontend`
- Estado: `docker compose ps`
- Detener: `docker compose down`

## Builds con Makefile

```bash
make build-prod
make docker-build
```

- `build-prod` genera artefactos de producción locales.
- `docker-build` construye las imágenes de Compose.
