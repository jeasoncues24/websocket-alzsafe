---
story_id: "6.3"
epic: "epic-6"
title: "Backend — Endpoint admin para listar webhooks de un teléfono"
status: review
estimated_days: 1
priority: high
branch: "feature/panel-telefonos-webhooks"
skills: ["golang-security", "bmad-code-review"]
affects:
  - backend/internal/http/admin.go
  - backend/internal/http/routes_admin.go
  - backend/internal/http/container.go
---

# Story 6.3: Backend — Endpoint admin para listar webhooks de un teléfono

## Story

Como admin del panel wsapi,
quiero poder consultar los webhooks registrados para un teléfono específico,
para diagnosticar problemas de entrega y apoyar a los integradores B2B.

## Contexto técnico

Ya existe `WebhookStore.ListByTelefono(telefonoID int64)` en `backend/internal/storage/webhook_store.go`. El campo `Secret` del struct `domain.Webhook` tiene `json:"-"` — nunca se serializa automáticamente en JSON. Solo hay que exponer el store a través de un nuevo handler admin.

Ruta propuesta: `GET /api/admin/telefonos/{id}/webhooks` — sigue el patrón de `GET /api/admin/telefonos/{id}/api-keys` que ya existe.

## Acceptance Criteria

**AC1 — Endpoint lista webhooks del teléfono:**
**Dado** un token de admin válido,
**Cuando** se hace `GET /api/admin/telefonos/{id}/webhooks`,
**Entonces** responde `200` con `{"ok": true, "webhooks": [...], "total": N}`
**Y** cada webhook incluye: `id`, `empresa_id`, `telefono_id`, `api_key_id`, `url`, `eventos`, `activo`, `failure_count`, `last_error`, `last_success_at`, `created_at`, `updated_at`
**Y** el campo `secret` NO aparece en la respuesta.

**AC2 — Teléfono no encontrado:**
**Dado** un ID de teléfono que no existe,
**Cuando** se hace la solicitud,
**Entonces** responde `404`.

**AC3 — Sin webhooks:**
**Dado** un teléfono válido sin webhooks registrados,
**Cuando** se hace la solicitud,
**Entonces** responde `200` con `{"ok": true, "webhooks": [], "total": 0}`.

**AC4 — Solo accesible con token admin:**
**Dado** una solicitud sin token o con token inválido,
**Cuando** se hace la solicitud,
**Entonces** responde `401`.

**AC5 — Verificación de build:**
**Dado** los cambios aplicados,
**Cuando** se ejecutan `go build ./...` y `go test ./...`,
**Entonces** ambos terminan sin errores.

## Tasks / Subtasks

- [x] **T1 — Verificar inyección de `WebhookStore` en `AdminHandler`** (AC1)
  - [x] Story 6.1 ya inyectó `webhookStore` — reutilizado

- [x] **T2 — Implementar handler `ListTelefonoWebhooks`** (AC1, AC2, AC3)
  - [x] Extraer `telefonoID` del path con `r.PathValue("id")`
  - [x] Validar token admin con `getPanelAdminAccess`
  - [x] Verificar que el teléfono existe (usar `telefonoStore.GetByID`)
  - [x] Llamar `webhookStore.ListByTelefono(telefonoID)`
  - [x] Responder con `writeAdminJSON`

- [x] **T3 — Registrar ruta en `routes_admin.go`** (AC1, AC4)
  - [x] Añadido: `mux.Handle("GET /api/admin/telefonos/{id}/webhooks", adminStack(...))`

- [x] **T4 — Verificación final** (AC5)
  - [x] `go build ./...` sin errores
  - [x] `go test ./...` sin regresiones

## Dev Notes

### Patrón de handler a seguir

Ver `ApiKeysHandler.ListByTelefono` como referencia directa — misma lógica de extracción de ID y validación de acceso.

### Estructura de respuesta

```go
type adminWebhooksListResponse struct {
    OK       bool             `json:"ok"`
    Webhooks []domain.Webhook `json:"webhooks"`
    Total    int              `json:"total"`
}
```

`domain.Webhook.Secret` tiene `json:"-"` → no se serializa. No hay que hacer nada extra.

### Protección de acceso

Usar `getPanelAdminAccess` igual que otros handlers. Si el admin tiene restricción por empresa, verificar que el teléfono pertenece a una empresa a la que tiene acceso (consultar `telefonoStore.GetByID` y comparar `phone.EmpresaID` con `access.CanAccessEmpresa`).

### Ejemplo de respuesta

```json
{
  "ok": true,
  "webhooks": [
    {
      "id": 1,
      "empresa_id": 3,
      "telefono_id": 7,
      "api_key_id": 42,
      "url": "https://app-loyo.com/hooks/whatsapp",
      "eventos": ["message.received", "session.connected"],
      "activo": true,
      "failure_count": 0,
      "last_error": null,
      "last_success_at": "2026-05-23T14:00:00Z",
      "created_at": "2026-05-20T10:00:00Z",
      "updated_at": "2026-05-23T14:00:00Z"
    }
  ],
  "total": 1
}
```

### References

- WebhookStore: `backend/internal/storage/webhook_store.go` (`ListByTelefono`)
- Patrón de handler: `backend/internal/http/admin.go` (buscar `ListByTelefono` de api-keys)
- Rutas admin: `backend/internal/http/routes_admin.go:62`
- Domain Webhook: `backend/internal/domain/webhook.go`

## Dev Agent Record

### Agent Model Used
claude-sonnet-4-6

### Debug Log References
- `webhookStore` ya inyectado por story 6.1 — T1 trivial
- Usado `r.PathValue("id")` consistente con `handlers/v1_webhooks.go` y `admin_sessions.go`
- `domain.Webhook.Secret` tiene `json:"-"` — confirmado, no se expone

### Completion Notes List
- T1: webhookStore reutilizado de 6.1
- T2: `ListTelefonoWebhooks` implementado con auth check, 404 si no existe teléfono, nil-guard en webhookStore
- T3: ruta registrada en routes_admin.go junto al bloque de AdminHandler
- T4: build y tests OK

### File List
- backend/internal/http/admin.go
- backend/internal/http/routes_admin.go
