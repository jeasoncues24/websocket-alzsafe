---
title: 'Story 2.4 — Módulo sesiones: herramientas y logs'
type: 'feature+bugfix'
created: '2026-05-06'
status: 'done'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: 'cee26ae'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como administrador del panel,
quiero que la página de sesiones muestre la empresa, el estado real en runtime, el historial de eventos recientes, y tenga herramientas para reconectar sesiones caídas,
para que soporte pueda diagnosticar problemas de conexión y actuar desde el panel sin escalar a desarrollo.

## Acceptance Criteria

**AC1:** `GET /api/admin/sesiones` retorna un objeto `summary` en la raíz con conteos exactos: `{total, active, disconnected, mismatch, qr_pending, initializing}`. Los conteos se calculan del listado de sesiones — sin queries adicionales a DB.

**AC2:** Cada sesión en la respuesta incluye: `telefono_id`, `empresa_id`, `empresa_nombre`, `account_id`, `status` (de DB), `runtime_connected` (bool), `mismatch` (bool), `last_connected` (de DB, nullable), `updated_at`, `qr_string` (solo si qr_pending), `events` (últimos 10 eventos de lifecycle, del ring buffer en memoria).

**AC3:** `SessionStore` mantiene un ring buffer de hasta 20 `SessionEventEntry` por cuenta. Se registran eventos: `initializing`, `qr_generated`, `connected`, `disconnected` (con reason). El ring buffer no persiste entre reinicios — solo refleja la sesión del proceso actual.

**AC4:** Los handlers `HandleGetAdminSessions` y `HandlePostAdminSessions` son métodos de `AdminSessionsHandler` con `empresaStore`, `telefonoStore`, `manager`, `sessionStore` inyectados — no recrean DB ni Manager por request. El struct `SessionInfo` de `router.go` se elimina junto a las funciones libres.

**AC5:** La página `/sessions` muestra una barra de métricas en la parte superior con 4 tiles: Activas (verde), Desconectadas (rojo), Inconsistentes (amarillo), QR Pendiente (azul). Los tiles usan los valores del campo `summary` de la respuesta API.

**AC6:** Cada card de sesión muestra: nombre de empresa (prominente), account_id (subtítulo muted), badge de estado (color según status), badge amarillo "Inconsistente" si `mismatch === true`, última conexión formateada (`last_connected` como "hace 2h" o "Nunca"), y un indicador de punto verde/rojo del estado runtime.

**AC7:** Sesiones con `status !== "active"` tienen botón "Reconectar" que llama a `POST /api/admin/telefonos/{telefono_id}/connect` via `fetchWithAuth`. El estado de carga es individual por sesión (no bloquea el resto). Tras éxito, se recarga el listado.

**AC8:** Cada card tiene un toggle expandible "Ver eventos" que muestra los últimos eventos del campo `events` como una mini-línea de tiempo: ícono de tipo + timestamp relativo + descripción. Si `events` está vacío o no disponible, el toggle no aparece.

**AC9:** `cd backend && go build ./...` pasa sin errores.

**AC10:** `cd backend && go test ./...` pasa sin nuevas regresiones.

**AC11:** `cd frontend && npm run lint` pasa sin nuevos errores.

## Tasks / Subtasks

- [x] **Tarea 1: Agregar ring buffer de eventos al SessionStore** (AC: 3)
  - [x] En `backend/internal/storage/sessions.go`, definir struct `SessionEventEntry{Timestamp time.Time, Type string, Details string}`
  - [x] Agregar campo `Events []SessionEventEntry` a `SessionState`
  - [x] Agregar método `AppendEvent(accountID, eventType, details string)` al `SessionStore` — máximo 20 eventos, descarta el más antiguo al superar el límite (ring buffer manual con slice)
  - [x] No modificar los métodos existentes de `SessionStore` (`SetActive`, `SetDisconnected`, etc.) — `AppendEvent` es llamado desde afuera de `sessions.go`

