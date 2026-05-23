---
title: 'Story 2.7 — QR vía API token: WS primario + fallback REST'
type: 'bugfix+feature'
created: '2026-05-07'
status: 'done'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '3fe8a86'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como empresa cliente (B2B) que consume la API de wsapi,
quiero poder conectarme al canal WebSocket con mi JWT de empresa, suscribirme a un teléfono y recibir eventos de sesión en tiempo real (QR, conectado, desconectado),
para que pueda integrar el flujo de conexión WhatsApp en mi propio sistema sin polling constante, con fallback REST cuando WS no esté disponible.

## Acceptance Criteria

**AC1 — WS bridge de sesión:**
Cuando la empresa se conecta a `GET /v1/ws` con JWT válido y envía `{"type": "subscribe", "data": {"phone_id": N}}`, el servidor:
- Verifica que el teléfono pertenece a la empresa del JWT
- Llama `whatsapp.StartSession(h.manager, phone.NumeroCompleto)`
- Entra en modo bridge: reenvía eventos del canal de sesión al WS cliente

**AC2 — Formato de eventos V1:**
Los eventos bridgeados usan el formato existente de V1 (`{"type": "...", "data": {...}}`):
- QR: `{"type": "qr", "data": {"qr_string": "...", "expires_in": 60, "message": "..."}}`
- Activo: `{"type": "connected", "data": {"isActive": true, "message": "..."}}`
- Desconectado: `{"type": "disconnected", "data": {"isActive": false, "reason": "...", ...}}`
- El campo `type` se deriva del `Event` del `SessionEvent` con mapeo simple (ver Dev Notes)

**AC3 — Cleanup de goroutine en QR:**
Cuando el WS se cierra (cualquier causa) y la sesión estaba en `"initializing"` o `"qr_pending"` en el `sessionStore`, el handler llama `h.manager.Delete(accountID)` en su defer, igual que `ConnectCompanyPhoneWS` (admin, story 2-6).

**AC4 — Keepalive ping:**
El servidor envía `{"type": "ping"}` cada 25 segundos mientras el WS está activo. Si el write falla, el handler retorna.

**AC5 — Suscripción inválida rechazada:**
Si el teléfono no existe, no pertenece a la empresa, o el JWT es inválido al momento del subscribe, el servidor envía `{"type": "error", "data": {"message": "..."}}` y retorna (cerrando el WS).

**AC6 — expires_in corregido en REST:**
Los tres endpoints REST que devuelven `expires_in` lo corrigen de 300 a 60:
- `POST /api/telefonos/{id}/qr` → `expires_in: 60`
- `POST /api/sesiones` → `expires_in: 60`
- `POST /api/sesiones/{id}/connect` → `expires_in: 60`

**AC7 — V1WSHandler con dependencias inyectadas:**
`V1WSHandler` recibe `telefonoStore *storage.TelefonoStore`, `sessionStore *storage.SessionStore` adicionales al `manager` y `jwtCfg` que ya tiene. `NewV1WSHandler` y `container.go` se actualizan.

**AC8 — Log de apertura/cierre:**
El handler imprime `[INFO] V1 WS opened empresa=%d` al inicio y `[INFO] V1 WS closed empresa=%d reason=%v` al cerrar.

**AC9:** `cd backend && go build ./...` pasa sin errores.

**AC10:** `cd backend && go test ./...` pasa sin nuevas regresiones.

**AC11:** `cd frontend && npm run lint` pasa sin nuevos errores.

## Tasks / Subtasks

- [x] **Tarea 1: Actualizar `V1WSHandler` struct y constructor** (AC: 7)
  - [x] Agregar campos `telefonoStore *storage.TelefonoStore` y `sessionStore *storage.SessionStore` al struct
  - [x] Actualizar `NewV1WSHandler` para aceptar los nuevos parámetros
  - [x] En `container.go`, pasar `telefonoStore` y `sessionStore` a `NewV1WSHandler`

