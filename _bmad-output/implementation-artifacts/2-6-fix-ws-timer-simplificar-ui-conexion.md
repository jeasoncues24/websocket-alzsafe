---
title: 'Story 2.6 — Fix WS timer + simplificar UI de conexión'
type: 'bugfix+ux'
created: '2026-05-07'
status: 'ready-for-dev'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '1960bec'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como administrador del panel,
quiero que la página de conexión WS de un teléfono sea estable, sin fugas de memoria, con logs útiles cuando algo falla, y con una UI clara que no exponga detalles internos del WebSocket,
para que soporte pueda conectar/reconectar teléfonos sin confusión y el equipo de desarrollo pueda diagnosticar problemas desde los logs de Go.

## Acceptance Criteria

**AC1 — Logging del ciclo de vida del WS:**
En `ConnectCompanyPhoneWS`, cada apertura y cierre de conexión WS produce un log de Go. El log de apertura incluye `telefonoID` y `accountID`. El log de cierre incluye la razón (`ctx.Err()` si es cancelación, o el error de write si fue un fallo de escritura). Los logs usan el logger existente del paquete `whatsapp` (o `fmt.Printf` al estilo del resto de `admin.go` — **no introducir un logger nuevo**).

**AC2 — Limpieza de sesión al abandonar WS durante QR:**
Cuando la conexión WS se cierra (por cualquier causa: cliente cierra tab, proxy corta la conexión, error de red) y la sesión estaba en estado `"initializing"` o `"qr_pending"` en el `sessionStore`, el handler llama `h.manager.Delete(accountID)` en su `defer`. Esto cancela el goroutine `runSession` y libera el runtime WS de la sesión (SQLite container, contexto, canal de eventos). Las sesiones en estado `"active"` NO se interrumpen — el cliente WhatsApp permanece conectado.

**AC3 — Sin sesiones duplicadas al reconectar:**
Cuando el WS se reconecta mientras una sesión ya está activa (runtime en `s.runtimes[accountID]`), `StartSession` devuelve un canal sintético con el estado actual y lo cierra inmediatamente — no inicia un segundo goroutine ni una segunda conexión a WhatsApp. Este comportamiento ya existe en `service.go` (guard por `runtimes` + `starting`); esta story lo documenta con un comentario explícito en `ConnectCompanyPhoneWS` indicando por qué no es necesario evitar el doble llamado.

**AC4 — Keepalive ping del servidor:**
El servidor envía un frame `{event: "ping"}` al cliente WS cada 25 segundos mientras la conexión está activa. Si el write falla, el handler retorna y el `defer` ejecuta la limpieza. El cliente frontend ignora silenciosamente los eventos de tipo `"ping"`.

**AC5 — Timer de QR corregido:**
El evento `qr-{accountID}` incluye `"expires_in": 60` en su `data`. El frontend usa `data.expires_in ?? 60` (no 300) para inicializar el countdown al recibir un QR por WS. Al recibir respuesta del fallback REST, usa `response.data.expires_in ?? 60`.

**AC6 — UI simplificada, sin badges técnicos:**
Los badges "WS activo" / "WS inactivo" y los spans "Empresa #X Teléfono #Y" (con IDs numéricos) se eliminan. Se muestra el `phone.numeroCompleto` cuando esté disponible. Los botones "Iniciar / regenerar" y "Reconectar WS" se fusionan en un único botón inteligente (ver Dev Notes).

**AC7 — Comentarios de flujo en el backend:**
El handler `ConnectCompanyPhoneWS` tiene comentarios de bloque que explican: (a) autenticación, (b) inicio de sesión y su guard contra duplicados, (c) el loop de eventos como puente WS↔manager, (d) el keepalive ping, (e) el cleanup al cerrar.