- [x] **Tarea 2: Registrar eventos de lifecycle en whatsapp/service.go** (AC: 3)
  - [x] En `markConnected(accountID)`: llamar `s.sessionStore.AppendEvent(accountID, "connected", "")`
  - [x] En `markDisconnected(accountID, reason)`: llamar `s.sessionStore.AppendEvent(accountID, "disconnected", reason)`
  - [x] En `runSession`, cuando `s.sessionStore.SetInitializing`: llamar `AppendEvent(accountID, "initializing", "")`
  - [x] En `runSession`, cuando `s.sessionStore.SetQRPending` (case "code"): llamar `AppendEvent(accountID, "qr_generated", "")`
  - [x] Todos los `AppendEvent` son no-ops si `s.sessionStore == nil`

- [x] **Tarea 3: Crear `AdminSessionsHandler` struct** (AC: 4)
  - [x] Crear `backend/internal/http/handlers/admin_sessions.go` con `package http`
  - [x] Definir struct `AdminSessionsHandler{empresaStore, telefonoStore, manager, sessionStore}`
  - [x] Definir constructor `NewAdminSessionsHandler(...) *AdminSessionsHandler`
  - [x] Definir `sessionInfoDTO` (privado, lowercase) con todos los campos del AC2
  - [x] Definir `sessionSummaryDTO` (privado) con campos del AC1

- [x] **Tarea 4: Implementar `GetSessions` con respuesta enriquecida** (AC: 1, 2, 4)
  - [x] Iterar empresas → teléfonos usando `h.empresaStore` y `h.telefonoStore`
  - [x] Por cada teléfono: consultar runtime via `h.manager.Get(whatsapp.NormalizeAccountID(t.NumeroCompleto))`
  - [x] Por cada teléfono: consultar events via `h.sessionStore.Get(t.NumeroCompleto)` y extraer `.Events` (últimos 10)
  - [x] Calcular `mismatch = (t.Status == domain.TelefonoStatusActive) != runtimeConnected`
  - [x] Solo incluir `QRString` si `t.Status == domain.TelefonoStatusQRPending`
  - [x] `EmpresaNombre`: usar `empresa.NombreComercial` si no está vacío, sino `empresa.Nombre`
  - [x] Calcular `summary` contando desde el slice de resultados (no queries adicionales)
  - [x] Retornar `{"ok": true, "summary": {...}, "sessions": [...]}`

- [x] **Tarea 5: Implementar `PostSession` (disconnect)** (AC: 4)
  - [x] Mover lógica de `HandlePostAdminSessions` (router.go:250-276) al método `PostSession`
  - [x] Usar `h.telefonoStore.GetByNumeroCompleto` y `h.telefonoStore.SetDisconnected`
  - [x] Registrar evento `AppendEvent(req.AccountID, "disconnected", "manual_admin")` via `h.sessionStore`

- [x] **Tarea 6: Registrar handler en Container y routes, eliminar funciones libres** (AC: 4)
  - [x] Agregar `AdminSessionsHandler *handlers.AdminSessionsHandler` al struct `Container`
  - [x] Instanciar en `NewContainer()`: `adminSessionsHandler := handlers.NewAdminSessionsHandler(empresaStore, telefonoStore, manager, sessionStore)`
  - [x] Actualizar `routes_admin.go` líneas 81-82 para usar `c.AdminSessionsHandler`
  - [x] Eliminar de `router.go`: struct `SessionInfo`, `HandleGetAdminSessions`, `HandlePostAdminSessions`

- [x] **Tarea 7: Actualizar tipos y funciones en `frontend/lib/api.ts`** (AC: 2, 7)
  - [x] Extender `SessionInfo` con campos opcionales: `telefono_id?`, `empresa_id?`, `empresa_nombre?`, `runtime_connected?`, `mismatch?`, `last_connected?`, `events?`
  - [x] Definir `SessionEvent{timestamp, type, details?}` y `SessionSummary{total, active, disconnected, mismatch, qr_pending, initializing}`
  - [x] Extender `SessionsResponse` con `summary?: SessionSummary`
  - [x] Agregar `reconnectAdminSession(telefonoId: number)` usando `fetchWithAuth` (NO `fetchWithEmpresaAuth`)
  - [x] NO modificar `connectEmpresaTelefono` existente — se usa en el flujo de empresa separado