- [x] **Tarea 2: Reescribir `HandleWS` con bridge real** (AC: 1, 2, 3, 4, 5, 8)
  - [x] Fase 1 — auth: extraer JWT de query param o header, validar con `auth.ParseEmpresaJWT`
  - [x] Fase 2 — esperar subscribe: leer primer mensaje del WS, verificar `type == "subscribe"`, extraer `phone_id`
  - [x] Verificar que el teléfono existe y pertenece a la empresa (usar `telefonoStore.BelongsToEmpresa`)
  - [x] Agregar log apertura: `fmt.Printf("[INFO] V1 WS opened empresa=%d\n", claims.EmpresaID)`
  - [x] Registrar `defer` de cleanup (igual que admin handler): si sessionStore reporta `initializing`/`qr_pending` → `h.manager.Delete(accountID)`
  - [x] Agregar log de cierre en el defer
  - [x] Llamar `whatsapp.StartSession(h.manager, phone.NumeroCompleto)` para obtener canal de eventos
  - [x] Ticker de 25s para keepalive ping
  - [x] Loop `select` sobre `events chan`, `ticker.C`, `ctx.Done()` (ver Dev Notes para código exacto)
  - [x] Mapear `SessionEvent.Event` → tipo V1 al reenviar (ver Dev Notes)

- [x] **Tarea 3: Corregir `expires_in` en V1 REST** (AC: 6)
  - [x] `v1_phones.go` → `PostPhoneQr`: cambiar `"expires_in": 300` → `"expires_in": 60`
  - [x] `v1_sessions.go` → `PostSessions`: cambiar `"expires_in": 300` → `"expires_in": 60`
  - [x] `v1_sessions.go` → `StartPhoneConnection`: cambiar `"expires_in": 300` → `"expires_in": 60`

- [x] **Tarea 4: Verificar build, tests y lint** (AC: 9, 10, 11)
  - [x] `cd backend && go build ./...`
  - [x] `cd backend && go test ./...`
  - [x] `cd frontend && npm run lint`

### Review Findings

- [x] [Review][Decision] Mapeo de Eventos de Desconexión Roto (AC2) (Solucionado: mapV1EventType corregido para evaluar isActive y retornar disconnected)
- [x] [Review][Patch] Carrera en defer cleanup de V1 WS [backend/internal/http/handlers/v1_ws.go:~125] (Solucionado: guard IsConnected integrado en defer)
- [x] [Review][Patch] Silencio en fallos de escritura de WebSocket V1 [backend/internal/http/handlers/v1_ws.go:~149] (Solucionado: agregados logs de error explicitos para writes/pings)
- [x] [Review][Patch] Registrar ws_closed con razones de desconexión en el historial [backend/internal/http/handlers/v1_ws.go:~125] (Solucionado: AppendEvent en sessionStore agregado a los defers de admin y v1_ws para visibilidad del frontend)

## Dev Notes

### 🔍 Diagnóstico del estado actual — qué falla y por qué

**Bug 1 — V1WSHandler.HandleWS no bridgea eventos:**
```go
// v1_ws.go (actual) — subscribe no hace nada útil
case "subscribe":
    var req struct{ PhoneID int64 `json:"phone_id"` }
    if json.Unmarshal(payload.Data, &req) == nil {
        _ = writeWSEvent(c, "subscribed", map[string]int64{"phone_id": req.PhoneID})
    }
// Solo echoea "subscribed" — no llama StartSession, no bridgea eventos
```
El cliente empresa se conecta al WS, se suscribe, y no recibe nada. Los eventos de QR, conexión y desconexión se pierden.

**Bug 2 — StartPhoneConnection drena eventos silenciosamente:**
```go
// v1_sessions.go (actual)
go func() {
    for range events {
    }
}()
```
Esto es un goroutine de drenaje que descarta todos los eventos de sesión. El cliente REST que llama `StartPhoneConnection` no tiene forma de recibir el QR en tiempo real — tendría que hacer polling. El goroutine existe para que el canal no se bloquee (buffer de 8 se llenaría), pero la consecuencia es que el WS cliente tampoco recibe eventos aunque se conecte después (porque `StartSession` devuelve un **canal sintético cerrado** si la sesión ya existe).

**Bug 3 — expires_in: 300 incorrecto:**
Tres endpoints devuelven `"expires_in": 300`. El QR real de whatsmeow expira en ~60s. El cliente que use este valor tendrá un countdown visual 5x más largo que el QR real.

### 🔍 Cómo funciona `StartSession` (crítico para entender el bridge)

```go
func StartSession(manager *Manager, accountID string) (<-chan SessionEvent, error)
```

