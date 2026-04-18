package http

import (
	"bytes"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type mockEmpresaStore struct {
	empresas   map[int64]*domain.Empresa
	lastPage   int
	lastLimit  int
	lastSearch string
	lastActivo *bool
	nextID     int64
}

func newMockEmpresaStore() *mockEmpresaStore {
	return &mockEmpresaStore{
		empresas: map[int64]*domain.Empresa{},
		nextID:   3,
	}
}

func (m *mockEmpresaStore) GetByID(id int64) (*domain.Empresa, error) {
	if e, ok := m.empresas[id]; ok {
		copy := *e
		return &copy, nil
	}
	return nil, nil
}

func (m *mockEmpresaStore) GetByRUC(ruc string) (*domain.Empresa, error) {
	for _, e := range m.empresas {
		if e.RUC == ruc {
			copy := *e
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *mockEmpresaStore) GetAll(page, limit int, search string, activo *bool) ([]domain.Empresa, int, error) {
	m.lastPage = page
	m.lastLimit = limit
	m.lastSearch = search
	m.lastActivo = activo

	result := make([]domain.Empresa, 0, len(m.empresas))
	for _, e := range m.empresas {
		if search != "" && !strings.Contains(strings.ToLower(e.Nombre), strings.ToLower(search)) && !strings.Contains(strings.ToLower(e.RUC), strings.ToLower(search)) {
			continue
		}
		if activo != nil && e.Activo != *activo {
			continue
		}
		result = append(result, *e)
	}
	return result, len(result), nil
}

func (m *mockEmpresaStore) Create(empresa *domain.Empresa) (int64, error) {
	id := m.nextID
	m.nextID++
	copy := *empresa
	copy.ID = id
	m.empresas[id] = &copy
	return id, nil
}

func (m *mockEmpresaStore) Update(empresa *domain.Empresa) error {
	copy := *empresa
	m.empresas[empresa.ID] = &copy
	return nil
}

func (m *mockEmpresaStore) Delete(id int64) error {
	if e, ok := m.empresas[id]; ok {
		e.Activo = false
	}
	return nil
}

func (m *mockEmpresaStore) IncrementTokenVersion(id int64) (int, error) {
	if e, ok := m.empresas[id]; ok {
		e.TokenVersion++
		return e.TokenVersion, nil
	}
	return 0, nil
}

func TestCompaniesListFiltersAndPagination(t *testing.T) {
	store := newMockEmpresaStore()
	store.empresas[1] = &domain.Empresa{ID: 1, RUC: "20100000001", Nombre: "Empresa Uno", Activo: true}
	store.empresas[2] = &domain.Empresa{ID: 2, RUC: "20100000002", Nombre: "Empresa Dos", Activo: false}
	h := &CompaniesHandler{empresaStore: store, jwtConfig: &config.JWTConfig{}}

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/admin/empresas?page=2&limit=5&busqueda=Uno&estado=activo", nil)
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{UserID: 1, Username: "root", Rol: domain.RoleSuperAdmin, IsRoot: true}))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if store.lastPage != 2 || store.lastLimit != 5 {
		t.Fatalf("unexpected pagination args: page=%d limit=%d", store.lastPage, store.lastLimit)
	}
	if store.lastSearch != "Uno" || store.lastActivo == nil || !*store.lastActivo {
		t.Fatalf("unexpected filters: search=%q activo=%v", store.lastSearch, store.lastActivo)
	}
}

func TestCompaniesCreateUpdateDeleteAndTokens(t *testing.T) {
	store := newMockEmpresaStore()
	store.empresas[1] = &domain.Empresa{ID: 1, RUC: "20100000001", Nombre: "Empresa Uno", Activo: true, TokenVersion: 1}
	store.empresas[2] = &domain.Empresa{ID: 2, RUC: "20100000002", Nombre: "Empresa Dos", Activo: true, TokenVersion: 1}
	sessionStore := storage.NewSessionStore()
	h := &CompaniesHandler{empresaStore: store, sessionStore: sessionStore, jwtConfig: &config.JWTConfig{Secret: "secret", Issuer: "wsapi"}}
	claims := &domain.AdminJWTClaims{UserID: 1, Username: "root", Rol: domain.RoleSuperAdmin, IsRoot: true}

	createReq := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/empresas", bytes.NewBufferString(`{"ruc":"20100000003","nombre":"Empresa Tres","nombre_comercial":"Tres SA","telefono_contacto":"999","direccion":"Calle 1"}`))
	createReq = createReq.WithContext(domain.WithAdminJWTClaims(createReq.Context(), claims))
	createRR := httptest.NewRecorder()
	h.Create(createRR, createReq)
	if createRR.Code != stdhttp.StatusCreated {
		t.Fatalf("expected 201 creating company, got %d", createRR.Code)
	}

	dupReq := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/empresas", bytes.NewBufferString(`{"ruc":"20100000001","nombre":"Dup"}`))
	dupReq = dupReq.WithContext(domain.WithAdminJWTClaims(dupReq.Context(), claims))
	dupRR := httptest.NewRecorder()
	h.Create(dupRR, dupReq)
	if dupRR.Code != stdhttp.StatusConflict {
		t.Fatalf("expected 409 for duplicate RUC, got %d", dupRR.Code)
	}

	updateReq := httptest.NewRequest(stdhttp.MethodPut, "/api/admin/empresas/1", bytes.NewBufferString(`{"nombre":"Empresa Uno Updated","telefono_contacto":"111"}`))
	updateReq = updateReq.WithContext(domain.WithAdminJWTClaims(updateReq.Context(), claims))
	updateRR := httptest.NewRecorder()
	h.Update(updateRR, updateReq)
	if updateRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 updating company, got %d", updateRR.Code)
	}

	updated, _ := store.GetByID(1)
	if updated.Nombre != "Empresa Uno Updated" || updated.Telefono != "111" {
		t.Fatalf("update was not applied: %+v", updated)
	}

	activeReq := httptest.NewRequest(stdhttp.MethodDelete, "/api/admin/empresas/1", nil)
	activeReq = activeReq.WithContext(domain.WithAdminJWTClaims(activeReq.Context(), claims))
	activeRR := httptest.NewRecorder()
	sessionStore.SetActive("20100000001")
	h.Delete(activeRR, activeReq)
	if activeRR.Code != stdhttp.StatusConflict {
		t.Fatalf("expected 409 when sessions active, got %d", activeRR.Code)
	}

	sessionStore.SetDisconnected("20100000001", "closed")
	deleteReq := httptest.NewRequest(stdhttp.MethodDelete, "/api/admin/empresas/1", nil)
	deleteReq = deleteReq.WithContext(domain.WithAdminJWTClaims(deleteReq.Context(), claims))
	deleteRR := httptest.NewRecorder()
	h.Delete(deleteRR, deleteReq)
	if deleteRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 deleting company, got %d", deleteRR.Code)
	}

	inactive, _ := store.GetByID(1)
	if inactive.Activo {
		t.Fatalf("expected company to be inactive after delete")
	}

	tokenReq := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/empresas/2/token", nil)
	tokenReq = tokenReq.WithContext(domain.WithAdminJWTClaims(tokenReq.Context(), claims))
	tokenRR := httptest.NewRecorder()
	h.GenerateToken(tokenRR, tokenReq)
	if tokenRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 generating token, got %d", tokenRR.Code)
	}
	var tokenResp map[string]any
	if err := json.Unmarshal(tokenRR.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("invalid token response: %v", err)
	}
	if tokenResp["token"] == "" {
		t.Fatal("expected non-empty token")
	}

	revokeReq := httptest.NewRequest(stdhttp.MethodPost, "/api/admin/empresas/2/token/revoke", nil)
	revokeReq = revokeReq.WithContext(domain.WithAdminJWTClaims(revokeReq.Context(), claims))
	revokeRR := httptest.NewRecorder()
	h.RevokeToken(revokeRR, revokeReq)
	if revokeRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 revoking token, got %d", revokeRR.Code)
	}
}

