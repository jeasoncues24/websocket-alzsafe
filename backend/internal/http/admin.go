package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"time"

	"github.com/coder/websocket"
	"golang.org/x/crypto/bcrypt"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	handlers "wsapi/internal/http/handlers"
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
	webhookStore  *storage.WebhookStore
	sessionStore  *storage.SessionStore
	manager       *whatsapp.Manager
	jwtCfg        *config.JWTConfig
}

func writeAdminJSON(w http.ResponseWriter, status int, payload interface{}) {
	writeJSON(w, status, payload)
}

func writeAdminError(w http.ResponseWriter, status int, message string) {
	writeAPIError(w, status, message)
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
		webhookStore:  storage.NewWebhookStore(db),
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

type panelAdminAccess = domain.PanelAccess

func getPanelAdminAccess(r *http.Request) (panelAdminAccess, bool) {
	return domain.GetPanelAccess(r.Context())
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
		writeAdminError(w, http.StatusInternalServerError, "error al obtener usuarios")
		return
	}

	result := make([]domain.AdminUser, len(users))
	for i, u := range users {
		u.PasswordHash = ""
		result[i] = u
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"users": result,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AdminHandler) ListUsuarioAdmins(w http.ResponseWriter, r *http.Request) {
	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
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
	if access.IsAdminJWT || access.IsRoot {
		users, total, err = h.userStore.GetAll(page, limit)
	} else {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener usuario_admin")
		return
	}

	result := make([]domain.AdminUser, len(users))
	for i, u := range users {
		u.PasswordHash = ""
		result[i] = u
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"users": result,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener usuario")
		return
	}
	if user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	user.PasswordHash = ""
	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user": user})
}

func (h *AdminHandler) GetUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener usuario_admin")
		return
	}
	if user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario_admin no encontrado")
		return
	}

	user.PasswordHash = ""
	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user": user})
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	RoleID   *int64 `json:"role_id"`
	// IsRoot se obtiene via rol - no se permite setear directamente
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeAdminError(w, http.StatusBadRequest, "username y password requeridos")
		return
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al crear hash de contraseña")
		return
	}
	hash := string(hashBytes)
	user := &domain.AdminUser{
		Username:     req.Username,
		PasswordHash: hash,
		Email:        req.Email,
		RoleID:       req.RoleID,
		Activo:       true,
	}

	id, err := h.userStore.Create(user)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al crear usuario: %v", err))
		return
	}

	user.ID = id
	user.PasswordHash = ""
	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user": user})
}

func (h *AdminHandler) CreateUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeAdminError(w, http.StatusBadRequest, "username y password requeridos")
		return
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al crear hash de contraseña")
		return
	}
	user := &domain.AdminUser{
		Username:     req.Username,
		PasswordHash: string(hashBytes),
		Email:        req.Email,
		RoleID:       req.RoleID,
		Activo:       true,
	}

	id, err := h.userStore.Create(user)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al crear usuario_admin: %v", err))
		return
	}
	user.ID = id
	user.PasswordHash = ""
	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user": user})
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	RoleID   *int64 `json:"role_id"`
	IsActive bool   `json:"is_active"`
	// IsRoot se obtiene via rol - no se permite setear directamente
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RoleID != nil {
		user.RoleID = req.RoleID
	}
	user.Activo = req.IsActive

	err = h.userStore.Update(user)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al actualizar usuario: %v", err))
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user": user})
}

func (h *AdminHandler) UpdateUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario_admin no encontrado")
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RoleID != nil {
		user.RoleID = req.RoleID
	}
	user.Activo = req.IsActive

	if err := h.userStore.Update(user); err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al actualizar usuario_admin: %v", err))
		return
	}

	user.PasswordHash = ""
	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user": user})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	action, err := h.userStore.DeleteWithPolicy(id)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al eliminar usuario: %v", err))
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": action})
}

func (h *AdminHandler) DeleteUsuarioAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario_admin no encontrado")
		return
	}

	action, err := h.userStore.DeleteWithPolicy(id)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al eliminar usuario_admin: %v", err))
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": action})
}

type PromoteUserRequest struct {
	RoleID *int64 `json:"role_id"`
}

