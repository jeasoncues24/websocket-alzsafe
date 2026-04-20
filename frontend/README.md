# WSAPI Frontend

## Entorno

1. Copia `frontend/.env.example` a `frontend/.env.local`.
2. Define `NEXT_PUBLIC_API_URL` con la URL del backend.

Ejemplo:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Desarrollo

```bash
npm install
npm run dev
```

## Producción

El flujo recomendado para producción se documenta en el README raíz y en `docker/production.md`.

Resumen:

```bash
make check-ports
make build-prod
make docker-build
docker compose up -d
```

## Reglas

- No hardcodear `localhost`, IPs o puertos en el código fuente.
- Mantener `NEXT_PUBLIC_API_URL` en `frontend/.env.local` o en el build de Docker.
