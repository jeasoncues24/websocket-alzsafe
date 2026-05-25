# Story 3-1: Backend — Módulos efectivos en /api/auth/me

**Estado:** review
**Epic padre:** Epic 3 — Módulos Dinámicos desde BD y Perfil de Usuario
**Story ID:** 3.1

---

## Contexto

El endpoint `GET /api/auth/me` (`backend/internal/http/handlers/auth.go:189`) actualmente devuelve solo datos básicos del usuario autenticado (`id`, `username`, `email`, `role_id`, `is_root`, `activo`). No incluye los módulos que el usuario tiene permitidos.

El frontend necesita esta información para construir la navegación dinámicamente (story 3-2 y 3-3). Sin este campo, el sidebar sigue siendo estático.

La infraestructura de permisos ya está completa en base de datos:
- Tabla `modules` — 8 módulos con slug canónico (seeds en migración 016).
- Tabla `roles` — campo `permissions` JSON con slugs o `["all"]`.
- Tabla `user_modules` — override específico por usuario (join table con índices correctos).
- `UserModuleStore.GetByUserID()` ya existe y funciona (`storage/user_module.go:29`).
- `ModuleStore.GetAll()` ya existe (`storage/module.go:72`).
- `RoleStore.GetByID()` ya existe (`storage/role.go:47`).

El único trabajo de esta story es conectar esas piezas en el handler `Me()` y pasar los stores al `AuthHandler`.

---

## Objetivo

Extender `GET /api/auth/me` para incluir en su respuesta el array `allowed_modules` con los slugs de módulos que el usuario autenticado tiene acceso efectivo, siguiendo la lógica de precedencia definida en el epic.

---

## Usuario / actor afectado

- **Frontend de wsapi** (consumer del endpoint): recibe el array y lo usa para renderizar el menú dinámico.
- **Todos los usuarios del panel** (indirecto): su menú de navegación reflejará sus permisos reales.

---

## Alcance

- Agregar `userModuleStore`, `roleStore` y `moduleStore` al struct `AuthHandler`.
- Actualizar `NewAuthHandler` para aceptar estos tres stores.
- Actualizar `container.go` para instanciarlos y pasarlos.
- Implementar la lógica de `allowed_modules` en el handler `Me()`.
- Mantener 100% de compatibilidad hacia atrás: los campos existentes en la respuesta no cambian.

## Fuera de alcance

- Cambios en el JWT (no se modifican los claims).
- Middleware de autorización por módulo (story 3-3).
- Frontend (stories 3-2 y 3-3).
- Cambios en `user_modules`, `roles` o `modules` stores existentes.

---

## Acceptance Criteria

**AC1 — Usuario root ve todos los módulos**
Dado un usuario autenticado con `is_root = true`,
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene los slugs de todos los módulos en la tabla `modules` (actualmente: `dashboard`, `companies`, `users`, `roles`, `modules`, `sessions`, `messages`, `broadcasts`).

**AC2 — Usuario con `user_modules` ve solo sus módulos asignados**
Dado un usuario no-root que tiene filas en `user_modules` (ej. soporte: `companies`, `messages`, `sessions`, `broadcasts`),
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene exactamente los slugs de sus módulos asignados en `user_modules`, sin más ni menos.

**AC3 — Usuario sin `user_modules` pero con rol que tiene permisos específicos**
Dado un usuario no-root sin filas en `user_modules` cuyo rol tiene `permissions = ["companies","messages"]`,
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene `["companies", "messages"]`.

**AC4 — Usuario sin `user_modules` y rol con `permissions = ["all"]`**
Dado un usuario no-root sin filas en `user_modules` cuyo rol tiene `permissions = ["all"]`,
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene todos los slugs de módulos (igual que un usuario root).

**AC5 — Fallback mínimo**
Dado un usuario sin `user_modules` y sin rol asignado (o con rol sin permisos),
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene al menos `["dashboard"]`.

**AC6 — Campos existentes no rotos**
La respuesta sigue incluyendo los campos actuales (`id`, `username`, `email`, `role_id`, `is_root`, `activo`) con los mismos nombres y tipos.

**AC7 — Build y tests pasan**
`cd backend && go build ./...` y `cd backend && go test ./...` pasan sin errores.

---

## Tareas técnicas

### T1 — Extender `AuthHandler` con los nuevos stores
**Archivo:** `backend/internal/http/handlers/auth.go`

Agregar al struct `AuthHandler`:
```go
type AuthHandler struct {
    userStore       *storage.AdminUserStore
    empresaStore    domain.EmpresaStoreInterface
    blacklistStore  *storage.TokenBlacklistStore
    jwtConfig       *config.JWTConfig
    userModuleStore *storage.UserModuleStore  // NUEVO
    roleStore       *storage.RoleStore        // NUEVO
    moduleStore     *storage.ModuleStore      // NUEVO
}
```

