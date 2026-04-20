package http

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	sqlite "modernc.org/sqlite"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

func init() {
	sqlite.MustRegisterScalarFunction("NOW", 0, func(_ *sqlite.FunctionContext, _ []driver.Value) (driver.Value, error) {
		return time.Now().Format("2006-01-02 15:04:05"), nil
	})
}

func TestV1MessagesGetMessageByReferenceReturnsRetryMetadata(t *testing.T) {
	db := newV1MessagesTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	msgRepo := storage.NewMessagesRepository(db)
	telefonoID := insertV1Telefono(t, db, 1, "+51", "999888777", "51999888777", string(domain.TelefonoStatusActive))
	insertV1Message(t, db, 1, telefonoID, "51911122233", "hola", "ref-1", string(domain.MessageStateFailed), "cliente caido", 2)

	h := NewV1MessagesHandler(msgRepo, telefonoStore, whatsapp.NewManager())
	req := httptest.NewRequest(http.MethodGet, "/api/mensajes/ref-1", nil)
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{EmpresaID: 1, TelefonoID: telefonoID}))
	rr := httptest.NewRecorder()

	h.GetMessageByReference(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			Message struct {
				ReferenceID string `json:"reference_id"`
				ErrorReason string `json:"error_reason"`
				RetryCount  int    `json:"retry_count"`
			} `json:"message"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok response")
	}
	if resp.Data.Message.ReferenceID != "ref-1" || resp.Data.Message.ErrorReason != "cliente caido" || resp.Data.Message.RetryCount != 2 {
		t.Fatalf("unexpected response payload: %+v", resp.Data.Message)
	}
}

func TestV1MessagesUpdateMessageUpdatesEditableMessage(t *testing.T) {
	db := newV1MessagesTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	msgRepo := storage.NewMessagesRepository(db)
	telefonoID := insertV1Telefono(t, db, 1, "+51", "999888777", "51999888777", string(domain.TelefonoStatusActive))
	insertV1Message(t, db, 1, telefonoID, "51911122233", "hola", "ref-edit", string(domain.MessageStateFailed), "fallo inicial", 0)

	h := NewV1MessagesHandler(msgRepo, telefonoStore, whatsapp.NewManager())
	body := `{"contenido":"contenido corregido"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/mensajes/ref-edit", strings.NewReader(body))
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{EmpresaID: 1, TelefonoID: telefonoID}))
	rr := httptest.NewRecorder()

	h.UpdateMessage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	msg, err := msgRepo.GetByReferenceID("ref-edit")
	if err != nil {
		t.Fatalf("get message: %v", err)
	}
	if msg == nil || msg.Contenido != "contenido corregido" {
		t.Fatalf("expected updated content, got %+v", msg)
	}
}

func TestV1MessagesUpdateMessageRejectsSentMessage(t *testing.T) {
	db := newV1MessagesTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	msgRepo := storage.NewMessagesRepository(db)
	telefonoID := insertV1Telefono(t, db, 1, "+51", "999888777", "51999888777", string(domain.TelefonoStatusActive))
	insertV1Message(t, db, 1, telefonoID, "51911122233", "hola", "ref-sent", string(domain.MessageStateSent), "", 0)

	h := NewV1MessagesHandler(msgRepo, telefonoStore, whatsapp.NewManager())
	req := httptest.NewRequest(http.MethodPatch, "/api/mensajes/ref-sent", strings.NewReader(`{"contenido":"no debe aplicar"}`))
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{EmpresaID: 1, TelefonoID: telefonoID}))
	rr := httptest.NewRecorder()

	h.UpdateMessage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	msg, err := msgRepo.GetByReferenceID("ref-sent")
	if err != nil {
		t.Fatalf("get message: %v", err)
	}
	if msg == nil || msg.Contenido != "hola" {
		t.Fatalf("sent message should remain unchanged, got %+v", msg)
	}
}

