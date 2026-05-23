---
story_id: "6.1"
epic: "epic-6"
title: "Backend — Conteos de API keys y webhooks en respuesta admin de teléfonos"
status: review
estimated_days: 1
priority: high
branch: "feature/panel-telefonos-webhooks"
skills: ["bmad-code-review"]
affects:
  - backend/internal/domain/telefono.go
  - backend/internal/http/admin.go
  - backend/internal/http/container.go
---

# Story 6.1: Backend — Conteos de API keys y webhooks en respuesta admin de teléfonos

## Story

Como admin del panel wsapi,
quiero que la lista de teléfonos de una empresa incluya cuántas API keys activas y cuántos webhooks tiene cada número,
para poder ver el estado operativo completo sin hacer clics adicionales.

## Contexto técnico

El endpoint `GET /api/admin/empresas/{id}/telefonos` actualmente enriquece cada teléfono con `RuntimeConnected` y `Mismatch`, pero no incluye conteos de API keys ni webhooks. El handler es `AdminHandler.ListCompanyPhones` en `backend/internal/http/admin.go:895`.

Ya existen los stores necesarios:
- `ApiKeyStore.GetByTelefonoID(telefonoID int64)` → `backend/internal/storage/api_key.go:297`
- `WebhookStore.ListByTelefono(telefonoID int64)` → `backend/internal/storage/webhook_store.go`

El `AdminHandler` ya tiene acceso a `apiKeyStore` (verificar si está inyectado). El `WebhookStore` puede necesitar inyección en el handler.

## Acceptance Criteria

**AC1 — Campos nuevos en la respuesta:**
**Dado** `GET /api/admin/empresas/{id}/telefonos`,
**Cuando** responde exitosamente,
**Entonces** cada teléfono incluye `api_key_count` (int, API keys con `activo = true`) y `webhook_count` (int, webhooks con `activo = true`).

**AC2 — Conteos correctos:**
**Dado** un teléfono con 3 API keys (2 activas, 1 revocada) y 1 webhook activo,
**Cuando** se consulta el endpoint,
**Entonces** `api_key_count = 2` y `webhook_count = 1`.

**AC3 — Degradación segura sin DB:**
**Dado** que el `WebhookStore` es nil (entorno sin DB),
**Cuando** se llama al endpoint,
**Entonces** `api_key_count = 0` y `webhook_count = 0` sin que el handler retorne error.

**AC4 — Verificación de build:**
**Dado** los cambios aplicados,
**Cuando** se ejecutan `go build ./...` y `go test ./...`,
**Entonces** ambos terminan sin errores ni regresiones.

## Tasks / Subtasks

- [x] **T1 — Extender `domain.Telefono`** (AC1)
  - [x] Añadir `ApiKeyCount int \`json:"api_key_count,omitempty"\`` al struct `Telefono`
  - [x] Añadir `WebhookCount int \`json:"webhook_count,omitempty"\`` al struct `Telefono`

- [x] **T2 — Inyectar `WebhookStore` en `AdminHandler`** (AC1, AC3)
  - [x] Verificar si `AdminHandler` ya tiene campo para `webhookStore`
  - [x] Si no existe: añadir campo `webhookStore *storage.WebhookStore` al struct
  - [x] Actualizar constructor/inicialización en `container.go` para pasar el store

- [x] **T3 — Poblar conteos en `ListCompanyPhones`** (AC1, AC2, AC3)
  - [x] En el loop de enriquecimiento (`enriched`): llamar `apiKeyStore.GetByTelefonoID` y contar activos
  - [x] Llamar `webhookStore.ListByTelefono` y contar activos (solo si `webhookStore != nil`)
  - [x] Asignar `enriched[i].ApiKeyCount` y `enriched[i].WebhookCount`

- [x] **T4 — Verificación final** (AC4)
  - [x] `go build ./...` sin errores
  - [x] `go test ./...` sin regresiones

## Dev Notes

### Patrón de enriquecimiento existente

El loop en `ListCompanyPhones` ya hace esto por cada teléfono:
```go
enriched[i] = phone
runtimeConnected := false
if h.manager != nil {
    // ...
}
enriched[i].RuntimeConnected = runtimeConnected
```
Añadir los conteos en el mismo loop, después del bloque de `Mismatch`.

### Filtrar solo activos

```go
// API keys activas
keys, _ := h.apiKeyStore.GetByTelefonoID(phone.ID)
activeKeys := 0
for _, k := range keys {
    if k.Activo {
        activeKeys++
    }
}
enriched[i].ApiKeyCount = activeKeys

// Webhooks activos
if h.webhookStore != nil {
    hooks, _ := h.webhookStore.ListByTelefono(phone.ID)
    activeHooks := 0
    for _, wh := range hooks {
        if wh.Activo {
            activeHooks++
        }
    }
    enriched[i].WebhookCount = activeHooks
}
```

### Verificar AdminHandler struct

```bash
grep -n "webhookStore\|WebhookStore\|apiKeyStore\|ApiKeyStore" backend/internal/http/admin.go | head -10
```

### References

- Handler: `backend/internal/http/admin.go:895` (`ListCompanyPhones`)
- Domain: `backend/internal/domain/telefono.go`
- ApiKeyStore: `backend/internal/storage/api_key.go:297`
- WebhookStore: `backend/internal/storage/webhook_store.go`
- Container: `backend/internal/http/container.go`

## Dev Agent Record

### Agent Model Used
claude-sonnet-4-6

### Debug Log References
- `AdminHandler` ya tenía `apiKeyStore` inyectado; solo faltaba `webhookStore`
- Constructor `NewAdminHandler` inicializa directamente stores — no se usa container.go para esto

### Completion Notes List
- T1: Añadidos `ApiKeyCount` y `WebhookCount` a `domain.Telefono`
- T2: Añadido `webhookStore *storage.WebhookStore` al struct y constructor `NewAdminHandler`
- T3: Loop de `ListCompanyPhones` enriquece con conteos de API keys activas y webhooks activos, con guard nil en ambos stores
- T4: `go build ./...` y `go test ./...` pasan sin errores

### File List
- backend/internal/domain/telefono.go
- backend/internal/http/admin.go
