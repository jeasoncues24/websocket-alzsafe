# Story 1.9: Frontend Next.js — servidor de producción con PM2

Status: ready-for-dev

## Story

As a operador de despliegue,
I want correr el frontend Next.js en producción con `npm run start` gestionado por PM2,
so that el panel de administración esté disponible en un puerto estable con reinicio automático ante fallos.

## Contexto

El frontend (Next.js 16.2.3) requiere un servidor Node.js para correr en producción (`next start`). Esto mantiene el proxy del `next.config.ts` activo (rewrites hacia el backend Go) y no requiere cambios en la arquitectura actual.

El frontend se gestiona con PM2 igual que el backend, siguiendo el patrón ya establecido en el proyecto.

## Variables de entorno requeridas

El frontend necesita `frontend/.env.local` (no commitear). Usar `frontend/.env.local.example` como base.

| Variable | Obligatoria | Descripción |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | **Sí** | URL del backend accesible desde el browser (ej: `http://mi-servidor:8080`) |
| `NEXT_INTERNAL_API_URL` | No | URL interna del backend para el proxy server-side. Si no se define, usa `NEXT_PUBLIC_API_URL` |
| `PORT` | No | Puerto del servidor Next.js (default: 3000) |

**Importante:** `NEXT_PUBLIC_API_URL` se incrusta en el bundle JS en build time. Si cambia la URL del backend, hay que rebuildar el frontend (`make build`).

## Cómo funciona el proxy

El `next.config.ts` reescribe las rutas `/api/*`, `/admin/*`, `/metrics` y `/ws` al backend Go. El browser habla siempre con Next.js (ej: `http://servidor:3000/api/...`) y Next.js lo reenvía al backend. Esto evita que el browser haga peticiones cross-origin directamente al backend.

## Acceptance Criteria

1. `make build` en `frontend/` compila el proyecto sin errores.
2. `make start` registra el proceso en PM2 como `wsapi-frontend` y arranca en el puerto configurado.
3. El servidor responde en `http://localhost:PORT` con el panel de administración.
4. `make restart`, `make stop` y `make logs` funcionan correctamente con PM2.
5. El proceso se reinicia automáticamente si cae (PM2 lo gestiona).
6. `frontend/.env.local.example` existe con todas las variables documentadas.

## Tasks / Subtasks

- [x] Crear `frontend/.env.local.example` con `NEXT_PUBLIC_API_URL`, `NEXT_INTERNAL_API_URL` y `PORT` documentados (AC: 6)
- [x] Crear `frontend/Makefile` con targets: `install`, `build`, `start`, `restart`, `stop`, `logs`, `dev` (AC: 2, 3, 4)
- [ ] Crear `frontend/.env.local` en el servidor de producción a partir del `.env.local.example` (AC: 1, 3) — **tarea manual del operador**
- [ ] Verificar `make build` compila sin errores en el entorno de producción (AC: 1)
- [ ] Verificar `make start` levanta el servidor y PM2 muestra `wsapi-frontend` como online (AC: 2, 3, 5)

## Archivos creados/modificados

| Archivo | Estado | Descripción |
|---|---|---|
| `frontend/.env.local.example` | **Nuevo** | Plantilla de variables de entorno con documentación |
| `frontend/Makefile` | **Nuevo** | Targets build/start/restart/stop/logs/dev para PM2 |

## Notas técnicas

- El proceso PM2 se llama `wsapi-frontend` para no colisionar con el proceso backend `wsapi`.
- El puerto es configurable: `make start PORT=4000` o definiendo `PORT` en `.env.local`.
- `next start` usa la variable `PORT` del entorno automáticamente. El Makefile también la pasa como flag `-p $(PORT)` para garantizar consistencia.
- PM2 debe estar instalado globalmente. El `make start` lo instala automáticamente si no está presente.
