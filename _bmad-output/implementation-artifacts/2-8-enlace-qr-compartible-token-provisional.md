---
title: 'Story 2.8 — Enlace QR compartible: token provisional'
type: 'feature'
created: '2026-05-08'
status: 'review'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '3c41157'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

## Story

Como administrador del panel de wsapi,
quiero generar un enlace compartible de corta duración para que un operador externo escanee el QR de WhatsApp desde su navegador,
para que no sea necesario darle acceso al panel administrativo ni a la API de empresa para conectar un teléfono.

## Acceptance Criteria

**AC1 — Endpoint admin para generar token provisional:**
`POST /api/admin/telefonos/{id}/qr-link` (protegido por adminStack) responde:
```json
{"ok": true, "token": "eyJ...", "phone_id": 42, "expires_in": 600}
```
El token es un JWT firmado con el mismo secret del sistema, dura 10 minutos, y contiene claims `sub=empresa_id`, `phone_id`, `scope="qr_link"`.

**AC2 — Middleware rechaza tokens QR link en endpoints REST:**
`EmpresaAuthMiddleware.RequireEmpresaAuth()` rechaza con HTTP 401 cualquier JWT que tenga `scope == "qr_link"` (verificado antes del check de token_version).

**AC3 — EmpresaJWTClaims extendido con Scope y PhoneID:**
`domain.EmpresaJWTClaims` incluye `Scope string` y `PhoneID int64` (zero values = token regular de empresa). `extractEmpresaJWTClaims` los extrae si están presentes en los claims del JWT.

**AC4 — V1WSHandler: auto-subscribe para tokens QR link:**
Cuando el token tiene `scope == "qr_link"` y `phone_id > 0`, el handler:
- Salta la espera del mensaje "subscribe" del cliente
- Llama `telefonoStore.GetByID(claims.PhoneID)` para obtener el teléfono
- Si el teléfono no existe → envía `{"type":"error","data":{"message":"teléfono no encontrado"}}` y cierra
- Procede directamente al bridge: logs de apertura, defer cleanup, `StartSession`, loop de eventos

**AC5 — V1WSHandler: path regular de empresa JWT sin cambios:**
El path de empresa JWT (sin scope, ya implementado en story 2-7) sigue funcionando exactamente igual: espera subscribe, valida BelongsToEmpresa, etc.

**AC6 — Página pública /qr sin login:**
`frontend/app/qr/page.tsx` accesible sin autenticación. Lee `?token=XXX` de los query params. Conecta al WS en `/api/service/v1/ws?token=TOKEN`. No envía mensaje de subscribe (el servidor auto-suscribe). Muestra:
- Estado "Conectando..." mientras establece WS
- QR en tiempo real con countdown cuando llega evento `type="qr"`
- Mensaje de éxito cuando llega `type="connected"` con `data.isActive=true`
- Mensaje de error cuando llega `type="error"` o `type="connected"` con `data.isActive=false`
- Fallback: si token vacío → "Enlace inválido o expirado"

**AC7 — Botón "Compartir enlace" en la connect page del admin:**
En `/empresas/[empresaId]/telefonos/[telefonoId]/connect`, un botón "Compartir enlace" que:
- Llama a `POST /api/admin/telefonos/{telefonoId}/qr-link`
- Construye la URL: `${window.location.origin}/qr?token=${data.token}`
- Muestra la URL con un botón "Copiar" (usando `navigator.clipboard.writeText`)
- Muestra mensaje de confirmación "Enlace copiado" por 2 segundos

**AC8 — Logs del handler:**
El handler imprime:
- `[INFO] V1 WS QR-link opened phone={phone_id} account={accountID}`
- `[INFO] V1 WS QR-link closed phone={phone_id} account={accountID} reason={ctx.Err()}`

**AC9:** `cd backend && go build ./...` pasa sin errores.

**AC10:** `cd backend && go test ./...` pasa sin nuevas regresiones.

**AC11:** `cd frontend && npm run lint` pasa sin nuevos errores (los 3 pre-existentes en `api.ts` son aceptables).