Actualizar `NewAuthHandler`:
```go
func NewAuthHandler(
    userStore *storage.AdminUserStore,
    empresaStore domain.EmpresaStoreInterface,
    blacklistStore *storage.TokenBlacklistStore,
    jwtConfig *config.JWTConfig,
    userModuleStore *storage.UserModuleStore,  // NUEVO
    roleStore *storage.RoleStore,              // NUEVO
    moduleStore *storage.ModuleStore,          // NUEVO
) *AuthHandler {
    return &AuthHandler{
        userStore:       userStore,
        empresaStore:    empresaStore,
        blacklistStore:  blacklistStore,
        jwtConfig:       jwtConfig,
        userModuleStore: userModuleStore,
        roleStore:       roleStore,
        moduleStore:     moduleStore,
    }
}
```

### T2 — Implementar `resolveAllowedModules` (función privada en handlers/auth.go)

Crear función helper que encapsula la lógica de precedencia:

```go
// resolveAllowedModules determina los slugs de módulos efectivos para el usuario.
// Precedencia: is_root → user_modules → role.permissions["all"] → role.permissions específicos → fallback dashboard
func (h *AuthHandler) resolveAllowedModules(user *domain.AdminUser) []string {
    // 1. Root: todos los módulos
    if user.IsRoot {
        return h.getAllModuleSlugs()
    }

    // 2. user_modules override: si el usuario tiene asignaciones específicas
    if h.userModuleStore != nil {
        modules, err := h.userModuleStore.GetByUserID(user.ID)
        if err == nil && len(modules) > 0 {
            slugs := make([]string, 0, len(modules))
            for _, m := range modules {
                slugs = append(slugs, m.Slug)
            }
            return slugs
        }
    }

    // 3. Rol con permissions
    if user.RoleID != nil && h.roleStore != nil {
        role, err := h.roleStore.GetByID(*user.RoleID)
        if err == nil && role != nil {
            for _, p := range role.Permissions {
                if p == "all" {
                    return h.getAllModuleSlugs()
                }
            }
            if len(role.Permissions) > 0 {
                return role.Permissions
            }
        }
    }

    // 4. Fallback mínimo
    return []string{"dashboard"}
}

func (h *AuthHandler) getAllModuleSlugs() []string {
    if h.moduleStore == nil {
        return []string{"dashboard"}
    }
    modules, err := h.moduleStore.GetAll()
    if err != nil || len(modules) == 0 {
        return []string{"dashboard"}
    }
    slugs := make([]string, 0, len(modules))
    for _, m := range modules {
        slugs = append(slugs, m.Slug)
    }
    return slugs
}
```

### T3 — Actualizar handler `Me()` para incluir `allowed_modules`

En `Me()` (línea ~189), después de obtener el usuario, agregar:

```go
allowedModules := h.resolveAllowedModules(user)

response := map[string]interface{}{
    "ok": true,
    "user": map[string]interface{}{
        "id":              user.ID,
        "username":        user.Username,
        "email":           user.Email,
        "role_id":         user.RoleID,
        "is_root":         user.IsRoot,
        "activo":          user.Activo,
        "allowed_modules": allowedModules,  // NUEVO
    },
}
```

### T4 — Actualizar `container.go`

**Archivo:** `backend/internal/http/container.go` (línea ~91)

Instanciar los nuevos stores y pasarlos a `NewAuthHandler`:

```go
userStore := storage.NewAdminUserStore(db)
blacklistStore := storage.NewTokenBlacklistStore(db)
userModuleStore := storage.NewUserModuleStore(db)  // NUEVO
roleStore := storage.NewRoleStore(db)              // NUEVO
moduleStore := storage.NewModuleStore(db)          // NUEVO

authHandler := handlers.NewAuthHandler(
    userStore, empresaStore, blacklistStore, jwtCfg,
    userModuleStore, roleStore, moduleStore,        // NUEVO
)
```

> **Nota:** `NewUserModuleStore`, `NewRoleStore` y `NewModuleStore` ya existen en sus respectivos archivos de storage. No hay que crearlos.

> **Nil safety:** los tres nuevos stores se instancian solo si `db != nil`. Si la DB no está disponible, pasarlos como `nil` es válido — `resolveAllowedModules` hace nil-check antes de usarlos y cae al fallback `["dashboard"]`. Esto mantiene el comportamiento actual sin DB.

### T5 — Test unitario

**Archivo nuevo:** `backend/internal/http/handlers/auth_modules_test.go`

Test tabla-driven para `resolveAllowedModules`. Usar mocks simples (structs que implementan la interfaz mínima necesaria) o instanciar con stores en memoria si aplica el patrón del proyecto.