func (h *AdminHandler) promoteUser(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	id, err := extractPanelUserID(path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req PromoteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.RoleID = nil
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	if req.RoleID != nil {
		user.RoleID = req.RoleID
	}

	err = h.userStore.Update(user)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al promover usuario: %v", err))
		return
	}

	user.PasswordHash = ""
	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user": user})
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
		writeAdminError(w, http.StatusBadRequest, "invalid path")
		return
	}
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req PromoteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.RoleID = nil
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	if req.RoleID != nil {
		user.RoleID = req.RoleID
	}

	err = h.userStore.Update(user)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al promover usuario: %v", err))
		return
	}

	user.PasswordHash = ""
	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user": user})
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
	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario_admin no encontrado")
		return
	}

	userModuleStore := storage.NewUserModuleStore(h.db)
	modules, err := userModuleStore.GetByUserID(id)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener módulos")
		return
	}

	moduleIDs := make([]int64, len(modules))
	for i, m := range modules {
		moduleIDs[i] = m.ID
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "module_ids": moduleIDs})
}

func (h *AdminHandler) GetUsuarioAdminModules(w http.ResponseWriter, r *http.Request) {
	h.GetUserModules(w, r)
}

func (h *AdminHandler) AssignUserModules(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id, err := extractPanelUserID(path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userStore.GetByID(id)
	if err != nil || user == nil {
		writeAdminError(w, http.StatusNotFound, "usuario_admin no encontrado")
		return
	}

	var req AssignModulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := h.validateModuleIDs(req.ModuleIDs); err != nil {
		if strings.Contains(err.Error(), "inválido") {
			writeAdminError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeAdminError(w, http.StatusInternalServerError, err.Error())
		return
	}

	userModuleStore := storage.NewUserModuleStore(h.db)
	if err := userModuleStore.AssignModules(id, req.ModuleIDs); err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al asignar módulos")
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": "ok"})
}

func (h *AdminHandler) AssignUsuarioAdminModules(w http.ResponseWriter, r *http.Request) {
	h.AssignUserModules(w, r)
}

func (h *AdminHandler) AssignUserModulesByID(w http.ResponseWriter, r *http.Request) {
	h.AssignUserModules(w, r)
}

func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	if _, ok := getPanelAdminAccess(r); !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}

	roles, err := h.roleStore.GetAll()
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener roles "+err.Error())
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
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
	if _, ok := getPanelAdminAccess(r); !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid role ID")
		return
	}

	role, err := h.roleStore.GetByID(id)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener rol")
		return
	}
	if role == nil {
		writeAdminError(w, http.StatusNotFound, "rol no encontrado")
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "role": role})
}

func (h *AdminHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	if _, ok := getPanelAdminAccess(r); !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}

	var req RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid request")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		writeAdminError(w, http.StatusBadRequest, "name requerido")
		return
	}

	if existing, err := h.roleStore.GetByName(req.Name); err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al validar nombre de rol")
		return
	} else if existing != nil {
		writeAdminError(w, http.StatusConflict, "el rol ya existe")
		return
	}

	if err := h.validateRolePermissions(req.Permissions); err != nil {
		writeAdminError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.IsRoot {
		if rootRole, err := h.roleStore.GetRootRole(); err != nil {
			writeAdminError(w, http.StatusInternalServerError, "error al validar rol root")
			return
		} else if rootRole != nil {
			// http.Error(w, "ya existe un rol root", http.StatusConflict)
			// return
		}
	}

	role := &domain.Role{
		Name:        req.Name,
		Description: req.Description,
		IsRoot:      req.IsRoot,
		Permissions: req.Permissions,
	}

	if _, err := h.roleStore.Create(role); err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al crear rol: %v", err))
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "role": role})
}

