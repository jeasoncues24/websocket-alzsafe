package whatsapp

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
	waEvents "go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

func TestServiceMarkConnectedEnqueuesSessionConnectedWebhook(t *testing.T) {
	db := newWebhookEmitterTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	webhookStore := storage.NewWebhookStore(db)
	manager := NewManager()
	service := NewService(manager, nil, telefonoStore, webhookStore, "")

	seedEmitterTelefono(t, db, 10, 7, "51", "999888777", "51999888777", domain.TelefonoStatusDisconnected)
	seedEmitterWebhook(t, db, 7, 10, 100, `https://ejemplo.com/connected`, `secret-a`, `["session.connected"]`, true)
	seedEmitterWebhook(t, db, 7, 10, 100, `https://ejemplo.com/other-event`, `secret-b`, `["message.received"]`, true)
	seedEmitterWebhook(t, db, 9, 99, 999, `https://ejemplo.com/other-company`, `secret-c`, `["session.connected"]`, true)

	service.markConnected("51999888777")

	items := listEmitterQueueItems(t, db)
	if len(items) != 1 {
		t.Fatalf("expected 1 queue item, got %d", len(items))
	}
	if items[0].EventType != domain.WebhookEventSessionConnected {
		t.Fatalf("unexpected event type: %s", items[0].EventType)
	}
	if items[0].Payload["telefono_id"] != float64(10) {
		t.Fatalf("unexpected telefono_id payload: %+v", items[0].Payload)
	}
	if items[0].Payload["phone"] != "51999888777" {
		t.Fatalf("unexpected phone payload: %+v", items[0].Payload)
	}
}

func TestServiceMarkDisconnectedEnqueuesSessionDisconnectedWebhook(t *testing.T) {
	db := newWebhookEmitterTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	webhookStore := storage.NewWebhookStore(db)
	manager := NewManager()
	service := NewService(manager, nil, telefonoStore, webhookStore, "")

	seedEmitterTelefono(t, db, 11, 7, "51", "999000111", "51999000111", domain.TelefonoStatusActive)
	seedEmitterWebhook(t, db, 7, 11, 110, `https://ejemplo.com/disconnected`, `secret-a`, `["session.disconnected"]`, true)

	service.markDisconnected("51999000111", "logged_out")

	items := listEmitterQueueItems(t, db)
	if len(items) != 1 {
		t.Fatalf("expected 1 queue item, got %d", len(items))
	}
	if items[0].EventType != domain.WebhookEventSessionDisconnected {
		t.Fatalf("unexpected event type: %s", items[0].EventType)
	}
	if items[0].Payload["reason"] != "logged_out" {
		t.Fatalf("unexpected reason payload: %+v", items[0].Payload)
	}
}

func TestServiceHandleWhatsAppEventEnqueuesMessageReceived(t *testing.T) {
	db := newWebhookEmitterTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	webhookStore := storage.NewWebhookStore(db)
	manager := NewManager()
	service := NewService(manager, nil, telefonoStore, webhookStore, "")

	seedEmitterTelefono(t, db, 12, 7, "51", "911122233", "51911122233", domain.TelefonoStatusActive)
	seedEmitterWebhook(t, db, 7, 12, 120, `https://ejemplo.com/message`, `secret-a`, `["message.received"]`, true)
	seedEmitterWebhook(t, db, 7, 12, 120, `https://ejemplo.com/connected`, `secret-b`, `["session.connected"]`, true)
	seedEmitterWebhook(t, db, 9, 99, 999, `https://ejemplo.com/other-company`, `secret-c`, `["message.received"]`, true)

	msg := &waEvents.Message{
		Info: waTypes.MessageInfo{
			MessageSource: waTypes.MessageSource{
				Chat:     waTypes.NewJID("51911122233", waTypes.DefaultUserServer),
				Sender:   waTypes.NewJID("51944455566", waTypes.DefaultUserServer),
				IsFromMe: false,
			},
			ID:        waTypes.MessageID("wamid-123"),
			Timestamp: time.Date(2026, 5, 22, 15, 4, 5, 0, time.UTC),
		},
		Message: &waE2E.Message{Conversation: proto.String("hola webhook")},
	}

	service.handleWhatsAppEvent("51911122233", msg)

	items := listEmitterQueueItems(t, db)
	if len(items) != 1 {
		t.Fatalf("expected 1 queue item, got %d", len(items))
	}
	if items[0].EventType != domain.WebhookEventMessageReceived {
		t.Fatalf("unexpected event type: %s", items[0].EventType)
	}
	if items[0].Payload["telefono_id"] != float64(12) {
		t.Fatalf("unexpected telefono_id payload: %+v", items[0].Payload)
	}
	if items[0].Payload["from"] != "51944455566" {
		t.Fatalf("unexpected from payload: %+v", items[0].Payload)
	}
	if items[0].Payload["message_id"] != "wamid-123" {
		t.Fatalf("unexpected message_id payload: %+v", items[0].Payload)
	}
	if items[0].Payload["content"] != "hola webhook" {
		t.Fatalf("unexpected content payload: %+v", items[0].Payload)
	}
	if items[0].Payload["type"] != "text" {
		t.Fatalf("unexpected type payload: %+v", items[0].Payload)
	}
}

