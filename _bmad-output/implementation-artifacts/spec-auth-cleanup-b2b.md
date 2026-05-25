---
status: done
slug: auth-cleanup-b2b
---

# Spec: Limpieza de Auth — B2B solo API Key, eliminar EmpresaJWT

## Contexto y problema

El proyecto tiene tres sistemas de autenticación:

1. **AdminJWT** → panel admin interno. ✅ No tocar.
2. **EmpresaJWT** (`empresaStack`) → JWT de larga duración (5 años) por empresa. ❌ Concepto incorrecto: existe en rutas B2B pero el B2B real es exclusivamente API Key por teléfono. El middleware de API Key incluso inyecta `EmpresaJWTClaims` sintéticos en el contexto, creando ambigüedad de identidad.
3. **ApiKey** (`clientStack`) → API Key por teléfono. ✅ Este ES el auth B2B real.

**Archivos muertos** (no registrados en ninguna ruta pero compilan y referencian `EmpresaJWTClaims`):
- `backend/internal/http/handlers.go` (struct `Handler`)
- `backend/internal/http/v1_handler.go` (struct `V1Handler`)

**Dependencia crítica de QR-link:** `backend/internal/auth/empresa_jwt.go` contiene TANTO `GenerateEmpresaJWT` (eliminar) COMO `GenerateQRLinkToken` y `ParseEmpresaJWT` (parcialmente necesarios para el WS handler del QR). Antes de eliminar este archivo, hay que extraer la lógica de QR-link a archivos propios.

**Frontend análisis:** El frontend (`frontend/lib/api.ts`) NO llama ninguna ruta `empresaStack`. Todas las llamadas van a `/api/admin/*`. La función `fetchWithEmpresaAuth` usa el admin JWT (misleading name, safe to ignore). La única ruta B2B que usa el frontend es `GET /api/service/v1/ws` en `app/qr/page.tsx`, con un token QR-link generado vía `POST /api/admin/telefonos/{id}/qr-link`. **El frontend no se ve afectado por esta limpieza**, salvo que el WS handler debe seguir aceptando tokens QR-link.

## Objetivo

**B2B = solo `clientStack` (ApiKeyClaims). El endpoint `/ws` acepta exclusivamente tokens QR-link.**

1. Extraer lógica de QR-link a archivos propios antes de eliminar `empresa_jwt.go`.
2. Eliminar `EmpresaJWTClaims` y toda su infraestructura del sistema.
3. Eliminar `empresaStack` de las rutas B2B. Esos endpoints se gestionan desde el panel admin.
4. Eliminar handlers B2B obsoletos y los admin handlers de generación/revocación de empresa JWT.
5. Limpiar la inyección sintética de `EmpresaJWTClaims` en el middleware de API Key.
6. Actualizar el WS handler para que solo acepte tokens QR-link (eliminar el flujo de subscribe con empresa JWT).
7. Al finalizar, ejecutar `/bmad-code-review`.

---

## PASO 1 — Crear archivos nuevos (hacer PRIMERO, antes de eliminar nada)

### 1a. CREAR `backend/internal/domain/qr_link_claims.go`

```go
package domain

import "context"

// QRLinkClaims representa los claims del token provisional de QR link (10 min).
type QRLinkClaims struct {
	EmpresaID int64  `json:"sub"`
	PhoneID   int64  `json:"phone_id"`
	Scope     string `json:"scope"` // siempre "qr_link"
}

type qrLinkClaimsKey struct{}

func WithQRLinkClaims(ctx context.Context, c *QRLinkClaims) context.Context {
	return context.WithValue(ctx, qrLinkClaimsKey{}, c)
}

func GetQRLinkClaims(ctx context.Context) (*QRLinkClaims, bool) {
	c, ok := ctx.Value(qrLinkClaimsKey{}).(*QRLinkClaims)
	return c, ok
}
```

### 1b. CREAR `backend/internal/auth/qr_link_jwt.go`

