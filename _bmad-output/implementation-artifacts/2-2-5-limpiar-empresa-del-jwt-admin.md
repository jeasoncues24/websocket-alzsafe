---
title: 'Story 2.2.5 — Eliminar empresa del JWT admin y corregir control de acceso'
type: 'fix'
created: '2026-05-04'
status: 'review'
epic: 'epic-2-mejoras-post-revision'
baseline_commit: '69509dc'
context:
  - '{project-root}/_bmad-output/project-context.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** `AdminJWTClaims` declara `EmpresaID`, `EmpresaRUC` y `EmpresaNombre`, pero estos campos nunca se populan al generar el token admin (comentario en `auth.go:69`: *"admin users don't have empresa_id"*). El resultado es una inconsistencia silenciosa con dos síntomas graves:

1. **Código muerto:** el middleware lee esos campos del JWT pero siempre los encuentra vacíos; la función `generateToken` acepta parámetros `empresaRUC`/`empresaNombre` que se pasan como `nil`.
2. **Bug de acceso roto para admins no-root:** todos los handlers que verifican `claims.EmpresaID == companyID` fallan con 403 para cualquier admin con JWT administrativo que no sea `is_root`, porque `EmpresaID` siempre es nil. En la práctica, solo el super_admin puede operar en los endpoints de empresa/teléfonos desde el JWT admin.

**Approach:**
- Eliminar `EmpresaID`, `EmpresaRUC` y `EmpresaNombre` del struct `AdminJWTClaims` y de la generación del token.
- Limpiar el middleware que los parseaba.
- Rediseñar `panelAdminAccess`: un JWT admin válido (cualquier rol) puede acceder a todas las empresas; solo el JWT de empresa (`EmpresaJWTClaims`) sigue limitado a su propia empresa. El flag `is_root` se preserva para operaciones privilegiadas.
- Resultado: los admins no-root del panel recuperan acceso completo a las operaciones de empresa/teléfonos.

## Boundaries & Constraints

**Always:** preservar el comportamiento de `EmpresaJWTClaims` intacto (sigue usando `EmpresaID` para escopar); mantener el flag `is_root` para operaciones que lo requieran; no modificar las rutas `/api/v1/` ni el JWT de empresa; limpiar todos los usos de `claims.EmpresaID` en el contexto de JWT admin sin dejar compilación rota.

**Never:** tocar `EmpresaJWTClaims`, `empresa_jwt.go`, `empresa_auth.go`, ni los handlers de v1; cambiar el esquema de la BD; crear migraciones; tocar tests de empresa JWT.

**Ask First:** si algún flujo activo depende de que un admin tenga su `empresa_id` en el JWT para auto-restringirse (p.ej. un admin multi-tenant explícito). Si no existe ese caso, proceder.

## Contexto técnico relevante

**Generación del JWT admin** (`handlers/auth.go:generateToken`):
```go
// línea 69: "admin users don't have empresa_id"
token, err := h.generateToken(user, nil, nil)  // empresaRUC=nil, empresaNombre=nil

func (h *AuthHandler) generateToken(user *domain.AdminUser, empresaRUC, empresaNombre *string) (string, error) {
    claims := jwt.MapClaims{
        // ...
        "empresa_ruc":    empresaRUC,     // siempre nil
        "empresa_nombre": empresaNombre,  // siempre nil
        // "empresa_id" nunca se incluye
    }
}
```

**Struct afectado** (`domain/admin_user.go`):
```go
type AdminJWTClaims struct {
    EmpresaID     *int64   `json:"empresa_id,omitempty"`   // ← eliminar
    EmpresaRUC    *string  `json:"empresa_ruc,omitempty"`  // ← eliminar
    EmpresaNombre *string  `json:"empresa_nombre,omitempty"` // ← eliminar
}
```

**Bug de acceso roto** (patrón en `admin.go` y `handlers/companies.go`):
```go
// Este check siempre falla para admin JWT no-root porque EmpresaID == nil
if claims != nil && !claims.IsRoot {
    if claims.EmpresaID == nil || *claims.EmpresaID != companyID {
        403 Forbidden  // ← todo admin no-root queda bloqueado
    }
}
```