func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	if _, ok := getPanelAdminAccess(r); !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid role ID")
		return
	}

	current, err := h.roleStore.GetByID(id)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener rol")
		return
	}
	if current == nil {
		writeAdminError(w, http.StatusNotFound, "rol no encontrado")
		return
	}

	var req RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid request")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		writeAdminError(w, http.StatusBadRequest, "name requerido")
		return
	}

	if existing, err := h.roleStore.GetByName(req.Name); err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al validar nombre de rol")
		return
	} else if existing != nil && existing.ID != id {
		writeAdminError(w, http.StatusConflict, "el rol ya existe")
		return
	}

	if err := h.validateRolePermissions(req.Permissions); err != nil {
		writeAdminError(w, http.StatusBadRequest, err.Error())
		return
	}

	if current.IsRoot && !req.IsRoot {
		writeAdminError(w, http.StatusBadRequest, "el rol root no puede perder is_root")
		return
	}
	if req.IsRoot && !current.IsRoot {
		if rootRole, err := h.roleStore.GetRootRole(); err != nil {
			writeAdminError(w, http.StatusInternalServerError, "error al validar rol root")
			return
		} else if rootRole != nil && rootRole.ID != current.ID {
			// http.Error(w, "ya existe un rol root", http.StatusConflict)
			// return
		}
	}

	current.Name = req.Name
	current.Description = req.Description
	current.IsRoot = req.IsRoot
	current.Permissions = req.Permissions

	if err := h.roleStore.Update(current); err != nil {
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al actualizar rol: %v", err))
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "role": current})
}

func (h *AdminHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	if _, ok := getPanelAdminAccess(r); !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}

	id, err := extractPanelUserID(r.URL.Path)
	if err != nil {
		writeAdminError(w, http.StatusBadRequest, "invalid role ID")
		return
	}

	role, err := h.roleStore.GetByID(id)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener rol")
		return
	}
	if role == nil {
		writeAdminError(w, http.StatusNotFound, "rol no encontrado")
		return
	}
	if role.IsRoot {
		writeAdminError(w, http.StatusConflict, "el rol root no puede eliminarse")
		return
	}

	if err := h.roleStore.DeleteIfUnused(id); err != nil {
		if err == storage.ErrRoleInUse {
			writeAdminError(w, http.StatusConflict, "el rol está en uso y no puede eliminarse")
			return
		}
		writeAdminError(w, http.StatusInternalServerError, fmt.Sprintf("error al eliminar rol: %v", err))
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": "deleted"})
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
	modules, err := h.moduleStore.GetAll()
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener módulos: "+err.Error())
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"modules": modules,
	})
}

func (h *AdminHandler) ListCompanyPhones(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.telefonoStore == nil {
		writeAdminError(w, http.StatusInternalServerError, "telefono store no disponible")
		return
	}

	companyID, err := extractCompanyIDFromPath(r.URL.Path, "/api/admin/empresas/", "/telefonos")
	if err != nil || companyID <= 0 {
		writeAdminError(w, http.StatusBadRequest, "invalid company ID")
		return
	}

	companyStore := storage.NewEmpresaStore(h.db)
	company, err := companyStore.GetByID(companyID)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al validar empresa")
		return
	}
	if company == nil {
		writeAdminError(w, http.StatusNotFound, "empresa no encontrada")
		return
	}

	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}
	if !access.CanAccessEmpresa(companyID) {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}

	phones, err := h.telefonoStore.GetByEmpresa(companyID)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener teléfonos")
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

		if h.apiKeyStore != nil {
			keys, _ := h.apiKeyStore.GetByTelefonoID(phone.ID)
			activeKeys := 0
			for _, k := range keys {
				if k.Activo {
					activeKeys++
				}
			}
			enriched[i].ApiKeyCount = activeKeys
		}

		if h.webhookStore != nil {
			hooks, _ := h.webhookStore.ListByTelefono(phone.ID)
			activeHooks := 0
			for _, wh := range hooks {
				if wh.Activo {
					activeHooks++
				}
			}
			enriched[i].WebhookCount = activeHooks
		}
	}

	writeAdminJSON(w, http.StatusOK, domain.TelefonosListResponse{
		OK:        true,
		Telefonos: enriched,
		Total:     len(phones),
	})
}