## Tasks / Subtasks

- [x] **Tarea 1: Extender domain y auth** (AC: 3, 1)
  - [x] Agregar `Scope string` y `PhoneID int64` a `EmpresaJWTClaims` en `domain/empresa_token.go`
  - [x] En `extractEmpresaJWTClaims` (`auth/empresa_jwt.go`): extraer `scope` y `phone_id` si presentes
  - [x] Agregar `GenerateQRLinkToken(empresaID, phoneID int64, secret string) (string, error)` en `auth/empresa_jwt.go`

- [x] **Tarea 2: Middleware — rechazar QR link tokens** (AC: 2)
  - [x] En `middleware/empresa_auth.go`, después de `ParseEmpresaJWT`, agregar: `if claims.Scope == "qr_link" { rechazar }` (antes del check de token_version)

- [x] **Tarea 3: V1WSHandler — branch para QR link** (AC: 4, 5, 8)
  - [x] Refactorizar `HandleWS` para separar el path de QR link del path de empresa JWT
  - [x] Path QR link: usar `claims.PhoneID`, saltar subscribe, llamar `GetByID`, log QR-link, defer, StartSession, bridge
  - [x] Path empresa JWT: comportamiento existente de story 2-7 sin cambios

- [x] **Tarea 4: Nuevo endpoint admin GenerateQRLink** (AC: 1)
  - [x] Agregar `jwtCfg *config.JWTConfig` a `AdminSessionsHandler` struct y `NewAdminSessionsHandler`
  - [x] Implementar `GenerateQRLink(w, r)` en `handlers/admin_sessions.go`
  - [x] Registrar `POST /api/admin/telefonos/{id}/qr-link` en `routes_admin.go` (bajo bloque AdminSessionsHandler)
  - [x] Actualizar `NewAdminSessionsHandler(...)` en `container.go` para pasar `jwtCfg`

- [x] **Tarea 5: Página pública /qr** (AC: 6)
  - [x] Crear `frontend/app/qr/page.tsx` con patrón `Suspense` + `useSearchParams`
  - [x] Conectar WS con `buildAdminWsUrl('/api/service/v1/ws', token)`
  - [x] Manejar eventos: `qr` (mostrar QR + countdown), `connected` (éxito/error según `isActive`), `ping` (ignorar), `error` (mostrar mensaje)
  - [x] Sin login ni localStorage — solo el token de URL

- [x] **Tarea 6: Botón compartir en connect page** (AC: 7)
  - [x] Agregar `generateQRLink(telefonoId: number)` a `frontend/lib/api.ts`
  - [x] Agregar botón "Compartir enlace" en `connect/page.tsx` con estado de copia

- [x] **Tarea 7: Build, tests y lint** (AC: 9, 10, 11)
  - [x] `cd backend && go build ./...`
  - [x] `cd backend && go test ./...`
  - [x] `cd frontend && npm run lint`

## Dev Notes

### 🏗️ Arquitectura del token provisional

El token QR link es un JWT estándar HS256 firmado con el mismo `jwtCfg.Secret` del sistema. Claims:
```json
{
  "sub":      42,           // empresa_id (int64, como en empresa JWT)
  "phone_id": 7,            // telefono_id restringido
  "scope":    "qr_link",   // distingue del empresa JWT normal
  "iat":      1234567890,
  "exp":      1234568490    // iat + 600 segundos (10 minutos)
}
```

No tiene `ver` (token_version) porque no se valida contra DB — su corta duración (10 min) es la protección suficiente.

### 📐 Tarea 1 — Código exacto: domain/empresa_token.go

```go
type EmpresaJWTClaims struct {
    EmpresaID     int64    `json:"sub"`
    TokenVersion  int      `json:"ver"`
    EmpresaRUC    string   `json:"ruc"`
    EmpresaNombre string   `json:"nombre"`
    Permissions   []string `json:"permissions,omitempty"`
    // Campos opcionales — zero values = token regular de empresa
    Scope   string `json:"scope,omitempty"`    // "qr_link" para tokens provisionales
    PhoneID int64  `json:"phone_id,omitempty"` // teléfono restringido
}
```

