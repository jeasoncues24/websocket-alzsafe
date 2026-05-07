---
title: 'Story 2.5 — Validar reconexión al reiniciar binario'
type: 'bugfix+feature'
created: '2026-05-07'
status: 'done'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: 'cee26ae'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como administrador del panel,
quiero que al reiniciar el binario las sesiones activas reconecten automáticamente y el panel refleje su estado real durante ese proceso,
para que soporte pueda distinguir entre una sesión **verdaderamente caída** y una **reconectándose tras reinicio**, sin intervenir innecesariamente.

## Acceptance Criteria

**AC1:** Durante el bootstrap de reconexión (i.e. el `StartupBootstrapper` está procesando un teléfono con `status=active` en DB pero no conectado en runtime), la respuesta de `GET /api/admin/sesiones` NO muestra `mismatch: true` para esas sesiones — en cambio, `mismatch` es `false` y se incluye un campo `reconnecting: true`.

**AC2:** El campo `reconnecting` en la respuesta de `GET /api/admin/sesiones` es `true` cuando `t.Status == active` en DB **y** el SessionStore tiene estado `"initializing"` o `"qr_pending"` para ese teléfono. En cualquier otro caso es `false` u omitido.

**AC3:** El `StartupBootstrapper` llama `sessionStore.AppendEvent(accountID, "initializing", "bootstrap_restart")` para cada candidato antes de lanzar su goroutine de reconexión (después del `sessionStore.SetInitializing` existente).

**AC4:** Cuando `startSessionWithRetry` falla definitivamente, el bootstrapper llama `sessionStore.AppendEvent(accountID, "disconnected", "startup_start_failed")` (además del `sessionStore.SetDisconnected` ya existente).

**AC5:** La página `/sessions` del frontend muestra un badge "Reconectando" (color azul/gris) en lugar del badge "Inconsistente" (amarillo) cuando `session.reconnecting === true`. El badge "Inconsistente" solo aparece cuando `session.mismatch === true` **y** `session.reconnecting` es falso o ausente.

**AC6:** El tipo `SessionInfo` en `frontend/lib/api.ts` incluye el campo opcional `reconnecting?: boolean`.

**AC7:** `cd backend && go build ./...` pasa sin errores.

**AC8:** `cd backend && go test ./...` pasa sin nuevas regresiones. Los tests existentes de `startup_bootstrap_test.go` siguen en verde.

**AC9:** `cd frontend && npm run lint` pasa sin nuevos errores.

## Tasks / Subtasks

- [x] **Tarea 1: Actualizar `sessionInfoDTO` y lógica de mismatch en `GetSessions`** (AC: 1, 2)
  - [x] Agregar campo `Reconnecting bool` a `sessionInfoDTO` en `backend/internal/http/handlers/admin_sessions.go`
  - [x] Agregar campo `"reconnecting"` al JSON: `json:"reconnecting,omitempty"` (solo serializar si `true`)
  - [x] En `GetSessions`, después de calcular `runtimeConnected`, consultar `h.sessionStore.Get(t.NumeroCompleto)` para obtener `storeState`
  - [x] Calcular `reconnecting = (t.Status == domain.TelefonoStatusActive) && !runtimeConnected && (storeState.Status == "initializing" || storeState.Status == "qr_pending")`
  - [x] Actualizar cálculo de mismatch: `mismatch = (t.Status == domain.TelefonoStatusActive) != runtimeConnected && !reconnecting`
  - [x] Pasar `Reconnecting: reconnecting` en la construcción de `sessionInfoDTO`

- [x] **Tarea 2: Actualizar `computeSessionSummary`** (AC: 1)
  - [x] En `computeSessionSummary`, no contar sesiones con `Reconnecting == true` en el contador `Mismatch`
  - [x] Verificar que el `summary.mismatch` solo refleja mismatches reales (no sesiones en bootstrap)

- [x] **Tarea 3: Integrar `AppendEvent` en `StartupBootstrapper`** (AC: 3, 4)
  - [x] En `backend/internal/whatsapp/startup_bootstrap.go`, en el bloque donde se agrega un candidato (después de `b.sessionStore.SetInitializing(accountID)`), agregar:
    ```go
    if b.sessionStore != nil {
        b.sessionStore.AppendEvent(accountID, "initializing", "bootstrap_restart")
    }
    ```
  - [x] En la goroutine de `startSessionWithRetry`, en el bloque de error (después de `b.sessionStore.SetDisconnected(c.accountID, "startup_start_failed")`), agregar:
    ```go
    b.sessionStore.AppendEvent(c.accountID, "disconnected", "startup_start_failed")
    ```
  - [x] Verificar que ambas llamadas tienen nil-guard (`if b.sessionStore != nil`) — el existing code ya lo tiene en el bloque; agregar el mismo patrón