```go
package auth

import (
	"fmt"
	"time"

	"wsapi/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

const qrLinkTokenExpiry = 600 * time.Second // 10 minutos

// GenerateQRLinkToken genera un JWT de corta duración para el QR link de un teléfono.
func GenerateQRLinkToken(empresaID, phoneID int64, secret string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      empresaID,
		"phone_id": phoneID,
		"scope":    "qr_link",
		"iat":      now.Unix(),
		"exp":      now.Add(qrLinkTokenExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error al firmar QR link token: %w", err)
	}
	return signed, nil
}

// ParseQRLinkToken valida y parsea un JWT de QR link. Rechaza tokens con scope distinto a "qr_link".
func ParseQRLinkToken(tokenString, secret string) (*domain.QRLinkClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo inesperado: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("QR link token inválido: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token no válido")
	}
	m, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("claims con formato inesperado")
	}
	var empresaID, phoneID int64
	if v, ok := m["sub"].(float64); ok {
		empresaID = int64(v)
	}
	if v, ok := m["phone_id"].(float64); ok {
		phoneID = int64(v)
	}
	scope, _ := m["scope"].(string)
	if scope != "qr_link" {
		return nil, fmt.Errorf("scope inválido para QR link: %q", scope)
	}
	if phoneID <= 0 {
		return nil, fmt.Errorf("phone_id ausente o inválido en token QR")
	}
	return &domain.QRLinkClaims{EmpresaID: empresaID, PhoneID: phoneID, Scope: scope}, nil
}
```

---

## PASO 2 — Archivos a ELIMINAR completos

Eliminar en este orden exacto. Después de cada `rm`, ejecutar `cd backend && go build ./...` para detectar referencias rotas antes de continuar.

| Orden | Archivo | Razón |
|-------|---------|-------|
| 1 | `backend/internal/http/handlers.go` | Struct `Handler` muerto, no registrado. Usa `EmpresaJWTClaims`. |
| 2 | `backend/internal/http/v1_handler.go` | Struct `V1Handler` muerto, no registrado. Usa `EmpresaJWTClaims`. |
| 3 | `backend/internal/http/handlers/v1_sessions.go` | Usa `GetEmpresaJWTClaims` en todos sus métodos. Sus rutas se eliminan. |
| 4 | `backend/internal/http/handlers/v1_phones.go` | Usa `GetEmpresaJWTClaims` en todos sus métodos. Sus rutas se eliminan. |
| 5 | `backend/internal/http/handlers/v1_metrics.go` | Usa `GetEmpresaJWTClaims`. Su ruta se elimina. |
| 6 | `backend/internal/http/middleware/empresa_auth.go` | Middleware `EmpresaAuthMiddleware`. Sin EmpresaJWT, no existe. |
| 7 | `backend/internal/auth/empresa_jwt.go` | El paso 1 ya extrajo `GenerateQRLinkToken` y creó `ParseQRLinkToken`. Eliminar **solo después** de que el paso 1 esté completo. |
| 8 | `backend/internal/domain/empresa_token.go` | Define `EmpresaJWTClaims`, `WithEmpresaJWTClaims`, `GetEmpresaJWTClaims`, `EmpresaJWTResponse`. Eliminar **solo después** de que ningún archivo activo importe estos símbolos. |

> ⚠️ `WithEmpresaID` NO está en `empresa_token.go` — está en `domain/empresa_filter.go` y es usado por `router.go`. No eliminar `empresa_filter.go`.

---

## PASO 3 — Archivos a MODIFICAR

### 3a. `backend/internal/http/handlers/v1_ws.go`

Reemplazar el handler completo. Antes usaba `auth.ParseEmpresaJWT` y aceptaba tanto empresa JWT como QR-link. Ahora solo acepta QR-link.

**Eliminar importaciones:** `"wsapi/internal/auth"` (el import ya no se necesita si se usa `auth.ParseQRLinkToken` — ajustar import), `"encoding/json"` (el flujo subscribe que usaba json desaparece).

**Antes (lines 52–103):**
```go
claims, err := auth.ParseEmpresaJWT(token, h.jwtCfg.Secret)
if err != nil {
    writeV1Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token inválido o expirado")
    return
}
// ... accept websocket ...
var phoneID int64
if claims.Scope == "qr_link" {
    if claims.PhoneID <= 0 {
        _ = writeWSEvent(c, "error", map[string]string{"message": "token QR inválido: phone_id ausente"})
        return
    }
    phoneID = claims.PhoneID
} else {
    // Empresa JWT regular: esperar mensaje subscribe con phone_id
    _, data, err := c.Read(ctx)
    // ... leer subscribe, validar BelongsToEmpresa, etc ...
    phoneID = sub.PhoneID
}
```