func (h *AdminHandler) CreateCompanyPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	companyID, err := extractCompanyIDFromPath(r.URL.Path, "/api/admin/empresas/", "/telefonos")
	if err != nil || companyID <= 0 {
		writeAdminError(w, http.StatusBadRequest, "invalid company ID")
		return
	}

	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}
	if !access.CanAccessEmpresa(companyID) {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}

	var req adminTelefonoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	req.CodigoPais = strings.TrimSpace(req.CodigoPais)
	req.Numero = strings.TrimSpace(req.Numero)
	if req.CodigoPais == "" || req.Numero == "" {
		writeAdminError(w, http.StatusBadRequest, "codigo_pais y numero requeridos")
		return
	}

	status := domain.TelefonoStatusDisconnected
	if req.Status != "" {
		status = domain.TelefonoStatus(strings.TrimSpace(req.Status))
		switch status {
		case domain.TelefonoStatusActive, domain.TelefonoStatusQRPending, domain.TelefonoStatusDisconnected:
		default:
			writeAdminError(w, http.StatusBadRequest, "estado de teléfono inválido")
			return
		}
	}

	numeroCompleto := req.CodigoPais + req.Numero
	if existing, _ := h.telefonoStore.GetByNumeroCompleto(numeroCompleto); existing != nil {
		writeAdminError(w, http.StatusConflict, "ya existe un teléfono con ese número")
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
		writeAdminError(w, http.StatusInternalServerError, "error al crear teléfono")
		return
	}

	writeAdminJSON(w, http.StatusOK, domain.TelefonoResponse{OK: true, Telefono: phone})
}

func (h *AdminHandler) UpdateCompanyPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
	if err != nil || telefonoID <= 0 {
		writeAdminError(w, http.StatusBadRequest, "invalid telefono ID")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeAdminError(w, http.StatusNotFound, "teléfono no encontrado")
		return
	}

	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}
	if !access.CanAccessEmpresa(phone.EmpresaID) {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}

	var req adminTelefonoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminError(w, http.StatusBadRequest, "JSON inválido")
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
			writeAdminError(w, http.StatusBadRequest, "estado de teléfono inválido")
			return
		}
	}
	phone.NumeroCompleto = phone.CodigoPais + phone.Numero

	if existing, _ := h.telefonoStore.GetByNumeroCompleto(phone.NumeroCompleto); existing != nil && existing.ID != phone.ID {
		writeAdminError(w, http.StatusConflict, "ya existe un teléfono con ese número")
		return
	}

	if err := h.telefonoStore.Update(phone); err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al actualizar teléfono")
		return
	}

	writeAdminJSON(w, http.StatusOK, domain.TelefonoResponse{OK: true, Telefono: phone})
}

func (h *AdminHandler) DeleteCompanyPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
	if err != nil || telefonoID <= 0 {
		writeAdminError(w, http.StatusBadRequest, "invalid telefono ID")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeAdminError(w, http.StatusNotFound, "teléfono no encontrado")
		return
	}

	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}
	if !access.CanAccessEmpresa(phone.EmpresaID) {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}

	if h.apiKeyStore != nil {
		if err := h.apiKeyStore.RevokeByTelefonoID(phone.ID); err != nil {
			writeAdminError(w, http.StatusInternalServerError, "error al invalidar API keys")
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
		writeAdminError(w, http.StatusInternalServerError, "error al eliminar teléfono")
		return
	}

	writeAdminJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AdminHandler) GetCompanyPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
	if err != nil || telefonoID <= 0 {
		writeAdminError(w, http.StatusBadRequest, "invalid telefono ID")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeAdminError(w, http.StatusNotFound, "teléfono no encontrado")
		return
	}

	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}
	if !access.CanAccessEmpresa(phone.EmpresaID) {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}

	writeAdminJSON(w, http.StatusOK, domain.TelefonoResponse{OK: true, Telefono: phone})
}