**Solución para `panelAdminAccess`** — agregar `isAdminJWT bool`:
```go
type panelAdminAccess struct {
    EmpresaID  *int64
    IsRoot     bool
    isAdminJWT bool  // true cuando viene de AdminJWTClaims
}

func getPanelAdminAccess(r *http.Request) (panelAdminAccess, bool) {
    if claims, ok := domain.GetAdminJWTClaims(r.Context()); ok && claims != nil {
        return panelAdminAccess{IsRoot: claims.IsRoot, isAdminJWT: true}, true
    }
    if claims, ok := domain.GetEmpresaJWTClaims(r.Context()); ok && claims != nil {
        eid := claims.EmpresaID
        return panelAdminAccess{EmpresaID: &eid}, true
    }
    return panelAdminAccess{}, false
}

func (a panelAdminAccess) canAccessEmpresa(empresaID int64) bool {
    if a.IsRoot || a.isAdminJWT {
        return true  // admin JWT puede acceder a cualquier empresa
    }
    if a.EmpresaID == nil {
        return false
    }
    return *a.EmpresaID == empresaID
}
```

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Admin no-root accede a empresa | JWT admin con `is_root=false`, `GET /api/admin/empresas/1/telefonos` | 200 OK — acceso permitido por ser admin JWT | N/A |
| Admin root accede a empresa | JWT admin con `is_root=true` | 200 OK — sin cambio respecto al estado actual | N/A |
| Empresa JWT accede a su empresa | JWT empresa, `GET /api/admin/empresas/1/telefonos` con empresa_id=1 | 200 OK — sin cambio | N/A |
| Empresa JWT accede a empresa ajena | JWT empresa con empresa_id=1, accede a empresa 2 | 403 Forbidden — sin cambio | N/A |
| JWT admin generado tras el cambio | Login con admin no-root | Token no contiene `empresa_id`, `empresa_ruc`, `empresa_nombre` | N/A |
| JWT admin viejo (con campos nil) | Token emitido antes del cambio | Middleware los ignora igual que antes (ya eran nil) | N/A |

</frozen-after-approval>

## Code Map

**Backend — eliminar campos del struct y generación:**
- `backend/internal/domain/admin_user.go` — remover `EmpresaID *int64`, `EmpresaRUC *string`, `EmpresaNombre *string` de `AdminJWTClaims`
- `backend/internal/http/handlers/auth.go` — remover params `empresaRUC, empresaNombre *string` de `generateToken`; remover `"empresa_ruc"` y `"empresa_nombre"` de `jwt.MapClaims`; simplificar llamada en `Login`
- `backend/internal/http/middleware/auth.go` — remover bloque que parsea `empresa_id`, `empresa_ruc`, `empresa_nombre` del JWT (approx. líneas 81-111); actualizar construcción de `AdminJWTClaims` para no incluir esos campos

**Backend — corregir control de acceso:**
- `backend/internal/http/admin.go` — agregar `isAdminJWT bool` a `panelAdminAccess`; actualizar `getPanelAdminAccess` para setear `isAdminJWT: true` en path de admin JWT; actualizar `canAccessEmpresa` y `companyID()` según el diseño propuesto; eliminar los checks directos de `claims.EmpresaID` en `ListCompanyPhones`, `CreateCompanyPhone`, `UpdateCompanyPhone`, `DeleteCompanyPhone`, `GetCompanyPhone`, `StartCompanyPhoneConnection`, `ConnectCompanyPhoneWS`, `GetSessionsDiagnostics` (reemplazarlos con `getPanelAdminAccess` + `canAccessEmpresa`)
- `backend/internal/http/handlers/companies.go` — reemplazar todos los checks directos de `claims.EmpresaID` en `Delete`, `GetCurrent`, `UpdateCurrent` con el patrón `getPanelAdminAccess` + `canAccessEmpresa`

## Tasks & Acceptance

**Execution:**

- [x] `domain/admin_user.go` — eliminar `EmpresaID`, `EmpresaRUC`, `EmpresaNombre` de `AdminJWTClaims`

- [x] `handlers/auth.go` — simplificar `generateToken`: eliminar los dos parámetros opcionales, eliminar `empresa_ruc` y `empresa_nombre` de `jwt.MapClaims`, y simplificar la llamada en `Login`

- [x] `middleware/auth.go` — eliminar el bloque de parsing de `empresa_id`, `empresa_ruc`, `empresa_nombre`; actualizar la construcción del struct `AdminJWTClaims` resultante

- [x] `http/admin.go` — centralizar acceso con `PanelAccess`, permitir acceso global para JWT admin válido y mantener scope de `EmpresaJWTClaims`; reemplazar checks directos de `claims.EmpresaID`

- [x] `handlers/companies.go` — reemplazar checks directos de `claims.EmpresaID` con `PanelAccess` + `CanAccessEmpresa`