**AC8 — Tests del flujo WS:**
Existen tests en `backend/internal/http/` (mismo paquete `http`) que cubren:
- Auth requerida: WS sin token recibe evento `"error"` y se cierra
- Teléfono no encontrado: WS con token válido pero ID inexistente recibe evento `"error"`
- Cleanup en QR: WS cierra mientras sessionStore reporta `qr_pending` → `manager` ya no tiene el accountID registrado al terminar

**AC9:** `cd backend && go build ./...` pasa sin errores.

**AC10:** `cd backend && go test ./...` pasa sin nuevas regresiones.

**AC11:** `cd frontend && npm run lint` pasa sin nuevos errores.

## Tasks / Subtasks

- [ ] **Tarea 1: Backend — Logging + cleanup en ConnectCompanyPhoneWS** (AC: 1, 2, 3, 7)
  - [ ] Al inicio de `ConnectCompanyPhoneWS` (justo después de validar el teléfono), agregar log de apertura:
    `fmt.Printf("[INFO] WS connect opened telefono=%d account=%s\n", phone.ID, accountID)`
  - [ ] Agregar `defer` de cleanup ANTES del loop de eventos:
    - Log de cierre con razón (ver Dev Notes para código exacto)
    - Si `sessionStore.Get(phone.NumeroCompleto).Status` es `"initializing"` o `"qr_pending"` → llamar `h.manager.Delete(accountID)`
  - [ ] Agregar comentarios de bloque explicativos en cada sección del handler (auth, start session, loop, ping, defer cleanup)

- [ ] **Tarea 2: Backend — Keepalive ping en el loop de ConnectCompanyPhoneWS** (AC: 4)
  - [ ] Crear `ticker := time.NewTicker(25 * time.Second)` + `defer ticker.Stop()` justo antes del `for { select { ... } }`
  - [ ] Agregar `case <-ticker.C:` en el select que llame `writeEvent(ctx, wsConn, outboundPayload{Event: "ping"})` y retorne si hay error

- [ ] **Tarea 3: Backend — expires_in en evento qr-** (AC: 5)
  - [ ] En `service.go`, función `runSession`, bloque `case "code":` (línea ~322):
    - Definir `const qrExpiresInSec = 60` a nivel de paquete (o como constante local)
    - Agregar `"expires_in": qrExpiresInSec` al map del emit `qr-{accountID}`

- [ ] **Tarea 4: Frontend — Manejar ping y corregir countdown** (AC: 4, 5)
  - [ ] En `ws.onmessage`: agregar `if (type === "ping") return;` antes de cualquier otro handler
  - [ ] Bloque `qr-`: `const expiresIn = typeof data.expires_in === "number" ? data.expires_in : 60` → `setCountdown(expiresIn)`
  - [ ] `startFallback`: `setCountdown(response.data.expires_in ?? 60)`

- [ ] **Tarea 5: Frontend — UI simplificada** (AC: 6)
  - [ ] Eliminar el bloque `<div className="flex flex-wrap items-center gap-2 text-sm">` con badges WS y spans de IDs numéricos
  - [ ] Mostrar `phone?.numeroCompleto` si disponible (span mono debajo del título)
  - [ ] Reemplazar el grid de dos botones por botón único inteligente (ver Dev Notes para JSX exacto)
  - [ ] `startFallback` permanece en código pero solo aparece como botón ghost "Forzar conexión REST" cuando hay un `error` visible

- [ ] **Tarea 6: Backend — Tests del flujo WS** (AC: 8)
  - [ ] Crear o ampliar `backend/internal/http/admin_ws_test.go` con los 3 tests descritos en Dev Notes
  - [ ] Los tests usan `httptest.NewServer`, `websocket.Dial` de `github.com/coder/websocket` y `storage.NewSessionStore()`

- [ ] **Tarea 7: Verificar build, tests y lint** (AC: 9, 10, 11)
  - [ ] `cd backend && go build ./...`
  - [ ] `cd backend && go test ./...`
  - [ ] `cd frontend && npm run lint`

## Dev Notes

### 🔍 Arquitectura actual — qué falla y por qué

