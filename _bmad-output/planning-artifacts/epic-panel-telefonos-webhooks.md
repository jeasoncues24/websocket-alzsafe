---
project_name: 'wsapi'
user_name: 'Fulanito'
date: '2026-05-23'
status: 'listo para implementación'
target_branch: 'feature/panel-telefonos-webhooks'
---

# Epic 6: Panel de gestión — rediseño de teléfonos y módulo de webhooks

## Contexto

El panel admin de wsapi tiene una vista de teléfonos por empresa (`/empresas/[id]/telefonos`) con un diseño de lista plana que no comunica el estado operativo de cada número de un vistazo. Un admin de soporte que entra a diagnosticar un problema tiene que hacer varios clics antes de ver si hay webhooks registrados, si están fallando, o cuántas API keys activas tiene ese teléfono.

Adicionalmente, el Epic 5 completó la implementación backend de webhooks salientes (modelo, store, endpoints REST, worker de entrega, emisión de eventos), pero no existe ninguna interfaz en el panel para que el admin pueda inspeccionar los webhooks de un teléfono. El administrador hoy no tiene visibilidad de qué integradores registraron webhooks, si están activos, o si están acumulando fallos de entrega.

## Objetivo del epic

1. **Rediseñar** la vista de teléfonos con un layout bento grid que exponga de un vistazo el estado de cada número, sus API keys activas y sus webhooks registrados.
2. **Crear** una vista de diagnóstico de webhooks por teléfono (read-only) accesible desde cada card.
3. **Enriquecer** las respuestas del backend admin con los conteos necesarios para el frontend.

## Definición de done del epic

- El admin puede ver en la lista de teléfonos cuántas API keys y webhooks tiene cada número, sin hacer clic adicional.
- El admin puede navegar a `/empresas/[id]/telefonos/[id]/webhooks` y ver todos los webhooks registrados para ese teléfono, con su estado, eventos suscritos, conteo de fallos y último éxito.
- Todo el texto de la interfaz está en español.
- El layout usa bento grid con `shadcn/ui Card` — sin cambiar el sistema de diseño existente.
- El módulo de webhooks es estrictamente read-only para el admin.

## Rama obligatoria

`feature/panel-telefonos-webhooks`

---

## Story 6.1: Backend — Conteos de API keys y webhooks en respuesta admin de teléfonos

**Objetivo**: El endpoint `GET /api/admin/empresas/{id}/telefonos` devuelve, por cada teléfono, el número de API keys activas y el número de webhooks registrados.

**Cambios**:
- Añadir campos `ApiKeyCount int` y `WebhookCount int` al struct `domain.Telefono` (con tags `json:"api_key_count,omitempty"` y `json:"webhook_count,omitempty"`).
- En `AdminHandler.ListCompanyPhones`: para cada teléfono, consultar `ApiKeyStore.GetByTelefonoID(phone.ID)` y `WebhookStore.ListByTelefono(phone.ID)` y poblar los conteos.
- El `WebhookStore` ya está disponible en `AdminHandler` (se inyectará en esta story).

**Criterios de aceptación**:
- [ ] `GET /api/admin/empresas/{id}/telefonos` incluye `api_key_count` y `webhook_count` por teléfono
- [ ] Los conteos reflejan solo API keys `activo = true` y webhooks `activo = true`
- [ ] Si el `WebhookStore` es nil (sin DB), los conteos son 0 sin romper el handler
- [ ] `go build ./...` y `go test ./...` sin regresiones

**Esfuerzo**: pequeño (backend puro, no hay migración)

---

## Story 6.2: Frontend — Rediseño bento grid de lista de teléfonos

**Objetivo**: Reemplazar la lista plana actual con un bento grid que muestre el estado operativo completo de cada teléfono en su card.

**Cambios**:
- `frontend/app/empresas/[empresaId]/telefonos/page.tsx`: reescribir el renderizado de cards con CSS Grid (`grid-cols-1 md:grid-cols-2 lg:grid-cols-3`).
- Cada card muestra: número completo, badge de estado (Conectado / Desconectado / En espera / Desajuste), contadores mini de claves activas y webhooks, y los botones de acción.
- Nuevo botón `[Webhooks]` en cada card que navega a `/empresas/[id]/telefonos/[id]/webhooks`.
- Lógica de tamaño: teléfono activo con webhooks → `col-span-2`; resto → `col-span-1`.
- Actualizar `AdminTelefono` interface en `frontend/lib/api.ts` con `api_key_count` y `webhook_count`.
- Todos los labels en español.

