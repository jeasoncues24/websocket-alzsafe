package domain

import (
	"context"
	"testing"
)

type mockEmpresa struct {
	ID     int64
	RUC    string
	Nombre string
}

type mockEmpresaStore struct {
	empresas map[int64]*mockEmpresa
	byRUC    map[string]*mockEmpresa
}

func (m *mockEmpresaStore) GetByID(id int64) (*Empresa, error) {
	if e, ok := m.empresas[id]; ok {
		return &Empresa{ID: e.ID, RUC: e.RUC, Nombre: e.Nombre}, nil
	}
	return nil, nil
}

func (m *mockEmpresaStore) GetByRUC(ruc string) (*Empresa, error) {
	if e, ok := m.byRUC[ruc]; ok {
		return &Empresa{ID: e.ID, RUC: e.RUC, Nombre: e.Nombre}, nil
	}
	return nil, nil
}

func (m *mockEmpresaStore) GetAll(page, limit int, search string, activo *bool) ([]Empresa, int, error) {
	result := []Empresa{}
	for _, e := range m.empresas {
		result = append(result, Empresa{ID: e.ID, RUC: e.RUC, Nombre: e.Nombre})
	}
	return result, len(result), nil
}

func (m *mockEmpresaStore) Create(empresa *Empresa) (int64, error) {
	return 0, nil
}

func (m *mockEmpresaStore) Update(empresa *Empresa) error {
	return nil
}

func (m *mockEmpresaStore) Delete(id int64) error {
	return nil
}

func (m *mockEmpresaStore) IncrementTokenVersion(id int64) (int, error) {
	return 2, nil
}

func (m *mockEmpresaStore) Restore(id int64) error { return nil }

func TestGetEmpresaFilter_AdminJWTGlobal(t *testing.T) {
	claims := &AdminJWTClaims{
		UserID:   1,
		Username: "user1",
		IsRoot:   false,
	}
	ctx := WithAdminJWTClaims(context.Background(), claims)

	filter, ok := GetEmpresaFilter(ctx, "")

	if !ok {
		t.Fatal("Expected filter to be valid")
	}
	if filter.IsRoot {
		t.Error("Expected IsRoot to be false for normal user")
	}
	if filter.EmpresaID != nil {
		t.Error("Expected EmpresaID to be nil for admin JWT without header")
	}
}

func TestGetEmpresaFilter_RootSinEmpresa(t *testing.T) {
	claims := &AdminJWTClaims{
		UserID:   1,
		Username: "root",
		IsRoot:   true,
	}
	ctx := WithAdminJWTClaims(context.Background(), claims)

	filter, ok := GetEmpresaFilter(ctx, "")

	if !ok {
		t.Fatal("Expected filter to be valid")
	}
	if !filter.IsRoot {
		t.Error("Expected IsRoot to be true")
	}
	if filter.EmpresaID != nil {
		t.Error("Expected EmpresaID to be nil for root without header")
	}
}

func TestGetEmpresaFilter_RootConHeader(t *testing.T) {
	claims := &AdminJWTClaims{
		UserID:   1,
		Username: "root",
		IsRoot:   true,
	}
	ctx := WithAdminJWTClaims(context.Background(), claims)

	filter, ok := GetEmpresaFilter(ctx, "20")

	if !ok {
		t.Fatal("Expected filter to be valid")
	}
	if !filter.IsRoot {
		t.Error("Expected IsRoot to be true")
	}
	if filter.EmpresaID == nil || *filter.EmpresaID != 20 {
		t.Error("Expected EmpresaID to be 20 from header")
	}
}

func TestGetRUCFromContext_UsuarioNormal(t *testing.T) {
	mockStore := &mockEmpresaStore{
		empresas: map[int64]*mockEmpresa{10: {ID: 10, RUC: "12345678", Nombre: "Empresa1"}},
		byRUC:    map[string]*mockEmpresa{"12345678": {ID: 10, RUC: "12345678", Nombre: "Empresa1"}},
	}

	filter := &EmpresaFilter{
		IsRoot:    false,
		EmpresaID: int64Ptr(10),
	}

	ruc, err := GetRUCFromContext(context.Background(), filter, mockStore)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ruc != "12345678" {
		t.Errorf("Expected RUC '12345678', got '%s'", ruc)
	}
}

func TestGetRUCFromContext_RootConHeader(t *testing.T) {
	mockStore := &mockEmpresaStore{
		empresas: map[int64]*mockEmpresa{20: {ID: 20, RUC: "87654321", Nombre: "Empresa2"}},
		byRUC:    map[string]*mockEmpresa{"87654321": {ID: 20, RUC: "87654321", Nombre: "Empresa2"}},
	}

	filter := &EmpresaFilter{
		IsRoot:    true,
		EmpresaID: nil,
	}

	ctx := context.WithValue(context.Background(), "x-empresa-id", "20")
	ruc, err := GetRUCFromContext(ctx, filter, mockStore)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ruc != "87654321" {
		t.Errorf("Expected RUC '87654321', got '%s'", ruc)
	}
}

func TestGetRUCFromContext_RootSinHeader(t *testing.T) {
	mockStore := &mockEmpresaStore{}

	filter := &EmpresaFilter{
		IsRoot:    true,
		EmpresaID: nil,
	}

	ruc, err := GetRUCFromContext(context.Background(), filter, mockStore)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ruc != "" {
		t.Errorf("Expected empty RUC, got '%s'", ruc)
	}
}

func TestGetAllEmpresaRUCs(t *testing.T) {
	mockStore := &mockEmpresaStore{
		empresas: map[int64]*mockEmpresa{
			1: {ID: 1, RUC: "11111111", Nombre: "Empresa1"},
			2: {ID: 2, RUC: "22222222", Nombre: "Empresa2"},
		},
		byRUC: map[string]*mockEmpresa{
			"11111111": {ID: 1, RUC: "11111111", Nombre: "Empresa1"},
			"22222222": {ID: 2, RUC: "22222222", Nombre: "Empresa2"},
		},
	}

	rucs, err := GetAllEmpresaRUCs(mockStore)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(rucs) != 2 {
		t.Errorf("Expected 2 RUCs, got %d", len(rucs))
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}
