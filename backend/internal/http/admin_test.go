package http

import (
	"database/sql"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

func TestAdminGetCompanyPhone(t *testing.T) {
	db := newAdminPhoneTestDB(t)
	store := storage.NewTelefonoStore(db)
	phoneID := insertAdminPhone(t, db, 1, "+51", "999888777", "+51999888777")

	h := &AdminHandler{telefonoStore: store}
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/admin/telefonos/1", nil)
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{
		UserID:   1,
		Username: "root",
		Rol:      domain.RoleSuperAdmin,
		IsRoot:   true,
	}))
	rr := httptest.NewRecorder()

	h.GetCompanyPhone(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp domain.TelefonoResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if !resp.OK || resp.Telefono == nil || resp.Telefono.ID != phoneID {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAdminStartCompanyPhoneConnection(t *testing.T) {
	db := newAdminPhoneTestDB(t)
	store := storage.NewTelefonoStore(db)
	insertAdminPhone(t, db, 1, "+51", "999888777", "+51999888777")

	h := &AdminHandler{
		telefonoStore: store,
		sessionStore:  storage.NewSessionStore(),
		manager:       whatsapp.NewManager(),
	}
	req := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/telefonos/1/connect", nil)
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{
		UserID:   1,
		Username: "admin",
		Rol:      domain.RoleAdmin,
		IsRoot:   false,
	}))
	rr := httptest.NewRecorder()

	h.StartCompanyPhoneConnection(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if resp["status"] != "initializing" {
		t.Fatalf("expected initializing status, got %+v", resp)
	}
}



func TestBuildAdminSessionDiagnosticMismatch(t *testing.T) {
	phone := &domain.Telefono{
		ID:             11,
		EmpresaID:      2,
		NumeroCompleto: "+51999888777",
		Status:         domain.TelefonoStatusActive,
	}

	diag := buildAdminSessionDiagnostic(phone, false)
	if !diag.Mismatch {
		t.Fatalf("expected mismatch to be true")
	}
	if diag.MismatchReason != "db_active_runtime_disconnected" {
		t.Fatalf("unexpected mismatch reason: %s", diag.MismatchReason)
	}
	if diag.RecommendedAction != "reanudar_conexion" {
		t.Fatalf("unexpected action: %s", diag.RecommendedAction)
	}
}

func TestBuildAdminSessionDiagnosticHealthy(t *testing.T) {
	phone := &domain.Telefono{
		ID:             12,
		EmpresaID:      3,
		NumeroCompleto: "51977596225",
		Status:         domain.TelefonoStatusActive,
	}

	diag := buildAdminSessionDiagnostic(phone, true)
	if diag.Mismatch {
		t.Fatalf("expected mismatch false")
	}
	if diag.StatusRuntime != "connected" {
		t.Fatalf("expected connected runtime status, got %s", diag.StatusRuntime)
	}
	if diag.RecommendedAction != "none" {
		t.Fatalf("unexpected action: %s", diag.RecommendedAction)
	}
}

func TestExtractPanelUserIDSupportsUsuarioAdminPath(t *testing.T) {
	id, err := extractPanelUserID("/api/admin/usuario_admin/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

// TestListUsuarioAdminsUsesEmpresaScope removed - empresa_id ya no existe en admin_users

func TestListUsuarioAdminsUnauthorizedReturnsJSONError(t *testing.T) {
	db := newAdminUserScopeTestDB(t)
	store := storage.NewAdminUserStore(db)
	h := &AdminHandler{userStore: store}
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/admin/usuario_admin?page=1&limit=20", nil)
	rr := httptest.NewRecorder()

	h.ListUsuarioAdmins(rr, req)

	if rr.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected json content-type, got %q", got)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp["ok"] != false || resp["error"] == "" {
		t.Fatalf("unexpected error payload: %+v", resp)
	}
}

func TestDeleteUsuarioAdminHardDeletesWhenNoDependencies(t *testing.T) {
	db := newAdminUserScopeTestDB(t)
	store := storage.NewAdminUserStore(db)
	userID := insertAdminUser(t, db, "alice", "$2a$10$abcdefghijklmnopqrstuv", "alice@a.com", 1)

	action, err := store.DeleteWithPolicy(userID)
	if err != nil {
		t.Fatalf("delete policy: %v", err)
	}
	if action != "deleted" {
		t.Fatalf("expected deleted action, got %s", action)
	}

	user, err := store.GetByID(userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user != nil {
		t.Fatalf("expected user removed, got %+v", user)
	}
}

func TestDeleteUsuarioAdminDisablesWhenDependenciesExist(t *testing.T) {
	db := newAdminUserScopeTestDB(t)
	store := storage.NewAdminUserStore(db)
	userID := insertAdminUser(t, db, "alice", "$2a$10$abcdefghijklmnopqrstuv", "alice@a.com", 1)
	if _, err := db.Exec(`INSERT INTO user_modules (user_id, module_id) VALUES (?, 1)`, userID); err != nil {
		t.Fatalf("seed dependency: %v", err)
	}

	action, err := store.DeleteWithPolicy(userID)
	if err != nil {
		t.Fatalf("delete policy: %v", err)
	}
	if action != "disabled" {
		t.Fatalf("expected disabled action, got %s", action)
	}

	user, err := store.GetByID(userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user == nil || user.Activo {
		t.Fatalf("expected disabled user, got %+v", user)
	}
}

func TestCreateAndUpdateUsuarioAdmin(t *testing.T) {
	db := newAdminUserScopeTestDB(t)
	store := storage.NewAdminUserStore(db)
	h := &AdminHandler{userStore: store}

	createReq := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/usuario_admin", strings.NewReader(`{"username":"newadmin","password":"Secret123!","email":"new@a.com","role_id":1}`))
	createReq = createReq.WithContext(domain.WithAdminJWTClaims(createReq.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	createRR := httptest.NewRecorder()

	h.CreateUsuarioAdmin(createRR, createReq)

	if createRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected create 200, got %d body=%s", createRR.Code, createRR.Body.String())
	}

	var createResp map[string]any
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("invalid create response: %v", err)
	}
	usuarioAdmin, ok := createResp["user"].(map[string]any)
	if !ok {
		t.Fatalf("missing user payload: %+v", createResp)
	}
	userID := int64(usuarioAdmin["id"].(float64))
	stored, err := store.GetByID(userID)
	if err != nil {
		t.Fatalf("get created user: %v", err)
	}
	if stored == nil {
		t.Fatalf("created user not found")
	}
	if stored.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("Secret123!")) != nil {
		t.Fatalf("password hash not stored correctly")
	}

	updateReq := httptest.NewRequest(stdhttp.MethodPut, "/api/admin/usuario_admin/"+strconv.FormatInt(userID, 10), strings.NewReader(`{"email":"updated@a.com","role_id":2,"is_active":false}`))
	updateReq = updateReq.WithContext(domain.WithAdminJWTClaims(updateReq.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	updateRR := httptest.NewRecorder()

	h.UpdateUsuarioAdmin(updateRR, updateReq)

	if updateRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected update 200, got %d body=%s", updateRR.Code, updateRR.Body.String())
	}

	updated, err := store.GetByID(userID)
	if err != nil {
		t.Fatalf("get updated user: %v", err)
	}
	if updated == nil || updated.Email != "updated@a.com" || updated.RoleID == nil || *updated.RoleID != 2 || updated.Activo {
		t.Fatalf("unexpected updated user: %+v", updated)
	}
}

func TestCreateRoleValidatesPermissionsAndRejectsDuplicates(t *testing.T) {
	db := newAdminRoleTestDB(t)
	store := storage.NewRoleStore(db)
	h := &AdminHandler{roleStore: store, moduleStore: storage.NewModuleStore(db)}

	createReq := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/roles", strings.NewReader(`{"name":"auditor","description":"Auditor","permissions":["messages","broadcasts:read"]}`))
	createReq = createReq.WithContext(domain.WithAdminJWTClaims(createReq.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	createRR := httptest.NewRecorder()

	h.CreateRole(createRR, createReq)

	if createRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected create 200, got %d body=%s", createRR.Code, createRR.Body.String())
	}

	var createResp map[string]any
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("invalid create response: %v", err)
	}
	rolePayload, ok := createResp["role"].(map[string]any)
	if !ok {
		t.Fatalf("missing role payload: %+v", createResp)
	}
	roleID := int64(rolePayload["id"].(float64))
	created, err := store.GetByID(roleID)
	if err != nil {
		t.Fatalf("get created role: %v", err)
	}
	if created == nil || created.Name != "auditor" || len(created.Permissions) != 2 {
		t.Fatalf("unexpected created role: %+v", created)
	}

	dupReq := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/roles", strings.NewReader(`{"name":"auditor","description":"Duplicado","permissions":["messages"]}`))
	dupReq = dupReq.WithContext(domain.WithAdminJWTClaims(dupReq.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	dupRR := httptest.NewRecorder()

	h.CreateRole(dupRR, dupReq)

	if dupRR.Code != stdhttp.StatusConflict {
		t.Fatalf("expected duplicate conflict, got %d body=%s", dupRR.Code, dupRR.Body.String())
	}
	var dupResp map[string]any
	if err := json.Unmarshal(dupRR.Body.Bytes(), &dupResp); err != nil {
		t.Fatalf("expected duplicate response json: %v", err)
	}
	if dupResp["ok"] != false || dupResp["error"] != "el rol ya existe" {
		t.Fatalf("unexpected duplicate payload: %+v", dupResp)
	}
}

func TestUpdateRoleRejectsDroppingRootFlag(t *testing.T) {
	db := newAdminRoleTestDB(t)
	store := storage.NewRoleStore(db)
	rootRole := &domain.Role{Name: "super_admin", Description: "Super Admin", IsRoot: true, Permissions: []string{"all"}}
	if _, err := store.Create(rootRole); err != nil {
		t.Fatalf("seed root role: %v", err)
	}

	h := &AdminHandler{roleStore: store, moduleStore: storage.NewModuleStore(db)}
	req := httptest.NewRequest(stdhttp.MethodPut, "/api/admin/roles/1", strings.NewReader(`{"name":"super_admin","description":"Root","is_root":false,"permissions":["all"]}`))
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	rr := httptest.NewRecorder()

	h.UpdateRole(rr, req)

	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestDeleteRoleRejectsInUseRoles(t *testing.T) {
	db := newAdminRoleTestDB(t)
	store := storage.NewRoleStore(db)
	role := &domain.Role{Name: "support", Description: "Support", Permissions: []string{"messages"}}
	roleID, err := store.Create(role)
	if err != nil {
		t.Fatalf("seed role: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO admin_users (username, password_hash, email, role_id, activo) VALUES (?, ?, ?, ?, 1)`, "alice", "$2a$10$abcdefghijklmnopqrstuv", "alice@a.com", roleID); err != nil {
		t.Fatalf("seed user dependency: %v", err)
	}

	h := &AdminHandler{roleStore: store, moduleStore: storage.NewModuleStore(db)}
	req := httptest.NewRequest(stdhttp.MethodDelete, "/api/admin/roles/1", nil)
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	rr := httptest.NewRecorder()

	h.DeleteRole(rr, req)

	if rr.Code != stdhttp.StatusConflict {
		t.Fatalf("expected conflict, got %d body=%s", rr.Code, rr.Body.String())
	}
	remaining, err := store.GetByID(roleID)
	if err != nil {
		t.Fatalf("get role after delete attempt: %v", err)
	}
	if remaining == nil {
		t.Fatalf("expected role to remain after failed delete")
	}
}

func TestGetUsuarioAdminModulesReturnsAssignedModules(t *testing.T) {
	db := newAdminRoleTestDB(t)
	userStore := storage.NewAdminUserStore(db)
	userID := insertAdminUser(t, db, "alice", "$2a$10$abcdefghijklmnopqrstuv", "alice@a.com", 1)
	if err := storage.NewUserModuleStore(db).AssignModules(userID, []int64{1, 3}); err != nil {
		t.Fatalf("seed modules: %v", err)
	}

	h := &AdminHandler{db: db, userStore: userStore}
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/admin/usuario_admin/"+itoa(userID)+"/modulos", nil)
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	rr := httptest.NewRecorder()

	h.GetUsuarioAdminModules(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		ModuleIDs []int64 `json:"module_ids"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if len(resp.ModuleIDs) != 2 {
		t.Fatalf("expected 2 module ids, got %+v", resp.ModuleIDs)
	}
	if !containsInt64(resp.ModuleIDs, 1) || !containsInt64(resp.ModuleIDs, 3) {
		t.Fatalf("unexpected module ids: %+v", resp.ModuleIDs)
	}
}

func TestAssignUsuarioAdminModulesReplacesFullSet(t *testing.T) {
	db := newAdminRoleTestDB(t)
	userStore := storage.NewAdminUserStore(db)
	userModuleStore := storage.NewUserModuleStore(db)
	userID := insertAdminUser(t, db, "alice", "$2a$10$abcdefghijklmnopqrstuv", "alice@a.com", 1)
	if err := userModuleStore.AssignModules(userID, []int64{1}); err != nil {
		t.Fatalf("seed modules: %v", err)
	}

	h := &AdminHandler{db: db, userStore: userStore, moduleStore: storage.NewModuleStore(db)}
	req := httptest.NewRequest(stdhttp.MethodPut, "/api/admin/usuario_admin/"+itoa(userID)+"/modulos", strings.NewReader(`{"module_ids":[2,3]}`))
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	rr := httptest.NewRecorder()

	h.AssignUsuarioAdminModules(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	modules, err := userModuleStore.GetByUserID(userID)
	if err != nil {
		t.Fatalf("get modules: %v", err)
	}
	if len(modules) != 2 || modules[0].ID == 1 || modules[1].ID == 1 {
		t.Fatalf("expected modules to be replaced, got %+v", modules)
	}
}

func TestAssignUsuarioAdminModulesRejectsInvalidModuleID(t *testing.T) {
	db := newAdminRoleTestDB(t)
	userStore := storage.NewAdminUserStore(db)
	userModuleStore := storage.NewUserModuleStore(db)
	userID := insertAdminUser(t, db, "alice", "$2a$10$abcdefghijklmnopqrstuv", "alice@a.com", 1)
	if err := userModuleStore.AssignModules(userID, []int64{1}); err != nil {
		t.Fatalf("seed modules: %v", err)
	}

	h := &AdminHandler{db: db, userStore: userStore, moduleStore: storage.NewModuleStore(db)}
	req := httptest.NewRequest(stdhttp.MethodPut, "/api/admin/usuario_admin/"+itoa(userID)+"/modulos", strings.NewReader(`{"module_ids":[1,99]}`))
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{UserID: 1, Rol: domain.RoleSuperAdmin, IsRoot: true}))
	rr := httptest.NewRecorder()

	h.AssignUsuarioAdminModules(rr, req)

	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}

	modules, err := userModuleStore.GetByUserID(userID)
	if err != nil {
		t.Fatalf("get modules: %v", err)
	}
	if len(modules) != 1 || modules[0].ID != 1 {
		t.Fatalf("expected original modules to remain, got %+v", modules)
	}
}

func TestListModulesReturnsCatalog(t *testing.T) {
	db := newAdminRoleTestDB(t)
	h := &AdminHandler{moduleStore: storage.NewModuleStore(db)}
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/admin/modules", nil)
	rr := httptest.NewRecorder()

	h.ListModules(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Modules []domain.Module `json:"modules"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if len(resp.Modules) != 3 {
		t.Fatalf("expected 3 modules, got %+v", resp.Modules)
	}
	if resp.Modules[0].Slug == "" {
		t.Fatalf("expected module slugs in response: %+v", resp.Modules[0])
	}
}

func TestAdminModulesRouteIsReadOnly(t *testing.T) {
	mux := stdhttp.NewServeMux()
	mux.Handle("GET /api/admin/modules", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
	}))

	req := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/modules", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func newAdminPhoneTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:admin-phone-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	_, err = db.Exec(`
CREATE TABLE telefonos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empresa_id INTEGER NOT NULL,
    codigo_pais TEXT NOT NULL,
    numero TEXT NOT NULL,
    numero_completo TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'disconnected',
    session_data BLOB,
    qr_string TEXT,
    last_connected TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		db.Close()
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func insertAdminPhone(t *testing.T, db *sql.DB, empresaID int64, codigoPais, numero, numeroCompleto string) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO telefonos (empresa_id, codigo_pais, numero, numero_completo, status) VALUES (?, ?, ?, ?, 'disconnected')`,
		empresaID, codigoPais, numero, numeroCompleto,
	)
	if err != nil {
		t.Fatalf("insert phone: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func newAdminUserScopeTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:admin-user-scope-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	_, err = db.Exec(`
CREATE TABLE admin_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email TEXT,
    role_id INTEGER,
    activo INTEGER NOT NULL DEFAULT 1,
    created_by INTEGER,
    updated_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL
);
CREATE TABLE user_modules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    module_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_by_user_id INTEGER NULL
);
CREATE TABLE api_key_audit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key_id INTEGER NOT NULL,
    empresa_id INTEGER NOT NULL,
    telefono_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    actor_user_id INTEGER,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    is_root INTEGER NOT NULL DEFAULT 0,
    permissions TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		_ = db.Close()
		t.Fatalf("create admin_users schema: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func newAdminRoleTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:admin-role-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	_, err = db.Exec(`
CREATE TABLE modules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    slug TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    is_root INTEGER NOT NULL DEFAULT 0,
    permissions TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE admin_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email TEXT,
    role_id INTEGER,
    activo INTEGER NOT NULL DEFAULT 1,
    created_by INTEGER,
    updated_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL
);
CREATE TABLE user_modules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    module_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		_ = db.Close()
		t.Fatalf("create role schema: %v", err)
	}

	for _, module := range []struct {
		name, description, slug string
	}{
		{name: "messages", description: "Mensajes", slug: "messages"},
		{name: "broadcasts", description: "Difusiones", slug: "broadcasts"},
		{name: "roles", description: "Roles", slug: "roles"},
	} {
		if _, err := db.Exec(`INSERT INTO modules (name, description, slug) VALUES (?, ?, ?)`, module.name, module.description, module.slug); err != nil {
			_ = db.Close()
			t.Fatalf("seed module %s: %v", module.name, err)
		}
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func containsInt64(values []int64, target int64) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func insertAdminUser(t *testing.T, db *sql.DB, username, passwordHash, email string, roleID int64) int64 {
	t.Helper()
	res, err := db.Exec(`
		INSERT INTO admin_users (username, password_hash, email, role_id, activo) 
		VALUES (?, ?, ?, ?, 1)`, username, passwordHash, email, roleID)
	if err != nil {
		t.Fatalf("insert admin user: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func itoa(v int64) string { return strconv.FormatInt(v, 10) }
