# WSAPI Producción

Guía rápida para desplegar WSAPI con Docker Compose.

## Inicio Rápido

1. Verifica puertos ocupados:

```bash
make check-ports
```

2. Copia `.env.example` a `.env` y completa estos valores:
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

3. Levanta la solución:

```bash
docker compose up -d --build
```

4. Verifica el estado:

```bash
docker compose ps
docker compose logs -f backend frontend mariadb
```

## Notas

- El backend usa MariaDB dentro de Compose.
- El navegador usa `NEXT_PUBLIC_API_URL`.
- Next.js usa `NEXT_INTERNAL_API_URL` para rewrites SSR.
- La documentación más completa está en `docker/production.md`.