**Después:** Reemplazar ese bloque por:
```go
claims, err := auth.ParseQRLinkToken(token, h.jwtCfg.Secret)
if err != nil {
    writeV1Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token QR inválido o expirado")
    return
}
// ... accept websocket ...
phoneID := claims.PhoneID
```

**Eliminar también** los bloques condicionales `if claims.Scope == "qr_link"` en los logs y defer. Reemplazar por logs directos sin condición:
```go
// Antes (conditional logs):
if claims.Scope == "qr_link" {
    fmt.Printf("[INFO] V1 WS QR-link opened phone=%d account=%s\n", phone.ID, accountID)
} else {
    fmt.Printf("[INFO] V1 WS opened empresa=%d phone=%d account=%s\n", claims.EmpresaID, phone.ID, accountID)
}

// Después:
fmt.Printf("[INFO] V1 WS QR-link opened phone=%d account=%s\n", phone.ID, accountID)
```

**Campos eliminados del struct `V1WSHandler`:** `telefonoStore` ya no se usa (solo se usaba para `BelongsToEmpresa` en el flujo subscribe). Eliminar el campo y el parámetro en `NewV1WSHandler`.

**Imports que quedan en `v1_ws.go`:** `context`, `encoding/json`, `fmt`, `net/http`, `strings`, `time`, `wsapi/internal/auth`, `wsapi/internal/config`, `wsapi/internal/storage`, `wsapi/internal/whatsapp`, `github.com/coder/websocket`. Retirar `encoding/json` y `wsapi/internal/storage` si ya no se usan tras eliminar el flujo subscribe y `telefonoStore`.

---

### 3b. `backend/internal/http/middleware/api_key_auth.go`

Eliminar las líneas que inyectan `EmpresaJWTClaims` y `EmpresaID` sintéticos.

**Antes (bloque de contexto, ~lines 57–75):**
```go
ctx := domain.WithApiKeyClaims(r.Context(), &domain.ApiKeyClaims{
    ApiKeyID:   key.ID,
    EmpresaID:  key.EmpresaID,
    TelefonoID: key.TelefonoID,
    KeyPrefix:  key.KeyPrefix,
    Scopes:     key.Scopes,
})
ctx = domain.WithEmpresaID(ctx, key.EmpresaID)
ctx = domain.WithEmpresaJWTClaims(ctx, &domain.EmpresaJWTClaims{
    EmpresaID:     key.EmpresaID,
    TokenVersion:  0,
    EmpresaRUC:    empresa.RUC,
    EmpresaNombre: empresa.Nombre,
    Permissions:   key.Scopes,
})
next.ServeHTTP(w, r.WithContext(ctx))
```

**Después:**
```go
ctx := domain.WithApiKeyClaims(r.Context(), &domain.ApiKeyClaims{
    ApiKeyID:   key.ID,
    EmpresaID:  key.EmpresaID,
    TelefonoID: key.TelefonoID,
    KeyPrefix:  key.KeyPrefix,
    Scopes:     key.Scopes,
})
next.ServeHTTP(w, r.WithContext(ctx))
```

Eliminar el import de `"wsapi/internal/domain"` solo si ya no se usa en ese archivo (verificar — `domain.WithApiKeyClaims` sí se usa, así que el import queda). Solo eliminar las dos líneas de inyección.

---

### 3c. `backend/internal/http/kernel.go`

**Antes:**
```go
type Kernel struct {
    AdminAuth    func(http.Handler) http.Handler
    EmpresaAuth  func(http.Handler) http.Handler
    ClientAuth   func(http.Handler) http.Handler
    ServiceStack func(http.Handler) http.Handler
    Global       []func(http.Handler) http.Handler
}

func NewKernel(auth AdminAuthProvider, empresaAuth EmpresaAuthProvider, apiKeyAuth ClientAuthProvider, telemetryMW func(http.Handler) http.Handler) *Kernel {
    ...
    k := &Kernel{
        AdminAuth:   ...,
        EmpresaAuth: identityMiddleware,
        ...
    }
    if empresaAuth != nil {
        k.EmpresaAuth = empresaAuth.RequireEmpresaAuth()
    }
    ...
}

type EmpresaAuthProvider interface {
    RequireEmpresaAuth() func(http.Handler) http.Handler
}
```