func (h *AdminHandler) GetSessionsDiagnostics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		writeAdminError(w, http.StatusMethodNotAllowed, "metodo no permitido")
		return
	}

	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}

	mismatchOnly, _ := strconv.ParseBool(strings.TrimSpace(r.URL.Query().Get("mismatch_only")))
	empresaIDRaw := strings.TrimSpace(r.URL.Query().Get("empresa_id"))

	var (
		telefonos []domain.Telefono
		err       error
	)

	if access.IsAdminJWT || access.IsRoot {
		if empresaIDRaw != "" {
			empresaID, parseErr := strconv.ParseInt(empresaIDRaw, 10, 64)
			if parseErr != nil || empresaID <= 0 {
				writeAdminError(w, http.StatusBadRequest, "empresa_id invalido")
				return
			}
			telefonos, err = h.telefonoStore.GetByEmpresa(empresaID)
		} else {
			telefonos, err = h.telefonoStore.ListAll()
		}
	} else {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener sesiones")
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

	writeAdminJSON(w, http.StatusOK, map[string]any{
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
		writeAdminError(w, http.StatusBadRequest, "invalid telefono ID")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeAdminError(w, http.StatusNotFound, "teléfono no encontrado")
		return
	}

	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}
	if !access.CanAccessEmpresa(phone.EmpresaID) {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}

	if h.sessionStore != nil {
		if state, ok := h.sessionStore.Get(phone.NumeroCompleto); ok && state.Status == "active" {
			writeAdminJSON(w, http.StatusOK, map[string]interface{}{
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
		writeAdminError(w, http.StatusServiceUnavailable, "whatsapp manager no disponible")
		return
	}

	events, unsubscribe, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al iniciar conexión: "+err.Error())
		return
	}

	go func() {
		defer unsubscribe()
		for range events {
		}
	}()

	writeAdminJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"telefono_id":    phone.ID,
		"numeroCompleto": phone.NumeroCompleto,
		"status":         "initializing",
		"qr_string":      phone.QRString,
		"expires_in":     300,
	})
}

func (h *AdminHandler) ConnectCompanyPhoneWS(w http.ResponseWriter, r *http.Request) {
	// — Autenticar token admin (query param, header Authorization o Sec-WebSocket-Protocol) —
	token := r.URL.Query().Get("token")
	if token == "" {
		if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if token == "" {
		secProtocols := r.Header.Get("Sec-WebSocket-Protocol")
		if secProtocols != "" {
			parts := strings.Split(secProtocols, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					token = p
					break
				}
			}
		}
	}

	acceptOpts := &websocket.AcceptOptions{InsecureSkipVerify: true}
	if r.Header.Get("Sec-WebSocket-Protocol") != "" && token != "" {
		acceptOpts.Subprotocols = []string{token}
	}

	// — Upgrade a WebSocket —
	wsConn, err := websocket.Accept(w, r, acceptOpts)
	if err != nil {
		return
	}
	defer wsConn.CloseNow()

	// — Validar configuración JWT —
	if h.jwtCfg == nil {
		_ = handlers.WriteWSJSON(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "configuracion JWT no disponible"}})
		return
	}

	if token == "" {
		_ = handlers.WriteWSJSON(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "Token requerido"}})
		return
	}
	claims, err := middleware.NewAuthMiddleware(h.jwtCfg, nil).ValidateToken(token)
	if err != nil {
		_ = handlers.WriteWSJSON(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "Token inválido"}})
		return
	}
	_ = claims

	// — Resolver teléfono desde la ruta —
	telefonoID, err := extractTelefonoIDFromAdminPath(r.URL.Path)
	if err != nil || telefonoID <= 0 {
		_ = handlers.WriteWSJSON(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "ID de teléfono inválido"}})
		return
	}
	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		_ = handlers.WriteWSJSON(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "teléfono no encontrado"}})
		return
	}

	access := domain.PanelAccess{IsRoot: claims.IsRoot, IsAdminJWT: true}
	if !access.CanAccessEmpresa(phone.EmpresaID) {
		_ = handlers.WriteWSJSON(r.Context(), wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "acceso denegado a esta empresa"}})
		return
	}
	accountID := whatsapp.NormalizeAccountID(phone.NumeroCompleto)

	fmt.Printf("[INFO] WS connect opened telefono=%d account=%s\n", phone.ID, accountID)

	// El contexto del request se cancela cuando el cliente WS cierra la conexión.
	ctx := r.Context()

	// — Cleanup al cerrar el WS (por cualquier causa) —
	// El runtime WhatsApp es compartido entre todos los observadores WS; cerrar
	// este WS solo da de baja a este observador, nunca cancela la sesión. Una
	// sesión abandonada durante el QR la termina whatsmeow al expirar el código.
	var unsubscribe func()
	defer func() {
		if unsubscribe != nil {
			unsubscribe()
		}
		fmt.Printf("[INFO] WS connect closed telefono=%d account=%s reason=%v\n", phone.ID, accountID, ctx.Err())
		if h.sessionStore != nil {
			reasonStr := "normal"
			if ctx.Err() != nil {
				reasonStr = ctx.Err().Error()
			}
			h.sessionStore.AppendEvent(phone.NumeroCompleto, "ws_closed", "WS admin cerrado: "+reasonStr)
		}
	}()

	// — Enviar estado inicial del teléfono —
	if err := handlers.WriteWSJSON(ctx, wsConn, outboundPayload{
		Event: "phone-info",
		Data: map[string]any{
			"telefono_id":    phone.ID,
			"numeroCompleto": phone.NumeroCompleto,
			"status":         phone.Status,
			"qr_string":      phone.QRString,
			"lastConnected":  phone.LastConnected,
		},
	}); err != nil {
		fmt.Printf("[WARN] WS initial phone-info failed account=%s: %v\n", accountID, err)
		return
	}

	// — Iniciar o unirse a sesión existente —
	// StartSession es idempotente: si ya existe un runtime para este accountID,
	// este WS se suscribe como observador adicional al MISMO cliente WhatsApp.
	// No se abre una segunda conexión WhatsApp; el fan-out reenvía los eventos en
	// vivo (QR, connected, disconnected) a todos los observadores simultáneos.
	events, unsub, err := whatsapp.StartSession(h.manager, phone.NumeroCompleto)
	if err != nil {
		_ = handlers.WriteWSJSON(ctx, wsConn, outboundPayload{Event: "error", Data: map[string]any{"message": "error al iniciar conexión: " + err.Error()}})
		return
	}
	unsubscribe = unsub

	// — Keepalive: enviar ping cada 25s para evitar que proxies corten la conexión idle —
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	// — Loop principal: este WS es un puente entre el manager de la sesión y el cliente —
	// Los eventos del runtime (QR, connected, disconnected) se reenvían directamente al browser.
	// El loop termina cuando: el canal de eventos se cierra (sesión terminó),
	// el cliente WS se desconecta (ctx.Done), o falla un write.
	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Canal cerrado — la sesión terminó (conectó, desconectó, o timeout QR)
				return
			}
			if err := handlers.WriteWSJSON(ctx, wsConn, outboundPayload{Event: event.Event, Data: event.Data}); err != nil {
				fmt.Printf("[WARN] WS write event failed account=%s: %v\n", accountID, err)
				return
			}
		case <-ticker.C:
			// Keepalive ping — mantiene el WS activo a través de proxies con idle timeout
			if err := handlers.WriteWSJSON(ctx, wsConn, outboundPayload{Event: "ping", Data: map[string]any{}}); err != nil {
				fmt.Printf("[WARN] WS ping failed account=%s: %v\n", accountID, err)
				return
			}
		case <-ctx.Done():
			// El cliente WS cerró la conexión (cierre normal del browser, timeout de red, etc.)
			return
		}
	}
}

