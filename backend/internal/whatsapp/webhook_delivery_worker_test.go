package whatsapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

func TestWebhookDeliveryWorkerProcessDueItemsMarksDoneAndResetsWebhook(t *testing.T) {
	db := newWebhookDeliveryWorkerTestDB(t)
	store := storage.NewWebhookStore(db)

	var receivedSignature string
	var receivedEvent string
	var receivedDelivery string
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("X-Wsapi-Signature")
		receivedEvent = r.Header.Get("X-Wsapi-Event")
		receivedDelivery = r.Header.Get("X-Wsapi-Delivery")
		body, _ := io.ReadAll(r.Body)
		receivedBody = body
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	webhookID := seedWorkerWebhook(t, db, 7, server.URL, "supersecret", 3, "boom", true)
	queueID := seedWorkerQueueItem(t, db, webhookID, buildEnvelope(t, domain.WebhookEventMessageReceived, map[string]any{
		"telefono_id": 10,
		"from":        "51999999999",
	}), domain.WebhookQueuePending, 0, time.Now().Add(-time.Minute), "")

	worker := NewWebhookDeliveryWorker(store, server.Client(), WebhookDeliveryWorkerConfig{
		PollInterval:        5 * time.Second,
		RequestTimeout:      time.Second,
		BatchSize:           10,
		MaxAttempts:         6,
		DeactivateThreshold: 20,
		RetrySchedule:       []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 6 * time.Hour},
	})

	if err := worker.processDueItems(context.Background()); err != nil {
		t.Fatalf("process due items: %v", err)
	}

	item := getWorkerQueueItem(t, db, queueID)
	if item.Estado != domain.WebhookQueueDone {
		t.Fatalf("expected queue item done, got %+v", item)
	}
	if item.Intentos != 1 {
		t.Fatalf("expected 1 intento, got %d", item.Intentos)
	}

	webhook := getWorkerWebhook(t, db, webhookID)
	if webhook.FailureCount != 0 {
		t.Fatalf("expected failure_count reset, got %d", webhook.FailureCount)
	}
	if webhook.LastError != nil {
		t.Fatalf("expected last_error cleared, got %v", *webhook.LastError)
	}
	if webhook.LastSuccessAt == nil {
		t.Fatalf("expected last_success_at to be set")
	}

	expectedBody := `{"from":"51999999999","telefono_id":10}`
	if string(receivedBody) != expectedBody {
		t.Fatalf("expected body %s, got %s", expectedBody, string(receivedBody))
	}
	if receivedEvent != string(domain.WebhookEventMessageReceived) {
		t.Fatalf("unexpected X-Wsapi-Event: %q", receivedEvent)
	}
	if receivedDelivery != "1" {
		t.Fatalf("unexpected X-Wsapi-Delivery: %q", receivedDelivery)
	}

	mac := hmac.New(sha256.New, []byte("supersecret"))
	mac.Write([]byte(expectedBody))
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if receivedSignature != expectedSignature {
		t.Fatalf("unexpected signature: got %q want %q", receivedSignature, expectedSignature)
	}
}

func TestWebhookDeliveryWorkerProcessDueItemsRequeuesOnServerError(t *testing.T) {
	db := newWebhookDeliveryWorkerTestDB(t)
	store := storage.NewWebhookStore(db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	webhookID := seedWorkerWebhook(t, db, 7, server.URL, "supersecret", 0, "", true)
	queueID := seedWorkerQueueItem(t, db, webhookID, buildEnvelope(t, domain.WebhookEventSessionConnected, map[string]any{"ok": true}), domain.WebhookQueuePending, 0, time.Now().Add(-time.Minute), "")

	worker := NewWebhookDeliveryWorker(store, server.Client(), WebhookDeliveryWorkerConfig{
		PollInterval:        5 * time.Second,
		RequestTimeout:      time.Second,
		BatchSize:           10,
		MaxAttempts:         6,
		DeactivateThreshold: 20,
		RetrySchedule:       []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 6 * time.Hour},
	})
	before := time.Now()

	if err := worker.processDueItems(context.Background()); err != nil {
		t.Fatalf("process due items: %v", err)
	}

	item := getWorkerQueueItem(t, db, queueID)
	if item.Estado != domain.WebhookQueuePending {
		t.Fatalf("expected queue item pending, got %+v", item)
	}
	if item.Intentos != 1 {
		t.Fatalf("expected intentos=1, got %d", item.Intentos)
	}
	if item.LastError == nil || !strings.Contains(*item.LastError, "503") {
		t.Fatalf("expected last_error to mention 503, got %+v", item.LastError)
	}
	if diff := item.ProximoIntentoAt.Sub(before); diff < 55*time.Second || diff > 65*time.Second {
		t.Fatalf("expected next retry around 1 minute, got %s", diff)
	}

	webhook := getWorkerWebhook(t, db, webhookID)
	if webhook.FailureCount != 0 {
		t.Fatalf("expected webhook failure_count unchanged on retryable error, got %d", webhook.FailureCount)
	}
}