### 📐 Tarea 1 — Código exacto: auth/empresa_jwt.go

Añadir al final de `extractEmpresaJWTClaims`:
```go
// Campos opcionales para tokens provisionales
if v, ok := m["scope"].(string); ok {
    claims.Scope = v
}
if v, ok := m["phone_id"].(float64); ok {
    claims.PhoneID = int64(v)
}
```

Nueva función `GenerateQRLinkToken` (al final del archivo):
```go
const qrLinkTokenExpirySeconds = 600 // 10 minutos

func GenerateQRLinkToken(empresaID, phoneID int64, secret string) (string, error) {
    now := time.Now()
    claims := jwt.MapClaims{
        "sub":      empresaID,
        "phone_id": phoneID,
        "scope":    "qr_link",
        "iat":      now.Unix(),
        "exp":      now.Add(qrLinkTokenExpirySeconds * time.Second).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString([]byte(secret))
    if err != nil {
        return "", fmt.Errorf("error al firmar QR link token: %w", err)
    }
    return signed, nil
}
```

### 📐 Tarea 2 — Código exacto: middleware/empresa_auth.go

Dentro de `RequireEmpresaAuth()`, después de `claims, err := auth.ParseEmpresaJWT(...)`:
```go
claims, err := auth.ParseEmpresaJWT(parts[1], m.jwtConfig.Secret)
if err != nil {
    writeEmpresaError(w, http.StatusUnauthorized, "INVALID_TOKEN", "JWT inválido o expirado")
    return
}

// Rechazar tokens de corta duración (QR link) en endpoints REST
if claims.Scope == "qr_link" {
    writeEmpresaError(w, http.StatusUnauthorized, "INVALID_TOKEN", "JWT inválido o expirado")
    return
}

// Verificar token_version contra DB (revocación).
empresa, err := m.empresaStore.GetByID(claims.EmpresaID)
// ... resto sin cambios ...
```

### 📐 Tarea 3 — Código exacto: handlers/v1_ws.go

`HandleWS` refactorizado (mantiene `writeWSEvent` y `mapV1EventType` sin cambios):

```go
func (h *V1WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    if token == "" {
        authHeader := r.Header.Get("Authorization")
        if strings.HasPrefix(authHeader, "Bearer ") {
            token = strings.TrimPrefix(authHeader, "Bearer ")
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

    c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
    if err != nil {
        return
    }
    defer c.CloseNow()

    ctx := r.Context()

    // — Resolver phone según tipo de token —
    var phoneID int64

    if claims.Scope == "qr_link" {
        // Token provisional: auto-suscribir al teléfono del token
        if claims.PhoneID <= 0 {
            _ = writeWSEvent(c, "error", map[string]string{"message": "token QR inválido: phone_id ausente"})
            return
        }
        phoneID = claims.PhoneID
    } else {
        // Empresa JWT regular: esperar mensaje subscribe con phone_id
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
        belongs, _ := h.telefonoStore.BelongsToEmpresa(sub.PhoneID, claims.EmpresaID)
        if !belongs {
            _ = writeWSEvent(c, "error", map[string]string{"message": "forbidden"})
            return
        }
        phoneID = sub.PhoneID
    }

    // — Cargar teléfono (path común) —
    phone, err := h.telefonoStore.GetByID(phoneID)
    if err != nil || phone == nil {
        _ = writeWSEvent(c, "error", map[string]string{"message": "teléfono no encontrado"})
        return
    }
    accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)

    if claims.Scope == "qr_link" {
        fmt.Printf("[INFO] V1 WS QR-link opened phone=%d account=%s\n", phone.ID, accountID)
    } else {
        fmt.Printf("[INFO] V1 WS opened empresa=%d phone=%d account=%s\n", claims.EmpresaID, phone.ID, accountID)
    }

    defer func() {
        if claims.Scope == "qr_link" {
            fmt.Printf("[INFO] V1 WS QR-link closed phone=%d account=%s reason=%v\n", phone.ID, accountID, ctx.Err())
        } else {
            fmt.Printf("[INFO] V1 WS closed empresa=%d account=%s reason=%v\n", claims.EmpresaID, accountID, ctx.Err())
        }
        if h.sessionStore != nil && h.manager != nil {
            if state, ok := h.sessionStore.Get(phone.NumeroCompleto); ok {
                if state.Status == "initializing" || state.Status == "qr_pending" {
                    h.manager.Delete(accountID)
                }
            }
        }
    }()

    events, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
    if err != nil {
        _ = writeWSEvent(c, "error", map[string]string{"message": "error al iniciar sesión: " + err.Error()})
        return
    }

    ticker := time.NewTicker(25 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case event, ok := <-events:
            if !ok {
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
```

