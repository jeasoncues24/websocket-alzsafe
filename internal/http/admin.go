package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type AdminHandler struct {
	userStore   *storage.AdminUserStore
	roleStore   *storage.RoleStore
	moduleStore *storage.ModuleStore
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	if db == nil {
		return nil
	}
	return &AdminHandler{
		userStore:   storage.NewAdminUserStore(db),
		roleStore:   storage.NewRoleStore(db),
		moduleStore: storage.NewModuleStore(db),
	}
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

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

	hash := fmt.Sprintf("$plain$%s$", req.Password)
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

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.userStore.Delete(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error al eliminar usuario: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
		id, err = h.extractUserID(path, "/api/admin/users/promote/")
	} else if strings.Contains(path, "/users/") {
		id, err = h.extractUserID(path, "/api/admin/users/")
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

func (h *AdminHandler) PromoteUserByID(w http.ResponseWriter, r *http.Request) {
	h.PromoteUser(w, r)
}

type AssignModulesRequest struct {
	ModuleIDs []int64 `json:"module_ids"`
}

func (h *AdminHandler) AssignUserModules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	id, err := h.extractUserID(path, "/api/admin/users/modules/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req AssignModulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	cfg := config.Load()
	db, err := storage.NewDB(cfg)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM user_modules WHERE user_id = ?", id)
	if err != nil {
		http.Error(w, "error al limpiar módulos", http.StatusInternalServerError)
		return
	}

	for _, modID := range req.ModuleIDs {
		_, err = db.Exec("INSERT INTO user_modules (user_id, module_id) VALUES (?, ?)", id, modID)
		if err != nil {
			http.Error(w, "error al asignar módulo", http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AdminHandler) AssignUserModulesByID(w http.ResponseWriter, r *http.Request) {
	h.AssignUserModules(w, r)
}

func (h *AdminHandler) extractUserID(path, prefix string) (int64, error) {
	base := strings.TrimPrefix(path, prefix)
	parts := strings.Split(base, "/")
	if len(parts) < 1 {
		return 0, fmt.Errorf("id required")
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID")
	}
	return id, nil
}

func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roles, err := h.roleStore.GetAll()
	if err != nil {
		http.Error(w, "error al obtener roles", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"roles": roles,
	})
}

func (h *AdminHandler) ListModules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	modules, err := h.moduleStore.GetAll()
	if err != nil {
		http.Error(w, "error al obtener módulos", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"modules": modules,
	})
}

func getAdminHandler() *AdminHandler {
	cfg := config.Load()
	if cfg.DBHost == "" {
		return nil
	}
	db, err := storage.NewDB(cfg)
	if err != nil {
		return nil
	}
	return NewAdminHandler(db)
}
