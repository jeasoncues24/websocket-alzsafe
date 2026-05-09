---
title: 'Story 2.11 — Documentar arquitectura WS'
type: 'documentation'
created: '2026-05-08'
status: 'done'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '3c41157'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como desarrollador o integrador que trabaja con wsapi (y que NO necesariamente sabe Go),
quiero un documento de referencia en lenguaje sencillo que explique la arquitectura del canal WebSocket (sesiones, QR, eventos, lifecycle),
para poder integrar o mantener el sistema sin tener que leer el código fuente ni conocer Go.

## Acceptance Criteria

**AC1 — Documento creado en `docs/ws-arquitectura.md`:**
El archivo existe y está en español.

**AC2 — Sección: Endpoints WS:**
Documenta los dos endpoints WS con ruta, auth y propósito:
- Admin: `GET /api/admin/telefonos/{id}/connect/ws` — para el panel admin, JWT admin
- Service: `GET /api/service/v1/ws` — para integraciones B2B, JWT de empresa o JWT qr_link

**AC3 — Sección: Flujos de autenticación en V1 WS:**
Documenta los dos paths del `V1WSHandler`:
- **JWT empresa regular** (`scope` vacío): espera un mensaje `{"type":"subscribe","data":{"phone_id":N}}`, valida que el teléfono pertenece a la empresa
- **JWT qr_link** (`scope=="qr_link"`): auto-suscribe al `phone_id` del token sin esperar mensaje, dura 10 minutos

**AC4 — Sección: Máquina de estados de sesión:**
Diagrama o tabla que muestra los estados (`initializing` → `qr_pending` → `active` → `disconnected`) con las transiciones y qué las dispara.

**AC5 — Sección: Formato de eventos WS:**
Documenta todos los tipos de eventos que el servidor puede enviar al cliente:

| type | data | descripción |
|------|------|-------------|
| `qr` | `{qr_string, expires_in, message}` | código QR nuevo |
| `connected` | `{isActive:true, message}` | sesión activa |
| `disconnected` | `{isActive:false, reason, requiresNewQR}` | sesión caída |
| `ping` | — | keepalive cada 25 s |
| `error` | `{message}` | error fatal, WS se cierra |

**AC6 — Sección: Lifecycle de goroutines y cleanup:**
Explica que cada sesión corre en una goroutine (`runSession`) y que los handlers WS tienen un `defer` que llama `manager.Delete(accountID)` si el estado es `initializing` o `qr_pending` al cerrar, evitando sesiones zombi.

**AC7 — Sección: Flujo QR Link (token provisional):**
Documenta el flujo completo:
1. Admin llama `POST /api/admin/telefonos/{id}/qr-link` → recibe token JWT (10 min, scope=qr_link)
2. Frontend construye URL `${origin}/qr?token=TOKEN`
3. Operador abre el enlace en su navegador (sin login)
4. La página `/qr` conecta al WS con el token, el servidor auto-suscribe al teléfono
5. El operador escanea el QR y la sesión queda activa

**AC8 — Sección: Startup bootstrap:**
Explica que al iniciar el servidor, `StartupBootstrapper` reconecta automáticamente todos los teléfonos con `activo=true` en DB (máx 4 concurrentes, 2 reintentos, 1.2 s entre reintentos). El resultado se loguea como `[INFO] startup bootstrap sesiones: ...`.

**AC9 — Precisión técnica:**
Toda la información del documento es consistente con el código actual (no inventada). Los nombres de campos, rutas y comportamientos deben coincidir con la implementación real.

## Tasks / Subtasks

- [x] **Tarea 1: Crear `docs/ws-arquitectura.md`** (AC: 1–9)
  - [x] Sección: Endpoints WS (AC2)
  - [x] Sección: Autenticación V1 WS — dos paths (AC3)
  - [x] Sección: Máquina de estados de sesión (AC4)
  - [x] Sección: Formato de eventos WS (AC5)
  - [x] Sección: Lifecycle de goroutines y cleanup (AC6)
  - [x] Sección: Flujo QR Link (AC7)
  - [x] Sección: Startup bootstrap al arrancar (AC8)