**Criterios de aceptación**:
- [ ] La vista muestra un grid bento responsivo (1 col mobile, 2 col tablet, 3 col desktop)
- [ ] Cada card muestra: número, estado, `N claves activas`, `N webhooks`
- [ ] El botón `[Webhooks]` aparece en cada card y navega correctamente
- [ ] Los estados tienen labels en español: `Conectado`, `Desconectado`, `En espera`, `Desajuste`
- [ ] Loading state usa `Skeleton` en grid
- [ ] Empty state en español
- [ ] No hay regresión en las acciones existentes (Editar, Eliminar, API Keys, Conectar)

**Esfuerzo**: mediano (frontend, no hay cambios de backend)

---

## Story 6.3: Backend — Endpoint admin para listar webhooks de un teléfono

**Objetivo**: Exponer un endpoint admin que devuelve todos los webhooks registrados para un teléfono específico.

**Cambios**:
- Nuevo handler en `backend/internal/http/admin.go` (o archivo separado): `ListTelefonoWebhooks`
- Ruta: `GET /api/admin/telefonos/{id}/webhooks` (adminStack)
- Usa `WebhookStore.ListByTelefono(telefonoID)` — ya existe
- Respuesta: lista de webhooks con todos sus campos excepto `secret` (nunca exponer el secret en admin)
- Registrar en `routes_admin.go`

**Criterios de aceptación**:
- [ ] `GET /api/admin/telefonos/{id}/webhooks` responde `200` con lista de webhooks del teléfono
- [ ] El campo `secret` no aparece en la respuesta (json:"-" ya lo protege en el domain)
- [ ] Responde `404` si el teléfono no existe
- [ ] Responde `[]` (lista vacía) si el teléfono no tiene webhooks
- [ ] Solo accesible con token de admin (adminStack)
- [ ] `go build ./...` y `go test ./...` sin regresiones

**Esfuerzo**: pequeño

---

## Story 6.4: Frontend — Vista de webhooks por teléfono (read-only)

**Objetivo**: Nueva página en el panel admin que muestra todos los webhooks de un teléfono con información de diagnóstico.

**Cambios**:
- Nuevo archivo: `frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/webhooks/page.tsx`
- Nueva función en `frontend/lib/api.ts`: `getAdminTelefonoWebhooks(telefonoId)`
- Nuevo interface `AdminWebhook` en `frontend/lib/api.ts`
- Layout de la página:
  - Header: número de teléfono + badge de estado + breadcrumb de vuelta
  - Fila bento mini (3 cards): total activos, total inactivos, webhooks con fallos
  - Tabla: URL (truncada), eventos (badges), estado, conteo de fallos, último éxito, fecha creación
  - Accordion por fila: URL completa, prefijo de API key que lo registró, todos los eventos en badges
- Labels de eventos en español: `Mensaje recibido`, `Estado de mensaje`, `Sesión conectada`, `Sesión desconectada`
- Labels de estado en español: `Activo`, `Con fallos` (failure_count ≥ 5), `Inactivo`
- Empty state si no hay webhooks

**Criterios de aceptación**:
- [ ] La página carga correctamente desde el botón `[Webhooks]` de la card del teléfono
- [ ] Muestra resumen estadístico (3 mini-cards bento)
- [ ] Tabla con URL truncada, eventos como badges, estado, fallos y último éxito
- [ ] Al expandir una fila se ve la URL completa y el prefijo de la API key
- [ ] Estados y eventos en español
- [ ] Loading con `Skeleton`, empty state con `DataEmptyState`
- [ ] Breadcrumb funcional: volver a la lista de teléfonos

**Esfuerzo**: mediano

---

## Dependencias entre stories

```
6.1 (backend conteos) ──► 6.2 (frontend bento grid)
6.3 (backend endpoint webhooks) ──► 6.4 (frontend vista webhooks)
```

`6.1` y `6.3` pueden desarrollarse en paralelo. `6.2` y `6.4` dependen de sus respectivas stories de backend.
