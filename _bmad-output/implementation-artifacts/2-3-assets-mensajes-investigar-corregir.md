---
title: 'Story 2.3 — Investigar y corregir assets en mensajes'
type: 'bugfix'
created: '2026-05-06'
status: 'review'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: 'cee26ae'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como administrador del panel,
quiero que la sección de mensajes muestre correctamente los adjuntos (assets) de cada mensaje y que el reintento desde el panel funcione realmente,
para que soporte pueda diagnosticar y operar mensajes con adjuntos sin errores silenciosos.

## Acceptance Criteria

**AC1:** `GET /api/admin/mensajes` incluye el campo `adjuntos` en cada mensaje que lo tiene, con estructura `[{nombre, sha256_hash, tamano_bytes}]`. Mensajes sin adjuntos retornan `adjuntos` vacío o ausente.

**AC2:** El panel admin (página `/messages`) muestra los adjuntos correctamente en la tabla y en el Sheet de detalle (el campo `adjuntos` llega del backend — ya hay código frontend para mostrarlo, el bug es solo backend).

**AC3:** `POST /api/admin/mensajes/{reference_id}` usa el WhatsApp Manager compartido (el mismo que tiene las sesiones activas). Un mensaje sin adjuntos, con sesión activa, se reenvía exitosamente desde el panel admin.

**AC4:** Los handlers de mensajes admin son métodos de un struct con dependencias inyectadas (`msgRepo`, `empresaStore`, `telefonoStore`, `manager`) — no recrean conexión DB ni Manager por cada request.

**AC5:** `cd backend && go build ./...` pasa sin errores.

**AC6:** `cd backend && go test ./...` pasa sin errores (o regresiones).

**AC7:** `cd frontend && npm run lint` pasa sin errores.

## Tasks / Subtasks

- [x] **Tarea 1: Crear `AdminMessagesHandler` struct con dependencias inyectadas** (AC: 4)
  - [x] Crear `backend/internal/http/handlers/admin_messages.go` con `package http`
  - [x] Definir struct `AdminMessagesHandler` con campos: `msgRepo storage.MessagesRepository`, `empresaStore domain.EmpresaStoreInterface`, `telefonoStore *storage.TelefonoStore`, `manager *whatsapp.Manager`
  - [x] Definir constructor `NewAdminMessagesHandler(msgRepo, empresaStore, telefonoStore, manager) *AdminMessagesHandler`

- [x] **Tarea 2: Migrar `HandleGetAdminMessages` al nuevo struct e incluir adjuntos** (AC: 1, 2, 4)
  - [x] Mover la lógica de `HandleGetAdminMessages` (router.go:219-288) al método `(h *AdminMessagesHandler) GetMessages(w, r)`
  - [x] Definir `adminMessageDTO` local que incluya `Adjuntos []domain.AttachmentInfo json:"adjuntos,omitempty"`
  - [x] Al construir cada `adminMessageDTO`, copiar `Adjuntos: m.Adjuntos` desde el `domain.Message`
  - [x] Usar `h.msgRepo` y `h.empresaStore` en lugar de recrear DB/stores

- [x] **Tarea 3: Migrar `HandleAdminRetryMessage` al nuevo struct con manager real** (AC: 3, 4)
  - [x] Mover la lógica de `HandleAdminRetryMessage` (router.go:290-375) al método `(h *AdminMessagesHandler) RetryMessage(w, r)`
  - [x] Reemplazar `manager := whatsapp.NewManager()` con `h.manager`
  - [x] Usar `h.msgRepo` y `h.telefonoStore` en lugar de recrear DB/stores
  - [x] Preservar el check de autorización con `domain.GetPanelAccess`

- [x] **Tarea 4: Registrar nuevo handler en Container y routes** (AC: 4)
  - [x] Agregar `AdminMessagesHandler *handlers.AdminMessagesHandler` al struct `Container` en `container.go`
  - [x] Instanciar en `NewContainer()`: `adminMessagesHandler := handlers.NewAdminMessagesHandler(msgRepo, empresaStore, telefonoStore, manager)`
  - [x] Actualizar `routes_admin.go` para usar `c.AdminMessagesHandler.GetMessages` y `c.AdminMessagesHandler.RetryMessage`
  - [x] Eliminar las funciones libres `HandleGetAdminMessages` y `HandleAdminRetryMessage` de `router.go` y el struct `AdminMessage`

- [x] **Tarea 5: Verificar build y tests** (AC: 5, 6, 7)
  - [x] `cd backend && go build ./...` — pasa sin errores
  - [x] `cd backend && go test ./...` — sin nuevas regresiones (fallos pre-existentes en mocks de EmpresaStoreInterface confirmados como anteriores a esta story)
  - [x] `cd frontend && npm run lint` — sin nuevos errores (3 errores pre-existentes en lib/api.ts confirmados como anteriores a esta story)

## Dev Notes

### Bugs raíz identificados (no especular, estos son hechos verificados)