Comportamiento según estado:
- **Sesión activa** (`runtimes[accountID]` existe): devuelve canal sintético con un evento de estado actual y lo cierra inmediatamente. El WS recibirá ese estado y luego `!ok` en el select → retorna.
- **Sesión iniciando** (`starting[accountID]` true): devuelve canal sintético con estado initializing y lo cierra. Mismo comportamiento.
- **Sin sesión**: crea `sessionRuntime`, lanza goroutine `runSession`, devuelve `runtime.events` (buffer 8). El canal permanece abierto mientras la sesión corre; se cierra cuando `runSession` termina.

**Consecuencia para el bridge:**
- Cuando el WS empresa llama `StartSession` y ya existe una sesión activa → recibe un snapshot del estado (1 evento) y el canal cierra → el loop del WS termina. Esto es correcto: el cliente sabe que la sesión está activa.
- Cuando no hay sesión → StartSession la lanza, el canal permanece abierto, los eventos (QR, connected, disconnected) fluyen al WS.

**Consecuencia del goroutine de drenaje en StartPhoneConnection:**
Si `StartPhoneConnection` (REST) se llamó antes que el WS, la sesión YA está corriendo. Cuando el WS llama `StartSession` después, recibe el canal sintético (snapshot). Si la sesión estaba en QR, el snapshot incluirá el estado QR. El WS recibirá 1 evento con el QR actual y luego el canal cierra. **Esto está bien** — el cliente recibe el QR actual.

El goroutine de drenaje en `StartPhoneConnection` se puede mantener (no es un leak real — termina cuando el canal se cierra). No es necesario cambiarlo en esta story.

### 📐 Código exacto: V1WSHandler reescrito

```go
package http

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "wsapi/internal/auth"
    "wsapi/internal/config"
    "wsapi/internal/storage"
    "wsapi/internal/whatsapp"

    "github.com/coder/websocket"
)

type V1WSHandler struct {
    manager       *whatsapp.Manager
    jwtCfg        *config.JWTConfig
    telefonoStore *storage.TelefonoStore
    sessionStore  *storage.SessionStore
}

func NewV1WSHandler(
    manager *whatsapp.Manager,
    jwtCfg *config.JWTConfig,
    telefonoStore *storage.TelefonoStore,
    sessionStore *storage.SessionStore,
) *V1WSHandler {
    return &V1WSHandler{
        manager:       manager,
        jwtCfg:        jwtCfg,
        telefonoStore: telefonoStore,
        sessionStore:  sessionStore,
    }
}

func (h *V1WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
    // — Autenticar empresa JWT (query param o header) —
    token := r.URL.Query().Get("token")
    if token == "" {
        if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
            token = strings.TrimPrefix(auth, "Bearer ")
        }
    }
    if token == "" {
        writeV1Error(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "Token requerido")
        return
    }
    claims, err := auth.ParseEmpresaJWT(token, h.jwtCfg.Secret)
    if err != nil {
        writeV1Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token inválido o expirado")
        return
    }

    // — Upgrade a WebSocket —
    c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
    if err != nil {
        return
    }
    defer c.CloseNow()

    ctx := r.Context()

    // — Esperar primer mensaje: debe ser subscribe con phone_id —
    _, data, err := c.Read(ctx)
    if err != nil {
        return
    }
    var payload struct {
        Type string          `json:"type"`
        Data json.RawMessage `json:"data"`
    }
    if err := json.Unmarshal(data, &payload); err != nil || payload.Type != "subscribe" {
        _ = writeWSEvent(c, "error", map[string]string{"message": "primer mensaje debe ser subscribe"})
        return
    }
    var sub struct {
        PhoneID int64 `json:"phone_id"`
    }
    if err := json.Unmarshal(payload.Data, &sub); err != nil || sub.PhoneID <= 0 {
        _ = writeWSEvent(c, "error", map[string]string{"message": "phone_id requerido"})
        return
    }

    // — Validar que el teléfono pertenece a la empresa —
    belongs, _ := h.telefonoStore.BelongsToEmpresa(sub.PhoneID, claims.EmpresaID)
    if !belongs {
        _ = writeWSEvent(c, "error", map[string]string{"message": "forbidden"})
        return
    }
    phone, err := h.telefonoStore.GetByID(sub.PhoneID)
    if err != nil || phone == nil {
        _ = writeWSEvent(c, "error", map[string]string{"message": "teléfono no encontrado"})
        return
    }
    accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)

    fmt.Printf("[INFO] V1 WS opened empresa=%d phone=%d account=%s\n", claims.EmpresaID, phone.ID, accountID)

    // — Cleanup al cerrar el WS —
    // Si la sesión estaba en QR o initializing, cancelamos el runtime para evitar fuga de goroutine.
    // Sesiones activas no se interrumpen.
    defer func() {
        fmt.Printf("[INFO] V1 WS closed empresa=%d account=%s reason=%v\n", claims.EmpresaID, accountID, ctx.Err())
        if h.sessionStore != nil && h.manager != nil {
            if state, ok := h.sessionStore.Get(phone.NumeroCompleto); ok {
                if state.Status == "initializing" || state.Status == "qr_pending" {
                    h.manager.Delete(accountID)
                }
            }
        }
    }()

    // — Iniciar o unirse a sesión existente —
    // StartSession es idempotente: devuelve canal sintético si la sesión ya existe.
    events, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
    if err != nil {
        _ = writeWSEvent(c, "error", map[string]string{"message": "error al iniciar sesión: " + err.Error()})
        return
    }

    // — Keepalive: ping cada 25s para proxies con idle timeout —
    ticker := time.NewTicker(25 * time.Second)
    defer ticker.Stop()

    // — Loop principal: puente entre el manager de sesión y el cliente empresa —
    for {
        select {
        case event, ok := <-events:
            if !ok {
                // Canal cerrado — la sesión terminó
                return
            }
            if err := writeWSEvent(c, mapV1EventType(event.Event), event.Data); err != nil {
                return
            }
        case <-ticker.C:
            if err := writeWSEvent(c, "ping", nil); err != nil {
                return
            }
        case <-ctx.Done():
            return
        }
    }
}

// mapV1EventType convierte el Event string del SessionEvent al tipo V1.
// Los eventos del manager usan el formato "qr-{accountID}", "active-{accountID}", etc.
// Los eventos V1 usan "qr", "connected", "disconnected".
func mapV1EventType(event string) string {
    switch {
    case strings.HasPrefix(event, "qr-"):
        return "qr"
    case strings.HasPrefix(event, "active-"):
        return "connected"
    default:
        return event
    }
}
```