**Después:**
```go
type Kernel struct {
    AdminAuth    func(http.Handler) http.Handler
    ClientAuth   func(http.Handler) http.Handler
    ServiceStack func(http.Handler) http.Handler
    Global       []func(http.Handler) http.Handler
}

func NewKernel(auth AdminAuthProvider, apiKeyAuth ClientAuthProvider, telemetryMW func(http.Handler) http.Handler) *Kernel {
    ...
    k := &Kernel{
        AdminAuth: ...,
        ...
    }
    ...
}
// Eliminar EmpresaAuthProvider interface completa
```

---

### 3d. `backend/internal/http/router.go`

**Antes (line 46):**
```go
k := NewKernel(c.AuthMiddleware, c.EmpresaAuthMiddleware, c.ApiKeyAuthMiddleware, c.TelemetryMW)
```

**Después:**
```go
k := NewKernel(c.AuthMiddleware, c.ApiKeyAuthMiddleware, c.TelemetryMW)
```

---

### 3e. `backend/internal/http/routes_api.go`

Eliminar TODAS las líneas de `empresaStack` y sus rutas. Eliminar la declaración de la variable.

**Antes (líneas a eliminar):**
```go
empresaStack := k.EmpresaAuth

mux.Handle("POST /api/service/v1/auth/empresa/validate", empresaStack(...))
mux.Handle("GET /api/service/v1/empresas", empresaStack(http.HandlerFunc(c.CompaniesHandler.GetCurrent)))
mux.Handle("PUT /api/service/v1/empresas", empresaStack(http.HandlerFunc(c.CompaniesHandler.UpdateCurrent)))
mux.Handle("GET /api/service/v1/metricas", empresaStack(http.HandlerFunc(c.V1MetricsHandler.GetMetrics)))
mux.Handle("GET /api/service/v1/telefonos", empresaStack(http.HandlerFunc(c.V1PhonesHandler.GetPhones)))
mux.Handle("POST /api/service/v1/telefonos/{id}/qr", empresaStack(http.HandlerFunc(c.V1PhonesHandler.PostPhoneQr)))
mux.Handle("GET /api/service/v1/telefonos/{id}/estado", empresaStack(http.HandlerFunc(c.V1PhonesHandler.GetPhoneStatus)))
mux.Handle("GET /api/service/v1/sesiones", empresaStack(http.HandlerFunc(c.V1SessionsHandler.GetSessions)))
mux.Handle("POST /api/service/v1/sesiones", empresaStack(http.HandlerFunc(c.V1SessionsHandler.PostSessions)))
mux.Handle("GET /api/service/v1/sesiones/{id}", empresaStack(http.HandlerFunc(c.V1SessionsHandler.GetSession)))
mux.Handle("DELETE /api/service/v1/sesiones/{id}", empresaStack(http.HandlerFunc(c.V1SessionsHandler.DeleteSession)))
mux.Handle("POST /api/service/v1/sesiones/{id}/connect", empresaStack(http.HandlerFunc(c.V1SessionsHandler.StartPhoneConnection)))
mux.Handle("GET /api/service/v1/empresas/webhooks", empresaStack(http.HandlerFunc(c.V1WebhooksHandler.ListByEmpresa)))
```

**Verificar** que las referencias a `c.V1SessionsHandler`, `c.V1PhonesHandler`, `c.V1MetricsHandler`, `c.CompaniesHandler.GetCurrent`, `c.CompaniesHandler.UpdateCurrent`, `c.V1WebhooksHandler.ListByEmpresa` queden eliminadas del archivo junto con sus rutas.

**Rutas que permanecen (no tocar):**
```
GET  /api/service/v1/health
GET  /api/service/v1/me
GET  /api/service/v1/sesion
GET  /api/service/v1/mensajes
POST /api/service/v1/mensajes
GET  /api/service/v1/mensajes/{id}
PATCH /api/service/v1/mensajes/{id}
POST /api/service/v1/mensajes/{id}
GET  /api/service/v1/difusiones
POST /api/service/v1/difusiones
GET  /api/service/v1/difusiones/{id}
GET  /api/service/v1/ws
POST /api/service/v1/webhooks
GET  /api/service/v1/webhooks
DELETE /api/service/v1/webhooks/{id}
```

---

