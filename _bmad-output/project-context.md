---
project_name: 'wsapi'
user_name: 'Fulanito'
date: '2026-04-26'
sections_completed: ['technology_stack', 'project_structure', 'implementation_rules', 'bmad_rules']
existing_patterns_found: 12
---

# Project Context for AI Agents

Este archivo contiene reglas críticas y patrones que los agentes IA deben seguir al implementar código en este proyecto. Prioriza detalles que el agente podría olvidar o inferir mal.

## Technology Stack & Versions

### Backend

- Lenguaje: Go `1.25.0`.
- Módulo: `wsapi`.
- Servidor HTTP estándar: `net/http`.
- Router/contenedor propio bajo `backend/internal/http`.
- Base de datos principal: MySQL/MariaDB vía `github.com/go-sql-driver/mysql v1.9.3`.
- Migraciones: runner propio bajo `backend/internal/storage/migration.go`; dependencia `github.com/golang-migrate/migrate/v4 v4.19.1` está disponible.
- WhatsApp: `go.mau.fi/whatsmeow`.
- SQLite para sesiones WhatsApp: `modernc.org/sqlite v1.18.1`.
- JWT: `github.com/golang-jwt/jwt/v5 v5.3.1`.
- Logging: `github.com/rs/zerolog v1.34.0`.
- Docker backend: `golang:1.25-bookworm` build stage y `debian:bookworm-slim` runtime.

### Frontend

- Next.js `16.2.3`.
- React `19.2.4` y React DOM `19.2.4`.
- TypeScript `^5` con `strict: true`.
- Tailwind CSS `^4.2.2`.
- ESLint `^9` con `eslint-config-next 16.2.3`.
- Estado cliente: `zustand ^5.0.12`.
- Componentes UI estilo shadcn: `class-variance-authority`, `clsx`, `tailwind-merge`, Radix Dialog.
- Iconos: `lucide-react`.

## Project Structure

- Backend: `backend/`.
- Frontend: `frontend/`.
- Docker: `docker/` y `docker-compose.yml`.
- Artefactos BMad: `_bmad-output/`.
- Reglas BMad del proyecto: `docs/bmad-project-rules.md`.
- Overrides BMad: `_bmad/custom/`.

### Backend structure

- `backend/main.go`: entrypoint real del binario `wsapi`; maneja server y comandos de migración.
- `backend/cmd/api/main.go`: actualmente solo declara paquete `api`; no asumir que es entrypoint principal.
- `backend/internal/config`: configuración y JWT/logger.
- `backend/internal/http`: handlers, router, middleware, rutas admin/API, responses.
- `backend/internal/domain`: modelos/reglas de dominio.
- `backend/internal/storage`: persistencia, migraciones, repositorios.
- `backend/internal/auth`: auth, API keys y JWT de empresa.
- `backend/internal/whatsapp`: WhatsApp client/session/QR/send/broadcast.
- `backend/internal/metrics`: contadores/métricas.

### Frontend structure

- `frontend/app`: App Router de Next.js; páginas administrativas y layout.
- `frontend/components/ui`: componentes UI reutilizables.
- `frontend/components/*`: componentes por dominio visual.
- `frontend/lib/api.ts`: cliente API centralizado.
- `frontend/lib/utils.ts`: utilidades compartidas como `cn`.
- `frontend/stores`: stores Zustand.

## 🚨 Regla de Rama por Epic — IMPERATIVA

Cada epic tiene una rama Git dedicada. El código fuente (`frontend/`, `backend/`) **solo puede modificarse en la rama asignada al epic**.

| Epic | Rama obligatoria |
|------|-----------------|
| Epic 3 — Hardening de Seguridad y Calidad Frontend | `feature/security` |
| Epic 5 — Integración Loyo (provider B2B webhooks) | `feature/integracion-loyo` |

**Verificación previa obligatoria para agentes de implementación:**
```bash
git branch --show-current  # debe coincidir con la rama del epic activo
```
Si no coincide → detener implementación, notificar al usuario, cambiar de rama primero.

Los artefactos BMad (`_bmad-output/`, `docs/`) pueden editarse en cualquier rama.

---

## Critical Implementation Rules

### General

- Mantener idioma español en textos de usuario, errores existentes, rutas administrativas y documentos BMad, salvo nombres técnicos/código.
- Antes de cambiar patrones, leer archivos vecinos y copiar convenciones existentes.
- No introducir librerías nuevas sin necesidad clara y sin registrar la decisión en story/arquitectura.
- Preferir cambios pequeños, testeables y reversibles.
- Mantener compatibilidad con los nombres de endpoints y payloads existentes.

### MySQL / Migraciones