## Dev Notes

### 🗂️ Archivo de salida

```
docs/ws-arquitectura.md
```
Junto a `docs/deploy-backend.md` y `docs/bmad-project-rules.md`.

### 🔌 Endpoints WS — código real

**Admin** (`backend/internal/http/admin.go`):
- Ruta: `GET /api/admin/telefonos/{id}/connect/ws`
- Auth: JWT admin (panel administrativo)
- Handler: `AdminHandler.ConnectCompanyPhoneWS`
- Cleanup: defer `manager.Delete(accountID)` si `qr_pending` o `initializing`

**Service V1** (`backend/internal/http/handlers/v1_ws.go`):
- Ruta: `GET /api/service/v1/ws`
- Auth: JWT empresa (`?token=TOKEN` o `Authorization: Bearer TOKEN`)
- Handler: `V1WSHandler.HandleWS`
- Keepalive: `time.NewTicker(25 * time.Second)` → envía `{"type":"ping"}`
- Cleanup: mismo patrón de defer que admin

### 🔐 Dos paths de auth en V1WSHandler

```go
if claims.Scope == "qr_link" {
    // auto-suscribir: usar claims.PhoneID directamente
    phoneID = claims.PhoneID
} else {
    // esperar mensaje subscribe del cliente
    _, data, err := c.Read(ctx)
    // ... unmarshal, validar phone_id, BelongsToEmpresa(sub.PhoneID, claims.EmpresaID)
    phoneID = sub.PhoneID
}
```

El JWT `qr_link` se genera en `POST /api/admin/telefonos/{id}/qr-link` con claims:
```go
// sub = empresa_id, phone_id = N, scope = "qr_link", exp = now + 10min
```

### 📊 Máquina de estados — código real (`whatsapp/service.go`)

```
[inicio] → SetInitializing() → "initializing"
    ↓ GetQRChannel
"qr_pending"  ← SetQRPending(code)   (cada código QR)
    ↓ escaneo exitoso
"active"      ← SetActive()
    ↓ disconnect event (Disconnected/LoggedOut/StreamReplaced/TemporaryBan/ConnectFailure)
"disconnected" ← SetDisconnected(reason)
```

Transiciones guardadas en `sessionStore` (in-memory `storage.SessionStore`) y auditadas en `sessionStore.AppendEvent(accountID, tipo, detalle)`.

### 📨 Eventos WS — código real (`whatsapp/service.go`, `handlers/v1_ws.go`)

El `Service` emite `SessionEvent{Event, Data}` al canal. El handler mapea con `mapV1EventType`:

```go
func mapV1EventType(event string) string {
    switch {
    case strings.HasPrefix(event, "qr-"):    return "qr"
    case strings.HasPrefix(event, "active-"): return "connected"
    default:                                  return event
    }
}
```

Payload exacto de cada tipo:
- **qr**: `{"qr_string": "2@abc...", "expires_in": 60, "message": "Escanee el codigo QR para iniciar sesion."}`
- **connected** (activo): `{"isActive": true, "message": "Sesion activa"}`
- **connected** (no activo): `{"isActive": false, "reason": "qr_timeout|logged_out|...", "requiresNewQR": true}`
- **ping**: sin campo `data`
- **error**: `{"message": "descripción del error"}`

Razones de desconexión conocidas: `disconnect`, `stream_replaced`, `logged_out`, `temporary_ban`, `connect_failure`, `qr_timeout`, `qr_error`, `qr_channel_closed`, `connect_timeout`, `connect_error`.

### 🔄 Goroutines y cleanup

Cada sesión activa tiene:
- Una goroutine `runSession(accountID, runtime)` en `Service`
- Un `sessionRuntime` con `ctx`, `cancel`, `client`, `events chan SessionEvent`, `storage`
- El WS handler corre su propio loop `select` sobre el canal de eventos

