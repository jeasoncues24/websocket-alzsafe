package domain

import (
	"encoding/json"
	"time"
)

type WebhookEvent string

const (
	WebhookEventMessageReceived     WebhookEvent = "message.received"
	WebhookEventMessageStatus       WebhookEvent = "message.status_update"
	WebhookEventSessionConnected    WebhookEvent = "session.connected"
	WebhookEventSessionDisconnected WebhookEvent = "session.disconnected"
)

type WebhookQueueEstado string

const (
	WebhookQueuePending WebhookQueueEstado = "pending"
	WebhookQueueSending WebhookQueueEstado = "sending"
	WebhookQueueDone    WebhookQueueEstado = "done"
	WebhookQueueFailed  WebhookQueueEstado = "failed"
)

type Webhook struct {
	ID            int64          `json:"id"`
	EmpresaID     int64          `json:"empresa_id"`
	TelefonoID    int64          `json:"telefono_id"`
	ApiKeyID      int64          `json:"api_key_id"`
	URL           string         `json:"url"`
	Secret        string         `json:"-"`
	Eventos       []WebhookEvent `json:"eventos"`
	Activo        bool           `json:"activo"`
	FailureCount  int            `json:"failure_count"`
	LastError     *string        `json:"last_error,omitempty"`
	LastSuccessAt *time.Time     `json:"last_success_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type WebhookQueueItem struct {
	ID               int64              `json:"id"`
	WebhookID        int64              `json:"webhook_id"`
	Payload          json.RawMessage    `json:"payload"`
	Intentos         int                `json:"intentos"`
	ProximoIntentoAt time.Time          `json:"proximo_intento_at"`
	Estado           WebhookQueueEstado `json:"estado"`
	LastError        *string            `json:"last_error,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
}

type WebhookDeliveryEnvelope struct {
	EventType WebhookEvent    `json:"event_type"`
	Data      json.RawMessage `json:"data"`
}