**El flujo correcto (cuando funciona):**
```
Browser abre WS → ConnectCompanyPhoneWS
  → valida JWT admin
  → llama StartSession(manager, accountID)
      → service.StartSession crea sessionRuntime{ctx, cancel, client, events chan}
      → lanza goroutine runSession(accountID, runtime)
          → ConnectContext a WhatsApp
          → emite eventos al runtime.events chan
  → ConnectCompanyPhoneWS drena runtime.events → escribe al WS cliente
  → WhatsApp emite QR code → WS recibe {event:"qr-...", data:{qrString:...}}
  → Usuario escanea QR → WhatsApp emite success → WS recibe {event:"active-...", isActive:true}
  → runtime.events se cierra → loop WS termina → handler retorna
```

**Bug 1 — Fuga de goroutine cuando WS se abandona en QR:**
Si el browser cierra el tab mientras la sesión está en QR (antes de escanear):
- El `ctx` de `r.Context()` se cancela → el loop del handler retorna
- PERO `runSession` usa `runtime.ctx` = `context.WithCancel(context.Background())` → **no se cancela**
- El goroutine `runSession` queda vivo: tiene un SQLite abierto, el cliente whatsmeow intentando conectar, y el canal `runtime.events` bloqueado porque nadie lo lee
- El canal tiene buffer de 8 → eventualmente se llena → `runSession` se bloquea para siempre hasta que whatsmeow haga timeout por sí solo (puede tardar minutos)

**Fix:** En el `defer` del handler, si la sesión estaba en estado QR → `h.manager.Delete(accountID)`, que internamente llama `service.StopSession(accountID)` → cancela `runtime.ctx` → `runSession` sale por `case <-runtime.ctx.Done()` → limpia todo.

**Bug 2 — Timer de QR desincronizado:**
El countdown en frontend arranca en 300 segundos, pero el QR de whatsmeow expira en ~60 segundos. El handler del servidor emite `"qr_timeout"` y `requiresNewQR: true` cuando expira el QR real. La UI muestra "QR válido por 4:50" mientras el QR ya no funciona.

**Fix:** El servidor incluye `"expires_in": 60` en el evento `qr-`. El frontend lo usa para el countdown.

**Bug 3 — WS sin keepalive:**
Ningún lado envía frames de keepalive. Proxies con idle timeout de 30-60s cortan la conexión silenciosamente. El frontend detecta el cierre vía `ws.onclose` y debería reconectar, pero sin logging en el servidor no se puede diagnosticar.

**Sobre sesiones duplicadas (ya está correcto, solo falta documentarlo):**
`service.StartSession` tiene un guard doble:
1. Si `s.runtimes[accountID]` existe: devuelve canal sintético con estado actual → no crea segundo runtime
2. Si `s.starting[accountID]` es true: devuelve canal sintético → no crea segunda goroutine

Por lo tanto, si el WS reconecta mientras la sesión está activa, `StartSession` devuelve inmediatamente el estado actual (en un canal sintético que se cierra solo). El nuevo WS recibe el estado una vez y no recibe más eventos futuros (porque el `runtime.events` original ya tiene un consumer o está cerrado). Esto es correcto para sesiones activas — el nuevo WS solo necesita confirmar que la sesión existe.

### 📐 Código exacto: ConnectCompanyPhoneWS reescrito

