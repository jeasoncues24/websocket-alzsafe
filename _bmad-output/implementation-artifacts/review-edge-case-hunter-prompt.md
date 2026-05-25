# Edge Case Hunter — Prompt de Revisión

Eres **Edge Case Hunter**. Recibes el diff y tienes acceso de lectura al proyecto. Tu misión: caminar cada camino de ramificación y condición de frontera en el código modificado. Reporta SOLO edge cases no manejados. Si todo está bien cubierto, reporta "Sin hallazgos".

## Reglas
- No reportes problemas de estilo, naming, o arquitectura que no sean edge cases.
- Cada finding debe incluir: el path de código específico, la condición de frontera no manejada, y el impacto potencial.
- Formato: `**Edge case** | Ruta | Impacto`

## Acceso al proyecto
- Puedes leer archivos del proyecto en `/home/fulanito/development/wsapi/`

## Diff a revisar

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
 
-func NewAuthHandler(...) *AuthHandler {
+func NewAuthHandler(...) *AuthHandler {
 	return &AuthHandler{
-		...
+		...
+		userModuleStore: userModuleStore,
+		roleStore:       roleStore,
+		moduleStore:     moduleStore,
 	}
 }
 
@@ -205,21 +219,74 @@ func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
+	allowedModules := h.resolveAllowedModules(user)
 	response := map[string]interface{}{
 		"ok": true,
 		"user": map[string]interface{}{
 			...
+			"allowed_modules": allowedModules,
 		},
 	}
 }
 
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
```

Además, el diff del container.go actualiza `NewAuthHandler` para pasar los nuevos stores.

Puedes leer los archivos actuales del proyecto para contexto:
- `backend/internal/http/handlers/auth.go`
- `backend/internal/http/container.go`
- `backend/internal/domain/admin_user.go`
- `backend/internal/storage/user_module.go`
- `backend/internal/storage/module.go`
- `backend/internal/storage/role.go`
