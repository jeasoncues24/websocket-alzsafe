---
title: 'Story — WS Connect por API Key: iniciar sesión y emitir QR en tiempo real'
type: 'feature'
created: '2026-06-03'
status: 'done'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como empresa cliente (B2B) que posee una API key vinculada a un teléfono concreto,
quiero poder verificar el estado de conexión de mi teléfono con un endpoint REST y, si no está activo, conectarme a un WebSocket autenticado con mi misma API key para iniciar la sesión y recibir el código QR en tiempo real,
para que pueda integrar el flujo completo de conexión WhatsApp en mi propio frontend sin depender del panel de administración ni de tokens de corta duración generados por terceros.

---

## Contexto técnico

### Lo que ya existe (sin cambios)

Estos endpoints **ya están implementados y funcionan**. Solo se documentan aquí para que el frontend los consuma correctamente.

#### `GET /api/service/v1/sesion`

Autenticado con API key. Devuelve el estado de conexión del teléfono vinculado a la key.

**Request:**
```
GET /api/service/v1/sesion
X-API-Key: wsapikey_xxxx...
```

**Response exitosa:**
```json
{
  "ok": true,
  "data": {
    "telefono_id": 5,
    "account_id": "51999888777",
    "status_db": "disconnected",
    "status_runtime": "disconnected",
    "runtime_connected": false,
    "mismatch": false,
    "mismatch_reason": "",
    "recommended_action": "iniciar_conexion"
  }
}
```

**Valores de `status_db`:** `active` | `qr_pending` | `disconnected`

**Valores de `recommended_action`:**
- `"none"` → el teléfono está conectado, no hacer nada
- `"iniciar_conexion"` → el teléfono está desconectado, iniciar sesión
- `"reanudar_conexion"` → la DB dice activo pero el runtime no lo tiene, reconectar

**Campo clave para el frontend:** `runtime_connected: true` = conectado; `false` = hay que abrir el WS.

#### `GET /api/service/v1/me`

Devuelve lo mismo que `/sesion` pero además incluye los datos completos de empresa, teléfono y API key. Útil para cargar el contexto inicial de la sesión en el frontend.

---

### Lo que se agrega (nuevo)

Un endpoint WebSocket que el cliente B2B puede consumir directamente con su API key para iniciar la sesión WhatsApp y recibir el QR en tiempo real.

**Ruta nueva:** `GET /api/service/v1/ws/connect`

**Diferencia con el WS existente (`/api/service/v1/ws`):**

| | `/api/service/v1/ws` (existente) | `/api/service/v1/ws/connect` (nuevo) |
|--|--|--|
| Autenticación | JWT QR-link generado por el admin (corta duración) | API key del cliente B2B (larga duración) |
| ¿Quién lo usa? | El panel de admin interno | El cliente B2B en su propio frontend |
| Origen del teléfono | Codificado en el JWT | Implícito en la API key (`key.TelefonoID`) |

---

## Flujo completo para el frontend del cliente B2B

```
┌─────────────────────────────────────────────────────────┐
│  PASO 1 — Verificar estado (REST, ya existe)            │
│                                                         │
│  GET /api/service/v1/sesion                             │
│  X-API-Key: wsapikey_xxxx                               │
│                                                         │
│  ¿runtime_connected == true?                            │
│    SÍ → Teléfono activo, no hacer nada.                 │
│    NO → Ir a Paso 2.                                    │
└─────────────────────────────────────────────────────────┘
                          │ NO
                          ▼
┌─────────────────────────────────────────────────────────┐
│  PASO 2 — Abrir WS para conectar y obtener QR (nuevo)   │
│                                                         │
│  WS: GET /api/service/v1/ws/connect?api_key=wsapikey_xx │
│                                                         │
│  Eventos que llegan:                                    │
│                                                         │
│  {"type":"qr","data":{"qr_string":"2@xxx...","expires_in":60}}
│    → Mostrar QR en frontend para escanear               │
│                                                         │
│  {"type":"connected","data":{"isActive":true}}          │
│    → QR escaneado, teléfono conectado ✓                 │
│    → Cerrar WS (ya no se necesita)                      │
│                                                         │
│  {"type":"disconnected","data":{"isActive":false,...}}  │
│    → Sesión caída durante el flujo                      │
│    → Reintentar o notificar al usuario                  │
│                                                         │
│  {"type":"ping"}                                        │
│    → Keepalive del servidor, ignorar en frontend        │
│                                                         │
│  {"type":"error","data":{"message":"..."}}              │
│    → Error en el servidor, cerrar WS y manejar          │
└─────────────────────────────────────────────────────────┘
```

