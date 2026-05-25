# Acceptance Auditor — Prompt de Revisión

Eres **Acceptance Auditor**. Revisa el diff contra la spec y los documentos de contexto. Verifica: violaciones de acceptance criteria, desviaciones de la intención de la spec, implementación faltante de comportamiento especificado, contradicciones entre restricciones de la spec y el código real. Output findings como lista Markdown. Cada finding: título en una línea, qué AC/restricción viola, y evidencia del diff.

## Spec / Story: 3-1-backend-modulos-efectivos-auth-me

### Acceptance Criteria

**AC1 — Usuario root ve todos los módulos**
Dado un usuario autenticado con `is_root = true`,
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene los slugs de todos los módulos en la tabla `modules` (actualmente: `dashboard`, `companies`, `users`, `roles`, `modules`, `sessions`, `messages`, `broadcasts`).

**AC2 — Usuario con `user_modules` ve solo sus módulos asignados**
Dado un usuario no-root que tiene filas en `user_modules`,
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene exactamente los slugs de sus módulos asignados en `user_modules`.

**AC3 — Usuario sin `user_modules` pero con rol que tiene permisos específicos**
Dado un usuario no-root sin filas en `user_modules` cuyo rol tiene `permissions = ["companies","messages"]`,
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene `["companies", "messages"]`.

**AC4 — Usuario sin `user_modules` y rol con `permissions = ["all"]`**
Dado un usuario no-root sin filas en `user_modules` cuyo rol tiene `permissions = ["all"]`,
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene todos los slugs de módulos.

**AC5 — Fallback mínimo**
Dado un usuario sin `user_modules` y sin rol asignado (o con rol sin permisos),
cuando hace `GET /api/auth/me`,
entonces `allowed_modules` contiene al menos `["dashboard"]`.

**AC6 — Campos existentes no rotos**
La respuesta sigue incluyendo los campos actuales (`id`, `username`, `email`, `role_id`, `is_root`, `activo`) con los mismos nombres y tipos.

**AC7 — Build y tests pasan**
`cd backend && go build ./...` y `cd backend && go test ./...` pasan sin errores.

### Restricciones de la spec
- Los tres nuevos stores pueden ser `nil` si `db == nil`. El código debe tener nil-check.
- `getAllModuleSlugs()` lee desde BD — no hardcodear los 8 slugs.
- No modificar `generateToken()` — `allowed_modules` no debe ir en el JWT.
- Mantener compatibilidad hacia atrás: campos existentes en respuesta no cambian.
- No modificar archivos de storage existentes.

### Archivos afectados según spec
| Archivo | Tipo |
|---|---|
| `backend/internal/http/handlers/auth.go` | UPDATE |
| `backend/internal/http/container.go` | UPDATE |
| `backend/internal/http/handlers/auth_modules_test.go` | NEW |

## Diff