### 3f. `backend/internal/http/routes_admin.go`

Eliminar las dos rutas de empresa token (generación y revocación de EmpresaJWT). Sin EmpresaJWT, estas rutas no tienen sentido.

**Eliminar:**
```go
mux.Handle("POST /api/admin/empresas/{id}/token", adminStack(http.HandlerFunc(c.CompaniesHandler.GenerateToken)))
mux.Handle("POST /api/admin/empresas/{id}/token/revoke", adminStack(http.HandlerFunc(c.CompaniesHandler.RevokeToken)))
```

---

### 3g. `backend/internal/http/handlers/companies.go`

Eliminar cuatro métodos:
1. `GetCurrent` — usa `GetEmpresaJWTClaims`, era la ruta `GET /api/service/v1/empresas`.
2. `UpdateCurrent` — usa `GetEmpresaJWTClaims`, era la ruta `PUT /api/service/v1/empresas`.
3. `GenerateToken` — llama `auth.GenerateEmpresaJWT(empresa, ...)`. Era la ruta admin `POST /api/admin/empresas/{id}/token`.
4. `RevokeToken` — incrementa `token_version` de empresa para invalidar empresa JWTs. Era la ruta admin `POST /api/admin/empresas/{id}/token/revoke`.

Buscar y eliminar los cuatro métodos completos. Verificar que el import de `"wsapi/internal/auth"` en `companies.go` quede eliminado si ya no hay otras llamadas a `auth.*` en ese archivo.

Los métodos que **permanecen**: `List`, `Get`, `Create`, `Update`, `Delete`, `Restore` (todos usan `GetAdminJWTClaims` o acceso de panel).

---

### 3h. `backend/internal/http/handlers/v1_webhooks.go`

Eliminar solo el método `ListByEmpresa` (usa `GetEmpresaJWTClaims`, era la ruta `GET /api/service/v1/empresas/webhooks`).

Los métodos `Create`, `List`, `Delete` usan `GetApiKeyClaims` — no tocar.

---

### 3i. `backend/internal/http/handlers/v1_helpers.go`

Eliminar las funciones `getEmpresaIDFromContext` y `getAccessClaims` completas.

**Antes:**
```go
func getEmpresaIDFromContext(r *http.Request) (int64, bool) {
    if claims, ok := domain.GetEmpresaJWTClaims(r.Context()); ok {
        return claims.EmpresaID, true
    }
    // ...
}

func getAccessClaims(r *http.Request) (*domain.EmpresaJWTClaims, *domain.ApiKeyClaims, bool) {
    if claims, ok := domain.GetEmpresaJWTClaims(r.Context()); ok {
        return &domain.EmpresaJWTClaims{...}, nil, true
    }
    // ...
}
```

**Después:** Eliminar ambas funciones. Si algún handler activo las llama, reemplazar la llamada con `domain.GetApiKeyClaims(r.Context())` directamente.

Las funciones `writeV1Error`, `writeV1Success`, `extractTelefonoID` no usan `EmpresaJWTClaims` — no tocar.

---

### 3j. `backend/internal/domain/panel_access.go`

**Antes:**
```go
func GetPanelAccess(ctx context.Context) (PanelAccess, bool) {
    if claims, ok := GetAdminJWTClaims(ctx); ok && claims != nil {
        return PanelAccess{IsRoot: claims.IsRoot, IsAdminJWT: true}, true
    }
    if claims, ok := GetEmpresaJWTClaims(ctx); ok && claims != nil {
        eid := claims.EmpresaID
        return PanelAccess{EmpresaID: &eid}, true
    }
    return PanelAccess{}, false
}
```

**Después:**
```go
func GetPanelAccess(ctx context.Context) (PanelAccess, bool) {
    if claims, ok := GetAdminJWTClaims(ctx); ok && claims != nil {
        return PanelAccess{IsRoot: claims.IsRoot, IsAdminJWT: true}, true
    }
    return PanelAccess{}, false
}
```

Eliminar también el campo `EmpresaID *int64` del struct `PanelAccess` y los métodos `CompanyID()` y cualquier lógica de `CanAccessEmpresa` que dependa de `EmpresaID`. Verificar que ningún caller activo use `PanelAccess.EmpresaID`.

---

### 3k. `backend/internal/http/container.go`