**Bug 1 — `AdminMessage` struct sin campo `adjuntos`:**
```
backend/internal/http/router.go:207-217
```
El struct `AdminMessage` en `router.go` no tiene campo `Adjuntos`. El frontend ya espera `adjuntos?: AttachmentInfo[]` en su interface (definida en `frontend/lib/api.ts:351-362`) y el código de la página (`frontend/app/messages/page.tsx:303-314`) ya itera `selectedMessage.adjuntos`. El bug es que el backend nunca popula ese campo.

**Bug 2 — `HandleGetAdminMessages` no copia `Adjuntos` al construir la respuesta:**
```
backend/internal/http/router.go:243-259
```
El `domain.Message` scanneado desde DB (`scanMessages` en `storage/messages.go:444-515`) sí deserializa correctamente `adjuntos_json` → `m.Adjuntos []AttachmentInfo`. Pero al mapear a `AdminMessage`, el campo `Adjuntos` simplemente nunca se asigna.

**Bug 3 — Admin retry usa un Manager vacío nuevo, SIEMPRE falla:**
```
backend/internal/http/router.go:355
manager := whatsapp.NewManager()   ← BUG CRÍTICO
```
`whatsapp.NewManager()` crea una instancia vacía sin clientes. Los clientes WhatsApp reales están en `Container.Manager` (inicializado en `NewContainer()`, `container.go:55`). Este bug hace que TODOS los reintentos desde el panel admin fallen con `ErrClientNotConnected`, independientemente de que la sesión esté activa.

**Causa raíz estructural:** `HandleGetAdminMessages` y `HandleAdminRetryMessage` son funciones libres en `router.go` (no métodos de struct), lo que impide acceder a dependencias del Container. La solución correcta es convertirlos en métodos de un nuevo struct con inyección de dependencias, siguiendo el patrón de todos los otros handlers del proyecto.

### Patrón a seguir (obligatorio)

El proyecto tiene un patrón claro de handlers en `backend/internal/http/handlers/`:

```go
// Ejemplo: handlers/api_keys.go
package handlers

type ApiKeysHandler struct {
    apiKeyStore  *storage.ApiKeyStore
    telefonoStore *storage.TelefonoStore
    // ...
}

func NewApiKeysHandler(...) *ApiKeysHandler { ... }
func (h *ApiKeysHandler) Get(w http.ResponseWriter, r *http.Request) { ... }
```

El nuevo `AdminMessagesHandler` debe seguir **exactamente** este patrón.

### Definición del DTO de respuesta

El `AdminMessage` actual en `router.go` debe moverse al nuevo handler como un struct local o DTO. La estructura completa (incluyendo el campo faltante):

```go
type adminMessageDTO struct {
    ID          int                    `json:"id"`
    ReferenceID string                 `json:"reference_id,omitempty"`
    AccountID   string                 `json:"account_id"`
    To          string                 `json:"to"`
    Content     string                 `json:"content"`
    Status      string                 `json:"status"`
    ErrorReason *string                `json:"error_reason,omitempty"`
    RetryCount  *int                   `json:"retry_count,omitempty"`
    Adjuntos    []domain.AttachmentInfo `json:"adjuntos,omitempty"`  // ← campo faltante
    CreatedAt   time.Time              `json:"created_at"`
}
```

### Lógica de GetMessages (a preservar)

La lógica de `HandleGetAdminMessages` actual (router.go:219-288) que debe preservarse:
- Lee `limit` de query param (default 50)
- Lee `account_id` y `status` de query params
- Si `account_id` está: busca empresa por RUC, trae sus mensajes
- Si no: itera todas las empresas, trae mensajes de cada una, ordena por `CreatedAt DESC`, trunca a `limit`
- Usa `msgRepo.GetByEmpresa(empresaID, status, "", limit, 0)`
- Retorna `{ ok, messages, total }`

Al usar el struct, reemplazar la creación ad-hoc de DB:
```go
// ANTES (router.go) — crear DB en cada request:
cfg := config.Load()
db, err := storage.NewDB(cfg)
msgRepo := storage.NewMessagesRepository(db)

// DESPUÉS (método del struct) — usar dependencias inyectadas:
h.msgRepo.GetByEmpresa(...)
h.empresaStore.GetAll(...)
```

### Lógica de RetryMessage (a preservar)

La lógica de `HandleAdminRetryMessage` actual (router.go:290-375) que debe preservarse:
- Valida `domain.GetPanelAccess` para autorización
- Extrae `referenceID` del path (usar el mismo `extractReferenceID` de `router.go:377-385`)
- Busca mensaje por referenceID
- Valida `access.CanAccessEmpresa(msg.EmpresaID)`
- Rechaza si mensaje ya fue enviado/entregado
- Rechaza con `MEDIA_RETRY_UNSUPPORTED` si tiene adjuntos (`len(msg.Adjuntos) > 0`)
- Obtiene teléfono, verifica que esté activo
- Llama `IncrementRetryCount` + `SendRichMessage` + `UpdateEstado`