func TestV1MessagesRetryMessagePersistsRetryAndFailure(t *testing.T) {
	db := newV1MessagesTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	msgRepo := storage.NewMessagesRepository(db)
	telefonoID := insertV1Telefono(t, db, 1, "+51", "999888777", "51999888777", string(domain.TelefonoStatusActive))
	insertV1Message(t, db, 1, telefonoID, "51911122233", "hola", "ref-retry", string(domain.MessageStateFailed), "fallo inicial", 0)

	h := NewV1MessagesHandler(msgRepo, telefonoStore, whatsapp.NewManager())
	req := httptest.NewRequest(http.MethodPost, "/api/mensajes/ref-retry/reintentar", nil)
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{EmpresaID: 1, TelefonoID: telefonoID}))
	rr := httptest.NewRecorder()

	h.RetryMessage(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			ReferenceID string `json:"reference_id"`
			Estado      string `json:"estado"`
			Error       string `json:"error"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.OK {
		t.Fatalf("expected retry to fail without connected client")
	}
	if resp.Data.Estado != string(domain.MessageStateFailed) || resp.Data.ReferenceID != "ref-retry" {
		t.Fatalf("unexpected retry payload: %+v", resp.Data)
	}

	msg, err := msgRepo.GetByReferenceID("ref-retry")
	if err != nil {
		t.Fatalf("get message: %v", err)
	}
	if msg == nil || msg.RetryCount != 1 || msg.Estado != domain.MessageStateFailed || msg.ErrorReason == "" {
		t.Fatalf("retry metadata was not persisted: %+v", msg)
	}
	if msg.LastAttemptAt == nil {
		t.Fatalf("expected last_attempt_at to be set")
	}
}

func TestV1MessagesRetryMessageRejectsSentMessage(t *testing.T) {
	db := newV1MessagesTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	msgRepo := storage.NewMessagesRepository(db)
	telefonoID := insertV1Telefono(t, db, 1, "+51", "999888777", "51999888777", string(domain.TelefonoStatusActive))
	insertV1Message(t, db, 1, telefonoID, "51911122233", "hola", "ref-delivered", string(domain.MessageStateDelivered), "", 0)

	h := NewV1MessagesHandler(msgRepo, telefonoStore, whatsapp.NewManager())
	req := httptest.NewRequest(http.MethodPost, "/api/mensajes/ref-delivered/reintentar", nil)
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{EmpresaID: 1, TelefonoID: telefonoID}))
	rr := httptest.NewRecorder()

	h.RetryMessage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func newV1MessagesTestDB(t *testing.T) *sql.DB {
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

	_, err = db.Exec(`
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empresa_id INTEGER NOT NULL DEFAULT 0,
    telefono_id INTEGER NOT NULL DEFAULT 0,
    destino TEXT NOT NULL,
    contenido TEXT NOT NULL,
    adjuntos_json TEXT,
    estado TEXT NOT NULL DEFAULT 'pending',
    error_reason TEXT,
    reference_id TEXT UNIQUE NOT NULL,
    timestamp_created TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    timestamp_sent TIMESTAMP NULL,
    timestamp_confirmed TIMESTAMP NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP NULL,
    tiempo_envio TIMESTAMP NULL,
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

func insertV1Telefono(t *testing.T, db *sql.DB, empresaID int64, codigoPais, numero, numeroCompleto, status string) int64 {
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

func insertV1Message(t *testing.T, db *sql.DB, empresaID, telefonoID int64, destino, contenido, referenceID, estado, errorReason string, retryCount int) {
	t.Helper()
	var errorValue any
	if errorReason != "" {
		errorValue = errorReason
	}
	_, err := db.Exec(`
INSERT INTO messages (
    empresa_id, telefono_id, destino, contenido, adjuntos_json, estado, error_reason,
    reference_id, timestamp_created, retry_count, last_attempt_at
) VALUES (?, ?, ?, ?, 'null', ?, ?, ?, CURRENT_TIMESTAMP, ?, NULL)
`, empresaID, telefonoID, destino, contenido, estado, errorValue, referenceID, retryCount)
	if err != nil {
		t.Fatalf("insert message: %v", err)
	}
}