### 📐 Tarea 4 — Código exacto: handlers/admin_sessions.go

Cambios en struct y constructor:
```go
type AdminSessionsHandler struct {
    empresaStore  domain.EmpresaStoreInterface
    telefonoStore *storage.TelefonoStore
    manager       *whatsapp.Manager
    sessionStore  *storage.SessionStore
    jwtCfg        *config.JWTConfig   // ← nuevo
}

func NewAdminSessionsHandler(
    empresaStore domain.EmpresaStoreInterface,
    telefonoStore *storage.TelefonoStore,
    manager *whatsapp.Manager,
    sessionStore *storage.SessionStore,
    jwtCfg *config.JWTConfig,           // ← nuevo
) *AdminSessionsHandler {
    return &AdminSessionsHandler{
        empresaStore:  empresaStore,
        telefonoStore: telefonoStore,
        manager:       manager,
        sessionStore:  sessionStore,
        jwtCfg:        jwtCfg,          // ← nuevo
    }
}
```

Nueva función `GenerateQRLink`:
```go
// GenerateQRLink genera un token provisional de corta duración (10 min) para
// compartir el enlace QR de un teléfono sin requerir acceso al panel.
// POST /api/admin/telefonos/{id}/qr-link
func (h *AdminSessionsHandler) GenerateQRLink(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    telefonoID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || telefonoID <= 0 {
        writeAdminError(w, http.StatusBadRequest, "telefono_id inválido")
        return
    }

    if h.telefonoStore == nil || h.jwtCfg == nil {
        writeAdminError(w, http.StatusInternalServerError, "servicio no disponible")
        return
    }

    phone, err := h.telefonoStore.GetByID(telefonoID)
    if err != nil || phone == nil {
        writeAdminError(w, http.StatusNotFound, "Teléfono no encontrado")
        return
    }

    token, err := auth.GenerateQRLinkToken(phone.EmpresaID, phone.ID, h.jwtCfg.Secret)
    if err != nil {
        writeAdminError(w, http.StatusInternalServerError, "Error generando token")
        return
    }

    writeAdminJSON(w, http.StatusOK, map[string]interface{}{
        "ok":         true,
        "token":      token,
        "phone_id":   phone.ID,
        "expires_in": 600,
    })
}
```

**Imports a agregar en admin_sessions.go:**
```go
"strconv"
"wsapi/internal/auth"
"wsapi/internal/config"
```

### 📐 Tarea 4 — routes_admin.go

Dentro del bloque `if c.AdminSessionsHandler != nil`:
```go
if c.AdminSessionsHandler != nil {
    mux.Handle("GET /api/admin/sesiones", adminStack(http.HandlerFunc(c.AdminSessionsHandler.GetSessions)))
    mux.Handle("POST /api/admin/sesiones", adminStack(http.HandlerFunc(c.AdminSessionsHandler.PostSession)))
    mux.Handle("POST /api/admin/telefonos/{id}/qr-link", adminStack(http.HandlerFunc(c.AdminSessionsHandler.GenerateQRLink))) // ← nuevo
}
```

