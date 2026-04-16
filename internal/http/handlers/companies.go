package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"wsapi/internal/domain"
)

type CompaniesHandler struct {
	empresaStore domain.EmpresaStoreInterface
}

func NewCompaniesHandler(empresaStore domain.EmpresaStoreInterface) *CompaniesHandler {
	return &CompaniesHandler{empresaStore: empresaStore}
}

// List empresas con filtros
func (h *CompaniesHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	claims, _ := domain.GetTokenClaims(r.Context())

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
	if claims.Rol != domain.RoleSuperAdmin && claims.EmpresaID != nil {
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

// Get empresa by ID
func (h *CompaniesHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	claims, _ := domain.GetTokenClaims(r.Context())

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
	empresa.Direccion = req.Direccion

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

// Update empresa
func (h *CompaniesHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	claims, _ := domain.GetTokenClaims(r.Context())

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
		empresa.Direccion = req.Direccion
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

// Delete empresa (soft delete)
func (h *CompaniesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	claims, _ := domain.GetTokenClaims(r.Context())

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