> **NOTA IMPORTANTE sobre `writeWSEvent`:** Esta función ya existe en `v1_ws.go` con firma:
> ```go
> func writeWSEvent(c *websocket.Conn, eventType string, data interface{}) error
> ```
> El `data` de `SessionEvent` es `map[string]any` (compatible con `interface{}`). ✅
>
> **NOTA sobre import shadowing:** El parámetro `auth` en el handler colisiona con el import `wsapi/internal/auth`. Usar `authHeader := r.Header.Get("Authorization")` para evitar el conflicto, o nombrar el import: `authpkg "wsapi/internal/auth"`.

### 📐 Actualizar container.go

```go
// En NewContainer(), cambiar:
v1WSHandler := handlers.NewV1WSHandler(manager, jwtCfg)
// Por:
v1WSHandler := handlers.NewV1WSHandler(manager, jwtCfg, telefonoStore, sessionStore)
```

### 📐 Corrección de expires_in

```go
// v1_phones.go — PostPhoneQr
writeV1Success(w, map[string]interface{}{
    "telefono_id": telefonoID,
    "qr_string":   qrString,
    "expires_in":  60,   // era 300
}, claims.EmpresaID)

// v1_sessions.go — PostSessions
writeV1Success(w, map[string]interface{}{
    "telefono_id":    id,
    "numeroCompleto": numeroCompleto,
    "status":         "qr_pending",
    "expires_in":     60,   // era 300
    "qr_string":      qrString,
}, claims.EmpresaID)

// v1_sessions.go — StartPhoneConnection
writeV1Success(w, map[string]interface{}{
    "telefono_id":    phone.ID,
    "numeroCompleto": phone.NumeroCompleto,
    "status":         "initializing",
    "qr_string":      phone.QRString,
    "expires_in":     60,   // era 300
}, claims.EmpresaID)
```

### 📐 Estructura de archivos a modificar