⚠️ **RUTA POTENCIALMENTE CONFLICTIVA:** La ruta `POST /api/admin/telefonos/{id}/qr-link` bajo `AdminSessionsHandler` coexiste con `POST /api/admin/telefonos/{id}/connect` bajo `AdminHandler`. El mux de `net/http` en Go 1.22+ maneja rutas con prefijo diferente correctamente — `connect` y `qr-link` son sufijos distintos. ✅

### 📐 Tarea 4 — container.go

```go
// Cambiar:
adminSessionsHandler := handlers.NewAdminSessionsHandler(empresaStore, telefonoStore, manager, sessionStore)
// Por:
adminSessionsHandler := handlers.NewAdminSessionsHandler(empresaStore, telefonoStore, manager, sessionStore, jwtCfg)
```

### 📐 Tarea 5 — frontend/app/qr/page.tsx

```tsx
"use client"

import { Suspense, useCallback, useEffect, useRef, useState } from "react"
import { useSearchParams } from "next/navigation"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { QRRender } from "@/components/qr/qr-render"
import { buildAdminWsUrl } from "@/lib/api"
import { CheckCircle2, Loader2, WifiOff } from "lucide-react"

function QRPageContent() {
  const searchParams = useSearchParams()
  const token = searchParams.get("token") ?? ""

  const wsRef = useRef<WebSocket | null>(null)
  const [qrString, setQrString] = useState("")
  const [countdown, setCountdown] = useState(60)
  const [status, setStatus] = useState<"connecting" | "qr" | "connected" | "error" | "closed">("connecting")
  const [errorMsg, setErrorMsg] = useState("")

  const connect = useCallback(() => {
    if (!token) {
      setStatus("error")
      setErrorMsg("Enlace inválido o expirado")
      return
    }

    const ws = new WebSocket(buildAdminWsUrl("/api/service/v1/ws", token))
    wsRef.current = ws

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data as string) as { type: string; data?: Record<string, unknown> }
        if (msg.type === "ping") return
        if (msg.type === "qr") {
          const d = msg.data ?? {}
          setQrString(String(d.qrString ?? d.qr_string ?? ""))
          setCountdown(typeof d.expires_in === "number" ? d.expires_in : 60)
          setStatus("qr")
          return
        }
        if (msg.type === "connected") {
          const d = msg.data ?? {}
          if (d.isActive) {
            setStatus("connected")
          } else {
            setStatus("error")
            setErrorMsg(String(d.message ?? d.reason ?? "Sesión cerrada"))
          }
          return
        }
        if (msg.type === "error") {
          setStatus("error")
          setErrorMsg(String(msg.data?.message ?? "Error de conexión"))
        }
      } catch {
        // ignorar mensajes malformados
      }
    }

    ws.onerror = () => {
      setStatus("error")
      setErrorMsg("Error de conexión WebSocket")
    }

    ws.onclose = () => {
      wsRef.current = null
      setStatus((prev) => (prev === "connected" ? "connected" : "closed"))
    }
  }, [token])

  useEffect(() => {
    connect()
    return () => { wsRef.current?.close() }
  }, [connect])

  useEffect(() => {
    if (status !== "qr" || !qrString || countdown <= 0) return
    const t = setTimeout(() => setCountdown((n) => n - 1), 1000)
    return () => clearTimeout(t)
  }, [countdown, qrString, status])

  const formatTime = (s: number) => `${Math.floor(s / 60)}:${(s % 60).toString().padStart(2, "0")}`

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/30 p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle>Conectar WhatsApp</CardTitle>
          <CardDescription>Escanea el código QR con tu WhatsApp</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {status === "connecting" && (
            <div className="text-center py-8">
              <Loader2 className="mx-auto h-8 w-8 animate-spin text-muted-foreground" />
              <p className="mt-2 text-sm text-muted-foreground">Conectando...</p>
            </div>
          )}

          {status === "qr" && qrString && (
            <div className="space-y-3 text-center">
              <QRRender value={qrString} size={220} title="QR WhatsApp" />
              <p className="text-sm text-muted-foreground">Válido por {formatTime(countdown)}</p>
              <p className="text-xs text-muted-foreground">Abre WhatsApp → Dispositivos vinculados → Vincular dispositivo</p>
            </div>
          )}

          {status === "connected" && (
            <div className="text-center py-8">
              <CheckCircle2 className="mx-auto h-10 w-10 text-green-500" />
              <p className="mt-2 font-medium">¡Teléfono conectado!</p>
              <p className="text-sm text-muted-foreground mt-1">Puedes cerrar esta página</p>
            </div>
          )}

          {(status === "error" || status === "closed") && (
            <div>
              <Alert variant="destructive">
                <WifiOff className="h-4 w-4" />
                <AlertDescription>
                  {errorMsg || (status === "closed" ? "Enlace expirado o cerrado" : "Error desconocido")}
                </AlertDescription>
              </Alert>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default function QRPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    }>
      <QRPageContent />
    </Suspense>
  )
}
```