**Eliminar:**
```go
// En el struct Container:
EmpresaAuthMiddleware *middleware.EmpresaAuthMiddleware

// En el constructor:
empresaAuthMiddleware := middleware.NewEmpresaAuthMiddleware(jwtCfg, empresaStore, telefonoStore)
// ...
EmpresaAuthMiddleware: empresaAuthMiddleware,
```

**Eliminar también** los campos del container correspondientes a los handlers eliminados:
```go
V1SessionsHandler *handlers.V1SessionsHandler
V1PhonesHandler   *handlers.V1PhonesHandler
V1MetricsHandler  *handlers.V1MetricsHandler
```

Y sus instanciaciones en el constructor.

---

### 3l. `backend/internal/http/handlers/admin_sessions.go`

**Solo actualizar el import.** Este archivo llama `auth.GenerateQRLinkToken(...)` — la función se movió de `empresa_jwt.go` a `qr_link_jwt.go`, pero el paquete sigue siendo `auth`. No hay cambio funcional, solo verificar que el import `"wsapi/internal/auth"` siga presente y que `go build` compile correctamente.

---

### 3m. `docs/routes/contrato-b2b/`

El directorio ya fue renombrado de `contrato-empresa` a `contrato-b2b` en una sesión anterior. Verificar que exista en `docs/routes/contrato-b2b/`.

Eliminar los archivos que documentaban endpoints ahora inexistentes en B2B:
- `sesiones.md`
- `telefonos.md`
- `metricas.md`
- `empresa.md`

Actualizar `README.md` de `contrato-b2b`:
- Eliminar toda referencia a `empresaStack`, JWT de empresa, o `EmpresaJWTClaims`.
- Actualizar el índice de endpoints para reflejar solo los que quedan (los listados en el paso 3e).
- La sección de autenticación debe decir: "Autenticación: `X-API-Key: <api_key>` en el header. Las API Keys se crean por teléfono desde el panel admin."

---

## Criterios de Aceptación

### AC-1: No existe `EmpresaJWTClaims` en código activo
```bash
grep -r "EmpresaJWTClaims" backend/internal/ --include="*.go" | grep -v "_test.go"
# Resultado esperado: sin output (0 líneas)
```

### AC-2: No existe `empresaStack` ni `EmpresaAuth` en el kernel
```bash
grep -r "empresaStack\|EmpresaAuth\|EmpresaAuthProvider\|EmpresaAuthMiddleware" backend/internal/ --include="*.go"
# Resultado esperado: sin output
```

### AC-3: El middleware API Key no inyecta claims sintéticos
```bash
grep -n "WithEmpresaJWTClaims" backend/internal/http/middleware/api_key_auth.go
# Resultado esperado: sin output
```

### AC-4: El WS handler solo usa `ParseQRLinkToken`
```bash
grep -n "ParseEmpresaJWT\|ParseQRLinkToken" backend/internal/http/handlers/v1_ws.go
# Resultado esperado: exactamente 1 línea, con ParseQRLinkToken
```

### AC-5: `empresa_jwt.go` y `empresa_auth.go` no existen
```bash
ls backend/internal/auth/empresa_jwt.go backend/internal/http/middleware/empresa_auth.go
# Resultado esperado: error "No such file or directory"
```

### AC-6: Los nuevos archivos QR-link existen y compilan
```bash
ls backend/internal/auth/qr_link_jwt.go backend/internal/domain/qr_link_claims.go
# Resultado esperado: ambos archivos existen
cd backend && go build ./...
# Resultado esperado: 0 errores
```

### AC-7: Los tests pasan
```bash
cd backend && go test ./...
# Resultado esperado: todos pasan (los tests de handlers eliminados se eliminan junto con el handler)
```

### AC-8: Rutas admin de empresa token eliminadas
```bash
grep -n "empresas.*token\|token/revoke" backend/internal/http/routes_admin.go
# Resultado esperado: sin output
```

---

## Orden de implementación obligatorio