func TestCompaniesListForNonSuperAdminReturnsOwnCompany(t *testing.T) {
	store := newMockEmpresaStore()
	store.empresas[1] = &domain.Empresa{ID: 1, RUC: "20100000001", Nombre: "Empresa Uno", Activo: true}
	store.empresas[2] = &domain.Empresa{ID: 2, RUC: "20100000002", Nombre: "Empresa Dos", Activo: true}
	h := &CompaniesHandler{empresaStore: store, jwtConfig: &config.JWTConfig{}}
	claims := &domain.AdminJWTClaims{UserID: 2, Username: "admin", Rol: domain.RoleAdmin, EmpresaID: int64Ptr(2)}

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/admin/empresas", nil)
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), claims))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp domain.EmpresasListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if len(resp.Empresas) != 1 || resp.Empresas[0].ID != 2 {
		t.Fatalf("expected only own company, got %+v", resp.Empresas)
	}
}

func TestCompaniesCurrentEndpoints(t *testing.T) {
	store := newMockEmpresaStore()
	store.empresas[2] = &domain.Empresa{ID: 2, RUC: "20100000002", Nombre: "Empresa Dos", Activo: true}
	h := &CompaniesHandler{empresaStore: store, jwtConfig: &config.JWTConfig{}}
	claims := &domain.EmpresaJWTClaims{EmpresaID: 2, TokenVersion: 1, EmpresaRUC: "20100000002", EmpresaNombre: "Empresa Dos"}

	getReq := httptest.NewRequest(stdhttp.MethodGet, "/api/empresas", nil)
	getReq = getReq.WithContext(domain.WithEmpresaJWTClaims(getReq.Context(), claims))
	getRR := httptest.NewRecorder()

	h.GetCurrent(getRR, getReq)

	if getRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 for current company, got %d", getRR.Code)
	}

	var getResp domain.EmpresaResponse
	if err := json.Unmarshal(getRR.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if getResp.Empresa == nil || getResp.Empresa.ID != 2 {
		t.Fatalf("expected company 2, got %+v", getResp.Empresa)
	}

	putReq := httptest.NewRequest(stdhttp.MethodPut, "/api/empresas", bytes.NewBufferString(`{"ruc":"20100000002","nombre":"Empresa Dos Actualizada","telefono_contacto":"111","direccion":"Nueva Calle"}`))
	putReq = putReq.WithContext(domain.WithEmpresaJWTClaims(putReq.Context(), claims))
	putRR := httptest.NewRecorder()

	h.UpdateCurrent(putRR, putReq)

	if putRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 updating current company, got %d", putRR.Code)
	}

	updated, _ := store.GetByID(2)
	if updated.Nombre != "Empresa Dos Actualizada" || updated.Telefono != "111" || updated.RUC != "20100000002" {
		t.Fatalf("update was not applied correctly: %+v", updated)
	}

	blockedReq := httptest.NewRequest(stdhttp.MethodPut, "/api/empresas", bytes.NewBufferString(`{"ruc":"99999999999","nombre":"Hack"}`))
	blockedReq = blockedReq.WithContext(domain.WithEmpresaJWTClaims(blockedReq.Context(), claims))
	blockedRR := httptest.NewRecorder()

	h.UpdateCurrent(blockedRR, blockedReq)

	if blockedRR.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400 for readonly ruc, got %d", blockedRR.Code)
	}
}

func int64Ptr(v int64) *int64 { return &v }