### 📐 Tarea 6 — frontend/lib/api.ts

Agregar función (después de las existentes de sesiones):
```ts
export async function generateQRLink(telefonoId: number): Promise<{
  ok: boolean;
  token?: string;
  phone_id?: number;
  expires_in?: number;
  error?: string;
}> {
  const token = localStorage.getItem("admin_token");
  const res = await fetch(`${API_BASE}/api/admin/telefonos/${telefonoId}/qr-link`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  return res.json();
}
```

### 📐 Tarea 6 — connect/page.tsx (cambios)

Agregar estado y botón:
```tsx
// Estados a agregar:
const [shareUrl, setShareUrl] = useState("")
const [sharing, setSharing] = useState(false)
const [copied, setCopied] = useState(false)

// Función:
const handleShare = async () => {
  setSharing(true)
  try {
    const data = await generateQRLink(telefonoId)
    if (data.ok && data.token) {
      const url = `${window.location.origin}/qr?token=${data.token}`
      setShareUrl(url)
    }
  } finally {
    setSharing(false)
  }
}

const handleCopy = async () => {
  await navigator.clipboard.writeText(shareUrl)
  setCopied(true)
  setTimeout(() => setCopied(false), 2000)
}
```

Botón a agregar en el área de botones (solo cuando status !== "active"):
```tsx
{status !== "active" && (
  <Button variant="outline" onClick={handleShare} disabled={sharing} size="sm">
    {sharing ? <Loader2 className="mr-2 h-3 w-3 animate-spin" /> : <Share2 className="mr-2 h-3 w-3" />}
    Compartir enlace QR
  </Button>
)}
{shareUrl && (
  <div className="rounded border p-2 text-xs space-y-1">
    <p className="break-all text-muted-foreground">{shareUrl}</p>
    <Button size="sm" variant="ghost" onClick={handleCopy} className="h-6 text-xs">
      {copied ? "✓ Copiado" : "Copiar enlace"}
    </Button>
  </div>
)}
```

Agregar import: `import { generateQRLink } from "@/lib/api"` y `Share2` de `lucide-react`.

### ⚠️ Lo que NO cambia esta story

- `whatsapp.StartSession` — no se modifica
- El path de empresa JWT en V1WSHandler — solo refactoring, misma lógica
- `LegacyWSHandler` (admin WS) — no se toca
- `GenerateEmpresaJWT` — no se modifica
- El middleware `ApiKeyAuthMiddleware` — no se toca
- `telefonoStore.BelongsToEmpresa` — sigue usándose para empresa JWT, no para QR link
- Las rutas REST de V1 — no se tocan

### 🧠 Cosas a verificar antes de implementar

1. **Import `auth` en `admin_sessions.go`:** El paquete es `wsapi/internal/auth`. Verificar que no hay conflicto con el variable de panel access en el mismo handler.

2. **`r.PathValue("id")` disponible:** Go 1.25 + `net/http.ServeMux` → `r.PathValue("id")` funciona para rutas con `{id}`. ✅

3. **El `jwtCfg` puede ser nil si la DB no está disponible:** Agregar nil-guard en `GenerateQRLink`. ✅ (ya incluido en el código de arriba)