- [x] **Tarea 8: Actualizar página de sesiones en frontend** (AC: 5, 6, 7, 8)
  - [x] Agregar barra de métricas en la parte superior con 4 tiles usando `data.summary`
  - [x] Mostrar `session.empresa_nombre` como título principal del card, `session.account_id` como subtítulo
  - [x] Implementar badge de estado con colores: verde (active), rojo (disconnected), amarillo (qr_pending/initializing)
  - [x] Agregar badge amarillo "Inconsistente" si `session.mismatch === true`
  - [x] Mostrar `last_connected` formateado: usar `formatDistanceToNow` de `date-fns` o implementación propia sin librería nueva
  - [x] Agregar indicador visual runtime: punto verde si `runtime_connected`, punto rojo si no
  - [x] Implementar botón "Reconectar" con estado de carga individual por `account_id`
  - [x] Implementar toggle "Ver eventos" con mini-timeline si `session.events` tiene items

- [x] **Tarea 9: Verificar build y tests** (AC: 9, 10, 11)
  - [x] `cd backend && go build ./...` — OK
  - [x] `cd backend && go test ./...` — storage, whatsapp, http OK; fallos en handlers/middleware son pre-existentes (story 2-2)
  - [x] `cd frontend && npm run lint` — sin nuevos errores (3 pre-existentes en api.ts líneas 42, 637, 641)

## Dev Notes

### 🔴 Bug raíz — igual que story 2-3

`HandleGetAdminSessions` y `HandlePostAdminSessions` son **funciones libres** en `router.go:213-276` que recrean la conexión DB en cada request (`storage.NewDB(cfg)` dentro del handler). La solución es el mismo patrón ya aplicado en story 2-3 con `AdminMessagesHandler`.

### ⚠️ Estandarización — impacto en otros archivos

**`SessionInfo` struct vive en `router.go:206-211`.** Solo es referenciada internamente por `HandleGetAdminSessions` en el mismo archivo. Al moverla como `sessionInfoDTO` (privado) al nuevo handler, no hay callers externos que se rompan.

Verificar antes de eliminar:
```bash
grep -rn "SessionInfo" backend/internal/
```
El único resultado esperado son las 3 líneas en `router.go`. Si aparece en otro archivo, documentarlo y adaptar.

`ApiKeysHandler.runtimeSessionInfo` en `handlers/api_keys.go:427` tiene nombre similar pero es un **método diferente** que retorna `map[string]any` — no está relacionado con el struct `SessionInfo`. No modificar.

**Contrato JSON de `GET /api/admin/sesiones`:** La respuesta actual tiene `{ok, sessions:[{account_id, status, qr_string, updated_at}]}`. El nuevo formato agrega campos y un `summary` en la raíz. Los campos nuevos son aditivos — el frontend existente (`getAdminSessions` en api.ts) seguirá funcionando porque `SessionInfo` usa campos opcionales.

### 📐 Estructura del ring buffer de eventos

```go
// storage/sessions.go — AGREGAR (no modificar SessionState existente excepto agregar Events)
type SessionEventEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Type      string    `json:"type"`    // "connected","disconnected","qr_generated","initializing"
    Details   string    `json:"details,omitempty"` // reason de disconnect, o vacío
}

// Agregar a SessionState:
Events []SessionEventEntry

// Nuevo método en SessionStore:
func (s *SessionStore) AppendEvent(accountID, eventType, details string) {
    accountID = normalizeSessionAccountID(accountID)
    s.mu.Lock()
    defer s.mu.Unlock()
    state, ok := s.state[accountID]
    if !ok {
        state = SessionState{AccountID: accountID}
    }
    entry := SessionEventEntry{Timestamp: time.Now(), Type: eventType, Details: details}
    state.Events = append(state.Events, entry)
    if len(state.Events) > 20 {
        state.Events = state.Events[len(state.Events)-20:] // keep last 20
    }
    s.state[accountID] = state
}
```

### 📐 Llamadas a AppendEvent en whatsapp/service.go

Insertar DESPUÉS de cada llamada existente a sessionStore:

```go
// En runSession, después de s.sessionStore.SetInitializing(accountID):
if s.sessionStore != nil {
    s.sessionStore.AppendEvent(accountID, "initializing", "")
}

// En runSession case "code" (QR), después de s.sessionStore.SetQRPending:
if s.sessionStore != nil {
    s.sessionStore.AppendEvent(accountID, "qr_generated", "")
}

// En markConnected, después de s.sessionStore.SetActive(accountID):
if s.sessionStore != nil {
    s.sessionStore.AppendEvent(accountID, "connected", "")
}

// En markDisconnected, después de s.sessionStore.SetDisconnected(accountID, reason):
if s.sessionStore != nil {
    s.sessionStore.AppendEvent(accountID, "disconnected", reason)
}
```

Valores de `reason` existentes: `"disconnect"`, `"stream_replaced"`, `"logged_out"`, `"temporary_ban"`, `"connect_failure"`, `"connect_error"`, `"connect_timeout"`, `"qr_channel_closed"`, `"qr_timeout"`, `"qr_error"`.

### 📐 sessionInfoDTO y sessionSummaryDTO

```go
// backend/internal/http/handlers/admin_sessions.go

type sessionEventDTO struct {
    Timestamp time.Time `json:"timestamp"`
    Type      string    `json:"type"`
    Details   string    `json:"details,omitempty"`
}

type sessionInfoDTO struct {
    TelefonoID      int64             `json:"telefono_id"`
    EmpresaID       int64             `json:"empresa_id"`
    EmpresaNombre   string            `json:"empresa_nombre"`
    AccountID       string            `json:"account_id"`
    Status          string            `json:"status"`          // de DB
    RuntimeConnected bool             `json:"runtime_connected"`
    Mismatch        bool              `json:"mismatch"`
    QRString        string            `json:"qr_string,omitempty"`
    LastConnected   *time.Time        `json:"last_connected,omitempty"`
    UpdatedAt       time.Time         `json:"updated_at"`
    Events          []sessionEventDTO `json:"events,omitempty"` // últimos 10
}

type sessionSummaryDTO struct {
    Total        int `json:"total"`
    Active       int `json:"active"`
    Disconnected int `json:"disconnected"`
    Mismatch     int `json:"mismatch"`
    QRPending    int `json:"qr_pending"`
    Initializing int `json:"initializing"`
}
```

### 📐 Lógica de GetSessions

```go
func (h *AdminSessionsHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
    sessions := []sessionInfoDTO{}
    if h.empresaStore != nil && h.telefonoStore != nil {
        empresas, _, err := h.empresaStore.GetAll(1, 1000, "", nil)
        if err == nil {
            for i := range empresas {
                telefonos, err := h.telefonoStore.GetByEmpresa(empresas[i].ID)
                if err != nil { continue }
                nombre := empresas[i].NombreComercial
                if nombre == "" { nombre = empresas[i].Nombre }
                for _, t := range telefonos {
                    accountID := whatsapp.NormalizeAccountID(t.NumeroCompleto)
                    runtimeConnected := false
                    if h.manager != nil {
                        if client, ok := h.manager.Get(accountID); ok && client != nil && client.IsConnected() {
                            runtimeConnected = true
                        }
                    }
                    mismatch := (t.Status == domain.TelefonoStatusActive) != runtimeConnected
                    var events []sessionEventDTO
                    if h.sessionStore != nil {
                        if state, ok := h.sessionStore.Get(t.NumeroCompleto); ok {
                            last := state.Events
                            if len(last) > 10 { last = last[len(last)-10:] }
                            for _, e := range last {
                                events = append(events, sessionEventDTO{Timestamp: e.Timestamp, Type: e.Type, Details: e.Details})
                            }
                        }
                    }
                    qr := ""
                    if t.Status == domain.TelefonoStatusQRPending { qr = t.QRString }
                    sessions = append(sessions, sessionInfoDTO{
                        TelefonoID: t.ID, EmpresaID: empresas[i].ID,
                        EmpresaNombre: nombre, AccountID: t.NumeroCompleto,
                        Status: string(t.Status), RuntimeConnected: runtimeConnected,
                        Mismatch: mismatch, QRString: qr,
                        LastConnected: t.LastConnected, UpdatedAt: t.UpdatedAt,
                        Events: events,
                    })
                }
            }
        }
    }
    summary := computeSessionSummary(sessions)
    writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
        "ok": true, "summary": summary, "sessions": sessions,
    })
}

func computeSessionSummary(sessions []sessionInfoDTO) sessionSummaryDTO {
    s := sessionSummaryDTO{Total: len(sessions)}
    for _, sess := range sessions {
        switch sess.Status {
        case "active":       s.Active++
        case "disconnected": s.Disconnected++
        case "qr_pending":   s.QRPending++
        case "initializing": s.Initializing++
        }
        if sess.Mismatch { s.Mismatch++ }
    }
    return s
}
```

