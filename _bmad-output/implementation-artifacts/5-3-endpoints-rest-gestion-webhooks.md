---
story_id: "5.3"
epic: "epic-5"
title: "Endpoints REST para gestión de webhooks"
status: review
estimated_days: 1
priority: high
branch: "feature/integracion-loyo"
skills: ["bmad-code-review"]
affects:
  - backend/internal/http/handlers/v1_webhooks.go
  - backend/internal/http/routes_api.go
  - backend/internal/http/container.go
  - backend/internal/http/handlers/v1_webhooks_test.go
  - docs/webhooks-integracion.md
  - _bmad-output/implementation-artifacts/sprint-status.yaml
---

# Story 5.3: Endpoints REST para gestión de webhooks

Status: in-progress

## Story

Como integrador B2B que usa wsapi como provider de WhatsApp,
quiero poder registrar, listar y eliminar mis webhooks vía API REST autenticada con API key,
para configurar notificaciones en tiempo real sin necesidad del panel admin.

## Contexto técnico

El WebhookStore ya existe (`backend/internal/storage/webhook_store.go`) con métodos Create, ListByEmpresa, GetByID, Delete.
El modelo de dominio ya existe (`backend/internal/domain/webhook.go`) con tipos Webhook, WebhookEvent, WebhookQueueItem.
Las migraciones 017 y 018 ya están creadas y validadas con sql-optimization.

Esta story expone ese store vía HTTP usando el mismo patrón que los handlers V1 existentes.

## Acceptance Criteria

**AC1 — POST /api/service/v1/webhooks crea webhook:**
**Dado** un request autenticado con API key,
**Cuando** se hace POST a `/api/service/v1/webhooks` con body `{"url": "https://ejemplo.com/hook", "eventos": ["message.received"]}`,
**Entonces** responde `HTTP 201` con `{"ok": true, "data": {"id": <nuevo_id>, "secret": "<hex_32_bytes>"}}`
**Y** el secret se genera con `crypto/rand` (32 bytes → 64 chars hex)
**Y** el webhook queda persistido en `webhooks_outbound` con `empresa_id` del token.

**AC2 — POST valida URL HTTPS:**
**Dado** un request con `url` que no empieza con `https://`,
**Cuando** se hace POST,
**Entonces** responde `HTTP 400` con error `INVALID_URL`.

**AC3 — POST valida eventos:**
**Dado** un request con `eventos` que contiene un valor no válido (ej: `["invalid_event"]`),
**Cuando** se hace POST,
**Entonces** responde `HTTP 400` con error `INVALID_EVENTOS`.

**AC4 — POST respeta límite de webhooks por empresa:**
**Dado** que la empresa ya tiene `WEBHOOKS_MAX_PER_EMPRESA` webhooks activos (default 10),
**Cuando** se hace POST,
**Entonces** responde `HTTP 400` con error `MAX_WEBHOOKS_REACHED`.

**AC5 — GET /api/service/v1/webhooks lista webhooks:**
**Dado** un request autenticado,
**Cuando** se hace GET a `/api/service/v1/webhooks`,
**Entonces** responde `HTTP 200` con lista de webhooks de la empresa
**Y** ningún webhook incluye el campo `secret` en la respuesta.

**AC6 — DELETE /api/service/v1/webhooks/{id} elimina webhook:**
**Dado** un webhook existente perteneciente a la empresa del token,
**Cuando** se hace DELETE a `/api/service/v1/webhooks/{id}`,
**Entonces** responde `HTTP 200` con `{"ok": true}`
**Y** el webhook es eliminado de la BD (con CASCADE a la cola).

**AC7 — DELETE 404 si webhook no existe o no pertenece:**
**Dado** un ID que no existe o no pertenece a la empresa del token,
**Cuando** se hace DELETE,
**Entonces** responde `HTTP 404`.

**AC8 — Verificación de build:**
**Dado** los cambios aplicados,
**Cuando** se ejecutan `cd backend && go build ./...` y `cd backend && go test ./...`,
**Entonces** ambos comandos terminan sin errores ni regresiones.

## Tasks / Subtasks