**Formas de pasar la API key en el WS** (en orden de preferencia):
1. Query param: `?api_key=wsapikey_xxx` — la más simple para clientes WS
2. Header `X-API-Key: wsapikey_xxx` — si el cliente permite headers en el handshake
3. Header `Authorization: ApiKey wsapikey_xxx` o `Authorization: Bearer wsapikey_xxx`
4. Header `Sec-WebSocket-Protocol: wsapikey_xxx` — fallback para browsers que no permiten headers personalizados

---

## Acceptance Criteria

**AC1 — API key ausente rechazada antes del upgrade:**
Dado un cliente que no incluye API key en ningún lugar,
cuando intenta conectarse a `GET /api/service/v1/ws/connect`,
entonces el servidor responde HTTP 401 (antes del upgrade WebSocket) con `{"ok":false,"error":"API_KEY_REQUIRED","message":"API key requerida"}`.

**AC2 — API key inválida rechazada antes del upgrade:**
Dado un cliente con una API key incorrecta, expirada o revocada,
cuando intenta conectarse,
entonces el servidor responde HTTP 401 con `{"ok":false,"error":"INVALID_API_KEY","message":"API key inválida o expirada"}`.

**AC3 — Teléfono no encontrado rechazado antes del upgrade:**
Dado una API key válida cuyo teléfono fue eliminado,
cuando intenta conectarse,
entonces el servidor responde HTTP 401 con `{"ok":false,"error":"TELEFONO_NOT_FOUND","message":"Teléfono no encontrado"}`.

**AC4 — Upgrade exitoso e inicio de sesión:**
Dado un cliente con API key válida (teléfono en cualquier estado: `disconnected`, `qr_pending` o `active`),
cuando se establece la conexión WebSocket,
entonces el handler llama `whatsapp.StartSession(h.manager, phone.NumeroCompleto)` y hace bridge de todos los eventos de sesión al WS cliente.

**AC5 — Evento QR emitido al cliente:**
Cuando la sesión genera un QR (estado `qr_pending`),
el WS cliente recibe `{"type":"qr","data":{"qr_string":"<datos>","expires_in":60,"message":"Escanea el código QR"}}`.

**AC6 — Evento connected emitido al cliente:**
Cuando el teléfono completa la conexión,
el WS cliente recibe `{"type":"connected","data":{"isActive":true,"message":"..."}}`.

**AC7 — Keepalive ping cada 25 segundos:**
Mientras la conexión WS esté activa, el servidor envía `{"type":"ping"}` cada 25 segundos. Si el write falla, el handler retorna limpiamente.

**AC8 — Cleanup al cerrar conexión:**
Cuando el WS se cierra (por cualquier causa),
el handler llama `unsubscribe()` y registra `sessionStore.AppendEvent(phone.NumeroCompleto, "ws_closed", "WS cliente V1 connect cerrado: <razón>")`.

**AC9 — Logs de apertura y cierre:**
- Apertura: `[INFO] V1 WS connect opened phone=<id> account=<accountID> empresa=<empresaID>`
- Cierre: `[INFO] V1 WS connect closed phone=<id> account=<accountID> reason=<err>`

**AC10 — `V1WSHandler` actualizado con `apiKeyStore`:**
- El struct `V1WSHandler` incluye `apiKeyStore *storage.ApiKeyStore`.
- `NewV1WSHandler` acepta `apiKeyStore *storage.ApiKeyStore` como parámetro adicional.
- `container.go` pasa `apiKeyStore` al constructor (ya existe la variable `apiKeyStore` en `NewContainer`).

