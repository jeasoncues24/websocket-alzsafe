package domain

import (
	"time"

	"github.com/google/uuid"
)

// MessageState represents the lifecycle state of a message
type MessageState string

const (
	MessageStatePending   MessageState = "pending"
	MessageStateSent      MessageState = "sent"
	MessageStateDelivered MessageState = "delivered"
	MessageStateFailed    MessageState = "failed"
	MessageStateRejected  MessageState = "rejected"
)

// Message represents a direct message to be sent.
// Incluye tanto los campos del dominio de negocio como los campos de persistencia en DB.
type Message struct {
	// [QUÉ] ID es la PK autoincremental de la DB, no se expone en JSON.
	// [POR QUÉ] El cliente no necesita el ID interno; usa ReferenceID para trazabilidad.
	ID int64 `json:"-"`

	ReferenceID string           `json:"reference_id"`
	EmpresaID   int64            `json:"empresa_id"`
	TelefonoID  int64            `json:"telefono_id"`
	Destino     string           `json:"destino"`
	Contenido   string           `json:"contenido"`
	TiempoEnvio time.Time        `json:"tiempo_envio"`
	Estado      MessageState     `json:"estado"`
	Adjuntos    []AttachmentInfo `json:"adjuntos,omitempty"`

	// [QUÉ] Campos de trazabilidad del ciclo de vida del envío.
	// [POR QUÉ] Permiten saber exactamente cuándo cada etapa ocurrió (auditoría y reintentos).
	// Son punteros (*time.Time) porque son opcionales: NULL en DB cuando la etapa no ocurrió aún.
	TimestampSent      *time.Time `json:"timestamp_sent,omitempty"`
	TimestampConfirmed *time.Time `json:"timestamp_confirmed,omitempty"`

	// [QUÉ] Razón de fallo para mensajes en estado 'failed' o 'rejected'.
	// [POR QUÉ] Permite al operador saber exactamente qué falló (número inválido, ban, etc.)
	ErrorReason string `json:"error_reason,omitempty"`

	// [QUÉ] Campos para control de reintentos.
	// [POR QUÉ] Permite saber cuántos intentos se han hecho y cuándo fue el último.
	RetryCount     int        `json:"retry_count,omitempty"`
	LastAttemptAt  *time.Time `json:"last_attempt_at,omitempty"`
}

// MessageRequest represents the HTTP POST request payload for direct messages
type MessageRequest struct {
	TelefonoID int64               `json:"telefono_id"`
	Destino    string              `json:"destino"`
	Mensaje    string              `json:"mensaje"`
	Adjuntos   []AttachmentPayload `json:"adjuntos,omitempty"`
}

// MessageResponse represents the HTTP response for message creation
type MessageResponse struct {
	OK            bool   `json:"ok"`
	Message       string `json:"message"`
	ReferenceID   string `json:"reference_id,omitempty"`
	EmpresaID     int64  `json:"empresa_id,omitempty"`
	EmpresaNombre string `json:"empresa_nombre,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	Error         string `json:"error,omitempty"`
	Details       string `json:"details,omitempty"`
}

// NewMessage creates a new Message instance with default values
func NewMessage(empresaID, telefonoID int64, destino, contenido string) *Message {
	return &Message{
		ReferenceID: uuid.New().String(),
		EmpresaID:   empresaID,
		TelefonoID:  telefonoID,
		Destino:     destino,
		Contenido:   contenido,
		TiempoEnvio: time.Now(),
		Estado:      MessageStatePending,
	}
}

// ValidationError represents a validation failure
type ValidationError struct {
	Code    string
	Message string
}

// ErrorCodeInvalidPhoneFormat indicates phone number validation failure
const ErrorCodeInvalidPhoneFormat = "INVALID_PHONE_FORMAT"

// ErrorCodeEmptyMessage indicates message content is empty
const ErrorCodeEmptyMessage = "EMPTY_MESSAGE"

// ErrorCodeSessionNotActive indicates empresa session is not active
const ErrorCodeSessionNotActive = "SESSION_NOT_ACTIVE_FOR_EMPRESA"

// ErrorCodeInvalidJSON indicates malformed JSON
const ErrorCodeInvalidJSON = "INVALID_JSON"

// ErrorCodeMissingField indicates required field is missing
const ErrorCodeMissingField = "MISSING_FIELD"

// ErrorCodeValidation indicates a general validation failure (e.g. wrong type or empty list)
const ErrorCodeValidation = "VALIDATION_ERROR"

// MessagesListResponse represents the HTTP response for GET /messages.
// [QUÉ] Envuelve la lista de mensajes con metadatos de paginación.
// [POR QUÉ] Estandarizar el envelope de respuesta facilita al frontend saber cuántas páginas hay.
type MessagesListResponse struct {
	OK            bool      `json:"ok"`
	Messages      []Message `json:"messages"`
	Total         int       `json:"total"`
	Page          int       `json:"page"`
	Limit         int       `json:"limit"`
	TotalPages    int       `json:"total_pages"`
	EmpresaID     int64     `json:"empresa_id,omitempty"`
	EmpresaNombre string    `json:"empresa_nombre,omitempty"`
	Error         string    `json:"error,omitempty"`
	Details       string    `json:"details,omitempty"`
}