- [x] **T1 — Handler v1_webhooks.go** (AC1-AC7)
  - [x] Struct `V1WebhooksHandler` con `webhookStore *storage.WebhookStore`, `maxWebhooks int`
  - [x] Constructor `NewV1WebhooksHandler(store *storage.WebhookStore, maxWebhooks int) *V1WebhooksHandler`
  - [x] Método `List` → GET, leer empresa de context, llamar ListByEmpresa, armar respuesta sin secret
  - [x] Método `Create` → POST, validar JSON, validar HTTPS, validar eventos, contar activos vs límite, generar secret con `crypto/rand`, llamar Create, responder 201
  - [x] Método `Delete` → DELETE, extraer ID del path, validar pertenencia a empresa, llamar Delete
  - [x] Método router dispatch único cubierto con métodos por ruta registrados en `routes_api.go`

- [x] **T2 — Container** (AC1)
  - [x] Añadir campo `V1WebhooksHandler *handlers.V1WebhooksHandler` al struct Container
  - [x] Instanciar en bloque `if cfg.DBHost != "" { ... }` con webhookStore y WEBHOOKS_MAX_PER_EMPRESA (default 10)
  - [x] Cargar `WEBHOOKS_MAX_PER_EMPRESA` del entorno

- [x] **T3 — Routes** (AC1, AC5, AC6)
  - [x] En `routes_api.go`, añadir bloque con clientStack para `POST`, `GET` y `DELETE` de `/api/service/v1/webhooks`

- [x] **T4 — Tests** (AC8)
  - [x] Tests de handler con `httptest`
  - [x] Cubrir POST exitoso, validación URL, validación eventos, límite, GET, DELETE exitoso, DELETE 404

- [x] **T5 — Verificación final** (AC8)
  - [x] `cd backend && go build ./...` sin errores
  - [x] `cd backend && go test ./...` sin regresiones

## Dev Notes

### Patrón de handlers V1 del proyecto

Verificado en `v1_phones.go`, `v1_sessions.go`:

1. Archivo en `backend/internal/http/handlers/`, package `http` (NO sub-package)
2. Struct con dependencias inyectadas + constructor `NewXxxHandler(...)`. Receiver methods.
3. Cada método valida método HTTP con `writeV1Error` arriba
4. Claims de auth: `domain.GetApiKeyClaims(r.Context())` → `claims.EmpresaID`
5. Response exitosa: `writeV1Success(w, map[string]interface{}{...}, empresaID)`
6. Response error: `writeV1Error(w, status, "ERROR_CODE", "mensaje")`

### Extracción de ID de path

Para `DELETE /api/service/v1/webhooks/{id}`, Go 1.22+ `http.ServeMux` expone `r.PathValue("id")`:

```go
idStr := r.PathValue("id")
id, err := strconv.ParseInt(idStr, 10, 64)
```

No usar `extractTelefonoID` ni `extractAPIKeyID` (usan string splitting para rutas admin legacy).

### Generación de secret

```go
import "crypto/rand"
import "encoding/hex"

b := make([]byte, 32)
if _, err := rand.Read(b); err != nil {
    // error handling
}
secret := hex.EncodeToString(b)
```

### Límite de webhooks

- Variable de entorno `WEBHOOKS_MAX_PER_EMPRESA`, default `10`
- Se pasa al handler desde container.go
- Para contar webhooks activos actuales: llamar `ListByEmpresa` y contar `len(webhooks)` (o filtrar activos)

### Protección del secret

- `Webhook.Secret` tiene tag `json:"-"` → nunca serializa en JSON automáticamente
- En Create: devolver secret en campo separado de la respuesta (solo una vez)
- En List: no incluir secret en la respuesta

### Validación de eventos

```go
var validEvents = map[domain.WebhookEvent]bool{
    domain.WebhookEventMessageReceived:    true,
    domain.WebhookEventMessageStatus:      true,
    domain.WebhookEventSessionConnected:   true,
    domain.WebhookEventSessionDisconnected: true,
}
```

### Validación de URL

```go
if !strings.HasPrefix(req.URL, "https://") {
    writeV1Error(w, http.StatusBadRequest, "INVALID_URL", "URL debe ser HTTPS")
    return
}
```

