# WSAPI Frontend

## Entorno

1. Copia `frontend/.env.example` a `frontend/.env.local`.
2. Define `NEXT_PUBLIC_API_URL` con la URL pública del backend accesible desde el navegador.
3. Define `NEXT_INTERNAL_API_URL` solo si el servidor Next.js necesita hablar con el backend por una URL interna distinta.
4. Si lo necesitas, define `PORT` para cambiar el puerto de `next dev` / `next start`.

Ejemplo:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_INTERNAL_API_URL=http://localhost:8080
PORT=3000
```

Notas:

- `frontend/.env.local` es local y no debe commitearse.
- Si omites `NEXT_INTERNAL_API_URL`, `next.config.ts` reutiliza `NEXT_PUBLIC_API_URL`.
- `NEXT_PUBLIC_API_URL` es obligatoria porque `frontend/lib/api.ts` falla al iniciar si no existe.

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