func TestWebhookDeliveryWorkerProcessDueItemsRequeuesOnTimeout(t *testing.T) {
	db := newWebhookDeliveryWorkerTestDB(t)
	store := storage.NewWebhookStore(db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhookID := seedWorkerWebhook(t, db, 7, server.URL, "supersecret", 0, "", true)
	queueID := seedWorkerQueueItem(t, db, webhookID, buildEnvelope(t, domain.WebhookEventSessionDisconnected, map[string]any{"reason": "timeout"}), domain.WebhookQueuePending, 0, time.Now().Add(-time.Minute), "")

	worker := NewWebhookDeliveryWorker(store, &http.Client{Timeout: 50 * time.Millisecond}, WebhookDeliveryWorkerConfig{
		PollInterval:        5 * time.Second,
		RequestTimeout:      50 * time.Millisecond,
		BatchSize:           10,
		MaxAttempts:         6,
		DeactivateThreshold: 20,
		RetrySchedule:       []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 6 * time.Hour},
	})

	if err := worker.processDueItems(context.Background()); err != nil {
		t.Fatalf("process due items: %v", err)
	}

	item := getWorkerQueueItem(t, db, queueID)
	if item.Estado != domain.WebhookQueuePending {
		t.Fatalf("expected queue item pending, got %+v", item)
	}
	if item.Intentos != 1 {
		t.Fatalf("expected intentos=1, got %d", item.Intentos)
	}
	if item.LastError == nil || !strings.Contains(strings.ToLower(*item.LastError), "timeout") {
		t.Fatalf("expected timeout in last_error, got %+v", item.LastError)
	}
}

func TestWebhookDeliveryWorkerProcessDueItemsMarksTerminalFailureAndDeactivatesWebhook(t *testing.T) {
	db := newWebhookDeliveryWorkerTestDB(t)
	store := storage.NewWebhookStore(db)

	webhookID := seedWorkerWebhook(t, db, 7, "https://ejemplo.com/hook", "supersecret", 19, "previo", true)
	queueID := seedWorkerQueueItem(t, db, webhookID, []byte(`{"data":{"telefono_id":10}}`), domain.WebhookQueuePending, 0, time.Now().Add(-time.Minute), "")

	worker := NewWebhookDeliveryWorker(store, &http.Client{Timeout: time.Second}, WebhookDeliveryWorkerConfig{
		PollInterval:        5 * time.Second,
		RequestTimeout:      time.Second,
		BatchSize:           10,
		MaxAttempts:         6,
		DeactivateThreshold: 20,
		RetrySchedule:       []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 6 * time.Hour},
	})

	if err := worker.processDueItems(context.Background()); err != nil {
		t.Fatalf("process due items: %v", err)
	}

	item := getWorkerQueueItem(t, db, queueID)
	if item.Estado != domain.WebhookQueueFailed {
		t.Fatalf("expected queue item failed, got %+v", item)
	}
	if item.Intentos != 1 {
		t.Fatalf("expected intentos=1, got %d", item.Intentos)
	}

	webhook := getWorkerWebhook(t, db, webhookID)
	if webhook.FailureCount != 20 {
		t.Fatalf("expected failure_count=20, got %d", webhook.FailureCount)
	}
	if webhook.Activo {
		t.Fatalf("expected webhook to be deactivated")
	}
	if webhook.LastError == nil || !strings.Contains(*webhook.LastError, "event_type") {
		t.Fatalf("expected webhook last_error to mention event_type, got %+v", webhook.LastError)
	}
}

func TestWebhookDeliveryWorkerProcessQueueItemSkipsIfAlreadyClaimed(t *testing.T) {
	db := newWebhookDeliveryWorkerTestDB(t)
	store := storage.NewWebhookStore(db)

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhookID := seedWorkerWebhook(t, db, 7, server.URL, "supersecret", 0, "", true)
	queueID := seedWorkerQueueItem(t, db, webhookID, buildEnvelope(t, domain.WebhookEventSessionConnected, map[string]any{"ok": true}), domain.WebhookQueueSending, 0, time.Now().Add(-time.Minute), "")

	worker := NewWebhookDeliveryWorker(store, server.Client(), WebhookDeliveryWorkerConfig{
		PollInterval:        5 * time.Second,
		RequestTimeout:      time.Second,
		BatchSize:           10,
		MaxAttempts:         6,
		DeactivateThreshold: 20,
		RetrySchedule:       []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 6 * time.Hour},
	})

	item := getWorkerQueueItem(t, db, queueID)
	if err := worker.processQueueItem(context.Background(), item); err != nil {
		t.Fatalf("process queue item: %v", err)
	}
	if hits.Load() != 0 {
		t.Fatalf("expected no HTTP deliveries when item already claimed")
	}
}