### 📐 Mismatch — definición precisa

Un mismatch es cuando el estado en DB y el runtime no coinciden:
- **Caso A (crítico):** `t.Status == active` pero `runtimeConnected == false` → sesión "activa" en DB pero caída en memoria. Soporte la ve como activa pero los envíos fallan.
- **Caso B (menor):** `t.Status != active` pero `runtimeConnected == true` → runtime conectado pero DB no lo sabe. Raro, ocurre si el proceso reconectó sin actualizar DB.

`mismatch = (t.Status == domain.TelefonoStatusActive) != runtimeConnected` cubre ambos casos.

### 📐 Función reconnect en api.ts

```ts
// frontend/lib/api.ts — AGREGAR (no modificar connectEmpresaTelefono)
export async function reconnectAdminSession(
  telefonoId: number,
): Promise<{ ok: boolean; status?: string; qr_string?: string; error?: string }> {
  return fetchWithAuth(
    `${API_BASE}/api/admin/telefonos/${telefonoId}/connect`,
    { method: "POST" },
  );
}
```

**Por qué NO usar `connectEmpresaTelefono`:** usa `fetchWithEmpresaAuth` (JWT de empresa). El panel admin usa `fetchWithAuth` (JWT de admin). El endpoint `StartCompanyPhoneConnection` ya acepta admin JWT via `domain.GetPanelAccess`.

### 🎨 UX Design — Barra de métricas

Cuatro tiles horizontales en la parte superior de `/sessions`:

```
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│  3 Activas  │ │ 1 Desconec. │ │ 1 Inconsist.│ │  0 QR Pend. │
│  ●●● verde  │ │  ●●● rojo   │ │ ●●● amarillo│ │  ●●● azul   │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
```

Usar `Card` + `CardContent` de `@/components/ui/card`. Clases Tailwind para colores:
- Activas: `text-green-600`, border verde suave
- Desconectadas: `text-red-600`
- Inconsistentes: `text-yellow-600`
- QR Pendiente: `text-blue-600`

### 🎨 UX Design — Cards de sesión

```
┌──────────────────────────────────────────┐
│ Nombre Empresa S.A.         [● Activa]   │
│ +5219999999999               ● runtime   │
│ Última conexión: hace 2 horas            │
├──────────────────────────────────────────┤
│ [Desconectar]          [▼ Ver eventos]   │
└──────────────────────────────────────────┘

Con mismatch:
┌──────────────────────────────────────────┐
│ Empresa Cliente        [● Activa] [⚠ !] │
│ +5219888888888               ○ runtime   │ ← punto rojo = desconectado
│ Última conexión: hace 3 días             │
├──────────────────────────────────────────┤
│ [Reconectar]           [▼ Ver eventos]   │
└──────────────────────────────────────────┘
```

Badge mismatch: `<Badge variant="outline" className="border-yellow-500 text-yellow-600">Inconsistente</Badge>`

Punto runtime: `<span className={session.runtime_connected ? "text-green-500" : "text-red-400"}>●</span>`

### 🎨 UX Design — Mini timeline de eventos

```
▼ Ver eventos (click para expandir)
  ─────────────────────────────────
  ● connected    hace 2h
  ○ disconnected hace 3h  stream_replaced
  ○ qr_generated hace 3h
  ○ initializing hace 3h
```

Usar `useState<boolean>` por card para controlar el expand. Íconos sugeridos de lucide-react:
- `connected` → `Wifi`
- `disconnected` → `WifiOff`
- `qr_generated` → `QrCode`
- `initializing` → `Loader2`

Formato de timestamp: diferencia relativa simple sin librería nueva:
```ts
function relativeTime(ts: string): string {
  const diff = Date.now() - new Date(ts).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return "ahora"
  if (m < 60) return `hace ${m}m`
  const h = Math.floor(m / 60)
  if (h < 24) return `hace ${h}h`
  return `hace ${Math.floor(h / 24)}d`
}
```