4. **Formato de eventos en V1 WS vs admin WS:**
   - Admin WS: `{event: "qr-{accountID}", data: {...}}` (campo `event`, no `type`)
   - V1 WS: `{type: "qr", data: {...}}` (campo `type`, mapeado por `mapV1EventType`)
   - La página pública usa V1 WS → escuchar `type`, no `event`. ✅ (ya reflejado en el código)

5. **`buildAdminWsUrl` reutilizable:** La función ya existe en `api.ts` y acepta cualquier path + token opcional. La página pública la usa con `/api/service/v1/ws`. ✅

6. **No usar `localStorage` en página pública:** La página `/qr` no tiene panel auth; el token viene de la URL. Asegurarse de que el componente NO llama a `localStorage`. ✅

7. **`Share2` de lucide-react:** Verificar que el ícono existe en la versión instalada. Alternativa: `Link2` o `ExternalLink`. Usar el que esté disponible.

8. **Suspense en `/qr/page.tsx`:** `useSearchParams()` requiere `Suspense` en Next.js App Router. El patrón ya está reflejado en el código. ✅

9. **Ruta `/qr` sin layout de admin:** La página `/qr` vive en `app/qr/page.tsx` fuera del layout del panel admin. Verificar que el `layout.tsx` raíz del proyecto no aplique auth check que bloquee esta ruta. Revisar `frontend/app/layout.tsx` antes de implementar.

### Learnings de stories anteriores

- Package en `handlers/` es `package http` (NO `package handlers`)
- `fmt.Printf` para logs (no zerolog)
- `InsecureSkipVerify: true` en websocket.Accept — patrón establecido
- `r.PathValue("id")` para extraer path variables en Go 1.22+
- `writeAdminJSON` y `writeAdminError` disponibles en el mismo package (`http`) para handlers admin
- `nil-guard` en stores es obligatorio
- `time.NewTicker(25 * time.Second)` + `defer ticker.Stop()` — patrón de keepalive
- `whatsapp.NormalizeAccountID(phone.NumeroCompleto)` para accountID canónico
- El package de handlers NO es el mismo package que `container.go` — los handlers están en `internal/http/handlers/` (package `http`) y el container en `internal/http/` (package `http` también). Son el mismo package pero directorios distintos.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- ✅ Tarea 1: EmpresaJWTClaims extendido con Scope/PhoneID; extractEmpresaJWTClaims actualizado; GenerateQRLinkToken añadido
- ✅ Tarea 2: Middleware rechaza scope=="qr_link" antes del check de token_version (evita falso positivo con TokenVersion=0)
- ✅ Tarea 3: HandleWS refactorizado — branch QR link (auto-subscribe) vs empresa JWT (espera subscribe); logs diferenciados
- ✅ Tarea 4: AdminSessionsHandler recibe jwtCfg; GenerateQRLink implementado; ruta registrada; container actualizado
- ✅ Tarea 5: app/qr/page.tsx creada — pública sin auth, Suspense+useSearchParams, conecta V1 WS con token de URL
- ✅ Tarea 5: AdminAuthCheck actualizado para bypass /qr sin login
- ✅ Tarea 6: generateQRLink en api.ts; botón "Compartir enlace QR" + copia en connect/page.tsx
- ✅ Tarea 7: go build, go test, npm lint — todos limpios (3 errores pre-existentes en api.ts aceptados por AC11)

### File List

- backend/internal/domain/empresa_token.go (modificado)
- backend/internal/auth/empresa_jwt.go (modificado)
- backend/internal/http/middleware/empresa_auth.go (modificado)
- backend/internal/http/handlers/v1_ws.go (modificado)
- backend/internal/http/handlers/admin_sessions.go (modificado)
- backend/internal/http/routes_admin.go (modificado)
- backend/internal/http/container.go (modificado)
- frontend/app/qr/page.tsx (nuevo)
- frontend/components/admin-auth-check.tsx (modificado)
- frontend/lib/api.ts (modificado)
- frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/connect/page.tsx (modificado)

### Change Log

- 2026-05-08: Implementación completa de story 2-8 — token provisional QR link, endpoint admin, página pública /qr, botón compartir
