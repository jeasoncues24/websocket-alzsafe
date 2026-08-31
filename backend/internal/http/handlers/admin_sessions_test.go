package http

import (
	"database/sql"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/coder/websocket"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/socket"
	_ "modernc.org/sqlite"

	"wsapi/internal/config"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

func TestGetSessionsSorting(t *testing.T) {
	db := newAdminSessionsTestDB(t)
	empresaStore := storage.NewEmpresaStore(db)
	telefonoStore := storage.NewTelefonoStore(db)
	manager := whatsapp.NewManager()
	sessionStore := storage.NewSessionStore()
	jwtCfg := &config.JWTConfig{Secret: "test-secret"}

	// Insertar empresa
	empresaID := insertAdminSessionsEmpresa(t, db, "Empresa Alpha", "Alpha", "20100000001")

	// Teléfono A: disconnected en DB, pero runtime online en memoria
	accA := "51900000001"
	insertAdminSessionsTelefono(t, db, empresaID, "+51", "900000001", accA, "disconnected")
	manager.Set(accA, connectedWhatsAppClient(t))

	// Teléfono B: active en DB, offline en runtime
	accB := "51900000002"
	insertAdminSessionsTelefono(t, db, empresaID, "+51", "900000002", accB, "active")

	// Teléfono C: disconnected en DB, offline en runtime
	accC := "51900000003"
	insertAdminSessionsTelefono(t, db, empresaID, "+51", "900000003", accC, "disconnected")

	// Teléfono D: active en DB Y runtime online en memoria (máxima prioridad)
	accD := "51900000004"
	insertAdminSessionsTelefono(t, db, empresaID, "+51", "900000004", accD, "active")
	manager.Set(accD, connectedWhatsAppClient(t))

	handler := NewAdminSessionsHandler(empresaStore, telefonoStore, manager, sessionStore, jwtCfg)

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/admin/sesiones", nil)
	rec := httptest.NewRecorder()

	handler.GetSessions(rec, req)

	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		OK       bool             `json:"ok"`
		Sessions []sessionInfoDTO `json:"sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(resp.Sessions) != 4 {
		t.Fatalf("expected 4 sessions, got %d", len(resp.Sessions))
	}

	// Orden esperado:
	// 1. accD (runtime_connected: true, status: active)
	// 2. accA (runtime_connected: true, status: disconnected)
	// 3. accB (runtime_connected: false, status: active)
	// 4. accC (runtime_connected: false, status: disconnected)
	expectedOrder := []string{accD, accA, accB, accC}
	for i, expectedAcc := range expectedOrder {
		if resp.Sessions[i].AccountID != expectedAcc {
			t.Errorf("pos %d: expected account %s (runtimeConnected=%v, status=%s), got %s (runtimeConnected=%v, status=%s)",
				i, expectedAcc, (i < 2), resp.Sessions[i].Status,
				resp.Sessions[i].AccountID, resp.Sessions[i].RuntimeConnected, resp.Sessions[i].Status)
		}
	}
}

func newAdminSessionsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	schema := `
CREATE TABLE empresas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ruc TEXT NOT NULL UNIQUE,
    nombre TEXT NOT NULL,
    nombre_comercial TEXT,
    telefono_contacto TEXT,
    direccion TEXT,
    token_version INTEGER NOT NULL DEFAULT 1,
    permissions TEXT,
    activo INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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
);
`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func insertAdminSessionsEmpresa(t *testing.T, db *sql.DB, nombre, nombreComercial, ruc string) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO empresas (nombre, nombre_comercial, ruc, telefono_contacto, direccion, activo) VALUES (?, ?, ?, '', '', 1)`, nombre, nombreComercial, ruc)
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func insertAdminSessionsTelefono(t *testing.T, db *sql.DB, empresaID int64, codigoPais, numero, numeroCompleto, status string) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO telefonos (empresa_id, codigo_pais, numero, numero_completo, status) VALUES (?, ?, ?, ?, ?)`, empresaID, codigoPais, numero, numeroCompleto, status)
	if err != nil {
		t.Fatalf("insert telefono: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func connectedWhatsAppClient(t *testing.T) *whatsmeow.Client {
	t.Helper()
	client := &whatsmeow.Client{}
	fs := &socket.FrameSocket{}
	setUnexportedAtomicPointer(t, fs, "conn", &websocket.Conn{})
	noise := &socket.NoiseSocket{}
	setUnexportedField(t, noise, "fs", fs)
	setUnexportedField(t, client, "socket", noise)
	return client
}

func setUnexportedField(t *testing.T, target any, field string, value any) {
	t.Helper()
	rv := reflectField(t, target, field)
	reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

func setUnexportedAtomicPointer[T any](t *testing.T, target any, field string, value *T) {
	t.Helper()
	rv := reflectField(t, target, field)
	ptr := reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Interface().(*atomic.Pointer[T])
	ptr.Store(value)
}

func reflectField(t *testing.T, target any, field string) reflect.Value {
	t.Helper()
	rv := reflect.ValueOf(target).Elem().FieldByName(field)
	if !rv.IsValid() {
		t.Fatalf("field %s not found", field)
	}
	return rv
}