```go
func (h *AdminHandler) ConnectCompanyPhoneWS(w http.ResponseWriter, r *http.Request) {
    // — Upgrade a WebSocket —
    wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
    if err != nil {
        return
    }
    defer wsConn.CloseNow()

    // — Validar configuración JWT —
    if h.jwtCfg == nil {
        _ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "configuracion JWT no disponible"}})
        return
    }

    // — Autenticar token admin —
    token := r.URL.Query().Get("token")
    if token == "" {
        if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
            token = strings.TrimPrefix(authHeader, "Bearer ")
        }
    }
    if token == "" {
        _ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "Token requerido"}})
        return
    }
    claims, err := middleware.NewAuthMiddleware(h.jwtCfg, nil).ValidateToken(token)
    if err != nil {
        _ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "Token inválido"}})
        return
    }
    _ = claims

    // — Resolver teléfono —
    telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
    if err != nil || telefonoID <= 0 {
        _ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "ID de teléfono inválido"}})
        return
    }
    phone, err := h.telefonoStore.GetByID(telefonoID)
    if err != nil || phone == nil {
        _ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "teléfono no encontrado"}})
        return
    }
    accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)

    fmt.Printf("[INFO] WS connect opened telefono=%d account=%s\n", phone.ID, accountID)

    // — Cleanup al cerrar el WS (por cualquier causa) —
    // Si la sesión estaba en QR o initializing cuando el WS se cerró, cancelamos
    // el runtime de la sesión para liberar el goroutine, SQLite y canal de eventos.
    // Si la sesión ya estaba activa, la dejamos correr — el cliente WhatsApp
    // permanece conectado aunque el WS del panel se haya cerrado.
    defer func() {
        reason := ctx.Err()
        fmt.Printf("[INFO] WS connect closed telefono=%d account=%s reason=%v\n", phone.ID, accountID, reason)
        if h.sessionStore != nil && h.manager != nil {
            if state, ok := h.sessionStore.Get(phone.NumeroCompleto); ok {
                if state.Status == "initializing" || state.Status == "qr_pending" {
                    // Sesión abandonada durante QR — cancelar runtime para evitar fuga de goroutine
                    h.manager.Delete(accountID)
                }
            }
        }
    }()

    // — Enviar estado inicial del teléfono —
    _ = writeEvent(r.Context(), wsConn, outboundPayload{
        Event: "phone-info",
        Data: map[string]any{
            "telefono_id":    phone.ID,
            "numeroCompleto": phone.NumeroCompleto,
            "status":         phone.Status,
            "qr_string":      phone.QRString,
            "lastConnected":  phone.LastConnected,
        },
    })

    // — Iniciar o unirse a sesión existente —
    // StartSession es idempotente: si ya existe un runtime para este accountID,
    // devuelve un canal sintético con el estado actual y no crea una segunda goroutine.
    // Esto evita duplicar sesiones cuando el WS reconecta sobre una sesión viva.
    events, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
    if err != nil {
        _ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "error al iniciar conexión: " + err.Error()}})
        return
    }

    // — Keepalive: enviar ping cada 25s para evitar que proxies corten la conexión idle —
    ticker := time.NewTicker(25 * time.Second)
    defer ticker.Stop()

    ctx := r.Context()

    // — Loop principal: este WS es un puente entre el manager de la sesión y el cliente —
    // Los eventos del runtime WS (QR, connected, disconnected) se reenvían directamente.
    // El loop termina cuando: el canal de eventos se cierra (sesión terminó),
    // el cliente WS se desconecta (ctx.Done), o falla un write.
    for {
        select {
        case event, ok := <-events:
            if !ok {
                // Canal cerrado — la sesión terminó (conectó, desconectó, o timeout QR)
                return
            }
            if err := writeEvent(ctx, wsConn, outboundPayload{Event: event.Event, Data: event.Data}); err != nil {
                // Write falló — el cliente WS probablemente se desconectó
                return
            }
        case <-ticker.C:
            // Keepalive ping — mantiene el WS activo a través de proxies con idle timeout
            if err := writeEvent(ctx, wsConn, outboundPayload{Event: "ping"}); err != nil {
                return
            }
        case <-ctx.Done():
            // El cliente WS cerró la conexión (cierre normal del browser, timeout de red, etc.)
            return
        }
    }
}
```

> ⚠️ NOTA: `ctx` se declara después del `defer` que lo usa. El `defer` captura `ctx` por referencia (es variable en el closure), así que funciona correctamente — cuando el defer ejecuta, `ctx` ya está asignado.

