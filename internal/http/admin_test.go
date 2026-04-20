package http

import (
	"database/sql"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

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
		UserID:    1,
		Username:  "admin",
		Rol:       domain.RoleAdmin,
		EmpresaID: adminTestInt64Ptr(1),
		IsRoot:    false,
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

func TestAdminPhoneAccessDeniedForOtherCompany(t *testing.T) {
	db := newAdminPhoneTestDB(t)
	store := storage.NewTelefonoStore(db)
	insertAdminPhone(t, db, 1, "+51", "999888777", "+51999888777")

	h := &AdminHandler{telefonoStore: store}
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/admin/telefonos/1", nil)
	req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), &domain.AdminJWTClaims{
		UserID:    2,
		Username:  "admin",
		Rol:       domain.RoleAdmin,
		EmpresaID: adminTestInt64Ptr(2),
	}))
	rr := httptest.NewRecorder()

	h.GetCompanyPhone(rr, req)

	if rr.Code != stdhttp.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
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

func adminTestInt64Ptr(v int64) *int64 { return &v }