**AC11 — Ruta registrada sin middleware de empresa:**
```go
mux.Handle("GET /api/service/v1/ws/connect", http.HandlerFunc(c.V1WSHandler.HandleConnectWS))
```
Sin `clientStack`, igual que el WS existente.

**AC12 — Build pasa:** `cd backend && go build ./...` sin errores.

**AC13 — Tests no regresionan:** `cd backend && go test ./...` sin nuevas regresiones.

---

## Tasks / Subtasks

- [ ] **Tarea 1: Añadir `apiKeyStore` a `V1WSHandler`** (AC: 10)
  - [ ] `backend/internal/http/handlers/v1_ws.go`: añadir campo `apiKeyStore *storage.ApiKeyStore` al struct
  - [ ] Actualizar `NewV1WSHandler` para aceptar `apiKeyStore *storage.ApiKeyStore` como último parámetro
  - [ ] `backend/internal/http/container.go`: pasar `apiKeyStore` (ya existe en scope) a `NewV1WSHandler`

- [ ] **Tarea 2: Implementar `HandleConnectWS`** (AC: 1, 2, 3, 4, 5, 6, 7, 8, 9)
  - [ ] Añadir método `func (h *V1WSHandler) HandleConnectWS(w http.ResponseWriter, r *http.Request)` en `v1_ws.go`
  - [ ] **Fase 1 — Extraer API key** (antes del upgrade, ver Dev Notes):
    - `X-API-Key` header → `Authorization: ApiKey/Bearer` → `?api_key=` query param → `Sec-WebSocket-Protocol`
    - Si vacío: `writeV1Error(w, 401, "API_KEY_REQUIRED", "API key requerida")` y `return`
  - [ ] **Fase 2 — Validar API key** (antes del upgrade):
    - `key, err := h.apiKeyStore.Validate(rawKey)`
    - Si error/nil: `writeV1Error(w, 401, "INVALID_API_KEY", "API key inválida o expirada")` y `return`
  - [ ] **Fase 3 — Cargar teléfono** (antes del upgrade):
    - `phone, err := h.telefonoStore.GetByID(key.TelefonoID)`
    - Si error/nil: `writeV1Error(w, 401, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")` y `return`
    - `accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)`
  - [ ] **Fase 4 — Upgrade WebSocket**:
    - `acceptOpts := &websocket.AcceptOptions{InsecureSkipVerify: true}`
    - Si vino de `Sec-WebSocket-Protocol`: `acceptOpts.Subprotocols = []string{rawKey}`
    - `c, err := websocket.Accept(w, r, acceptOpts)` — si error, `return`
    - `defer c.CloseNow()`
  - [ ] **Fase 5 — Log apertura + StartSession + defer cleanup**:
    - Log apertura (AC9)
    - `events, unsubscribe, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)`
    - Si error: `writeWSEvent(c, "error", ...)` y `return`
    - `defer` con `unsubscribe()`, log cierre y `sessionStore.AppendEvent`
  - [ ] **Fase 6 — Event loop**:
    - `ticker := time.NewTicker(25 * time.Second)` + `defer ticker.Stop()`
    - `select` sobre `events`, `ticker.C`, `ctx.Done()` — misma lógica que `HandleWS`

- [ ] **Tarea 3: Registrar ruta** (AC: 11)
  - [ ] `backend/internal/http/routes_api.go`: dentro del bloque `if c.V1WSHandler != nil`, añadir:
    ```go
    mux.Handle("GET /api/service/v1/ws/connect", http.HandlerFunc(c.V1WSHandler.HandleConnectWS))
    ```

- [ ] **Tarea 4: Verificar build y tests** (AC: 12, 13)
  - [ ] `cd backend && go build ./...`
  - [ ] `cd backend && go test ./...`

### Review Findings

