# WSAPI Frontend

## Entorno

1. Copia `frontend/.env.example` a `frontend/.env.local`.
2. Define `NEXT_PUBLIC_API_URL` con la URL publica del backend.
3. Define `NEXT_INTERNAL_API_URL` con la URL interna para rewrites SSR.

Ejemplo:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_INTERNAL_API_URL=http://localhost:8080
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
docker compose up -d --build
docker compose ps
```

Si ejecutas el frontend fuera de Docker, puedes apuntar ambas URLs al backend local.

## Reglas

- No hardcodear `localhost`, IPs o puertos en el código fuente.
- Mantener `NEXT_PUBLIC_API_URL` en `frontend/.env.local` o en el build de Docker.
- Mantener `NEXT_INTERNAL_API_URL` para rewrites SSR cuando uses Docker o despliegue remoto.