| Archivo | Cambio | Tipo |
|---------|--------|------|
| `backend/internal/http/handlers/v1_ws.go` | Reescribir `HandleWS` con bridge real, agregar deps al struct | MODIFICAR |
| `backend/internal/http/handlers/v1_phones.go` | `expires_in: 60` en `PostPhoneQr` | MODIFICAR |
| `backend/internal/http/handlers/v1_sessions.go` | `expires_in: 60` en `PostSessions` y `StartPhoneConnection` | MODIFICAR |
| `backend/internal/http/container.go` | Agregar `telefonoStore`, `sessionStore` a `NewV1WSHandler` | MODIFICAR |

### ⚠️ Lo que NO cambia esta story

- Rutas API (`routes_api.go`) — el endpoint `GET /v1/ws` permanece igual, solo cambia el handler
- `V1SessionsHandler.StartPhoneConnection` — el goroutine de drenaje se mantiene (no es un leak real)
- `V1PhonesHandler.PostPhoneQr` — solo cambia `expires_in`, nada más
- `LegacyWSHandler` — completamente separado, no tocar
- Frontend — no hay UI nueva para esta story (es una API B2B)
- Cualquier otro handler de V1 (mensajes, difusiones, etc.)

### 🧠 Cosas a verificar antes de implementar

1. **Import collision:** En el handler, `auth` es tanto un import (`wsapi/internal/auth`) como un nombre de variable tentador. Usar `authpkg` como alias del import: `authpkg "wsapi/internal/auth"` y llamar `authpkg.ParseEmpresaJWT(...)`.

2. **`mapV1EventType`:** El formato exacto de `SessionEvent.Event` es `"qr-{accountID}"`, `"active-{accountID}"`, etc. (ver `service.go` emit calls). El mapeo con `strings.HasPrefix` es correcto.

3. **`writeWSEvent` con `data=nil`:** Para el ping, `data=nil` produce `{"type":"ping"}` sin campo `data`. Verificar que el cliente empresa no falle con eso.

4. **`telefonoStore.BelongsToEmpresa`:** Ya existe en `storage.TelefonoStore`. Verificar firma: `BelongsToEmpresa(telefonoID int64, empresaID int64) (bool, error)`.

5. **`sessionStore` puede ser nil:** En el cleanup defer, el guard `if h.sessionStore != nil && h.manager != nil` protege contra nil en configs de test.

### Learnings de stories anteriores

- Package en `handlers/` es `package http` (NO `package handlers`)
- `fmt.Printf` para logs (no introducir zerolog ni otros)
- `InsecureSkipVerify: true` en websocket.Accept — patrón ya establecido en todo el proyecto
- `storage.NewSessionStore()` para tests
- Nil-guard en todos los stores es obligatorio
- `whatsapp.NormalizeAccountID(phone.NumeroCompleto)` para obtener el accountID canónico
- `time.NewTicker(25 * time.Second)` + `defer ticker.Stop()` — patrón de keepalive de story 2-6

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Reescrito `V1WSHandler.HandleWS` con bridge real: auth JWT → subscribe → validación teléfono → `StartSession` → loop select con eventos/ticker/ctx
- `mapV1EventType` convierte `"qr-{id}"` → `"qr"`, `"active-{id}"` → `"connected"`, resto pasa directo
- Defer cleanup idéntico al admin handler (story 2-6): `manager.Delete(accountID)` solo si `initializing` o `qr_pending`
- Ticker de 25s para keepalive ping, igual que handler admin
- Logs de apertura y cierre con `empresa`, `phone` y `account`
- `NewV1WSHandler` extendido con `telefonoStore` y `sessionStore`; `container.go` actualizado
- `expires_in` corregido de 300 → 60 en los 3 endpoints REST (PostPhoneQr, PostSessions, StartPhoneConnection)
- Los 3 errores de lint pre-existentes en `frontend/lib/api.ts` no son nuevos — AC11 cumplido

### File List

- backend/internal/http/handlers/v1_ws.go
- backend/internal/http/container.go
- backend/internal/http/handlers/v1_phones.go
- backend/internal/http/handlers/v1_sessions.go

### Change Log

- 2026-05-08: Implementación completa story 2-7 — bridge WS empresa, corrección expires_in REST