func TestServiceHandleWhatsAppEventEnqueuesStatusUpdateWithReferenceIDWhenMapped(t *testing.T) {
	db := newWebhookEmitterTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	webhookStore := storage.NewWebhookStore(db)
	manager := NewManager()
	service := NewService(manager, nil, telefonoStore, webhookStore, "")

	seedEmitterTelefono(t, db, 13, 7, "51", "955667788", "51955667788", domain.TelefonoStatusActive)
	seedEmitterWebhook(t, db, 7, 13, 130, `https://ejemplo.com/status`, `secret-a`, `["message.status_update"]`, true)
	manager.RegisterOutboundMessageReference("51955667788", "wamid-out-1", "ref-123")

	receipt := &waEvents.Receipt{
		MessageSource: waTypes.MessageSource{
			Chat:     waTypes.NewJID("51999900011", waTypes.DefaultUserServer),
			Sender:   waTypes.NewJID("51999900011", waTypes.DefaultUserServer),
			IsFromMe: false,
		},
		MessageIDs: []waTypes.MessageID{waTypes.MessageID("wamid-out-1")},
		Timestamp:  time.Date(2026, 5, 22, 16, 0, 0, 0, time.UTC),
		Type:       waTypes.ReceiptTypeRead,
	}

	service.handleWhatsAppEvent("51955667788", receipt)

	items := listEmitterQueueItems(t, db)
	if len(items) != 1 {
		t.Fatalf("expected 1 queue item, got %d", len(items))
	}
	if items[0].EventType != domain.WebhookEventMessageStatus {
		t.Fatalf("unexpected event type: %s", items[0].EventType)
	}
	if items[0].Payload["message_id"] != "wamid-out-1" {
		t.Fatalf("unexpected message_id payload: %+v", items[0].Payload)
	}
	if items[0].Payload["reference_id"] != "ref-123" {
		t.Fatalf("unexpected reference_id payload: %+v", items[0].Payload)
	}
	if items[0].Payload["status"] != "read" {
		t.Fatalf("unexpected status payload: %+v", items[0].Payload)
	}
}

func TestWebhookEmitterNoActiveHooksDoesNotFail(t *testing.T) {
	db := newWebhookEmitterTestDB(t)
	telefonoStore := storage.NewTelefonoStore(db)
	webhookStore := storage.NewWebhookStore(db)
	emitter := NewWebhookEmitter(webhookStore, telefonoStore, WebhookEmitterConfig{})

	seedEmitterTelefono(t, db, 14, 7, "51", "900111222", "51900111222", domain.TelefonoStatusActive)
	seedEmitterWebhook(t, db, 7, 14, 140, `https://ejemplo.com/inactive`, `secret-a`, `["session.connected"]`, false)

	if err := emitter.EmitSessionConnectedByAccount("51900111222"); err != nil {
		t.Fatalf("expected no error when there are no active webhooks, got %v", err)
	}
	items := listEmitterQueueItems(t, db)
	if len(items) != 0 {
		t.Fatalf("expected 0 queue items, got %d", len(items))
	}
}

type emitterQueueItem struct {
	WebhookID int64
	EventType domain.WebhookEvent
	Payload   map[string]any
}

func newWebhookEmitterTestDB(t *testing.T) *sql.DB {
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
    numero_completo TEXT NOT NULL,
    status TEXT NOT NULL,
    session_data BLOB,
    qr_string TEXT,
    last_connected TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
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

func seedEmitterTelefono(t *testing.T, db *sql.DB, id, empresaID int64, codigoPais, numero, numeroCompleto string, status domain.TelefonoStatus) {
	t.Helper()
	_, err := db.Exec(`
INSERT INTO telefonos (
    id, empresa_id, codigo_pais, numero, numero_completo, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
`, id, empresaID, codigoPais, numero, numeroCompleto, string(status))
	if err != nil {
		t.Fatalf("insert telefono: %v", err)
	}
}

func seedEmitterWebhook(t *testing.T, db *sql.DB, empresaID, telefonoID, apiKeyID int64, url, secret, eventos string, activo bool) int64 {
	t.Helper()
	res, err := db.Exec(`
INSERT INTO webhooks_outbound (
    empresa_id, telefono_id, api_key_id, url, secret, eventos, activo, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
`, empresaID, telefonoID, apiKeyID, url, secret, eventos, activo)
	if err != nil {
		t.Fatalf("insert webhook: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func listEmitterQueueItems(t *testing.T, db *sql.DB) []emitterQueueItem {
	t.Helper()
	rows, err := db.Query(`SELECT webhook_id, payload FROM webhooks_outbound_queue ORDER BY id ASC`)
	if err != nil {
		t.Fatalf("query queue items: %v", err)
	}
	defer rows.Close()

	var items []emitterQueueItem
	for rows.Next() {
		var item emitterQueueItem
		var rawPayload string
		if err := rows.Scan(&item.WebhookID, &rawPayload); err != nil {
			t.Fatalf("scan queue item: %v", err)
		}

		var envelope struct {
			EventType domain.WebhookEvent `json:"event_type"`
			Data      map[string]any      `json:"data"`
		}
		if err := json.Unmarshal([]byte(rawPayload), &envelope); err != nil {
			t.Fatalf("decode queue payload: %v", err)
		}
		item.EventType = envelope.EventType
		item.Payload = envelope.Data
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate queue items: %v", err)
	}
	return items
}
