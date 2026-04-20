package whatsapp

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/coder/websocket"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/socket"
	_ "modernc.org/sqlite"
	sqlite "modernc.org/sqlite"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

func init() {
	sqlite.MustRegisterScalarFunction("NOW", 0, func(_ *sqlite.FunctionContext, _ []driver.Value) (driver.Value, error) {
		return time.Now().Format("2006-01-02 15:04:05"), nil
	})
}

func TestStartupBootstrapKeepsHealthyRuntimeStable(t *testing.T) {
	db := newStartupBootstrapTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	manager := NewManager()
	sessionStore := storage.NewSessionStore()
	accountID := "51999888777"
	telefonoID := insertBootstrapTelefono(t, db, 1, "+51", "999888777", accountID, string(domain.TelefonoStatusActive))
	_ = telefonoID
	manager.Set(accountID, connectedWhatsAppClient(t))

	bootstrap := NewStartupBootstrapper(manager, sessionStore, telefonoStore, StartupBootstrapConfig{MaxConcurrency: 1, MaxRetries: 0, RetryDelay: time.Millisecond})

	summary1 := bootstrap.Run(context.Background())
	summary2 := bootstrap.Run(context.Background())

	if summary1.TotalTelefonos != 1 || summary1.ActivosEnDB != 1 || summary1.RuntimeActivos != 1 || summary1.MismatchesDetectados != 0 || summary1.IntentosStart != 0 || summary1.ErroresStart != 0 {
		t.Fatalf("unexpected first summary: %+v", summary1)
	}
	if summary2.MismatchesDetectados != 0 || summary2.IntentosStart != 0 || summary2.ErroresStart != 0 {
		t.Fatalf("unexpected second summary: %+v", summary2)
	}

	phone, err := telefonoStore.GetByID(telefonoID)
	if err != nil {
		t.Fatalf("get telefono: %v", err)
	}
	if phone == nil || phone.Status != domain.TelefonoStatusActive {
		t.Fatalf("expected phone to remain active, got %+v", phone)
	}

	state, ok := sessionStore.Get(accountID)
	if !ok || state.Status != "active" {
		t.Fatalf("expected active session state, got %+v ok=%v", state, ok)
	}
}

func TestStartupBootstrapReconcilesDisconnectedDbWithRuntimeConnected(t *testing.T) {
	db := newStartupBootstrapTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	manager := NewManager()
	sessionStore := storage.NewSessionStore()
	accountID := "51999888777"
	telefonoID := insertBootstrapTelefono(t, db, 1, "+51", "999888777", accountID, string(domain.TelefonoStatusDisconnected))
	manager.Set(accountID, connectedWhatsAppClient(t))

	bootstrap := NewStartupBootstrapper(manager, sessionStore, telefonoStore, StartupBootstrapConfig{MaxConcurrency: 1, MaxRetries: 0, RetryDelay: time.Millisecond})
	summary := bootstrap.Run(context.Background())

	if summary.TotalTelefonos != 1 || summary.ActivosEnDB != 0 || summary.RuntimeActivos != 1 || summary.MismatchesDetectados != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	phone, err := telefonoStore.GetByID(telefonoID)
	if err != nil {
		t.Fatalf("get telefono: %v", err)
	}
	if phone == nil || phone.Status != domain.TelefonoStatusActive || phone.LastConnected == nil {
		t.Fatalf("expected reconciled active phone, got %+v", phone)
	}

	state, ok := sessionStore.Get(accountID)
	if !ok || state.Status != "active" {
		t.Fatalf("expected active session state, got %+v ok=%v", state, ok)
	}
}

func TestStartupBootstrapStartsMissingRuntimeSession(t *testing.T) {
	db := newStartupBootstrapTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	manager := NewManager()
	sessionStore := storage.NewSessionStore()
	accountID := "51999888777"
	insertBootstrapTelefono(t, db, 1, "+51", "999888777", accountID, string(domain.TelefonoStatusActive))

	bootstrap := NewStartupBootstrapper(manager, sessionStore, telefonoStore, StartupBootstrapConfig{MaxConcurrency: 1, MaxRetries: 0, RetryDelay: time.Millisecond})
	summary := bootstrap.Run(context.Background())

	if summary.TotalTelefonos != 1 || summary.ActivosEnDB != 1 || summary.RuntimeActivos != 0 || summary.MismatchesDetectados != 1 || summary.IntentosStart != 1 || summary.ErroresStart != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	state, ok := sessionStore.Get(accountID)
	if !ok || state.Status != "initializing" {
		t.Fatalf("expected initializing state, got %+v ok=%v", state, ok)
	}
}

func newStartupBootstrapTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared")
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

func insertBootstrapTelefono(t *testing.T, db *sql.DB, empresaID int64, codigoPais, numero, numeroCompleto, status string) int64 {
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
	setUnexportedField(t, fs, "conn", &websocket.Conn{})
	noise := &socket.NoiseSocket{}
	setUnexportedField(t, noise, "fs", fs)
	setUnexportedField(t, client, "socket", noise)
	return client
}

func setUnexportedField(t *testing.T, target any, field string, value any) {
	t.Helper()
	rv := reflect.ValueOf(target).Elem().FieldByName(field)
	if !rv.IsValid() {
		t.Fatalf("field %s not found", field)
	}
	reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}
