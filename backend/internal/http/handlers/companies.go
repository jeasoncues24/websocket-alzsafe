package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	_, ok := domain.GetPanelAccess(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "Token requerido")
		return
	}

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

	empresas, total, err := h.empresaStore.GetAll(page, limit, search, activo)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al obtener empresas")
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
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	access, ok := domain.GetPanelAccess(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "Token requerido")
		return
	}

	empresa, err := h.empresaStore.GetByID(id)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al obtener empresa")
		return
	}
	if empresa == nil {
		writeAPIError(w, http.StatusNotFound, "Empresa no encontrada")
		return
	}

	if !access.CanAccessEmpresa(empresa.ID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta empresa")
		return
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
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var req domain.EmpresaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	// Validaciones básicas
	if req.RUC == "" || req.Nombre == "" {
		writeAPIError(w, http.StatusBadRequest, "RUC y nombre son requeridos")
		return
	}

	// Verificar RUC único
	existing, _ := h.empresaStore.GetByRUC(req.RUC)
	if existing != nil {
		writeAPIError(w, http.StatusConflict, "Ya existe una empresa con este RUC")
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
		writeAPIError(w, http.StatusInternalServerError, "Error al crear empresa")
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
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	access, ok := domain.GetPanelAccess(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "Token requerido")
		return
	}

	empresa, err := h.empresaStore.GetByID(id)
	if err != nil || empresa == nil {
		writeAPIError(w, http.StatusNotFound, "Empresa no encontrada")
		return
	}

	if !access.CanAccessEmpresa(empresa.ID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta empresa")
		return
	}

	var req domain.EmpresaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "JSON inválido")
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
		writeAPIError(w, http.StatusInternalServerError, "Error al actualizar empresa")
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
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	access, ok := domain.GetPanelAccess(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "Token requerido")
		return
	}

	empresa, err := h.empresaStore.GetByID(id)
	if err != nil || empresa == nil {
		writeAPIError(w, http.StatusNotFound, "Empresa no encontrada")
		return
	}

	if !access.CanAccessEmpresa(empresa.ID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta empresa")
		return
	}

	// Verificar si tiene sesiones activas (simplificado)
	if h.sessionStore != nil {
		if state, ok := h.sessionStore.Get(empresa.RUC); ok && state.IsActive && state.Status == "active" {
			writeAPIError(w, http.StatusConflict, "La empresa tiene sesiones WhatsApp activas")
			return
		}
	}

	if !empresa.Activo {
		writeAPIError(w, http.StatusConflict, "La empresa ya está inactiva")
		return
	}

	if err := h.empresaStore.Delete(id); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al eliminar empresa")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// Restore empresa (soft delete undo)
func (h *CompaniesHandler) Restore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	id := h.extractIDFromPath(r.URL.Path)
	if id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	empresa, err := h.empresaStore.GetByID(id)
	if err != nil || empresa == nil {
		writeAPIError(w, http.StatusNotFound, "Empresa no encontrada")
		return
	}

	access, ok := domain.GetPanelAccess(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "Token requerido")
		return
	}

	if !access.CanAccessEmpresa(empresa.ID) {
		writeAPIError(w, http.StatusForbidden, "Acceso denegado a esta empresa")
		return
	}

	if err := h.empresaStore.Restore(id); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al restaurar empresa")
		return
	}

	empresa.Activo = true

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "empresa": empresa})
}



// extractIDFromPath extrae el ID del path desde rutas de admin como /api/admin/empresas/{id}/token.
func (h *CompaniesHandler) extractIDFromPath(path string) int64 {
	paths := []string{
		"/api/admin/empresas/",
	}
	for _, p := range paths {
		idStr := strings.TrimPrefix(path, p)
		if idStr != path {
			idStr = strings.TrimSuffix(idStr, "/restore")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil && id > 0 {
				return id
			}
			break
		}
	}
	return 0
}