- [x] **Tarea 4: Actualizar tipos en `frontend/lib/api.ts`** (AC: 6)
  - [x] Agregar `reconnecting?: boolean` a la interfaz `SessionInfo`

- [x] **Tarea 5: Actualizar lógica de badge en `frontend/app/sessions/page.tsx`** (AC: 5)
  - [x] En `getStatusBadge`, agregar caso específico para `reconnecting`:
    - Si `session.reconnecting === true` → badge azul/gris "Reconectando" (sobrescribe el badge de status)
  - [x] En el bloque de badges del card, mostrar el badge "Inconsistente" SOLO si `session.mismatch === true && !session.reconnecting`
  - [x] El botón "Reconectar" NO aparece cuando `session.reconnecting === true` (la sesión ya se está reconectando sola)

- [x] **Tarea 6: Verificar build, tests y lint** (AC: 7, 8, 9)
  - [x] `cd backend && go build ./...`
  - [x] `cd backend && go test ./...` — verificar especialmente `startup_bootstrap_test.go` y `internal/http`
  - [x] `cd frontend && npm run lint`

### Review Findings

- [x] [Review][Patch] Task checkboxes unchecked despite complete implementation — corregido
- [x] [Review][Defer] UX: botón "Reconectar" oculto sin feedback visual (spinner/disabled) — deferred, mejora futura en story de UI
- [x] [Review][Defer] Magic strings `"initializing"/"qr_pending"` sin constantes — deferred, patrón pre-existente en todo el store

## Dev Notes

### 🔍 Contexto crítico: el StartupBootstrapper ya existe

**NO crear un bootstrapper nuevo.** Ya existe en `backend/internal/whatsapp/startup_bootstrap.go` con tests completos en `startup_bootstrap_test.go`. Esta story solo lo _extiende_.

El bootstrap se activa en `buildStartupBootstrap` (`router.go:55`) controlado por env var:
```
WHATSAPP_BOOTSTRAP_ENABLED=true   (default: true)
WHATSAPP_BOOTSTRAP_TIMEOUT_SEC=60 (default: 60)
WHATSAPP_BOOTSTRAP_MAX_CONCURRENCY=4 (default: 4)
```

Se invoca desde `main.go:181` como goroutine paralela al servidor HTTP. Esto significa que el panel puede recibir requests mientras el bootstrap aún está corriendo.

### 🐛 Bug raíz que esta story corrige

**Descripción del problema:**

Cuando el binario reinicia:
1. `manager` (in-memory) → vacío
2. `SessionStore` (in-memory) → vacío
3. DB → todos los teléfonos previos siguen con `status = "active"`

El `StartupBootstrapper` detecta el mismatch y llama `SetInitializing(accountID)` + lanza `StartSession` para cada teléfono activo.

Pero mientras eso ocurre (puede tomar varios segundos por sesión), si el admin consulta `GET /api/admin/sesiones`, el handler actual calcula:
```go
mismatch = (t.Status == active) != runtimeConnected
         = true != false
         = true  ← FALSO POSITIVO
```

El admin ve el panel lleno de badges "Inconsistente" y puede pulsar "Reconectar" sobre sesiones que el sistema ya está reconectando. Una doble reconexión puede interferir con el proceso.

**La solución:** consultar `sessionStore.Get()` para saber si la sesión está en `"initializing"` (bootstrap en progreso). En ese caso, `mismatch = false` y `reconnecting = true`.

### 📐 Implementación exacta de la lógica de mismatch corregida

```go
// En AdminSessionsHandler.GetSessions — reemplazar bloque mismatch actual:

runtimeConnected := false
if h.manager != nil {
    if client, ok := h.manager.Get(accountID); ok && client != nil && client.IsConnected() {
        runtimeConnected = true
    }
}

// storeState puede no existir si el bootstrap aún no procesó esta sesión
var storeStatus string
if h.sessionStore != nil {
    if state, ok := h.sessionStore.Get(t.NumeroCompleto); ok {
        storeStatus = state.Status
    }
}

reconnecting := (t.Status == domain.TelefonoStatusActive) &&
    !runtimeConnected &&
    (storeStatus == "initializing" || storeStatus == "qr_pending")

mismatch := (t.Status == domain.TelefonoStatusActive) != runtimeConnected && !reconnecting
```

> ⚠️ NOTA: La consulta al `sessionStore` para los eventos (`h.sessionStore.Get(t.NumeroCompleto)`) ya existe en el código actual de `GetSessions`. **No duplicar la llamada** — reutilizar el `state` que ya se obtiene para los eventos.

### 📐 Refactor de la doble consulta sessionStore

El código actual de `GetSessions` ya llama `h.sessionStore.Get(t.NumeroCompleto)` para extraer eventos. Con esta story, también necesitamos el `storeStatus`. Consolidar en una sola llamada:

```go
var events []sessionEventDTO
var storeStatus string
if h.sessionStore != nil {
    if state, ok := h.sessionStore.Get(t.NumeroCompleto); ok {
        storeStatus = state.Status
        last := state.Events
        if len(last) > 10 {
            last = last[len(last)-10:]
        }
        for _, e := range last {
            events = append(events, sessionEventDTO{
                Timestamp: e.Timestamp,
                Type:      e.Type,
                Details:   e.Details,
            })
        }
    }
}
```

### 📐 Cambios exactos en `startup_bootstrap.go`

Ubicación del primer cambio — dentro del loop sobre candidatos, después de:
```go
if b.sessionStore != nil {
    b.sessionStore.SetInitializing(accountID)
}
```
Agregar inmediatamente después:
```go
if b.sessionStore != nil {
    b.sessionStore.AppendEvent(accountID, "initializing", "bootstrap_restart")
}
```

Ubicación del segundo cambio — en la goroutine, dentro del bloque `if err != nil` después de `b.sessionStore.SetDisconnected(c.accountID, "startup_start_failed")`:
```go
b.sessionStore.AppendEvent(c.accountID, "disconnected", "startup_start_failed")
```
(El nil-guard `if b.sessionStore != nil` ya existe en esa línea — el AppendEvent va dentro del mismo if.)

### 🎨 UX: Badge "Reconectando" en frontend

La prioridad de badges en el card de sesión:

```
1. Si session.reconnecting === true  → badge azul "Reconectando" (no mostrar "Inconsistente")
2. Si session.mismatch === true      → badge amarillo "Inconsistente"
3. (ambas pueden ser false simultáneamente)
```

Badge sugerido:
```tsx
{session.reconnecting && (
  <Badge variant="outline" className="border-blue-400 text-blue-500">
    Reconectando
  </Badge>
)}
{session.mismatch && !session.reconnecting && (
  <Badge variant="outline" className="border-yellow-500 text-yellow-600">
    Inconsistente
  </Badge>
)}
```

El botón "Reconectar" — condición actual:
```tsx
{session.status !== "active" && session.telefono_id != null && (
  <Button ...>Reconectar</Button>
)}
```
Actualizar a:
```tsx
{session.status !== "active" && session.telefono_id != null && !session.reconnecting && (
  <Button ...>Reconectar</Button>
)}
```

### 🧪 Tests existentes que NO deben romperse

En `backend/internal/whatsapp/startup_bootstrap_test.go`:
- `TestStartupBootstrapKeepsHealthyRuntimeStable`
- `TestStartupBootstrapReconcilesDisconnectedDbWithRuntimeConnected`
- `TestStartupBootstrapStartsMissingRuntimeSession`

Estos tests no mockean `SessionStore.AppendEvent` (el método no existía cuando se escribieron). Con esta story, `AppendEvent` será llamado en `Run()`. Los tests existentes pasan un `SessionStore` real, por lo que las llamadas a `AppendEvent` son válidas — no hay problema de interface.

Verificar que después de la story, `TestStartupBootstrapStartsMissingRuntimeSession` sigue pasando. Ese test verifica:
- `state.Status == "initializing"` → sigue correcto, `AppendEvent` no cambia `Status`

### ⚠️ Lo que NO cambia esta story

- El comportamiento del bootstrapper (cuándo y cómo reconecta) permanece igual
- La lógica de `startSessionWithRetry` (reintentos, delays) permanece igual
- Los campos existentes de `sessionInfoDTO` permanecen igual
- El `sessionSummaryDTO` campos permanecen igual (solo `Mismatch` se recalcula correctamente)
- No se agrega un nuevo endpoint

### Archivos a modificar

| Archivo | Cambio | Tipo |
|---------|--------|------|
| `backend/internal/http/handlers/admin_sessions.go` | Agregar `Reconnecting` a DTO, refactorizar consulta sessionStore, corregir mismatch | MODIFICAR |
| `backend/internal/whatsapp/startup_bootstrap.go` | Llamar `AppendEvent` en dos puntos | MODIFICAR |
| `frontend/lib/api.ts` | Agregar `reconnecting?: boolean` a `SessionInfo` | MODIFICAR |
| `frontend/app/sessions/page.tsx` | Badge "Reconectando", ocultar botón Reconectar durante bootstrap | MODIFICAR |

### Learnings de stories anteriores (2-3, 2-4)

- Package en `handlers/` es `package http` (NO `package handlers`)
- Usar `writeHandlerJSON` y `writeAPIError` — no inventar helpers propios
- Nil-guard para todos los stores en handlers
- El `set` privado de `SessionStore` preserva `Events` (modificado en 2-4) — `AppendEvent` no borra el estado
- `frontend/lib/api.ts` usa `fetchWithAuth` para rutas admin

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

### File List

### Change Log