Cambio clave: reemplazar `manager := whatsapp.NewManager()` con `h.manager`.

### Función extractReferenceID

`extractReferenceID` ya existe en `router.go:377-385` (package http):
```go
func extractReferenceID(path string) string {
    parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
    for i, p := range parts {
        if p == "mensajes" && i+1 < len(parts) {
            return parts[i+1]
        }
    }
    return ""
}
```

El nuevo handler está en `package handlers`, así que **NO puede llamar a esta función directamente**. Debe duplicarla localmente en `admin_messages.go` o usar `r.PathValue("id")` (disponible en Go 1.22+ con el router estándar que ya usa el proyecto — la ruta es `POST /api/admin/mensajes/{id}`).

**Usar `r.PathValue("id")`** es más limpio y evita duplicación:
```go
refID := r.PathValue("id")
if refID == "" {
    // error
}
```

Para GetMessages no hay path variable, no aplica.

### Archivos a modificar

| Archivo | Cambio |
|---------|--------|
| `backend/internal/http/handlers/admin_messages.go` | **NUEVO** — struct + constructor + métodos |
| `backend/internal/http/container.go` | Agregar `AdminMessagesHandler` al struct y a `NewContainer()` |
| `backend/internal/http/routes_admin.go` | Líneas 77-78: usar `c.AdminMessagesHandler` |
| `backend/internal/http/router.go` | Eliminar `AdminMessage` struct, `HandleGetAdminMessages`, `HandleAdminRetryMessage` (mantener resto) |

### Frontend: no requiere cambios

El frontend ya tiene todo lo necesario:
- `frontend/lib/api.ts:351-362`: `AdminMessage` interface ya tiene `adjuntos?: AttachmentInfo[]`
- `frontend/app/messages/page.tsx:303-314`: ya renderiza los adjuntos en el Sheet
- `frontend/app/messages/page.tsx:203`: ya muestra badge de adjuntos en la tabla

El único cambio es que ahora el backend envíe el campo. El frontend lo mostrará automáticamente.

### Imports necesarios en admin_messages.go

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "sort"
    "strconv"
    "time"

    "wsapi/internal/domain"
    "wsapi/internal/storage"
    "wsapi/internal/whatsapp"
)
```

### Verificación de comportamiento de empresaStore cuando DB no disponible

En `container.go:65-76`, `msgRepo` y `empresaStore` pueden ser `nil` si DB no está disponible. El handler debe manejarlo:
```go
func (h *AdminMessagesHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
    messages := []adminMessageDTO{}
    if h.msgRepo != nil && h.empresaStore != nil {
        // ... lógica de consulta
    }
    // retornar lista vacía si no hay DB
}
```

### Tests existentes relevantes

```bash
cd backend && go test ./internal/http/...       # tests de handlers
cd backend && go test ./internal/storage/...    # tests de mensajes
cd backend && go test ./internal/domain/...     # tests de dominio
```

Verificar que no hay regresiones. No se requieren tests nuevos para esta story (los bugs son de plumbing/wiring, no de lógica de negocio nueva), pero si se escribe algún test de integración para el handler, seguir el patrón de `handlers/*_test.go` existente.

### Project Structure Notes

- Package de handlers: `package handlers` (import path: `wsapi/internal/http/handlers`)
- El alias de import en container.go ya es: `handlers "wsapi/internal/http/handlers"`
- No crear archivos en `backend/internal/http/` directamente para handlers — ir a `handlers/` subdirectorio
- El struct `AdminMessage` en `router.go` es el que tiene el bug; al moverlo se convierte en `adminMessageDTO` (privado) dentro del nuevo handler
- No introducir nuevas dependencias externas

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Creado `backend/internal/http/handlers/admin_messages.go` (package http, como todos los archivos en ese directorio)
- `adminMessageDTO` incluye `Adjuntos []domain.AttachmentInfo` — fix del AC1/AC2
- `RetryMessage` usa `h.manager` (shared manager con sesiones activas) — fix del AC3
- Eliminados `AdminMessage` struct, `HandleGetAdminMessages`, `HandleAdminRetryMessage`, `extractReferenceID` de `router.go`
- Eliminado import `"sort"` de `router.go` (ya no usado tras la eliminación)
- Eliminado import `"encoding/json"` de `admin_messages.go` (no usado — writeHandlerJSON/writeAPIError están en response_helpers.go)
- Fallos de tests y lint pre-existentes confirmados con `git stash` antes/después: no se introdujeron regresiones

### File List

- `backend/internal/http/handlers/admin_messages.go` — NUEVO
- `backend/internal/http/container.go` — MODIFICADO (agregar AdminMessagesHandler)
- `backend/internal/http/routes_admin.go` — MODIFICADO (registrar rutas via AdminMessagesHandler)
- `backend/internal/http/router.go` — MODIFICADO (eliminar handlers libres y struct AdminMessage)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — MODIFICADO (status in-progress)