- **OBLIGATORIO:** Para cualquier tarea que involucre MySQL (escribir o modificar un CREATE TABLE, ALTER TABLE, consulta con JOIN/GROUP BY/agregaciones, decisión de índices), invocar primero la skill `/sql-optimization` instalada en `.agents/skills/sql-optimization/`.
- Las migraciones se definen en `backend/internal/storage/migrations/` con el patrón `NNN_descripcion_accion.up.sql` / `.down.sql`. Cada par de archivos define exactamente una acción (crear, eliminar o modificar una tabla). No distribuir columnas de una misma tabla en múltiples archivos de migración.
- El `.down.sql` de una migración debe revertir exactamente lo que hace el `.up.sql`, sin efectos secundarios sobre otras tablas.
- En fase de desarrollo activo se permite renombrar/eliminar archivos de migración existentes siempre que el mantenedor recree la BD desde cero. Nunca modificar migraciones históricas en producción.
- Tablas canónicas de telemetría (no usar las obsoletas `api_key_usage_events` ni `api_key_usage_daily`):
  - `telefono_request_logs` — trazas individuales de request.
  - `telefono_metrics_min` — agregados por minuto (bucket_min).
  - `api_key_audit_events` — eventos de ciclo de vida de una API key.

### Backend Go

- Usar el módulo `wsapi`; imports internos deben comenzar con `wsapi/internal/...`.
- Respetar separación: domain no debe depender de HTTP; storage no debe mezclar lógica de presentación; HTTP coordina request/response.
- Para nuevos endpoints, revisar `backend/internal/http/routes_api.go`, `routes_admin.go`, `handlers.go`, `admin.go`, `response.go`, `middleware.go` antes de implementar.
- Para persistencia, agregar o modificar storage/domain de forma coherente con repositorios existentes.
- Para cambios de esquema, usar el mecanismo de migraciones existente; no modificar producción implícitamente desde handlers.
- Manejar errores explícitamente; no ignorar errores de DB, JSON, HTTP o WhatsApp.
- Mantener comandos de verificación: `cd backend && go test ./...` y `cd backend && go build ./...`.

### Frontend Next.js / TypeScript

- TypeScript está en modo `strict`; evitar `any` salvo cuando se interactúe con payloads legacy y se justifique.
- Usar alias `@/*` para imports desde raíz de `frontend` cuando aplique.
- `NEXT_PUBLIC_API_URL` es obligatorio; `frontend/lib/api.ts` lanza error si falta.
- Centralizar llamadas API en `frontend/lib/api.ts` o módulos equivalentes; no duplicar lógica fetch/auth en páginas sin razón.
- Mantener headers auth existentes con `localStorage.getItem("admin_token")` para rutas admin cliente.
- Componentes UI deben seguir el patrón existente de `components/ui`: `React.forwardRef`, `cva`, `VariantProps`, `cn` cuando corresponda.
- Ejecutar `cd frontend && npm run lint` y, si el cambio lo amerita, `cd frontend && npm run build`.

### Docker / Build

- El binario backend esperado se llama `wsapi`.
- Docker compose actual compila backend con `docker compose run --rm backend-build` y copia el binario a `./dist/wsapi`.
- El backend runtime expone puerto `8080`, pero el server real usa `APP_PORT`; no asumir puerto si `.env` define otro.

## BMad Workflow Rules

- Reglas completas: `docs/bmad-project-rules.md`.
- Planning artifacts: `_bmad-output/planning-artifacts/`.
- Implementation artifacts: `_bmad-output/implementation-artifacts/`.
- Sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.

### Required sequence for solid work

1. Crear o actualizar PRD con `bmad-create-prd`.
2. Crear arquitectura con `bmad-create-architecture`.
3. Crear epics/stories con `bmad-create-epics-and-stories`.
4. Validar readiness con `bmad-check-implementation-readiness`.
5. Generar sprint status con `bmad-sprint-planning`.
6. Crear story con `bmad-create-story`.
7. Validar story con `bmad-create-story:validate`.
8. Implementar con `bmad-dev-story`.
9. Revisar con `bmad-code-review`.
10. Generar QA/E2E si corresponde con `bmad-qa-generate-e2e-tests`.

### Status policy

- Epic: `backlog`, `in-progress`, `done`.
- Story: `backlog`, `ready-for-dev`, `in-progress`, `review`, `changes-requested`, `done`, `blocked`, `deferred`.
- Retrospective: `optional`, `done`.

### Definition of Ready

Una story no está lista si le faltan ACs testeables, alcance, fuera de alcance, dependencias, pruebas o condición de desbloqueo.

### Definition of Done

Una story no está terminada si no cumple ACs, no compila, no tiene pruebas/verificación, no pasó review o no actualizó sprint status.

## Preferencias de salida para Sprint Status

Cuando el usuario pida estado de sprint, pendientes o "resumen", la respuesta debe incluir siempre tablas Markdown.

Formato mínimo:

1. Tabla de pendientes con columnas: `Orden`, `Story`, `Estado`, `Siguiente acción`.
2. Resumen de conteos de stories/epics en formato compacto.
3. Recomendación explícita del próximo workflow.

Reglas de estilo:

- Priorizar tablas sobre párrafos largos.
- Acciones concretas y cortas.
- Orden natural por story (1-1, 1-2, 2-1...).
- Si hay riesgos, incluirlos al final en lista breve o tabla.