### 📊 Métricas correctas — definición para operaciones

| Métrica | Fuente | Valor saludable | Alerta cuando |
|---------|--------|-----------------|---------------|
| `active` | sessions donde status=active | 100% de las configuradas | < 80% del total |
| `mismatch` | DB active + runtime disconnected | 0 | > 0 (siempre es un problema) |
| `disconnected` | status=disconnected | 0 | > 0 persistente (> 5 min sin reconectar) |
| `qr_pending` | status=qr_pending | 0 en producción | > 0 en producción significa sesión sin completar |
| `initializing` | status=initializing | 0 en estado estable | > 0 por más de 30 segundos |

El **mismatch** es el indicador más crítico porque es el que causa fallos silenciosos de envío: la DB cree que la sesión está activa (y los envíos se despachan al teléfono), pero el runtime ya no está conectado.

### Archivos a modificar

| Archivo | Cambio | Tipo |
|---------|--------|------|
| `backend/internal/storage/sessions.go` | Agregar `SessionEventEntry`, campo `Events` en `SessionState`, método `AppendEvent` | MODIFICAR |
| `backend/internal/whatsapp/service.go` | Llamar `AppendEvent` en 4 puntos de lifecycle | MODIFICAR |
| `backend/internal/http/handlers/admin_sessions.go` | Struct + constructor + GetSessions + PostSession | NUEVO |
| `backend/internal/http/container.go` | Agregar `AdminSessionsHandler` | MODIFICAR |
| `backend/internal/http/routes_admin.go` | Líneas 81-82: usar `c.AdminSessionsHandler` | MODIFICAR |
| `backend/internal/http/router.go` | Eliminar `SessionInfo`, `HandleGetAdminSessions`, `HandlePostAdminSessions` | MODIFICAR |
| `frontend/lib/api.ts` | Extender `SessionInfo`, agregar tipos, agregar `reconnectAdminSession` | MODIFICAR |
| `frontend/app/sessions/page.tsx` | Summary bar, empresa nombre, mismatch badge, reconectar, eventos | MODIFICAR |

### Learnings de story 2-3 (aplicar aquí)

- Package en `handlers/` es `package http` (NO `package handlers`)
- Usar `writeHandlerJSON` y `writeAPIError` — NO definir funciones propias de response
- Eliminar imports no usados antes de `go build` (encoding/json si no se usa directamente)
- Nil guard para stores (`if h.empresaStore != nil && h.telefonoStore != nil`)

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Ring buffer de 20 eventos implementado en `SessionStore.AppendEvent`. El método privado `set` fue modificado para preservar `Events` al hacer transiciones de estado (necesario para que el buffer sobreviva entre `SetActive`/`SetQRPending`/etc).
- `AdminSessionsHandler` sigue el mismo patrón que `AdminMessagesHandler` (story 2-3): dependencias inyectadas, `package http`, `writeHandlerJSON`/`writeAPIError`.
- Frontend: barra de métricas 4 tiles, cards con empresa nombre, badge mismatch, punto runtime, botón Reconectar con loading individual, toggle Ver eventos con mini-timeline.
- `reconnectAdminSession` usa `fetchWithAuth` (JWT admin), NO `fetchWithEmpresaAuth`.
- Fallos de test pre-existentes: mocks de `EmpresaStoreInterface` no implementan `Restore` (story 2-2). No introducidos por esta story.
- Errores de lint pre-existentes: 3 errores `@typescript-eslint/no-explicit-any` en api.ts (líneas 42, 637, 641). No introducidos por esta story.

### File List

- `backend/internal/storage/sessions.go` — MODIFICADO
- `backend/internal/whatsapp/service.go` — MODIFICADO
- `backend/internal/http/handlers/admin_sessions.go` — NUEVO
- `backend/internal/http/container.go` — MODIFICADO
- `backend/internal/http/routes_admin.go` — MODIFICADO
- `backend/internal/http/router.go` — MODIFICADO (eliminado SessionInfo + funciones libres)
- `frontend/lib/api.ts` — MODIFICADO
- `frontend/app/sessions/page.tsx` — MODIFICADO