> ⚠️ NOTA 2: `outboundPayload{Event: "ping"}` — el campo `Data` queda en nil. El JSON serializado será `{"event":"ping","data":null}`. Verificar que el frontend no falle con data null; si falla, usar `Data: map[string]any{}`.

### 📐 Código exacto: expires_in en service.go

```go
// Validez aproximada de cada código QR emitido por whatsmeow.
// whatsmeow genera un nuevo código ~cada 20s; el cliente usa este valor
// para el countdown visual. El timeout real lo maneja el servidor.
const qrExpiresInSec = 60

// En runSession, case "code":
case "code":
    if s.sessionStore != nil {
        s.sessionStore.SetQRPending(accountID, evt.Code)
        s.sessionStore.AppendEvent(accountID, "qr_generated", "")
    }
    s.syncTelefonoQR(accountID, evt.Code)
    emit("qr-"+accountID, map[string]any{
        "message":    "Escanee el codigo QR para iniciar sesion.",
        "qrString":   evt.Code,
        "expires_in": qrExpiresInSec,  // ← NUEVO
    })
```

### 📐 Botón único en frontend (connect/page.tsx)

```tsx
{/* Reemplaza el grid de dos botones */}
<div className="flex flex-col gap-2">
  {status !== "active" && (
    wsConnected ? (
      // WS abierto: ofrecer cancelar (cierra WS y vuelve a la lista)
      <Button variant="outline" onClick={() => {
        closeSocket()
        router.push(`/empresas/${empresaId}/telefonos`)
      }}>
        Cancelar
      </Button>
    ) : (
      // WS cerrado: reconectar
      <Button onClick={openSocket} disabled={starting}>
        {starting
          ? <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          : <RefreshCw className="mr-2 h-4 w-4" />}
        Reconectar
      </Button>
    )
  )}
  {/* Fallback REST — solo visible cuando hay un error, como último recurso */}
  {error && (
    <Button variant="ghost" size="sm" onClick={startFallback} disabled={starting}>
      Forzar conexión REST
    </Button>
  )}
</div>
```

### 📐 Tests exactos a escribir

Crear `backend/internal/http/admin_ws_test.go` (package `http`). Usar el helper `newAdminPhoneTestDB` y `insertAdminPhone` que ya existen en `admin_test.go`.

