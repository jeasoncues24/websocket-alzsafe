# WSAPI Producción

Guía rápida para desplegar WSAPI con Docker Compose y `Makefile`.

## Flujo Recomendado

1. Verifica puertos ocupados:

```bash
make check-ports
```

2. Copia `.env.example` a `.env` y completa:
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASS`
- `APP_PORT`
- `FRONTEND_PORT`

3. Compila artefactos de producción locales:

```bash
make build-prod
```

4. Construye las imágenes Docker:

```bash
make docker-build
```

5. Levanta la solución:

```bash
docker compose up -d
```

## Qué Hace Cada Comando

- `make check-ports`: revisa si `APP_PORT` y `FRONTEND_PORT` están ocupados y sugiere alternativas.
- `make build-prod`: compila el backend en `dist/wsapi` y ejecuta el build de producción del frontend.
- `make docker-build`: construye las imágenes definidas en `docker-compose.yml`.
- `docker compose up -d`: levanta backend y frontend en segundo plano.

## Validación

```bash
docker compose ps
docker compose logs -f backend frontend
```

## Notas

- Si la base de datos corre en el host, usa `DB_HOST=host.docker.internal` en `.env`.
- El frontend usa `NEXT_PUBLIC_API_URL` para apuntar al backend.
- La documentación más completa está en `docker/production.md`.