func TestWebhookDeliveryWorkerRunStopsOnContextCancel(t *testing.T) {
	db := newWebhookDeliveryWorkerTestDB(t)
	store := storage.NewWebhookStore(db)
	worker := NewWebhookDeliveryWorker(store, &http.Client{Timeout: time.Second}, WebhookDeliveryWorkerConfig{
		PollInterval:        5 * time.Second,
		RequestTimeout:      time.Second,
		BatchSize:           10,
		MaxAttempts:         6,
		DeactivateThreshold: 20,
		RetrySchedule:       []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 6 * time.Hour},
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}
}

func newWebhookDeliveryWorkerTestDB(t *testing.T) *sql.DB {
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
    estado TEXT NOT NULL,
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

func seedWorkerWebhook(t *testing.T, db *sql.DB, empresaID int64, url, secret string, failureCount int, lastError string, activo bool) int64 {
	t.Helper()
	var lastErrorValue any
	if lastError != "" {
		lastErrorValue = lastError
	}
	res, err := db.Exec(`
INSERT INTO webhooks_outbound (
    empresa_id, telefono_id, api_key_id, url, secret, eventos, activo, failure_count, last_error, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
`, empresaID, empresaID*10, empresaID*100, url, secret, `["message.received"]`, activo, failureCount, lastErrorValue)
	if err != nil {
		t.Fatalf("insert webhook: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func seedWorkerQueueItem(t *testing.T, db *sql.DB, webhookID int64, payload []byte, estado domain.WebhookQueueEstado, intentos int, nextRetryAt time.Time, lastError string) int64 {
	t.Helper()
	var lastErrorValue any
	if lastError != "" {
		lastErrorValue = lastError
	}
	res, err := db.Exec(`
INSERT INTO webhooks_outbound_queue (
    webhook_id, payload, intentos, proximo_intento_at, estado, last_error, created_at
) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
`, webhookID, string(payload), intentos, nextRetryAt, estado, lastErrorValue)
	if err != nil {
		t.Fatalf("insert queue item: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func getWorkerQueueItem(t *testing.T, db *sql.DB, id int64) domain.WebhookQueueItem {
	t.Helper()
	var item domain.WebhookQueueItem
	var payload string
	var lastError sql.NullString
	err := db.QueryRow(`SELECT id, webhook_id, payload, intentos, proximo_intento_at, estado, last_error, created_at FROM webhooks_outbound_queue WHERE id = ?`, id).Scan(
		&item.ID,
		&item.WebhookID,
		&payload,
		&item.Intentos,
		&item.ProximoIntentoAt,
		&item.Estado,
		&lastError,
		&item.CreatedAt,
	)
	item.Payload = json.RawMessage(payload)
	if err != nil {
		t.Fatalf("get queue item: %v", err)
	}
	if lastError.Valid {
		item.LastError = &lastError.String
	}
	return item
}

func getWorkerWebhook(t *testing.T, db *sql.DB, id int64) domain.Webhook {
	t.Helper()
	var webhook domain.Webhook
	var eventos string
	var lastError sql.NullString
	var lastSuccessAt sql.NullTime
	err := db.QueryRow(`SELECT id, empresa_id, url, secret, eventos, activo, failure_count, last_error, last_success_at, created_at, updated_at FROM webhooks_outbound WHERE id = ?`, id).Scan(
		&webhook.ID,
		&webhook.EmpresaID,
		&webhook.URL,
		&webhook.Secret,
		&eventos,
		&webhook.Activo,
		&webhook.FailureCount,
		&lastError,
		&lastSuccessAt,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("get webhook: %v", err)
	}
	_ = json.Unmarshal([]byte(eventos), &webhook.Eventos)
	if lastError.Valid {
		webhook.LastError = &lastError.String
	}
	if lastSuccessAt.Valid {
		webhook.LastSuccessAt = &lastSuccessAt.Time
	}
	return webhook
}

func buildEnvelope(t *testing.T, eventType domain.WebhookEvent, data map[string]any) []byte {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"event_type": eventType,
		"data":       data,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return payload
}
