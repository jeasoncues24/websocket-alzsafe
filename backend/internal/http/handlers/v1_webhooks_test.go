package http

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

func TestV1WebhooksCreateSuccessPersistsWebhook(t *testing.T) {
	db := newV1WebhooksTestDB(t)
	store := storage.NewWebhookStore(db)
	h := NewV1WebhooksHandler(store, 10)

	body := `{"url":"https://ejemplo.com/hook","eventos":["message.received"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/service/v1/webhooks", strings.NewReader(body))
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{ApiKeyID: 33, EmpresaID: 7, TelefonoID: 11}))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			ID     int64  `json:"id"`
			Secret string `json:"secret"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok response")
	}
	if resp.Data.ID <= 0 {
		t.Fatalf("expected created webhook id, got %d", resp.Data.ID)
	}
	if len(resp.Data.Secret) != 64 {
		t.Fatalf("expected 64-char hex secret, got %q", resp.Data.Secret)
	}
	if _, err := hex.DecodeString(resp.Data.Secret); err != nil {
		t.Fatalf("secret is not valid hex: %v", err)
	}

	persisted, err := store.GetByID(resp.Data.ID)
	if err != nil {
		t.Fatalf("get webhook: %v", err)
	}
	if persisted == nil {
		t.Fatalf("expected webhook to be persisted")
	}
	if persisted.EmpresaID != 7 || persisted.URL != "https://ejemplo.com/hook" {
		t.Fatalf("unexpected persisted webhook: %+v", persisted)
	}
	if persisted.Secret != resp.Data.Secret {
		t.Fatalf("secret mismatch between response and storage")
	}
	if len(persisted.Eventos) != 1 || persisted.Eventos[0] != domain.WebhookEventMessageReceived {
		t.Fatalf("unexpected persisted eventos: %+v", persisted.Eventos)
	}
}

func TestV1WebhooksCreateRejectsInvalidHTTPSURL(t *testing.T) {
	db := newV1WebhooksTestDB(t)
	h := NewV1WebhooksHandler(storage.NewWebhookStore(db), 10)

	cases := []string{
		`{"url":"http://ejemplo.com/hook","eventos":["message.received"]}`,
		`{"url":"https://","eventos":["message.received"]}`,
	}

	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/service/v1/webhooks", strings.NewReader(body))
			req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{ApiKeyID: 33, EmpresaID: 7, TelefonoID: 11}))
			rr := httptest.NewRecorder()

			h.Create(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rr.Code)
			}

			var resp map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if resp["error"] != "INVALID_URL" {
				t.Fatalf("expected INVALID_URL, got %+v", resp)
			}
		})
	}
}

func TestV1WebhooksCreateRejectsInvalidEvent(t *testing.T) {
	db := newV1WebhooksTestDB(t)
	h := NewV1WebhooksHandler(storage.NewWebhookStore(db), 10)

	req := httptest.NewRequest(http.MethodPost, "/api/service/v1/webhooks", strings.NewReader(`{"url":"https://ejemplo.com/hook","eventos":["invalid_event"]}`))
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{ApiKeyID: 33, EmpresaID: 7, TelefonoID: 11}))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] != "INVALID_EVENTOS" {
		t.Fatalf("expected INVALID_EVENTOS, got %+v", resp)
	}
}

func TestV1WebhooksCreateRejectsWhenMaxReached(t *testing.T) {
	db := newV1WebhooksTestDB(t)
	store := storage.NewWebhookStore(db)
	h := NewV1WebhooksHandler(store, 1)
	seedWebhook(t, db, 7, 11, 33, "https://ejemplo.com/existente", `["message.received"]`, true)

	req := httptest.NewRequest(http.MethodPost, "/api/service/v1/webhooks", strings.NewReader(`{"url":"https://ejemplo.com/hook","eventos":["message.received"]}`))
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{ApiKeyID: 33, EmpresaID: 7, TelefonoID: 11}))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] != "MAX_WEBHOOKS_REACHED" {
		t.Fatalf("expected MAX_WEBHOOKS_REACHED, got %+v", resp)
	}
}

