package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/coder/websocket"
	"golang.org/x/crypto/bcrypt"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/http/middleware"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

type AdminHandler struct {
	db            *sql.DB
	userStore     *storage.AdminUserStore
	roleStore     *storage.RoleStore
	moduleStore   *storage.ModuleStore
	telefonoStore *storage.TelefonoStore
	apiKeyStore   *storage.ApiKeyStore
	sessionStore  *storage.SessionStore
	manager       *whatsapp.Manager
	jwtCfg        *config.JWTConfig
}

func NewAdminHandler(db *sql.DB, sessionStore *storage.SessionStore, manager *whatsapp.Manager, jwtCfg *config.JWTConfig) *AdminHandler {
	if db == nil {
		return nil
	}
	return &AdminHandler{
		db:            db,
		userStore:     storage.NewAdminUserStore(db),
		roleStore:     storage.NewRoleStore(db),
		moduleStore:   storage.NewModuleStore(db),
		telefonoStore: storage.NewTelefonoStore(db),
		apiKeyStore:   storage.NewApiKeyStore(db),
		sessionStore:  sessionStore,
		manager:       manager,
		jwtCfg:        jwtCfg,
	}
}

type adminTelefonoRequest struct {
	CodigoPais string `json:"codigo_pais"`
	Numero     string `json:"numero"`
	Status     string `json:"status,omitempty"`
}

type AdminSessionDiagnostic struct {
	TelefonoID        int64  `json:"telefono_id"`
	EmpresaID         int64  `json:"empresa_id"`
	AccountID         string `json:"account_id"`
	StatusDB          string `json:"status_db"`
	StatusRuntime     string `json:"status_runtime"`
	RuntimeConnected  bool   `json:"runtime_connected"`
	Mismatch          bool   `json:"mismatch"`
	MismatchReason    string `json:"mismatch_reason"`
	RecommendedAction string `json:"recommended_action"`
}

type panelAdminAccess struct {
	EmpresaID *int64
	IsRoot    bool
}

func getPanelAdminAccess(r *http.Request) (panelAdminAccess, bool) {
	if claims, ok := domain.GetAdminJWTClaims(r.Context()); ok && claims != nil {
		return panelAdminAccess{EmpresaID: claims.EmpresaID, IsRoot: claims.IsRoot}, true
	}
	if claims, ok := domain.GetEmpresaJWTClaims(r.Context()); ok && claims != nil {
		eid := claims.EmpresaID
		return panelAdminAccess{EmpresaID: &eid, IsRoot: false}, true
	}
	return panelAdminAccess{}, false
}

func (a panelAdminAccess) canAccessEmpresa(empresaID int64) bool {
	if a.IsRoot {
		return true
	}
	if a.EmpresaID == nil {
		return false
	}
	return *a.EmpresaID == empresaID
}

func (a panelAdminAccess) companyID() (int64, bool) {
	if a.EmpresaID == nil || *a.EmpresaID <= 0 {
		return 0, false
	}
	return *a.EmpresaID, true
}

func extractPanelUserID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if id, err := strconv.ParseInt(parts[i], 10, 64); err == nil && id > 0 {
			return id, nil
		}
	}
	return 0, fmt.Errorf("invalid user ID")
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	users, total, err := h.userStore.GetAll(page, limit)
	if err != nil {
		http.Error(w, "error al obtener usuarios", http.StatusInternalServerError)
		return
	}

	result := make([]domain.AdminUser, len(users))
	for i, u := range users {
		u.PasswordHash = ""
		result[i] = u
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": result,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AdminHandler) ListUsuarioAdmins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	access, ok := getPanelAdminAccess(r)
	if !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var (
		users []domain.AdminUser
		total int
		err   error
	)
	if companyID, ok := access.companyID(); ok {
		users, total, err = h.userStore.GetAllByEmpresa(companyID, page, limit)
	} else if access.IsRoot {
		users, total, err = h.userStore.GetAll(page, limit)
	} else {
		http.Error(w, "acceso denegado", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "error al obtener usuario_admin", http.StatusInternalServerError)
		return
	}

	result := make([]domain.AdminUser, len(users))
	for i, u := range users {
		u.PasswordHash = ""
		result[i] = u
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"usuario_admin": result,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil {
		http.Error(w, "error al obtener usuario", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "usuario no encontrado", http.StatusNotFound)
		return
	}

	user.PasswordHash = ""
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) GetUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	access, ok := getPanelAdminAccess(r)
	if !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil {
		http.Error(w, "error al obtener usuario_admin", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "usuario_admin no encontrado", http.StatusNotFound)
		return
	}
	if user.EmpresaID != nil && !access.canAccessEmpresa(*user.EmpresaID) {
		http.Error(w, "acceso denegado", http.StatusForbidden)
		return
	}

	user.PasswordHash = ""
	json.NewEncoder(w).Encode(map[string]interface{}{"usuario_admin": user})
}

type CreateUserRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	EmpresaID *int64 `json:"empresa_id"`
	Role      string `json:"role"`
	RoleID    *int64 `json:"role_id"`
	IsRoot    bool   `json:"is_root"`
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username y password requeridos", http.StatusBadRequest)
		return
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error al crear hash de contraseña", http.StatusInternalServerError)
		return
	}
	hash := string(hashBytes)
	user := &domain.AdminUser{
		Username:     req.Username,
		PasswordHash: hash,
		Email:        req.Email,
		EmpresaID:    req.EmpresaID,
		Rol:          domain.UserRole(req.Role),
		RoleID:       req.RoleID,
		IsRoot:       req.IsRoot,
		Activo:       true,
	}

	id, err := h.userStore.Create(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("error al crear usuario: %v", err), http.StatusInternalServerError)
		return
	}

	user.ID = id
	user.PasswordHash = ""
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) CreateUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	access, ok := getPanelAdminAccess(r)
	if !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username y password requeridos", http.StatusBadRequest)
		return
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error al crear hash de contraseña", http.StatusInternalServerError)
		return
	}
	user := &domain.AdminUser{
		Username:     req.Username,
		PasswordHash: string(hashBytes),
		Email:        req.Email,
		EmpresaID:    req.EmpresaID,
		Rol:          domain.UserRole(req.Role),
		RoleID:       req.RoleID,
		IsRoot:       req.IsRoot,
		Activo:       true,
	}
	if user.EmpresaID == nil {
		if companyID, ok := access.companyID(); ok {
			user.EmpresaID = &companyID
		}
	}
	if user.EmpresaID != nil && !access.canAccessEmpresa(*user.EmpresaID) {
		http.Error(w, "acceso denegado", http.StatusForbidden)
		return
	}

	id, err := h.userStore.Create(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("error al crear usuario_admin: %v", err), http.StatusInternalServerError)
		return
	}
	user.ID = id
	user.PasswordHash = ""
	json.NewEncoder(w).Encode(map[string]interface{}{"usuario_admin": user})
}