type adminWebhooksListResponse struct {
	OK       bool             `json:"ok"`
	Webhooks []domain.Webhook `json:"webhooks"`
	Total    int              `json:"total"`
}

func (h *AdminHandler) ListTelefonoWebhooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	telefonoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || telefonoID <= 0 {
		writeAdminError(w, http.StatusBadRequest, "invalid telefono ID")
		return
	}

	phone, err := h.telefonoStore.GetByID(telefonoID)
	if err != nil || phone == nil {
		writeAdminError(w, http.StatusNotFound, "teléfono no encontrado")
		return
	}

	access, ok := getPanelAdminAccess(r)
	if !ok {
		writeAdminError(w, http.StatusUnauthorized, "token requerido")
		return
	}
	if !access.CanAccessEmpresa(phone.EmpresaID) {
		writeAdminError(w, http.StatusForbidden, "acceso denegado")
		return
	}

	if h.webhookStore == nil {
		writeAdminJSON(w, http.StatusOK, adminWebhooksListResponse{OK: true, Webhooks: []domain.Webhook{}, Total: 0})
		return
	}

	hooks, err := h.webhookStore.ListByTelefono(telefonoID)
	if err != nil {
		writeAdminError(w, http.StatusInternalServerError, "error al obtener webhooks")
		return
	}
	if hooks == nil {
		hooks = []domain.Webhook{}
	}

	writeAdminJSON(w, http.StatusOK, adminWebhooksListResponse{OK: true, Webhooks: hooks, Total: len(hooks)})
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

type outboundPayload struct {
	Event string         `json:"event"`
	Data  map[string]any `json:"data"`
}


