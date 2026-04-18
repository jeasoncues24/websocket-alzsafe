package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"wsapi/internal/auth"
	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type CompaniesHandler struct {
	empresaStore domain.EmpresaStoreInterface
	sessionStore *storage.SessionStore
	jwtConfig    *config.JWTConfig
}

func NewCompaniesHandler(empresaStore domain.EmpresaStoreInterface, sessionStore *storage.SessionStore, jwtConfig *config.JWTConfig) *CompaniesHandler {
	return &CompaniesHandler{empresaStore: empresaStore, sessionStore: sessionStore, jwtConfig: jwtConfig}
}

// List empresas con filtros
func (h *CompaniesHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())

	// Parse query params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	search := r.URL.Query().Get("busqueda")
	estado := r.URL.Query().Get("estado")

	var activo *bool
	if estado == "activo" {
		activo = &[]bool{true}[0]
	} else if estado == "inactivo" {
		activo = &[]bool{false}[0]
	}

	// Si no es super_admin, solo puede ver su empresa
	if claims.Rol != domain.RoleSuperAdmin {
		if claims.EmpresaID == nil {
			http.Error(w, `{"ok": false, "error": "Acceso denegado a esta empresa"}`, http.StatusForbidden)
			return
		}
		empresa, err := h.empresaStore.GetByID(*claims.EmpresaID)
		if err != nil || empresa == nil {
			http.Error(w, `{"ok": false, "error": "Empresa no encontrada"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(domain.EmpresasListResponse{
			OK:       true,
			Empresas: []domain.Empresa{*empresa},
			Total:    1,
			Page:     1,
			Limit:    1,
		})
		return
	}

	empresas, total, err := h.empresaStore.GetAll(page, limit, search, activo)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al obtener empresas"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.EmpresasListResponse{
		OK:       true,
		Empresas: empresas,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

// Get empresa by ID (para /api/admin/empresas/{id} y /api/companies/{id})
func (h *CompaniesHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())

	empresa, err := h.empresaStore.GetByID(id)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al obtener empresa"}`, http.StatusInternalServerError)
		return
	}
	if empresa == nil {
		http.Error(w, `{"ok": false, "error": "Empresa no encontrada"}`, http.StatusNotFound)
		return
	}

	// Verificar acceso
	if claims.Rol != domain.RoleSuperAdmin {
		if claims.EmpresaID == nil || *claims.EmpresaID != empresa.ID {
			http.Error(w, `{"ok": false, "error": "Acceso denegado a esta empresa"}`, http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.EmpresaResponse{
		OK:      true,
		Empresa: empresa,
	})
}

// Create empresa
func (h *CompaniesHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var req domain.EmpresaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"ok": false, "error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	// Validaciones básicas
	if req.RUC == "" || req.Nombre == "" {
		http.Error(w, `{"ok": false, "error": "RUC y nombre son requeridos"}`, http.StatusBadRequest)
		return
	}

	// Verificar RUC único
	existing, _ := h.empresaStore.GetByRUC(req.RUC)
	if existing != nil {
		http.Error(w, `{"ok": false, "error": "Ya existe una empresa con este RUC"}`, http.StatusConflict)
		return
	}

	empresa := domain.NewEmpresa(req.RUC, req.Nombre)
	empresa.NombreComercial = req.NombreComercial
	empresa.Telefono = req.Telefono
	if req.Direccion != "" {
		empresa.Direccion = &req.Direccion
	}

	id, err := h.empresaStore.Create(empresa)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al crear empresa"}`, http.StatusInternalServerError)
		return
	}

	empresa.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(domain.EmpresaResponse{
		OK:      true,
		Empresa: empresa,
	})
}

// Update empresa (para /api/admin/empresas/{id} y /api/companies/{id})
func (h *CompaniesHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())

	empresa, err := h.empresaStore.GetByID(id)
	if err != nil || empresa == nil {
		http.Error(w, `{"ok": false, "error": "Empresa no encontrada"}`, http.StatusNotFound)
		return
	}

	// Verificar acceso
	if claims.Rol != domain.RoleSuperAdmin {
		if claims.EmpresaID == nil || *claims.EmpresaID != empresa.ID {
			http.Error(w, `{"ok": false, "error": "Acceso denegado a esta empresa"}`, http.StatusForbidden)
			return
		}
	}

	var req domain.EmpresaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"ok": false, "error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if req.Nombre != "" {
		empresa.Nombre = req.Nombre
	}
	if req.NombreComercial != "" {
		empresa.NombreComercial = req.NombreComercial
	}
	if req.Telefono != "" {
		empresa.Telefono = req.Telefono
	}
	if req.Direccion != "" {
		empresa.Direccion = &req.Direccion
	}

	if err := h.empresaStore.Update(empresa); err != nil {
		http.Error(w, `{"ok": false, "error": "Error al actualizar empresa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.EmpresaResponse{
		OK:      true,
		Empresa: empresa,
	})
}

// Delete empresa (soft delete) (para /api/admin/empresas/{id} y /api/companies/{id})
func (h *CompaniesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	claims, _ := domain.GetAdminJWTClaims(r.Context())

	empresa, err := h.empresaStore.GetByID(id)
	if err != nil || empresa == nil {
		http.Error(w, `{"ok": false, "error": "Empresa no encontrada"}`, http.StatusNotFound)
		return
	}

	// Verificar acceso
	if claims.Rol != domain.RoleSuperAdmin {
		if claims.EmpresaID == nil || *claims.EmpresaID != empresa.ID {
			http.Error(w, `{"ok": false, "error": "Acceso denegado a esta empresa"}`, http.StatusForbidden)
			return
		}
	}

	// Verificar si tiene sesiones activas (simplificado)
	if h.sessionStore != nil {
		if state, ok := h.sessionStore.Get(empresa.RUC); ok && state.IsActive && state.Status == "active" {
			http.Error(w, `{"ok": false, "error": "La empresa tiene sesiones WhatsApp activas"}`, http.StatusConflict)
			return
		}
	}

	if !empresa.Activo {
		http.Error(w, `{"ok": false, "error": "La empresa ya está inactiva"}`, http.StatusConflict)
		return
	}

	if err := h.empresaStore.Delete(id); err != nil {
		http.Error(w, `{"ok": false, "error": "Error al eliminar empresa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// GetCurrent returns the authenticated company's own profile.
// GET /api/empresas
func (h *CompaniesHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		http.Error(w, `{"ok": false, "error": "Autenticación de empresa requerida"}`, http.StatusUnauthorized)
		return
	}

	empresa, err := h.empresaStore.GetByID(claims.EmpresaID)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al obtener empresa"}`, http.StatusInternalServerError)
		return
	}
	if empresa == nil {
		http.Error(w, `{"ok": false, "error": "Empresa no encontrada"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.EmpresaResponse{
		OK:      true,
		Empresa: empresa,
	})
}

// UpdateCurrent updates the authenticated company's own profile.
// RUC remains read-only for the company contract.
// PUT /api/empresas
func (h *CompaniesHandler) UpdateCurrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	claims, ok := domain.GetEmpresaJWTClaims(r.Context())
	if !ok {
		http.Error(w, `{"ok": false, "error": "Autenticación de empresa requerida"}`, http.StatusUnauthorized)
		return
	}

	empresa, err := h.empresaStore.GetByID(claims.EmpresaID)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al obtener empresa"}`, http.StatusInternalServerError)
		return
	}
	if empresa == nil {
		http.Error(w, `{"ok": false, "error": "Empresa no encontrada"}`, http.StatusNotFound)
		return
	}

	var req domain.EmpresaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"ok": false, "error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if req.RUC != "" && req.RUC != empresa.RUC {
		http.Error(w, `{"ok": false, "error": "El RUC es de solo lectura"}`, http.StatusBadRequest)
		return
	}

	if req.Nombre != "" {
		empresa.Nombre = req.Nombre
	}
	if req.NombreComercial != "" {
		empresa.NombreComercial = req.NombreComercial
	}
	if req.Telefono != "" {
		empresa.Telefono = req.Telefono
	}
	if req.Direccion != "" {
		empresa.Direccion = &req.Direccion
	}

	if err := h.empresaStore.Update(empresa); err != nil {
		http.Error(w, `{"ok": false, "error": "Error al actualizar empresa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.EmpresaResponse{
		OK:      true,
		Empresa: empresa,
	})
}

// generateToken emite un JWT de larga duración (5 años) para una empresa.
// Solo accesible por super_admin.
// POST /api/admin/empresas/{id}/token
func (h *CompaniesHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	claims, ok := domain.GetAdminJWTClaims(r.Context())
	if !ok || claims.Rol != domain.RoleSuperAdmin {
		http.Error(w, `{"ok": false, "error": "Solo super_admin puede generar JWT de empresa"}`, http.StatusForbidden)
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	empresa, err := h.empresaStore.GetByID(id)
	if err != nil || empresa == nil {
		http.Error(w, `{"ok": false, "error": "Empresa no encontrada"}`, http.StatusNotFound)
		return
	}
	if !empresa.Activo {
		http.Error(w, `{"ok": false, "error": "La empresa está inactiva"}`, http.StatusConflict)
		return
	}

	token, err := auth.GenerateEmpresaJWT(empresa, h.jwtConfig.Secret, h.jwtConfig.Issuer)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al generar JWT de empresa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.EmpresaJWTResponse{
		OK:      true,
		Token:   token,
		Message: "JWT de empresa generado exitosamente. Guárdalo en un lugar seguro.",
	})
}

// RevokeToken incrementa el token_version de la empresa e invalida todos los JWT activos.
// Solo accesible por super_admin.
// POST /api/admin/empresas/{id}/token/revoke
func (h *CompaniesHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	claims, ok := domain.GetAdminJWTClaims(r.Context())
	if !ok || claims.Rol != domain.RoleSuperAdmin {
		http.Error(w, `{"ok": false, "error": "Solo super_admin puede revocar JWT de empresa"}`, http.StatusForbidden)
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	newVersion, err := h.empresaStore.IncrementTokenVersion(id)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al revocar JWT de empresa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":            true,
		"token_version": newVersion,
		"message":       "Todos los JWT de empresa han sido revocados",
	})
}

// extractIDFromPath extrae el ID del path para rutas de compatibilidad como /api/admin/empresas/{id}/token.
func (h *CompaniesHandler) extractIDFromPath(path string) int64 {
	paths := []string{
		"/api/admin/empresas/",
		"/api/empresas/",
		"/api/companies/",
	}
	for _, p := range paths {
		idStr := strings.TrimPrefix(path, p)
		if idStr != path {
			idStr = strings.TrimSuffix(idStr, "/token")
			idStr = strings.TrimSuffix(idStr, "/token/revoke")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil && id > 0 {
				return id
			}
			break
		}
	}
	return 0
}