En desarrollo local se podría permitir HTTP (vía env flag), pero para MVP solo HTTPS.

### Dependencias

- **Depende de**: Story 5.2 (modelo + store + migraciones) — ✅ completada
- **Bloquea a**: Story 5.4 (worker), Story 5.5 (eventos)
- **Bloquea a**: Loyo Story 1.7 (registro de webhook desde panel Loyo)

### References

- Patrón handler: `backend/internal/http/handlers/v1_phones.go`
- Response helpers: `backend/internal/http/handlers/v1_helpers.go`
- Cableo de rutas: `backend/internal/http/routes_api.go`
- Container: `backend/internal/http/container.go:51-173`
- WebhookStore: `backend/internal/storage/webhook_store.go`
- Domain model: `backend/internal/domain/webhook.go`
- Epic origen: `_bmad-output/planning-artifacts/epic-integracion-loyo.md:115-133`
- Rama: `feature/integracion-loyo` (verificar con `git branch --show-current`)

## Dev Agent Record

### Debug Log References

- Se añadió `backend/internal/http/handlers/v1_webhooks_test.go` con cobertura de creación, validaciones, listado y borrado.
- Se endureció la validación de URL en `v1_webhooks.go` usando parseo real (`net/url`) + `scheme=https` + host requerido.
- Validaciones ejecutadas: `cd backend && go build ./...` y `cd backend && go test ./...`.

### Completion Notes

- AC1–AC8 verificados con código y suite de tests.
- POST ahora rechaza URLs HTTPS mal formadas como `https://`, evitando persistir destinos inválidos.
- Se validó generación de secret hex de 32 bytes, límite por empresa, ocultamiento de `secret` en listados y `DELETE` con control de pertenencia.

## File List

- `backend/internal/domain/webhook.go` — modelo de dominio para webhooks y cola
- `backend/internal/storage/webhook_store.go` — store CRUD y operaciones de cola
- `backend/internal/storage/migrations/017_create_webhooks_outbound.up.sql` — tabla `webhooks_outbound`
- `backend/internal/storage/migrations/017_create_webhooks_outbound.down.sql` — rollback de `webhooks_outbound`
- `backend/internal/storage/migrations/018_create_webhooks_outbound_queue.up.sql` — tabla `webhooks_outbound_queue`
- `backend/internal/storage/migrations/018_create_webhooks_outbound_queue.down.sql` — rollback de `webhooks_outbound_queue`
- `backend/internal/http/container.go` — instancia `V1WebhooksHandler` y límite configurable
- `backend/internal/http/routes_api.go` — rutas REST autenticadas con API key
- `backend/internal/http/handlers/v1_webhooks.go` — handler REST de webhooks
- `backend/internal/http/handlers/v1_webhooks_test.go` — tests `httptest` de creación, validaciones, listado y borrado
- `_bmad-output/implementation-artifacts/5-3-endpoints-rest-gestion-webhooks.md` — tracking de story actualizado
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — estado de sprint sincronizado

## Change Log

- 2026-05-22 — Story 5.3 implementada y validada: endpoints REST de webhooks, tests de handler y endurecimiento de validación HTTPS.
- **2026-05-23** — Reabierta. Ajustes derivados del cambio de ownership (ver 5-2 changelog):
  - `POST /api/service/v1/webhooks` ahora persiste `telefono_id` y `api_key_id` desde los claims (no solo `empresa_id`).
  - `GET /api/service/v1/webhooks` filtra por `api_key_id` (cada integrador ve solo sus webhooks, no los de otros teléfonos de la misma empresa).
  - `DELETE /api/service/v1/webhooks/{id}` valida ownership por `api_key_id`, no por `empresa_id`.
  - El límite `WEBHOOKS_MAX_PER_EMPRESA` ahora se aplica **por api_key** (rename semántico pendiente — la env var mantiene el nombre para compatibilidad; documentar en 5-6).
  - Nuevo endpoint admin read-only: `GET /api/service/v1/empresas/webhooks` montado en `empresaStack`, lista todos los webhooks de la empresa del JWT para soporte.