Casos mínimos a cubrir:
- Usuario root → todos los módulos
- Usuario con `user_modules` → exactamente esos slugs
- Usuario sin `user_modules`, rol con permissions específicos → esos slugs
- Usuario sin `user_modules`, rol con `permissions = ["all"]` → todos los módulos
- Usuario sin rol → `["dashboard"]`
- Stores `nil` (sin DB) → `["dashboard"]` (no panic)

---

## Archivos afectados

| Archivo | Tipo de cambio |
|---|---|
| `backend/internal/http/handlers/auth.go` | UPDATE — struct, constructor, Me(), +2 helpers |
| `backend/internal/http/container.go` | UPDATE — instanciar 3 stores, actualizar llamada a NewAuthHandler |
| `backend/internal/http/handlers/auth_modules_test.go` | NEW — tests de resolveAllowedModules |

**No modificar:**
- `backend/internal/storage/user_module.go` — sin cambios
- `backend/internal/storage/module.go` — sin cambios
- `backend/internal/storage/role.go` — sin cambios
- `backend/internal/domain/admin_user.go` — sin cambios
- `backend/internal/http/routes_admin.go` — sin cambios (la ruta GET /api/auth/me ya existe)

---

## Pruebas requeridas

1. **Tests unitarios** (`auth_modules_test.go`): todos los casos de AC1–AC5 cubiertos tabla-driven.
2. **Build completo**: `cd backend && go build ./...` sin errores.
3. **Suite completa**: `cd backend && go test ./...` verde.
4. **Verificación manual** (opcional si los tests pasan): hacer login con un usuario de soporte y llamar `GET /api/auth/me`; verificar que `allowed_modules` contiene `["companies","messages","sessions","broadcasts"]`.

---

## Riesgos y edge cases

| Edge case | Manejo |
|---|---|
| `user_modules` vacío Y rol es `nil` | Fallback `["dashboard"]` |
| `role.Permissions = ["all"]` | Expandir a todos los slugs de la tabla `modules` |
| `moduleStore.GetAll()` falla (DB down) | Fallback `["dashboard"]` — nunca panic |
| Usuario root con `user_modules` vacío | La condición `is_root` tiene prioridad 1 — ve todo |
| Nuevos módulos añadidos a BD en el futuro | Root y roles con `"all"` los recibirán automáticamente; usuarios con `user_modules` específicos no |
| Stores `nil` si `db == nil` (sin BD configurada) | Nil-check en `resolveAllowedModules` — retorna `["dashboard"]` |

**Seguridad (golang-security):**
- Trust boundary: `user.ID` viene del JWT claims validado por `adminStack` middleware, no de input del usuario — seguro.
- El endpoint es GET sin body — superficie de ataque mínima.
- `allowed_modules` solo expone lo que el usuario ya está autorizado a ver — no hay information disclosure.
- No se aceptan inputs del usuario en esta story.

**SQL (sql-optimization):**
- `GetByUserID`: JOIN sobre `idx_user_modules_user (user_id)` — índice ya existe, O(log n) en ~8-50 filas.
- `GetAll()` en modules: full scan de 8 filas estáticas — óptimo sin índice adicional.
- `GetByID()` en roles: PK lookup — óptimo.
- No se requieren migraciones ni índices nuevos.

---

## Dependencias y bloqueos

- **Ningún bloqueo**: todos los stores y métodos ya existen en el codebase.
- **Prerequisito para**: story 3-2 (frontend consume `allowed_modules` de este endpoint).

---

## Notas de implementación

### Rama de Git obligatoria

Verificar antes de escribir código:
```bash
git branch --show-current
```
El archivo `docs/bmad-project-rules.md` asigna `feature/security` al Epic 3 (el nombre es del epic anterior pero la rama se mantiene para este epic). Si la rama no existe aún, crearla desde `v1`:
```bash
git checkout -b feature/security
```

### Patrón de respuesta existente

El handler `Me()` actualmente construye la respuesta como `map[string]interface{}`. Mantener ese patrón — no introducir structs tipados para la respuesta en esta story (eso sería un refactor adicional fuera de alcance).

### nil-safety obligatoria

Los tres stores nuevos (`userModuleStore`, `roleStore`, `moduleStore`) pueden ser `nil` si la aplicación corre sin BD. Todos los accesos deben estar protegidos con nil-check. El código de ejemplo en T2 ya lo contempla.

### Módulos en seeds vs. en BD

Los 8 módulos del seed (`dashboard`, `companies`, `users`, `roles`, `modules`, `sessions`, `messages`, `broadcasts`) son los únicos existentes actualmente. `getAllModuleSlugs()` los lee desde BD para ser extensible — no hardcodearlos.

### No tocar `generateToken()`

La función `generateToken()` en `auth.go` no debe incluir `allowed_modules` en el JWT. El módulo de permisos debe resolverse en tiempo de request (no en el token) para reflejar cambios de BD sin requerir re-login.