```go
package http

import (
    "context"
    "encoding/json"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/coder/websocket"
    "github.com/golang-jwt/jwt/v5"

    "wsapi/internal/config"
    "wsapi/internal/storage"
    "wsapi/internal/whatsapp"
)

// helper: genera token admin válido para tests
func makeAdminToken(t *testing.T, secret string) string {
    t.Helper()
    tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": 1,
        "rol":     "super_admin",
        "exp":     time.Now().Add(time.Hour).Unix(),
    }).SignedString([]byte(secret))
    if err != nil {
        t.Fatalf("makeAdminToken: %v", err)
    }
    return tok
}

// helper: leer un evento WS con timeout
func readWSEvent(t *testing.T, ctx context.Context, conn *websocket.Conn) map[string]any {
    t.Helper()
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    _, data, err := conn.Read(ctx)
    if err != nil {
        t.Fatalf("readWSEvent: %v", err)
    }
    var evt map[string]any
    if err := json.Unmarshal(data, &evt); err != nil {
        t.Fatalf("readWSEvent unmarshal: %v", err)
    }
    return evt
}

// TestConnectCompanyPhoneWS_AuthRequired verifica que sin token el servidor
// envía un evento "error" y cierra la conexión.
func TestConnectCompanyPhoneWS_AuthRequired(t *testing.T) {
    db := newAdminPhoneTestDB(t)
    insertAdminPhone(t, db, 1, "+51", "999888777", "+51999888777")

    h := &AdminHandler{
        telefonoStore: storage.NewTelefonoStore(db),
        sessionStore:  storage.NewSessionStore(),
        manager:       whatsapp.NewManager(),
        jwtCfg:        &config.JWTConfig{Secret: "test-secret"},
    }

    srv := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        // Inyectar contexto de acceso admin para que el handler pase auth de panel
        h.ConnectCompanyPhoneWS(w, r)
    }))
    defer srv.Close()

    ctx := context.Background()
    wsURL := "ws" + srv.URL[4:] + "/api/admin/telefonos/1/connect/ws" // sin ?token=

    conn, _, err := websocket.Dial(ctx, wsURL, nil)
    if err != nil {
        t.Fatalf("dial: %v", err)
    }
    defer conn.CloseNow()

    evt := readWSEvent(t, ctx, conn)
    if evt["event"] != "error" {
        t.Errorf("expected event=error, got %v", evt["event"])
    }
}

// TestConnectCompanyPhoneWS_TelefonoNotFound verifica que con token válido
// pero teléfono inexistente el servidor envía "error".
func TestConnectCompanyPhoneWS_TelefonoNotFound(t *testing.T) {
    db := newAdminPhoneTestDB(t)
    // No insertamos teléfono con ID 999

    secret := "test-secret"
    h := &AdminHandler{
        telefonoStore: storage.NewTelefonoStore(db),
        sessionStore:  storage.NewSessionStore(),
        manager:       whatsapp.NewManager(),
        jwtCfg:        &config.JWTConfig{Secret: secret},
    }

    srv := httptest.NewServer(stdhttp.HandlerFunc(h.ConnectCompanyPhoneWS))
    defer srv.Close()

    token := makeAdminToken(t, secret)
    wsURL := "ws" + srv.URL[4:] + "/api/admin/telefonos/999/connect/ws?token=" + token

    ctx := context.Background()
    conn, _, err := websocket.Dial(ctx, wsURL, nil)
    if err != nil {
        t.Fatalf("dial: %v", err)
    }
    defer conn.CloseNow()

    evt := readWSEvent(t, ctx, conn)
    if evt["event"] != "error" {
        t.Errorf("expected event=error, got %v", evt["event"])
    }
}

// TestConnectCompanyPhoneWS_QRSessionCleanedOnDisconnect verifica que cuando
// el WS se cierra mientras la sesión está en qr_pending, el manager ya no
// tiene registrado el accountID (el goroutine de sesión fue cancelado).
func TestConnectCompanyPhoneWS_QRSessionCleanedOnDisconnect(t *testing.T) {
    db := newAdminPhoneTestDB(t)
    phoneID := insertAdminPhone(t, db, 1, "+51", "999888777", "+51999888777")
    _ = phoneID

    secret := "test-secret"
    sessionStore := storage.NewSessionStore()
    manager := whatsapp.NewManager()

    // Simular que la sesión está en qr_pending (como si ya hubiera arrancado)
    sessionStore.SetQRPending("+51999888777", "FAKE_QR_CODE")

    h := &AdminHandler{
        telefonoStore: storage.NewTelefonoStore(db),
        sessionStore:  sessionStore,
        manager:       manager,
        jwtCfg:        &config.JWTConfig{Secret: secret},
    }

    srv := httptest.NewServer(stdhttp.HandlerFunc(h.ConnectCompanyPhoneWS))
    defer srv.Close()

    token := makeAdminToken(t, secret)
    wsURL := "ws" + srv.URL[4:] + "/api/admin/telefonos/1/connect/ws?token=" + token

    ctx := context.Background()
    conn, _, err := websocket.Dial(ctx, wsURL, nil)
    if err != nil {
        t.Fatalf("dial: %v", err)
    }

    // Leer el phone-info inicial (o cualquier primer evento)
    readWSEvent(t, ctx, conn)

    // Cerrar WS desde el cliente — simula cierre de tab
    conn.Close(websocket.StatusNormalClosure, "test done")

    // Dar tiempo al defer del handler para ejecutarse
    time.Sleep(100 * time.Millisecond)

    // El manager NO debe tener el accountID registrado (fue limpiado por Delete)
    accountID := whatsapp.NormalizeAccountID("+51999888777")
    if manager.Exists(accountID) {
        t.Errorf("expected manager to have cleaned up account %s after QR disconnect", accountID)
    }
}
```

