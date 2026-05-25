# Blind Hunter — Prompt de Revisión

Eres **Blind Hunter**, un revisor adversarial. NO tienes contexto del proyecto ni spec. Solo recibes el diff unificado. Realiza una revisión adversarial (cynical review) del siguiente diff. Busca: bugs, vulnerabilidades, code smells, malas prácticas, lógica incorrecta, problemas de concurrencia, inconsistencias, y cualquier cosa que pueda salir mal. NO asumas nada del contexto del proyecto. Usa únicamente el diff.

## Reglas
- No tengo contexto del proyecto, spec, ni requisitos.
- Produzco findings como una **lista Markdown** con formato:
  `**Título del finding** | Severidad (critical/high/medium/low) | Evidencia desde el diff`
- Devuelvo SOLO la lista de findings, sin introducción ni resumen.

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
diff --git a/backend/internal/http/handlers/auth_modules_test.go b/backend/internal/http/handlers/auth_modules_test.go
new file mode 100644
index 0000000..ec92d7e
--- /dev/null
+++ b/backend/internal/http/handlers/auth_modules_test.go
@@ -0,0 +1,254 @@
+package http
+
+import (
+	"database/sql"
+	"testing"
+
+	_ "modernc.org/sqlite"
+
+	"wsapi/internal/domain"
+	"wsapi/internal/storage"
+)
+
+func setupAuthModulesTestDB(t *testing.T) *sql.DB {
+	t.Helper()
+	db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
+	if err != nil {
+		t.Fatalf("abrir sqlite: %v", err)
+	}
+	t.Cleanup(func() { db.Close() })
+
+	schema := `
+	CREATE TABLE IF NOT EXISTS roles (
+		id INTEGER PRIMARY KEY AUTOINCREMENT,
+		name TEXT UNIQUE NOT NULL,
+		description TEXT,
+		is_root INTEGER NOT NULL DEFAULT 0,
+		permissions TEXT,
+		created_by INTEGER NULL,
+		updated_by INTEGER NULL,
+		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
+		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
+	);
+	CREATE TABLE IF NOT EXISTS admin_users (
+		id INTEGER PRIMARY KEY AUTOINCREMENT,
+		username TEXT UNIQUE NOT NULL,
+		password_hash TEXT NOT NULL,
+		email TEXT,
+		role_id INTEGER NULL,
+		activo INTEGER NOT NULL DEFAULT 1,
+		created_by INTEGER NULL,
+		updated_by INTEGER NULL,
+		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
+		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
+		last_login_at TIMESTAMP NULL,
+		FOREIGN KEY (role_id) REFERENCES roles(id)
+	);
+	CREATE TABLE IF NOT EXISTS modules (
+		id INTEGER PRIMARY KEY AUTOINCREMENT,
+		name TEXT UNIQUE NOT NULL,
+		slug TEXT UNIQUE,
+		description TEXT,
+		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
+		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
+	);
+	CREATE TABLE IF NOT EXISTS user_modules (
+		id INTEGER PRIMARY KEY AUTOINCREMENT,
+		user_id INTEGER NOT NULL,
+		module_id INTEGER NOT NULL,
+		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
+		UNIQUE (user_id, module_id),
+		FOREIGN KEY (user_id) REFERENCES admin_users(id) ON DELETE CASCADE,
+		FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE
+	);`
+
+	if _, err := db.Exec(schema); err != nil {
+		t.Fatalf("crear esquema: %v", err)
+	}
+
+	// Módulos canónicos (8 slugs del seed)
+	mods := []string{"dashboard", "companies", "users", "roles", "modules", "sessions", "messages", "broadcasts"}
+	for _, slug := range mods {
+		if _, err := db.Exec(`INSERT INTO modules (name, slug, description) VALUES (?, ?, ?)`, slug, slug, slug); err != nil {
+			t.Fatalf("insertar módulo %s: %v", slug, err)
+		}
+	}
+
+	return db
+}
+
+func makeAuthHandlerForTest(db *sql.DB) *AuthHandler {
+	if db == nil {
+		return &AuthHandler{}
+	}
+	return &AuthHandler{
+		userModuleStore: storage.NewUserModuleStore(db),
+		roleStore:       storage.NewRoleStore(db),
+		moduleStore:     storage.NewModuleStore(db),
+	}
+}
+
+func insertTestRole(t *testing.T, db *sql.DB, name string, isRoot bool, permissions string) int64 {
+	t.Helper()
+	res, err := db.Exec(`INSERT INTO roles (name, description, is_root, permissions) VALUES (?, ?, ?, ?)`, name, name, isRoot, permissions)
+	if err != nil {
+		t.Fatalf("insertar rol %s: %v", name, err)
+	}
+	id, _ := res.LastInsertId()
+	return id
+}
+
+func insertTestUser(t *testing.T, db *sql.DB, username string, isRoot bool, roleID *int64) int64 {
+	t.Helper()
+	res, err := db.Exec(`INSERT INTO admin_users (username, password_hash, email, role_id, activo) VALUES (?, 'hash', 'test@test.com', ?, 1)`,
+		username, roleID)
+	if err != nil {
+		t.Fatalf("insertar usuario %s: %v", username, err)
+	}
+	id, _ := res.LastInsertId()
+	return id
+}
+
+func assignUserModules(t *testing.T, db *sql.DB, userID int64, slugs []string) {
+	t.Helper()
+	for _, slug := range slugs {
+		var modID int64
+		if err := db.QueryRow(`SELECT id FROM modules WHERE slug = ?`, slug).Scan(&modID); err != nil {
+			t.Fatalf("módulo %s no encontrado: %v", slug, err)
+		}
+		if _, err := db.Exec(`INSERT INTO user_modules (user_id, module_id) VALUES (?, ?)`, userID, modID); err != nil {
+			t.Fatalf("asignar módulo %s a usuario %d: %v", slug, userID, err)
+		}
+	}
+}
+
+func slugsToSet(slugs []string) map[string]bool {
+	m := make(map[string]bool, len(slugs))
+	for _, s := range slugs {
+		m[s] = true
+	}
+	return m
+}
+
+func TestResolveAllowedModules(t *testing.T) {
+	allSlugs := []string{"dashboard", "companies", "users", "roles", "modules", "sessions", "messages", "broadcasts"}
+
+	t.Run("AC1_root_ve_todos_los_modulos", func(t *testing.T) {
+		db := setupAuthModulesTestDB(t)
+		h := makeAuthHandlerForTest(db)
+		roleID := insertTestRole(t, db, "admin", true, `["all"]`)
+		userID := insertTestUser(t, db, "root_user", true, &roleID)
+		_ = userID
+
+		user := &domain.AdminUser{ID: userID, IsRoot: true, RoleID: &roleID}
+		got := h.resolveAllowedModules(user)
+
+		if len(got) != len(allSlugs) {
+			t.Errorf("esperaba %d módulos, got %d: %v", len(allSlugs), len(got), got)
+		}
+		gotSet := slugsToSet(got)
+		for _, s := range allSlugs {
+			if !gotSet[s] {
+				t.Errorf("falta slug %q en resultado root", s)
+			}
+		}
+	})
+
+	t.Run("AC2_user_modules_override", func(t *testing.T) {
+		db := setupAuthModulesTestDB(t)
+		h := makeAuthHandlerForTest(db)
+		roleID := insertTestRole(t, db, "soporte", false, `["companies","messages","sessions","broadcasts"]`)
+		userID := insertTestUser(t, db, "soporte_user", false, &roleID)
+		assigned := []string{"companies", "messages", "sessions", "broadcasts"}
+		assignUserModules(t, db, userID, assigned)
+
+		user := &domain.AdminUser{ID: userID, IsRoot: false, RoleID: &roleID}
+		got := h.resolveAllowedModules(user)
+
+		if len(got) != len(assigned) {
+			t.Errorf("esperaba %d módulos, got %d: %v", len(assigned), len(got), got)
+		}
+		gotSet := slugsToSet(got)
+		for _, s := range assigned {
+			if !gotSet[s] {
+				t.Errorf("falta slug %q en resultado user_modules", s)
+			}
+		}
+		// No debe tener módulos fuera de los asignados
+		for _, s := range got {
+			if !slugsToSet(assigned)[s] {
+				t.Errorf("slug inesperado %q en resultado", s)
+			}
+		}
+	})
+
+	t.Run("AC3_sin_user_modules_rol_permisos_especificos", func(t *testing.T) {
+		db := setupAuthModulesTestDB(t)
+		h := makeAuthHandlerForTest(db)
+		roleID := insertTestRole(t, db, "limitado", false, `["companies","messages"]`)
+		userID := insertTestUser(t, db, "limitado_user", false, &roleID)
+
+		user := &domain.AdminUser{ID: userID, IsRoot: false, RoleID: &roleID}
+		got := h.resolveAllowedModules(user)
+
+		expected := []string{"companies", "messages"}
+		if len(got) != len(expected) {
+			t.Errorf("esperaba %d módulos, got %d: %v", len(expected), len(got), got)
+		}
+		gotSet := slugsToSet(got)
+		for _, s := range expected {
+			if !gotSet[s] {
+				t.Errorf("falta slug %q en resultado de rol", s)
+			}
+		}
+	})
+
+	t.Run("AC4_sin_user_modules_rol_permissions_all", func(t *testing.T) {
+		db := setupAuthModulesTestDB(t)
+		h := makeAuthHandlerForTest(db)
+		roleID := insertTestRole(t, db, "administracion", false, `["all"]`)
+		userID := insertTestUser(t, db, "admin_user", false, &roleID)
+
+		user := &domain.AdminUser{ID: userID, IsRoot: false, RoleID: &roleID}
+		got := h.resolveAllowedModules(user)
+
+		if len(got) != len(allSlugs) {
+			t.Errorf("esperaba %d módulos (all), got %d: %v", len(allSlugs), len(got), got)
+		}
+	})
+
+	t.Run("AC5_fallback_sin_rol", func(t *testing.T) {
+		db := setupAuthModulesTestDB(t)
+		h := makeAuthHandlerForTest(db)
+		userID := insertTestUser(t, db, "sin_rol_user", false, nil)
+
+		user := &domain.AdminUser{ID: userID, IsRoot: false, RoleID: nil}
+		got := h.resolveAllowedModules(user)
+
+		if len(got) != 1 || got[0] != "dashboard" {
+			t.Errorf("esperaba [dashboard], got %v", got)
+		}
+	})
+
+	t.Run("nil_stores_no_panic_fallback_dashboard", func(t *testing.T) {
+		h := &AuthHandler{} // todos los stores nil
+		user := &domain.AdminUser{ID: 1, IsRoot: false, RoleID: nil}
+		got := h.resolveAllowedModules(user)
+
+		if len(got) != 1 || got[0] != "dashboard" {
+			t.Errorf("esperaba [dashboard] con stores nil, got %v", got)
+		}
+	})
+
+	t.Run("nil_stores_root_fallback_dashboard", func(t *testing.T) {
+		h := &AuthHandler{} // moduleStore nil
+		user := &domain.AdminUser{ID: 1, IsRoot: true}
+		got := h.resolveAllowedModules(user)
+
+		if len(got) != 1 || got[0] != "dashboard" {
+			t.Errorf("esperaba [dashboard] con moduleStore nil, got %v", got)
+		}
+	})
+}
+```