type UpdateUserRequest struct {
	Email     string `json:"email"`
	EmpresaID *int64 `json:"empresa_id"`
	Role      string `json:"role"`
	RoleID    *int64 `json:"role_id"`
	IsRoot    bool   `json:"is_root"`
	IsActive  bool   `json:"is_active"`
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		http.Error(w, "usuario no encontrado", http.StatusNotFound)
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	user.EmpresaID = req.EmpresaID
	if req.Role != "" {
		user.Rol = domain.UserRole(req.Role)
	}
	user.RoleID = req.RoleID
	user.IsRoot = req.IsRoot
	user.Activo = req.IsActive

	err = h.userStore.Update(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("error al actualizar usuario: %v", err), http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) UpdateUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	access, ok := getPanelAdminAccess(r)
	if !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		http.Error(w, "usuario_admin no encontrado", http.StatusNotFound)
		return
	}
	if user.EmpresaID != nil && !access.canAccessEmpresa(*user.EmpresaID) {
		http.Error(w, "acceso denegado", http.StatusForbidden)
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.EmpresaID != nil {
		if !access.canAccessEmpresa(*req.EmpresaID) && !access.IsRoot {
			http.Error(w, "acceso denegado", http.StatusForbidden)
			return
		}
		user.EmpresaID = req.EmpresaID
	} else if user.EmpresaID == nil {
		if companyID, ok := access.companyID(); ok {
			user.EmpresaID = &companyID
		}
	}
	if req.Role != "" {
		user.Rol = domain.UserRole(req.Role)
	}
	user.RoleID = req.RoleID
	user.IsRoot = req.IsRoot
	user.Activo = req.IsActive

	if err := h.userStore.Update(user); err != nil {
		http.Error(w, fmt.Sprintf("error al actualizar usuario_admin: %v", err), http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	json.NewEncoder(w).Encode(map[string]interface{}{"usuario_admin": user})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	action, err := h.userStore.DeleteWithPolicy(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error al eliminar usuario: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": action})
}

func (h *AdminHandler) DeleteUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	access, ok := getPanelAdminAccess(r)
	if !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		http.Error(w, "usuario_admin no encontrado", http.StatusNotFound)
		return
	}
	if user.EmpresaID != nil && !access.canAccessEmpresa(*user.EmpresaID) {
		http.Error(w, "acceso denegado", http.StatusForbidden)
		return
	}

	action, err := h.userStore.DeleteWithPolicy(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error al eliminar usuario_admin: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": action})
}

type PromoteUserRequest struct {
	Role string `json:"role"`
}

func (h *AdminHandler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	var id int64
	var err error

	if strings.Contains(path, "/promote") {
		id, err = extractPanelUserID(path)
	} else if strings.Contains(path, "/users/") || strings.Contains(path, "/usuario_admin/") {
		id, err = extractPanelUserID(path)
	} else {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req PromoteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Role = "user"
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		http.Error(w, "usuario no encontrado", http.StatusNotFound)
		return
	}

	if req.Role != "" {
		user.Rol = domain.UserRole(req.Role)
	}
	user.IsRoot = false

	err = h.userStore.Update(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("error al promover usuario: %v", err), http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) PromoteUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	h.PromoteUser(w, r)
}

func (h *AdminHandler) PromoteUserByID(w http.ResponseWriter, r *http.Request) {
	h.PromoteUser(w, r)
}

type AssignModulesRequest struct {
	ModuleIDs []int64 `json:"module_ids"`
}

func (h *AdminHandler) GetUserModules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	access, ok := getPanelAdminAccess(r)
	if !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		http.Error(w, "usuario_admin no encontrado", http.StatusNotFound)
		return
	}
	if user.EmpresaID != nil && !access.canAccessEmpresa(*user.EmpresaID) {
		http.Error(w, "acceso denegado", http.StatusForbidden)
		return
	}

	userModuleStore := storage.NewUserModuleStore(h.db)
	modules, err := userModuleStore.GetByUserID(id)
	if err != nil {
		http.Error(w, "error al obtener módulos", http.StatusInternalServerError)
		return
	}

	moduleIDs := make([]int64, len(modules))
	for i, m := range modules {
		moduleIDs[i] = m.ID
	}

	json.NewEncoder(w).Encode(map[string][]int64{"module_ids": moduleIDs})
}

func (h *AdminHandler) GetUsuarioAdminModules(w http.ResponseWriter, r *http.Request) {
	h.GetUserModules(w, r)
}

func (h *AdminHandler) AssignUserModules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	access, ok := getPanelAdminAccess(r)
	if !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	id, err := extractPanelUserID(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		http.Error(w, "usuario_admin no encontrado", http.StatusNotFound)
		return
	}
	if user.EmpresaID != nil && !access.canAccessEmpresa(*user.EmpresaID) {
		http.Error(w, "acceso denegado", http.StatusForbidden)
		return
	}

	var req AssignModulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.validateModuleIDs(req.ModuleIDs); err != nil {
		if strings.Contains(err.Error(), "inválido") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userModuleStore := storage.NewUserModuleStore(h.db)
	if err := userModuleStore.AssignModules(id, req.ModuleIDs); err != nil {
		http.Error(w, "error al asignar módulos", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AdminHandler) AssignUsuarioAdminModules(w http.ResponseWriter, r *http.Request) {
	h.AssignUserModules(w, r)
}

func (h *AdminHandler) AssignUserModulesByID(w http.ResponseWriter, r *http.Request) {
	h.AssignUserModules(w, r)
}

func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if _, ok := getPanelAdminAccess(r); !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	roles, err := h.roleStore.GetAll()
	if err != nil {
		http.Error(w, "error al obtener roles "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"roles": roles,
	})
}

type RoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsRoot      bool     `json:"is_root"`
	Permissions []string `json:"permissions"`
}