func TestV1WebhooksListOmitsSecret(t *testing.T) {
	db := newV1WebhooksTestDB(t)
	store := storage.NewWebhookStore(db)
	h := NewV1WebhooksHandler(store, 10)
	seedWebhook(t, db, 7, 11, 33, "https://ejemplo.com/uno", `["message.received"]`, true)
	seedWebhook(t, db, 7, 11, 33, "https://ejemplo.com/dos", `["session.connected"]`, false)

	req := httptest.NewRequest(http.MethodGet, "/api/service/v1/webhooks", nil)
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{ApiKeyID: 33, EmpresaID: 7, TelefonoID: 11}))
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			Webhooks []map[string]any `json:"webhooks"`
			Total    int              `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok response")
	}
	if resp.Data.Total != 2 || len(resp.Data.Webhooks) != 2 {
		t.Fatalf("unexpected total/list size: %+v", resp.Data)
	}
	for _, webhook := range resp.Data.Webhooks {
		if _, ok := webhook["secret"]; ok {
			t.Fatalf("secret should not be exposed in list response: %+v", webhook)
		}
	}
}

func TestV1WebhooksDeleteRemovesOwnedWebhook(t *testing.T) {
	db := newV1WebhooksTestDB(t)
	store := storage.NewWebhookStore(db)
	h := NewV1WebhooksHandler(store, 10)
	id := seedWebhook(t, db, 7, 11, 33, "https://ejemplo.com/delete", `["message.received"]`, true)
	idStr := "1"
	if id > 0 {
		idStr = strconv.FormatInt(id, 10)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/service/v1/webhooks/"+idStr, nil)
	req.SetPathValue("id", idStr)
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{ApiKeyID: 33, EmpresaID: 7, TelefonoID: 11}))
	rr := httptest.NewRecorder()

	h.Delete(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	deleted, err := store.GetByID(id)
	if err != nil {
		t.Fatalf("get webhook after delete: %v", err)
	}
	if deleted != nil {
		t.Fatalf("expected webhook to be deleted, got %+v", deleted)
	}
}

func TestV1WebhooksDeleteReturns404WhenWebhookDoesNotBelongToEmpresa(t *testing.T) {
	db := newV1WebhooksTestDB(t)
	store := storage.NewWebhookStore(db)
	h := NewV1WebhooksHandler(store, 10)
	id := seedWebhook(t, db, 9, 99, 999, "https://ejemplo.com/ajeno", `["message.received"]`, true)
	idStr := "1"
	if id > 0 {
		idStr = strconv.FormatInt(id, 10)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/service/v1/webhooks/"+idStr, nil)
	req.SetPathValue("id", idStr)
	req = req.WithContext(domain.WithApiKeyClaims(req.Context(), &domain.ApiKeyClaims{ApiKeyID: 33, EmpresaID: 7, TelefonoID: 11}))
	rr := httptest.NewRecorder()

	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func newV1WebhooksTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	_, err = db.Exec(`
CREATE TABLE webhooks_outbound (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    empresa_id INTEGER NOT NULL,
    telefono_id INTEGER NOT NULL,
    api_key_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    eventos TEXT NOT NULL,
    activo BOOLEAN NOT NULL DEFAULT 1,
    failure_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    last_success_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		_ = db.Close()
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedWebhook(t *testing.T, db *sql.DB, empresaID, telefonoID, apiKeyID int64, url, eventosJSON string, activo bool) int64 {
	t.Helper()
	res, err := db.Exec(`
INSERT INTO webhooks_outbound (
    empresa_id, telefono_id, api_key_id, url, secret, eventos, activo, failure_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
`, empresaID, telefonoID, apiKeyID, url, "secret-seed", eventosJSON, activo)
	if err != nil {
		t.Fatalf("insert webhook: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}