```diff
diff --git a/backend/internal/http/handlers/auth.go b/backend/internal/http/handlers/auth.go
index 66a1620..1a915e2 100644
--- a/backend/internal/http/handlers/auth.go
+++ b/backend/internal/http/handlers/auth.go
@@ -17,18 +17,32 @@ import (
 )
 
 type AuthHandler struct {
-	userStore      *storage.AdminUserStore
-	empresaStore   domain.EmpresaStoreInterface
-	blacklistStore *storage.TokenBlacklistStore
-	jwtConfig      *config.JWTConfig
+	userStore       *storage.AdminUserStore
+	empresaStore    domain.EmpresaStoreInterface
+	blacklistStore  *storage.TokenBlacklistStore
+	jwtConfig       *config.JWTConfig
+	userModuleStore *storage.UserModuleStore
+	roleStore       *storage.RoleStore
+	moduleStore     *storage.ModuleStore
 }
 
-func NewAuthHandler(userStore *storage.AdminUserStore, empresaStore domain.EmpresaStoreInterface, blacklistStore *storage.TokenBlacklistStore, jwtConfig *config.JWTConfig) *AuthHandler {
+func NewAuthHandler(
+	userStore *storage.AdminUserStore,
+	empresaStore domain.EmpresaStoreInterface,
+	blacklistStore *storage.TokenBlacklistStore,
+	jwtConfig *config.JWTConfig,
+	userModuleStore *storage.UserModuleStore,
+	roleStore *storage.RoleStore,
+	moduleStore *storage.ModuleStore,
+) *AuthHandler {
 	return &AuthHandler{
-		userStore:      userStore,
-		empresaStore:   empresaStore,
-		blacklistStore: blacklistStore,
-		jwtConfig:      jwtConfig,
+		userStore:       userStore,
+		empresaStore:    empresaStore,
+		blacklistStore:  blacklistStore,
+		jwtConfig:       jwtConfig,
+		userModuleStore: userModuleStore,
+		roleStore:       roleStore,
+		moduleStore:     moduleStore,
 	}
 }
 
@@ -205,21 +219,74 @@ func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
 		return
 	}
 
+	allowedModules := h.resolveAllowedModules(user)
+
 	response := map[string]interface{}{
 		"ok": true,
 		"user": map[string]interface{}{
-			"id":       user.ID,
-			"username": user.Username,
-			"email":    user.Email,
-			"role_id":  user.RoleID,
-			"is_root":  user.IsRoot,
-			"activo":   user.Activo,
+			"id":              user.ID,
+			"username":        user.Username,
+			"email":           user.Email,
+			"role_id":         user.RoleID,
+			"is_root":         user.IsRoot,
+			"activo":          user.Activo,
+			"allowed_modules": allowedModules,
 		},
 	}
 
 	writeHandlerJSON(w, http.StatusOK, response)
 }
 
+// resolveAllowedModules determina los slugs de módulos efectivos para el usuario.
+// Precedencia: is_root → user_modules override → role.permissions["all"] → role.permissions específicos → fallback dashboard.
+func (h *AuthHandler) resolveAllowedModules(user *domain.AdminUser) []string {
+	if user.IsRoot {
+		return h.getAllModuleSlugs()
+	}
+
+	if h.userModuleStore != nil {
+		modules, err := h.userModuleStore.GetByUserID(user.ID)
+		if err == nil && len(modules) > 0 {
+			slugs := make([]string, 0, len(modules))
+			for _, m := range modules {
+				slugs = append(slugs, m.Slug)
+			}
+			return slugs
+		}
+	}
+
+	if user.RoleID != nil && h.roleStore != nil {
+		role, err := h.roleStore.GetByID(*user.RoleID)
+		if err == nil && role != nil {
+			for _, p := range role.Permissions {
+				if p == "all" {
+					return h.getAllModuleSlugs()
+				}
+			}
+			if len(role.Permissions) > 0 {
+				return role.Permissions
+			}
+		}
+	}
+
+	return []string{"dashboard"}
+}
+
+func (h *AuthHandler) getAllModuleSlugs() []string {
+	if h.moduleStore == nil {
+		return []string{"dashboard"}
+	}
+	modules, err := h.moduleStore.GetAll()
+	if err != nil || len(modules) == 0 {
+		return []string{"dashboard"}
+	}
+	slugs := make([]string, 0, len(modules))
+	for _, m := range modules {
+		slugs = append(slugs, m.Slug)
+	}
+	return slugs
+}
+
 func (h *AuthHandler) generateToken(user *domain.AdminUser) (string, error) {
 	now := time.Now()
 	b := make([]byte, 16)
```

El diff de container.go (de commits anteriores) ya incluye:
- `userModuleStore := storage.NewUserModuleStore(db)`
- `roleStore := storage.NewRoleStore(db)`
- `moduleStore := storage.NewModuleStore(db)`
- `authHandler := handlers.NewAuthHandler(userStore, empresaStore, blacklistStore, jwtCfg, userModuleStore, roleStore, moduleStore)`

Y el archivo `auth_modules_test.go` es nuevo con tests tabla-driven para `resolveAllowedModules`.