- [x] Verificar que el proyecto compila sin errores: `go build ./...`

**Acceptance Criteria:**

- Given un admin con rol `admin` (no root) hace `GET /api/admin/empresas/1/telefonos`, when usa su JWT admin, then recibe `200 OK` con la lista de teléfonos.
- Given un admin no-root hace `POST /api/admin/empresas/1/telefonos`, when usa su JWT admin, then puede crear un teléfono (no recibe 403).
- Given el JWT generado tras el login, when se decodifica, then no contiene los campos `empresa_id`, `empresa_ruc` ni `empresa_nombre`.
- Given un token de empresa (JWT empresa) con `empresa_id=1`, when accede a `/api/admin/empresas/2/telefonos`, then recibe `403 Forbidden`.
- Given `go build ./...` en `backend/`, when se ejecuta, then compila sin errores de tipos.

## Spec Change Log

- 2026-05-04: implementación backend completada; story movida a `review` tras validar build y tests internos de `domain/http`

## Design Notes

El campo `isAdminJWT` en `panelAdminAccess` es un flag interno que distingue el origen del token sin exponer información adicional. No es parte del JWT ni del dominio público.

Los tokens admin ya existentes (emitidos antes de este cambio) no se ven afectados: el middleware simplemente ignorará los campos que ya no parsea, y como siempre fueron nil, el comportamiento es idéntico.

El `companyID()` de `panelAdminAccess` se usa en `ListUsuarioAdmins` para filtrar usuarios por empresa. Para admin JWT (sin empresa scoping), se debería devolver `(0, false)` para que el handler use `GetAll` en lugar de `GetAllByEmpresa`. Verificar este flujo al implementar.

## Verification

**Commands:**
- `cd backend && go build ./...` — result: OK
- `cd backend && go test ./internal/domain/... ./internal/http/...` — result: OK
- `grep -rn "EmpresaID\|EmpresaRUC\|EmpresaNombre" backend/internal/domain/admin_user.go` — expected: sin resultados
- `grep -n "empresa_ruc\|empresa_nombre\|empresa_id" backend/internal/http/handlers/auth.go` — expected: sin resultados
- `grep -n "empresa_id\|empresa_ruc\|empresa_nombre" backend/internal/http/middleware/auth.go` — expected: sin resultados

## Suggested Review Order

**Struct y generación de token:**
- Confirmar que `AdminJWTClaims` ya no tiene los tres campos y que el token no los incluye.
  [`backend/internal/domain/admin_user.go`](../../backend/internal/domain/admin_user.go)
  [`backend/internal/http/handlers/auth.go`](../../backend/internal/http/handlers/auth.go)

**Middleware:**
- Confirmar que el parsing de esos campos fue removido.
  [`backend/internal/http/middleware/auth.go`](../../backend/internal/http/middleware/auth.go)

**Control de acceso:**
- Confirmar que admin no-root puede acceder a operaciones de empresa y que empresa JWT sigue restringida.
  [`backend/internal/domain/panel_access.go`](../../backend/internal/domain/panel_access.go)
  [`backend/internal/http/admin.go`](../../backend/internal/http/admin.go)
  [`backend/internal/http/handlers/companies.go`](../../backend/internal/http/handlers/companies.go)

---

## Dev Agent Record

### Completion Notes

- ✅ Se creó `PanelAccess` en dominio para unificar el modelo de autorización del panel.
- ✅ `AdminJWTClaims` quedó sin `empresa_id`, `empresa_ruc` ni `empresa_nombre`.
- ✅ Endpoints admin y handlers de empresas migrados al nuevo criterio de acceso.
- ✅ Se removió del backend el registro incompleto de `restore` para no mezclar 2-2 dentro de 2-2.5.
- ✅ Build backend y tests de `internal/domain` + `internal/http` pasaron correctamente.

## File List

- `backend/internal/domain/panel_access.go` — nuevo
- `backend/internal/domain/admin_user.go` — modificado
- `backend/internal/http/handlers/auth.go` — modificado
- `backend/internal/http/middleware/auth.go` — modificado
- `backend/internal/http/admin.go` — modificado
- `backend/internal/http/handlers/companies.go` — modificado
- `backend/internal/http/handlers/api_keys.go` — modificado
- `backend/internal/http/router.go` — modificado
- `backend/internal/http/handlers.go` — modificado
- `backend/internal/domain/empresa_filter.go` — ajustado para no arrastrar `Restore` mezclado
- `backend/internal/http/routes_admin.go` — limpiado para no registrar `restore` incompleto