> **Nota sobre `insertAdminPhone`**: esta función ya existe en `admin_test.go`. Verifica su firma exacta antes de usarla; si recibe parámetros distintos, ajustar el test.

> **Nota sobre el mock de JWT**: el `middleware.NewAuthMiddleware` valida el token con el secret del `jwtCfg`. Asegurarse de que los claims del token generado en el test incluyan los campos que `ValidateToken` espera (ver `internal/http/middleware/auth.go`).

### 📐 Estructura de archivos a modificar

| Archivo | Cambio | Tipo |
|---------|--------|------|
| `backend/internal/http/admin.go` | Reescribir `ConnectCompanyPhoneWS` con logging, defer cleanup, keepalive, comentarios | MODIFICAR |
| `backend/internal/whatsapp/service.go` | Constante `qrExpiresInSec`, agregar `expires_in` en emit `qr-` | MODIFICAR |
| `backend/internal/http/admin_ws_test.go` | Crear con 3 tests WS | CREAR |
| `frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/connect/page.tsx` | Eliminar badges, botón único, countdown fix, manejo de ping | MODIFICAR |

### ⚠️ Lo que NO cambia esta story

- `service.go` — solo se agrega la constante y el campo `expires_in`. Toda la lógica de `runSession`, `StartSession`, guards de duplicados, etc. permanece intacta
- `manager.go` — no se agrega ningún método nuevo; se usa `manager.Delete` que ya existe
- `V1WSHandler` (handlers/v1_ws.go) — es el WS para empresas cliente, no el admin; completamente separado
- El endpoint REST `StartCompanyPhoneConnection` (POST sin WS)
- La página de sesiones admin (`/sessions`) — componente diferente
- La lógica del bootstrap (story 2-5)

### 🧠 Cosas que el dev debe verificar antes de implementar

1. **Firma exacta de `ValidateToken`**: el test usa `middleware.NewAuthMiddleware(h.jwtCfg, nil).ValidateToken(token)` — verificar que `ValidateToken` existe y acepta un string. Si la firma es diferente, ajustar el test.

2. **Claims del JWT en los tests**: `makeAdminToken` genera claims básicos. `ValidateToken` puede requerir claims específicos (campo `rol`, `user_id`, etc.). Ver `internal/http/middleware/auth.go` para la estructura exacta de claims esperada.

3. **`outboundPayload{Event: "ping"}` con Data nil**: verificar si `writeEvent` serializa correctamente `Data: nil`. Si produce JSON inválido, usar `Data: map[string]any{}`.

4. **Race condition en el test de cleanup**: el `time.Sleep(100ms)` es un hack aceptable para tests de integración, pero si el handler tarda más en ejecutar el defer (ej. DB lenta), el test fallará. Alternativa más robusta: polling con timeout corto.

5. **`insertAdminPhone` en `admin_test.go`**: verificar la firma real; la del ejemplo puede diferir.

### Learnings de stories anteriores

- Package en `handlers/` es `package http` (NO `package handlers`)
- `writeEvent` y `writeHandlerJSON` son los helpers estándar
- Nil-guard para todos los stores
- `outboundPayload` tiene `Event string` y `Data map[string]any` (ver handlers.go:48)
- `fmt.Printf` es el mecanismo de logging usado en admin.go y startup_bootstrap.go — no introducir zerolog aquí
- `storage.NewSessionStore()` crea un SessionStore in-memory listo para tests
- El test helper `newAdminPhoneTestDB` ya configura SQLite en memoria con schema

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

### File List

### Change Log