1. **Crear** `domain/qr_link_claims.go` (paso 1a).
2. **Crear** `auth/qr_link_jwt.go` (paso 1b).
3. **Eliminar** `handlers.go` y `v1_handler.go` (código muerto, fácil).
4. **Eliminar** `v1_sessions.go`, `v1_phones.go`, `v1_metrics.go`.
5. **Eliminar** `middleware/empresa_auth.go`.
6. **Eliminar** `auth/empresa_jwt.go` (ya extraído en paso 1).
7. **Eliminar** `domain/empresa_token.go` (ya reemplazado en paso 1).
8. **Modificar** `api_key_auth.go` — quitar inyección sintética.
9. **Modificar** `kernel.go` — quitar `EmpresaAuth` y `EmpresaAuthProvider`.
10. **Modificar** `router.go` — quitar `c.EmpresaAuthMiddleware` de `NewKernel(...)`.
11. **Modificar** `panel_access.go` — quitar fallback `EmpresaJWTClaims`.
12. **Modificar** `v1_helpers.go` — quitar `getEmpresaIDFromContext` y `getAccessClaims`.
13. **Modificar** `v1_webhooks.go` — quitar `ListByEmpresa`.
14. **Modificar** `companies.go` — quitar `GetCurrent`, `UpdateCurrent`, `GenerateToken`, `RevokeToken`.
15. **Modificar** `container.go` — quitar campos y constructores eliminados.
16. **Modificar** `routes_api.go` — quitar `empresaStack` y sus rutas.
17. **Modificar** `routes_admin.go` — quitar rutas de token de empresa.
18. **Modificar** `v1_ws.go` — usar `ParseQRLinkToken`, eliminar flujo subscribe, quitar `telefonoStore`.
19. **Ejecutar** `cd backend && go build ./...` → 0 errores.
20. **Ejecutar** `cd backend && go test ./...` → todos pasan.
21. **Actualizar** docs `contrato-b2b/` (eliminar archivos, actualizar README).

---

## Nota sobre `domain/empresa_filter.go`

**No tocar.** `WithEmpresaID` y `GetEmpresaID` están en `empresa_filter.go`, no en `empresa_token.go`. Son usados por `router.go` para filtros de empresa en el panel admin. Solo eliminar la llamada a `domain.WithEmpresaID` en `api_key_auth.go`.

---

### Review Findings — Grupo 3: Handlers (v1_ws, v1_helpers, v1_webhooks, companies, routes) (2026-05-25)

- [x] [Review][Defer] WS `phone == nil` tras accept no envía close frame explícito antes de error event — `defer c.CloseNow()` cierra, pero sin código de cierre WS [`backend/internal/http/handlers/v1_ws.go` L62-72] — deferred, pre-existing

### Review Findings — Grupo 2: Core (kernel, router, container, api_key_auth, panel_access) (2026-05-25)

- [x] [Review][Defer] `registeredRoutes` tiene gaps pre-existentes causando 404 en OPTIONS/CORS para rutas no listadas [`backend/internal/http/router.go`] — deferred, pre-existing

### Review Findings — Grupo 1: Archivos QR-link nuevos (2026-05-25)

- [x] [Review][Patch] `empresaID` sin validación: valor 0 pasa silenciosamente en `ParseQRLinkToken` [`backend/internal/auth/qr_link_jwt.go`]
- [x] [Review][Patch] Secret vacío aceptado sin error permite tokens forjables trivialmente [`backend/internal/auth/qr_link_jwt.go`]
- [x] [Review][Patch] Errores de setup de tokens en tests ignorados con `_` [`backend/internal/auth/qr_link_jwt_test.go`]
- [x] [Review][Defer] float64→int64 truncación silenciosa para IDs > 2^53 [`backend/internal/auth/qr_link_jwt.go` ~L50] — deferred, pre-existing pattern
- [x] [Review][Defer] Token WS entregado en query param aparece en logs del servidor — deferred, pre-existing design
- [x] [Review][Defer] Sin claim `nbf`, ventana de replay completa de 10 minutos — deferred, pre-existing

---

## Post-condición obligatoria

Al terminar toda la implementación, ejecutar `/bmad-code-review` para verificar:
- No queda ningún rastro de `EmpresaJWTClaims` en código activo.
- Ninguna ruta B2B acepta auth distinto a API Key (excepto `/ws` que solo acepta QR-link token).
- El sistema compila (`go build ./...`) y los tests pasan (`go test ./...`).
- No se introdujeron regresiones en el contrato admin.
- El frontend sigue funcionando: la página QR usa el endpoint `/api/service/v1/ws` con token QR-link, que ahora usa `ParseQRLinkToken`.
