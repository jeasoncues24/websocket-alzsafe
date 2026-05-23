package storage

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"wsapi/internal/domain"
)

func newWebhookStoreTestDB(t *testing.T) *sql.DB {
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
);
CREATE TABLE webhooks_outbound_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_id INTEGER NOT NULL,
    payload TEXT NOT NULL,
    intentos INTEGER NOT NULL DEFAULT 0,
    proximo_intento_at TIMESTAMP NOT NULL,
    estado TEXT NOT NULL DEFAULT 'pending',
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		_ = db.Close()
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestWebhookStoreCreateAndListByApiKey(t *testing.T) {
	db := newWebhookStoreTestDB(t)
	store := NewWebhookStore(db)

	wh := &domain.Webhook{
		EmpresaID:  1,
		TelefonoID: 10,
		ApiKeyID:   100,
		URL:        "https://ejemplo.com/hook",
		Secret:     "s3cr3t",
		Eventos:    []domain.WebhookEvent{domain.WebhookEventMessageReceived},
		Activo:     true,
	}

	if err := store.Create(wh); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if wh.ID == 0 {
		t.Fatal("expected non-zero ID after Create")
	}

	// otro api_key no debe aparecer
	other := &domain.Webhook{
		EmpresaID:  1,
		TelefonoID: 10,
		ApiKeyID:   999,
		URL:        "https://otro.com/hook",
		Secret:     "oth3r",
		Eventos:    []domain.WebhookEvent{domain.WebhookEventSessionConnected},
		Activo:     true,
	}
	if err := store.Create(other); err != nil {
		t.Fatalf("Create other: %v", err)
	}

	list, err := store.ListByApiKey(100)
	if err != nil {
		t.Fatalf("ListByApiKey: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 webhook for api_key 100, got %d", len(list))
	}
	if list[0].URL != "https://ejemplo.com/hook" {
		t.Fatalf("unexpected URL: %s", list[0].URL)
	}
	if len(list[0].Eventos) != 1 || list[0].Eventos[0] != domain.WebhookEventMessageReceived {
		t.Fatalf("unexpected Eventos: %v", list[0].Eventos)
	}
}

func TestWebhookStoreListByEmpresa(t *testing.T) {
	db := newWebhookStoreTestDB(t)
	store := NewWebhookStore(db)

	for i := 0; i < 3; i++ {
		w := &domain.Webhook{
			EmpresaID:  5,
			TelefonoID: int64(20 + i),
			ApiKeyID:   int64(200 + i),
			URL:        "https://empresa.com/hook",
			Secret:     "sec",
			Eventos:    []domain.WebhookEvent{domain.WebhookEventSessionConnected},
			Activo:     true,
		}
		if err := store.Create(w); err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
	}
	// distinta empresa
	w2 := &domain.Webhook{EmpresaID: 6, TelefonoID: 30, ApiKeyID: 300, URL: "https://otro.com", Secret: "s", Eventos: []domain.WebhookEvent{domain.WebhookEventMessageReceived}, Activo: true}
	_ = store.Create(w2)

	list, err := store.ListByEmpresa(5)
	if err != nil {
		t.Fatalf("ListByEmpresa: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 webhooks, got %d", len(list))
	}
}

func TestWebhookStoreDelete(t *testing.T) {
	db := newWebhookStoreTestDB(t)
	store := NewWebhookStore(db)

	wh := &domain.Webhook{
		EmpresaID:  1,
		TelefonoID: 10,
		ApiKeyID:   100,
		URL:        "https://ejemplo.com/hook",
		Secret:     "sec",
		Eventos:    []domain.WebhookEvent{domain.WebhookEventMessageReceived},
		Activo:     true,
	}
	if err := store.Create(wh); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := store.Delete(wh.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	list, err := store.ListByApiKey(100)
	if err != nil {
		t.Fatalf("ListByApiKey after delete: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 webhooks after delete, got %d", len(list))
	}
}

func TestWebhookStoreEnqueueAndPollPending(t *testing.T) {
	db := newWebhookStoreTestDB(t)
	store := NewWebhookStore(db)

	wh := &domain.Webhook{
		EmpresaID:  1,
		TelefonoID: 10,
		ApiKeyID:   100,
		URL:        "https://ejemplo.com/hook",
		Secret:     "sec",
		Eventos:    []domain.WebhookEvent{domain.WebhookEventMessageReceived},
		Activo:     true,
	}
	if err := store.Create(wh); err != nil {
		t.Fatalf("Create: %v", err)
	}

	payload, _ := json.Marshal(map[string]any{"event_type": "message.received", "data": map[string]any{"from": "51999"}})
	item := &domain.WebhookQueueItem{
		WebhookID:        wh.ID,
		Payload:          payload,
		ProximoIntentoAt: time.Now().Add(-time.Second), // ya vencido
	}
	if err := store.EnqueueEvent(item); err != nil {
		t.Fatalf("EnqueueEvent: %v", err)
	}
	if item.ID == 0 {
		t.Fatal("expected non-zero queue ID")
	}

	pending, err := store.PollPending(10)
	if err != nil {
		t.Fatalf("PollPending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending item, got %d", len(pending))
	}
	if pending[0].WebhookID != wh.ID {
		t.Fatalf("unexpected webhook_id: %d", pending[0].WebhookID)
	}
}

func TestWebhookStoreMarkSendingPreventsDoubleProcess(t *testing.T) {
	db := newWebhookStoreTestDB(t)
	store := NewWebhookStore(db)

	wh := &domain.Webhook{EmpresaID: 1, TelefonoID: 10, ApiKeyID: 100, URL: "https://ejemplo.com", Secret: "s", Eventos: []domain.WebhookEvent{domain.WebhookEventMessageReceived}, Activo: true}
	_ = store.Create(wh)

	payload, _ := json.Marshal(map[string]any{"event_type": "message.received", "data": map[string]any{}})
	item := &domain.WebhookQueueItem{WebhookID: wh.ID, Payload: payload, ProximoIntentoAt: time.Now().Add(-time.Second)}
	_ = store.EnqueueEvent(item)

	if err := store.MarkSending(item.ID); err != nil {
		t.Fatalf("MarkSending first time: %v", err)
	}
	if err := store.MarkSending(item.ID); err != ErrWebhookQueueNotPending {
		t.Fatalf("expected ErrWebhookQueueNotPending on second MarkSending, got: %v", err)
	}
}