- [x] [Review][Patch] Pánico por puntero nulo si apiKeyStore o telefonoStore son nulos [backend/internal/http/handlers/v1_ws.go:146]
- [x] [Review][Patch] Asunción de subprotocolos arbitrarios como API Key en extractConnectAPIKey [backend/internal/http/handlers/v1_ws.go:230]
- [x] [Review][Patch] Falta de validación para phone.NumeroCompleto vacío [backend/internal/http/handlers/v1_ws.go:158]
- [x] [Review][Patch] Clave del evento QR incorrecta (camelCase en vez de snake_case) [backend/internal/http/handlers/v1_ws.go:201]
- [x] [Review][Patch] Mensaje del evento QR no coincide con el texto requerido por AC5 [backend/internal/http/handlers/v1_ws.go:201]
- [x] [Review][Patch] Pérdida del error de escritura en logs y defer al cerrar [backend/internal/http/handlers/v1_ws.go:186]
- [x] [Review][Defer] Exposición a Cross-Site WebSocket Hijacking (CSWSH) por InsecureSkipVerify: true [backend/internal/http/handlers/v1_ws.go:159] — deferred, pre-existing
- [x] [Review][Defer] Falta de bucle de lectura (Read) en la conexión WebSocket [backend/internal/http/handlers/v1_ws.go:189] — deferred, pre-existing
- [x] [Review][Defer] Cierre abrupto de la conexión WebSocket con CloseNow() [backend/internal/http/handlers/v1_ws.go:167] — deferred, pre-existing
- [x] [Review][Defer] Registro de PII (número de teléfono) en los logs en texto plano [backend/internal/http/handlers/v1_ws.go:171] — deferred, pre-existing
- [x] [Review][Defer] Ausencia de límite de tiempo de escritura (Write Timeout) al enviar eventos [backend/internal/http/handlers/v1_ws.go:193] — deferred, pre-existing
- [x] [Review][Defer] Falta de propagación del contexto de la petición en StartSession [backend/internal/http/handlers/v1_ws.go:173] — deferred, pre-existing
- [x] [Review][Defer] Fuga de detalles de errores internos de infraestructura en StartSession [backend/internal/http/handlers/v1_ws.go:175] — deferred, pre-existing
- [x] [Review][Defer] Falta de rate limiting y control de conexiones WebSocket concurrentes [backend/internal/http/handlers/v1_ws.go:142] — deferred, pre-existing

---

## Dev Notes

### Extracción de API key (función auxiliar interna)

```go
// extractConnectAPIKey devuelve la raw key y si provino del Sec-WebSocket-Protocol header.
func extractConnectAPIKey(r *http.Request) (rawKey string, fromProtocol bool) {
    if v := strings.TrimSpace(r.Header.Get("X-API-Key")); v != "" {
        return v, false
    }
    authHdr := strings.TrimSpace(r.Header.Get("Authorization"))
    if authHdr != "" {
        parts := strings.SplitN(authHdr, " ", 2)
        if len(parts) == 2 && (strings.EqualFold(parts[0], "ApiKey") || strings.EqualFold(parts[0], "Bearer")) {
            return strings.TrimSpace(parts[1]), false
        }
    }
    if v := r.URL.Query().Get("api_key"); v != "" {
        return v, false
    }
    if sec := r.Header.Get("Sec-WebSocket-Protocol"); sec != "" {
        for _, p := range strings.Split(sec, ",") {
            if p = strings.TrimSpace(p); p != "" {
                return p, true
            }
        }
    }
    return "", false
}
```

### `apiKeyStore` en `container.go`

La variable `apiKeyStore` ya existe en `NewContainer` (línea donde se crea `apiKeysHandler`). Solo hay que pasarla adicionalmente a `NewV1WSHandler`.

### Reutilización de helpers existentes en `v1_ws.go`

- `mapV1EventType(event, data)` — ya existe, reutilizar sin cambios
- `writeWSEvent(c, eventType, data)` — ya existe, reutilizar sin cambios
- `writeV1Error(w, status, code, message)` — ya existe en el paquete

### `apiKeyStore.Validate` devuelve `*domain.ApiKey`

Campos útiles: `key.TelefonoID`, `key.EmpresaID`. La función ya verifica activo/expirado/revocado internamente.

---

## Spec Change Log

| Fecha | Cambio |
|-------|--------|
| 2026-06-03 | Spec inicial — enfoque en solo el WS con API key; documentados endpoints REST existentes para el flujo frontend |
