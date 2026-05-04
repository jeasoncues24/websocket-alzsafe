# Story 1.9: Frontend Next.js — `.env.example` del proyecto

Status: review

## Story

As a desarrollador del frontend,
I want un archivo `frontend/.env.example` versionado y alineado con las variables reales que usa Next.js,
so that pueda crear `frontend/.env.local` correctamente sin adivinar qué configuración necesita el proyecto.

## Epic padre

- `epic-1`

## Contexto

El frontend ya usa variables de entorno reales en dos puntos críticos:

- `frontend/lib/api.ts` exige `NEXT_PUBLIC_API_URL`.
- `frontend/next.config.ts` usa `NEXT_INTERNAL_API_URL` y, si no existe, reutiliza `NEXT_PUBLIC_API_URL`.

Hasta ahora el repositorio tenía una inconsistencia: existía `frontend/.env.local.example`, pero el `frontend/README.md` instruía copiar `frontend/.env.example`. Esta story corrige esa discrepancia y deja una sola plantilla oficial para el frontend.

## Objetivo

Dejar explícito el contrato de configuración del frontend para que cualquier desarrollador sepa qué variables debe colocar en `frontend/.env.local` sin tocar backend ni despliegue fuera de Next.js.

## Alcance

- Crear una plantilla oficial versionada `frontend/.env.example`.
- Documentar únicamente las variables reales consumidas por el frontend actual.
- Alinear `frontend/README.md` y `frontend/.gitignore` con esa plantilla.
- Reemplazar la story 1.9 anterior por esta nueva definición y sincronizar `sprint-status.yaml`.

## Fuera de alcance

- Cambios en backend, Docker, PM2 o infraestructura de despliegue.
- Agregar variables nuevas no consumidas por el frontend actual.
- Cambiar la lógica de `frontend/lib/api.ts` o `frontend/next.config.ts`.

## Variables de entorno requeridas

La plantilla oficial es `frontend/.env.example`. Cada desarrollador debe copiarla a `frontend/.env.local` y ajustar sus valores locales.

| Variable | Obligatoria | Descripción |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | **Sí** | URL del backend accesible desde el navegador. |
| `NEXT_INTERNAL_API_URL` | No | URL interna para rewrites/proxy server-side de Next.js; si no se define, se usa `NEXT_PUBLIC_API_URL`. |
| `PORT` | No | Puerto local del servidor Next.js (`3000` por defecto). |

## Acceptance Criteria

- AC1. Existe un único archivo plantilla oficial versionado en `frontend/.env.example`.
- AC2. `frontend/.env.example` documenta `NEXT_PUBLIC_API_URL`, `NEXT_INTERNAL_API_URL` y `PORT` con comentarios claros y un ejemplo seguro para entorno local.
- AC3. `frontend/README.md` indica copiar `frontend/.env.example` a `frontend/.env.local` y explica el propósito de las variables.
- AC4. El frontend sigue manteniendo `frontend/.env.local` como archivo local no commiteado.
- AC5. `sprint-status.yaml` y esta story quedan alineados con la nueva story 1.9 y su fase de revisión.
- AC6. La story incluye epic padre, alcance, fuera de alcance, pruebas requeridas, riesgos/edge cases y dependencias/bloqueos.

## Tasks / Subtasks

- [x] Reemplazar `frontend/.env.local.example` por `frontend/.env.example` como plantilla oficial (AC: 1, 2)
- [x] Mantener en la plantilla solo variables realmente usadas por el frontend actual (AC: 2)
- [x] Ajustar `frontend/.gitignore` para permitir versionar `frontend/.env.example` sin dejar de ignorar archivos `.env` locales (AC: 1, 4)
- [x] Actualizar `frontend/README.md` para que la guía de configuración quede consistente con la plantilla oficial (AC: 3)
- [x] Reemplazar la story 1.9 anterior por esta nueva definición enfocada en `.env.example` (AC: 5, 6)
- [x] Sincronizar `sprint-status.yaml` con la nueva clave de la story 1.9 (AC: 5)

## Archivos creados/modificados

| Archivo | Estado | Descripción |
|---|---|---|
| `frontend/.env.example` | **Nuevo** | Plantilla oficial versionada de variables de entorno del frontend |
| `frontend/.gitignore` | **Actualizado** | Excepción para versionar `.env.example` manteniendo ignorados los `.env` locales |
| `frontend/README.md` | **Actualizado** | Guía de configuración del frontend alineada con la plantilla oficial |
| `_bmad-output/implementation-artifacts/sprint-status.yaml` | **Actualizado** | Story 1.9 renombrada al nuevo objetivo |

## Pruebas requeridas

- Verificar que exista `frontend/.env.example`.
- Verificar que `frontend/README.md` ya no dependa de `.env.local.example` como plantilla oficial.
- Verificar que `sprint-status.yaml` use la clave `1-9-frontend-env-example`.

## Riesgos y edge cases

- Si `.gitignore` sigue ignorando `.env.example`, la plantilla puede desaparecer del repositorio.
- Si el ejemplo de `NEXT_INTERNAL_API_URL` sugiere una URL válida solo en Docker, puede confundir a quien configure el entorno local.
- Si quedan dos stories 1.9 activas o el sprint status sigue con la clave vieja, se rompe la trazabilidad del backlog.

## Dependencias y bloqueos

- Depende del contrato actual de configuración en `frontend/lib/api.ts` y `frontend/next.config.ts`.
- No tiene bloqueos externos; el cambio se limita al frontend y a artefactos BMad.

## Notas de implementación

- `NEXT_PUBLIC_API_URL` es obligatoria y se usa en el cliente para construir requests al backend.
- `NEXT_INTERNAL_API_URL` es opcional y permite separar la URL interna del backend de la URL pública consumida por el navegador.
- `PORT` es opcional; si no se define, Next.js usa `3000`.
- `frontend/.env.local` debe seguir siendo local y no versionado.