func (h *AdminHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if _, ok := getPanelAdminAccess(r); !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid role ID", http.StatusBadRequest)
		return
	}

	role, err := h.roleStore.GetByID(id)
	if err != nil {
		http.Error(w, "error al obtener rol", http.StatusInternalServerError)
		return
	}
	if role == nil {
		http.Error(w, "rol no encontrado", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"role": role})
}

func (h *AdminHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if _, ok := getPanelAdminAccess(r); !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	var req RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		http.Error(w, "name requerido", http.StatusBadRequest)
		return
	}

	if existing, err := h.roleStore.GetByName(req.Name); err != nil {
		http.Error(w, "error al validar nombre de rol", http.StatusInternalServerError)
		return
	} else if existing != nil {
		http.Error(w, "el rol ya existe", http.StatusConflict)
		return
	}

	if err := h.validateRolePermissions(req.Permissions); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.IsRoot {
		if rootRole, err := h.roleStore.GetRootRole(); err != nil {
			http.Error(w, "error al validar rol root", http.StatusInternalServerError)
			return
		} else if rootRole != nil {
			http.Error(w, "ya existe un rol root", http.StatusConflict)
			return
		}
	}

	role := &domain.Role{
		Name:        req.Name,
		Description: req.Description,
		IsRoot:      req.IsRoot,
		Permissions: req.Permissions,
	}

	if _, err := h.roleStore.Create(role); err != nil {
		http.Error(w, fmt.Sprintf("error al crear rol: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"role": role})
}

func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if _, ok := getPanelAdminAccess(r); !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid role ID", http.StatusBadRequest)
		return
	}

	current, err := h.roleStore.GetByID(id)
	if err != nil {
		http.Error(w, "error al obtener rol", http.StatusInternalServerError)
		return
	}
	if current == nil {
		http.Error(w, "rol no encontrado", http.StatusNotFound)
		return
	}

	var req RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		http.Error(w, "name requerido", http.StatusBadRequest)
		return
	}

	if existing, err := h.roleStore.GetByName(req.Name); err != nil {
		http.Error(w, "error al validar nombre de rol", http.StatusInternalServerError)
		return
	} else if existing != nil && existing.ID != id {
		http.Error(w, "el rol ya existe", http.StatusConflict)
		return
	}

	if err := h.validateRolePermissions(req.Permissions); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if current.IsRoot && !req.IsRoot {
		http.Error(w, "el rol root no puede perder is_root", http.StatusBadRequest)
		return
	}
	if req.IsRoot && !current.IsRoot {
		if rootRole, err := h.roleStore.GetRootRole(); err != nil {
			http.Error(w, "error al validar rol root", http.StatusInternalServerError)
			return
		} else if rootRole != nil && rootRole.ID != current.ID {
			http.Error(w, "ya existe un rol root", http.StatusConflict)
			return
		}
	}

	current.Name = req.Name
	current.Description = req.Description
	current.IsRoot = req.IsRoot
	current.Permissions = req.Permissions

	if err := h.roleStore.Update(current); err != nil {
		http.Error(w, fmt.Sprintf("error al actualizar rol: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"role": current})
}

func (h *AdminHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if _, ok := getPanelAdminAccess(r); !ok {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid role ID", http.StatusBadRequest)
		return
	}

	role, err := h.roleStore.GetByID(id)
	if err != nil {
		http.Error(w, "error al obtener rol", http.StatusInternalServerError)
		return
	}
	if role == nil {
		http.Error(w, "rol no encontrado", http.StatusNotFound)
		return
	}
	if role.IsRoot {
		http.Error(w, "el rol root no puede eliminarse", http.StatusConflict)
		return
	}

	if err := h.roleStore.DeleteIfUnused(id); err != nil {
		if err == storage.ErrRoleInUse {
			http.Error(w, "el rol está en uso y no puede eliminarse", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("error al eliminar rol: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *AdminHandler) validateRolePermissions(perms []string) error {
	if h.moduleStore == nil {
		return fmt.Errorf("módulos no disponibles para validar permisos")
	}

	modules, err := h.moduleStore.GetAll()
	if err != nil {
		return fmt.Errorf("error al cargar módulos: %w", err)
	}

	allowed := map[string]struct{}{"all": struct{}{}}
	for _, module := range modules {
		allowed[strings.ToLower(strings.TrimSpace(module.Name))] = struct{}{}
		allowed[strings.ToLower(strings.TrimSpace(module.Slug))] = struct{}{}
	}

	for _, raw := range perms {
		perm := strings.ToLower(strings.TrimSpace(raw))
		if perm == "" {
			continue
		}
		if perm == "all" {
			continue
		}
		if _, ok := allowed[perm]; ok {
			continue
		}
		if base, _, ok := strings.Cut(perm, ":"); ok {
			if _, ok := allowed[base]; ok {
				continue
			}
		}
		return fmt.Errorf("permiso inválido: %s", raw)
	}
	return nil
}

func (h *AdminHandler) validateModuleIDs(moduleIDs []int64) error {
	if h.moduleStore == nil {
		return fmt.Errorf("módulos no disponibles para validar")
	}

	seen := map[int64]struct{}{}
	for _, moduleID := range moduleIDs {
		if moduleID <= 0 {
			return fmt.Errorf("module_id inválido: %d", moduleID)
		}
		if _, ok := seen[moduleID]; ok {
			continue
		}
		seen[moduleID] = struct{}{}
		module, err := h.moduleStore.GetByID(moduleID)
		if err != nil {
			return fmt.Errorf("error al validar módulo %d: %w", moduleID, err)
		}
		if module == nil {
			return fmt.Errorf("module_id inválido: %d", moduleID)
		}
	}
	return nil
}

func (h *AdminHandler) ListModules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	modules, err := h.moduleStore.GetAll()
	if err != nil {
		http.Error(w, "error al obtener módulos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"modules": modules,
	})
}

func (h *AdminHandler) ListCompanyPhones(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.telefonoStore == nil {
		http.Error(w, "telefono store no disponible", http.StatusInternalServerError)
		return
	}

	companyID, err := extractCompanyIDFromPath(r.URL.Path, "/api/admin/empresas/", "/telefonos")
	if err != nil || companyID <= 0 {
		http.Error(w, "invalid company ID", http.StatusBadRequest)
		return
	}

	companyStore := storage.NewEmpresaStore(h.db)
	company, err := companyStore.GetByID(companyID)
	if err != nil {
		http.Error(w, "error al validar empresa", http.StatusInternalServerError)
		return
	}
	if company == nil {
		http.Error(w, "empresa no encontrada", http.StatusNotFound)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	if claims != nil && !claims.IsRoot {
		if claims.EmpresaID == nil || *claims.EmpresaID != companyID {
			http.Error(w, "acceso denegado", http.StatusForbidden)
			return
		}
	}

	phones, err := h.telefonoStore.GetByEmpresa(companyID)
	if err != nil {
		http.Error(w, "error al obtener teléfonos", http.StatusInternalServerError)
		return
	}

	enriched := make([]domain.Telefono, len(phones))
	for i, phone := range phones {
		enriched[i] = phone
		runtimeConnected := false
		if h.manager != nil {
			accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)
			if client, ok := h.manager.Get(accountID); ok && client != nil {
				runtimeConnected = client.IsConnected()
			}
		}
		enriched[i].RuntimeConnected = runtimeConnected

		expectedActive := phone.Status == domain.TelefonoStatusActive
		if expectedActive != runtimeConnected {
			enriched[i].Mismatch = true
			if expectedActive {
				enriched[i].MismatchReason = "db_active_runtime_disconnected"
			} else {
				enriched[i].MismatchReason = "db_not_active_runtime_connected"
			}
		}
	}

	json.NewEncoder(w).Encode(domain.TelefonosListResponse{
		OK:        true,
		Telefonos: enriched,
		Total:     len(phones),
	})
}

func (h *AdminHandler) CreateCompanyPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	companyID, err := extractCompanyIDFromPath(r.URL.Path, "/api/admin/empresas/", "/telefonos")
	if err != nil || companyID <= 0 {
		http.Error(w, "invalid company ID", http.StatusBadRequest)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	if claims != nil && !claims.IsRoot {
		if claims.EmpresaID == nil || *claims.EmpresaID != companyID {
			http.Error(w, "acceso denegado", http.StatusForbidden)
			return
		}
	}

	var req adminTelefonoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	req.CodigoPais = strings.TrimSpace(req.CodigoPais)
	req.Numero = strings.TrimSpace(req.Numero)
	if req.CodigoPais == "" || req.Numero == "" {
		http.Error(w, "codigo_pais y numero requeridos", http.StatusBadRequest)
		return
	}

	status := domain.TelefonoStatusDisconnected
	if req.Status != "" {
		status = domain.TelefonoStatus(strings.TrimSpace(req.Status))
		switch status {
		case domain.TelefonoStatusActive, domain.TelefonoStatusQRPending, domain.TelefonoStatusDisconnected:
		default:
			http.Error(w, "estado de teléfono inválido", http.StatusBadRequest)
			return
		}
	}

	numeroCompleto := req.CodigoPais + req.Numero
	if existing, _ := h.telefonoStore.GetByNumeroCompleto(numeroCompleto); existing != nil {
		http.Error(w, "ya existe un teléfono con ese número", http.StatusConflict)
		return
	}

	phone := &domain.Telefono{
		EmpresaID:      companyID,
		CodigoPais:     req.CodigoPais,
		Numero:         req.Numero,
		NumeroCompleto: numeroCompleto,
		Status:         status,
	}

	if _, err := h.telefonoStore.Create(phone); err != nil {
		http.Error(w, "error al crear teléfono", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(domain.TelefonoResponse{OK: true, Telefono: phone})
}

func (h *AdminHandler) UpdateCompanyPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
	if err != nil || telefonoID <= 0 {
		http.Error(w, "invalid telefono ID", http.StatusBadRequest)
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		http.Error(w, "teléfono no encontrado", http.StatusNotFound)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	if claims != nil && !claims.IsRoot {
		if claims.EmpresaID == nil || *claims.EmpresaID != phone.EmpresaID {
			http.Error(w, "acceso denegado", http.StatusForbidden)
			return
		}
	}

	var req adminTelefonoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.CodigoPais) != "" {
		phone.CodigoPais = strings.TrimSpace(req.CodigoPais)
	}
	if strings.TrimSpace(req.Numero) != "" {
		phone.Numero = strings.TrimSpace(req.Numero)
	}
	if strings.TrimSpace(req.Status) != "" {
		phone.Status = domain.TelefonoStatus(strings.TrimSpace(req.Status))
		switch phone.Status {
		case domain.TelefonoStatusActive, domain.TelefonoStatusQRPending, domain.TelefonoStatusDisconnected:
		default:
			http.Error(w, "estado de teléfono inválido", http.StatusBadRequest)
			return
		}
	}
	phone.NumeroCompleto = phone.CodigoPais + phone.Numero

	if existing, _ := h.telefonoStore.GetByNumeroCompleto(phone.NumeroCompleto); existing != nil && existing.ID != phone.ID {
		http.Error(w, "ya existe un teléfono con ese número", http.StatusConflict)
		return
	}

	if err := h.telefonoStore.Update(phone); err != nil {
		http.Error(w, "error al actualizar teléfono", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(domain.TelefonoResponse{OK: true, Telefono: phone})
}

func (h *AdminHandler) DeleteCompanyPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
	if err != nil || telefonoID <= 0 {
		http.Error(w, "invalid telefono ID", http.StatusBadRequest)
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		http.Error(w, "teléfono no encontrado", http.StatusNotFound)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	if claims != nil && !claims.IsRoot {
		if claims.EmpresaID == nil || *claims.EmpresaID != phone.EmpresaID {
			http.Error(w, "acceso denegado", http.StatusForbidden)
			return
		}
	}

	if h.apiKeyStore != nil {
		if err := h.apiKeyStore.RevokeByTelefonoID(phone.ID); err != nil {
			http.Error(w, "error al invalidar API keys", http.StatusInternalServerError)
			return
		}
	}
	if h.manager != nil {
		h.manager.Delete(phone.NumeroCompleto)
	}
	if h.sessionStore != nil {
		h.sessionStore.SetDisconnected(phone.NumeroCompleto, "admin_delete")
	}

	if err := h.telefonoStore.Delete(telefonoID); err != nil {
		http.Error(w, "error al eliminar teléfono", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *AdminHandler) GetCompanyPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
	if err != nil || telefonoID <= 0 {
		http.Error(w, "invalid telefono ID", http.StatusBadRequest)
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		http.Error(w, "teléfono no encontrado", http.StatusNotFound)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	if claims != nil && !claims.IsRoot {
		if claims.EmpresaID == nil || *claims.EmpresaID != phone.EmpresaID {
			http.Error(w, "acceso denegado", http.StatusForbidden)
			return
		}
	}

	json.NewEncoder(w).Encode(domain.TelefonoResponse{OK: true, Telefono: phone})
}

func (h *AdminHandler) GetSessionsDiagnostics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	if claims == nil {
		http.Error(w, "token requerido", http.StatusUnauthorized)
		return
	}

	mismatchOnly, _ := strconv.ParseBool(strings.TrimSpace(r.URL.Query().Get("mismatch_only")))

	var (
		telefonos []domain.Telefono
		err       error
	)

	if claims.IsRoot {
		empresaIDRaw := strings.TrimSpace(r.URL.Query().Get("empresa_id"))
		if empresaIDRaw != "" {
			empresaID, parseErr := strconv.ParseInt(empresaIDRaw, 10, 64)
			if parseErr != nil || empresaID <= 0 {
				http.Error(w, "empresa_id invalido", http.StatusBadRequest)
				return
			}
			telefonos, err = h.telefonoStore.GetByEmpresa(empresaID)
		} else {
			telefonos, err = h.telefonoStore.ListAll()
		}
	} else {
		if claims.EmpresaID == nil || *claims.EmpresaID <= 0 {
			http.Error(w, "acceso denegado", http.StatusForbidden)
			return
		}
		telefonos, err = h.telefonoStore.GetByEmpresa(*claims.EmpresaID)
	}
	if err != nil {
		http.Error(w, "error al obtener sesiones", http.StatusInternalServerError)
		return
	}

	diagnostics := make([]AdminSessionDiagnostic, 0, len(telefonos))
	totalMismatch := 0
	runtimeConnectedTotal := 0
	dbActiveTotal := 0

	for _, phone := range telefonos {
		runtimeConnected := false
		accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)
		if h.manager != nil {
			if client, ok := h.manager.Get(accountID); ok && client != nil && client.IsConnected() {
				runtimeConnected = true
			}
		}

		diag := buildAdminSessionDiagnostic(&phone, runtimeConnected)
		if diag.Mismatch {
			totalMismatch++
		}
		if diag.RuntimeConnected {
			runtimeConnectedTotal++
		}
		if diag.StatusDB == string(domain.TelefonoStatusActive) {
			dbActiveTotal++
		}

		if mismatchOnly && !diag.Mismatch {
			continue
		}
		diagnostics = append(diagnostics, diag)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"ok": true,
		"summary": map[string]any{
			"total_telefonos":       len(telefonos),
			"runtime_connected":     runtimeConnectedTotal,
			"db_active":             dbActiveTotal,
			"mismatches":            totalMismatch,
			"mismatch_only_applied": mismatchOnly,
		},
		"sessions": diagnostics,
	})
}

func buildAdminSessionDiagnostic(phone *domain.Telefono, runtimeConnected bool) AdminSessionDiagnostic {
	accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)
	statusRuntime := "disconnected"
	if runtimeConnected {
		statusRuntime = "connected"
	}

	dbActive := phone.Status == domain.TelefonoStatusActive
	mismatch := dbActive != runtimeConnected
	reason := ""
	if mismatch {
		if dbActive {
			reason = "db_active_runtime_disconnected"
		} else {
			reason = "db_not_active_runtime_connected"
		}
	}

	return AdminSessionDiagnostic{
		TelefonoID:        phone.ID,
		EmpresaID:         phone.EmpresaID,
		AccountID:         accountID,
		StatusDB:          string(phone.Status),
		StatusRuntime:     statusRuntime,
		RuntimeConnected:  runtimeConnected,
		Mismatch:          mismatch,
		MismatchReason:    reason,
		RecommendedAction: recommendedAdminSessionAction(string(phone.Status), runtimeConnected),
	}
}

func recommendedAdminSessionAction(statusDB string, runtimeConnected bool) string {
	if runtimeConnected {
		return "none"
	}
	if statusDB == string(domain.TelefonoStatusActive) {
		return "reanudar_conexion"
	}
	return "iniciar_conexion"
}

func (h *AdminHandler) StartCompanyPhoneConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
	if err != nil || telefonoID <= 0 {
		http.Error(w, "invalid telefono ID", http.StatusBadRequest)
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		http.Error(w, "teléfono no encontrado", http.StatusNotFound)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())
	if claims != nil && !claims.IsRoot {
		if claims.EmpresaID == nil || *claims.EmpresaID != phone.EmpresaID {
			http.Error(w, "acceso denegado", http.StatusForbidden)
			return
		}
	}

	if h.sessionStore != nil {
		if state, ok := h.sessionStore.Get(phone.NumeroCompleto); ok && state.Status == "active" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":             true,
				"telefono_id":    phone.ID,
				"numeroCompleto": phone.NumeroCompleto,
				"status":         "active",
				"lastConnected":  phone.LastConnected,
				"qr_string":      phone.QRString,
			})
			return
		}
	}

	if h.manager == nil {
		http.Error(w, "whatsapp manager no disponible", http.StatusServiceUnavailable)
		return
	}

	events, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
	if err != nil {
		http.Error(w, "error al iniciar conexión: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		for range events {
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":             true,
		"telefono_id":    phone.ID,
		"numeroCompleto": phone.NumeroCompleto,
		"status":         "initializing",
		"qr_string":      phone.QRString,
		"expires_in":     300,
	})
}

func (h *AdminHandler) ConnectCompanyPhoneWS(w http.ResponseWriter, r *http.Request) {
	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer wsConn.CloseNow()

	if h.jwtCfg == nil {
		_ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "configuracion JWT no disponible"}})
		return
	}

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
	if !claims.IsRoot {
		if claims.EmpresaID == nil || *claims.EmpresaID != phone.EmpresaID {
			_ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "acceso denegado"}})
			return
		}
	}

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

	events, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
	if err != nil {
		_ = writeEvent(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "error al iniciar conexión: " + err.Error()}})
		return
	}

	ctx := r.Context()
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeEvent(ctx, wsConn, outboundPayload{Event: event.Event, Data: event.Data}); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func extractTelefonoIDFromAdminPath(path string) (int64, error) {
	base := strings.TrimPrefix(path, "/api/admin/telefonos/")
	base = strings.TrimSuffix(base, "/connect/ws")
	base = strings.TrimSuffix(base, "/connect")
	base = strings.Trim(base, "/")
	if base == "" {
		return 0, fmt.Errorf("missing telefono id")
	}
	return strconv.ParseInt(base, 10, 64)
}

func extractCompanyIDFromPath(path, prefix, suffix string) (int64, error) {
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)
	path = strings.Trim(path, "/")
	if path == "" {
		return 0, fmt.Errorf("missing id")
	}
	return strconv.ParseInt(path, 10, 64)
}

func getAdminHandler() *AdminHandler {
	cfg := config.Load()
	if cfg.DBHost == "" {
		return nil
	}
	jwtCfg := config.LoadJWT()
	db, err := storage.NewDB(cfg)
	if err != nil {
		return nil
	}
	return NewAdminHandler(db, nil, nil, jwtCfg)
}