Cleanup en WS handlers (tanto admin como V1):
```go
defer func() {
    // ...logs...
    if state.Status == "initializing" || state.Status == "qr_pending" {
        h.manager.Delete(accountID)  // para la goroutine runSession
    }
}()
```
`manager.Delete` llama `runtime.cancel()` → `runSession` termina → cierra `events` chan → el WS handler sale del loop.

### 🚀 Startup bootstrap

Archivo: `backend/internal/whatsapp/startup_bootstrap.go`

Al arrancar (si `WHATSAPP_BOOTSTRAP_ENABLED=true` en `.env`):
1. Lista todos los teléfonos con `status="active"` en DB
2. Verifica cuáles ya tienen sesión runtime activa (no los toca)
3. Llama `StartSession()` para los mismatches — máx `WHATSAPP_BOOTSTRAP_MAX_CONCURRENCY` (default 4) en paralelo
4. 2 reintentos con 1.2 s de pausa entre intentos si falla
5. Loguea: `[INFO] startup bootstrap sesiones: total=N activos_db=A runtime_activos=R mismatches=M intentos_start=I errores_start=E duracion=Xs`

### 🔑 normalización de accountID

El `SessionStore` usa `phone.NumeroCompleto` como clave (ej: `"+51999888777"`).
`whatsapp.NormalizeAccountID(numero)` normaliza el formato para usarlo como clave en el `Manager` y en SQLite. Ambas keys deben ser consistentes — el código ya maneja esto internamente.

### ✍️ Tono y estilo — MUY IMPORTANTE

El documento es para dos audiencias:
1. **Integradores** (consumen el servicio vía API/WS, pueden no saber Go ni el internals)
2. **Desarrolladores** (mantienen el código, pero pueden ser nuevos en el proyecto)

Reglas de escritura:
- **Sin jerga Go**: no decir "goroutine", "chan", "defer" en el texto principal — si es necesario mencionarlos, explicarlos brevemente en una nota al pie o entre paréntesis
- **Diagramas en ASCII o tablas** siempre que sea posible para flujos y estados
- **Ejemplos JSON concretos** para todos los tipos de eventos — el integrador copia y pega
- **Lenguaje activo**: "el servidor envía", "el cliente manda", "si el QR expira, el servidor cierra la conexión"
- **Secciones para integradores primero**, implementación interna al final (opcional para quien mantiene el código)
- Evitar frases como "el runtime gestiona el ciclo de vida del sessionRuntime" — en su lugar: "el servidor mantiene la sesión activa mientras el cliente WS esté conectado"

### ⚠️ Lo que NO documenta esta story

- Endpoints REST (mensajes, empresas, api-keys) — fuera del scope
- Esquema completo de la DB — fuera del scope
- Configuración de entorno — ya en `docs/deploy-backend.md`
- Frontend interno del panel admin — fuera del scope

### Learnings de stories anteriores

- Los docs van en `docs/` (markdown en español)
- Baja prioridad: el documento debe ser útil y preciso, no exhaustivo
- El código es la fuente de verdad — no inventar comportamientos

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- `docs/ws-arquitectura.md` creado con 7 secciones: Endpoints WS, Autenticación (dos flujos), Estados de sesión (diagrama ASCII), Eventos con JSON completos, Ciclo de vida de la conexión, Flujo QR Link (diagrama ASCII con secuencia completa), Reconexión al arrancar.
- Tono: sin jerga Go, lenguaje activo, orientado a integradores y desarrolladores que no conocen Go.
- Datos verificados directamente contra: `v1_ws.go`, `admin_sessions.go`, `startup_bootstrap.go`.
- Keepalive: 25 segundos (confirmado en código).
- Bootstrap: maxConcurrency=4, maxRetries=2, retryDelay=1.2s (confirmado en código).

### File List

- docs/ws-arquitectura.md (nuevo)

### Change Log

- 2026-05-08: Story creada — documentar arquitectura WS completa (endpoints, auth, estados, eventos, goroutines, bootstrap, QR link)
- 2026-05-08: Implementación completa — `docs/ws-arquitectura.md` creado con todas las secciones requeridas
