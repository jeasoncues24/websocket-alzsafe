package whatsapp

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/rs/zerolog"
	waTypes "go.mau.fi/whatsmeow/types"
	waEvents "go.mau.fi/whatsmeow/types/events"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type WebhookEmitterConfig struct {
	NowFunc func() time.Time
	Logger  *zerolog.Logger
}

type WebhookEmitter struct {
	webhookStore  *storage.WebhookStore
	telefonoStore *storage.TelefonoStore
	logger        zerolog.Logger
	nowFunc       func() time.Time
}

type webhookMessageReceivedPayload struct {
	TelefonoID int64  `json:"telefono_id"`
	From       string `json:"from"`
	MessageID  string `json:"message_id"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	Timestamp  string `json:"timestamp"`
}

type webhookSessionConnectedPayload struct {
	TelefonoID int64  `json:"telefono_id"`
	Phone      string `json:"phone"`
	Timestamp  string `json:"timestamp"`
}

type webhookSessionDisconnectedPayload struct {
	TelefonoID int64  `json:"telefono_id"`
	Phone      string `json:"phone"`
	Reason     string `json:"reason"`
	Timestamp  string `json:"timestamp"`
}

type webhookMessageStatusPayload struct {
	TelefonoID  int64  `json:"telefono_id"`
	MessageID   string `json:"message_id"`
	ReferenceID string `json:"reference_id,omitempty"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
}

func NewWebhookEmitter(webhookStore *storage.WebhookStore, telefonoStore *storage.TelefonoStore, cfg WebhookEmitterConfig) *WebhookEmitter {
	if cfg.NowFunc == nil {
		cfg.NowFunc = time.Now
	}

	logger := config.GetLogger().With().Str("component", "webhook_emitter").Logger()
	if cfg.Logger != nil {
		logger = cfg.Logger.With().Str("component", "webhook_emitter").Logger()
	}

	return &WebhookEmitter{
		webhookStore:  webhookStore,
		telefonoStore: telefonoStore,
		logger:        logger,
		nowFunc:       cfg.NowFunc,
	}
}

func (e *WebhookEmitter) EmitSessionConnectedByAccount(accountID string) error {
	phone, err := e.lookupTelefono(accountID)
	if err != nil || phone == nil {
		return err
	}

	payload := webhookSessionConnectedPayload{
		TelefonoID: phone.ID,
		Phone:      phone.NumeroCompleto,
		Timestamp:  e.nowFunc().UTC().Format(time.RFC3339),
	}
	return e.emitWebhookEvent(phone.EmpresaID, phone.ID, domain.WebhookEventSessionConnected, payload)
}



func (e *WebhookEmitter) EmitSessionDisconnectedByAccount(accountID, reason string) error {
	phone, err := e.lookupTelefono(accountID)
	if err != nil || phone == nil {
		return err
	}
	if strings.TrimSpace(reason) == "" {
		reason = "disconnect"
	}

	payload := webhookSessionDisconnectedPayload{
		TelefonoID: phone.ID,
		Phone:      phone.NumeroCompleto,
		Reason:     reason,
		Timestamp:  e.nowFunc().UTC().Format(time.RFC3339),
	}
	return e.emitWebhookEvent(phone.EmpresaID, phone.ID, domain.WebhookEventSessionDisconnected, payload)
}

func (e *WebhookEmitter) EmitMessageReceivedByAccount(accountID string, evt *waEvents.Message) error {
	if evt == nil || evt.Info.IsFromMe {
		return nil
	}

	phone, err := e.lookupTelefono(accountID)
	if err != nil || phone == nil {
		return err
	}

	content, messageType := webhookMessageSummary(evt)
	from := strings.TrimSpace(evt.Info.Sender.User)
	if from == "" {
		from = strings.TrimSpace(evt.Info.Chat.User)
	}

	payload := webhookMessageReceivedPayload{
		TelefonoID: phone.ID,
		From:       from,
		MessageID:  string(evt.Info.ID),
		Content:    content,
		Type:       messageType,
		Timestamp:  evt.Info.Timestamp.UTC().Format(time.RFC3339),
	}
	return e.emitWebhookEvent(phone.EmpresaID, phone.ID, domain.WebhookEventMessageReceived, payload)
}

func (e *WebhookEmitter) EmitMessageStatusUpdateByAccount(accountID, messageID, referenceID string, receiptType waTypes.ReceiptType, timestamp time.Time) error {
	phone, err := e.lookupTelefono(accountID)
	if err != nil || phone == nil {
		return err
	}

	payload := webhookMessageStatusPayload{
		TelefonoID: phone.ID,
		MessageID:  messageID,
		Status:     webhookReceiptStatus(receiptType),
		Timestamp:  timestamp.UTC().Format(time.RFC3339),
	}
	if referenceID != "" {
		payload.ReferenceID = referenceID
	}
	return e.emitWebhookEvent(phone.EmpresaID, phone.ID, domain.WebhookEventMessageStatus, payload)
}

func (e *WebhookEmitter) lookupTelefono(accountID string) (*domain.Telefono, error) {
	if e == nil || e.telefonoStore == nil {
		return nil, nil
	}
	return e.telefonoStore.GetByNumeroCompletoNormalized(accountID)
}

func (e *WebhookEmitter) emitWebhookEvent(empresaID, telefonoID int64, eventType domain.WebhookEvent, payload any) error {
	if e == nil || e.webhookStore == nil {
		return nil
	}

	webhooks, err := e.webhookStore.ListActiveByTelefonoAndEvent(telefonoID, eventType)
	if err != nil {
		return err
	}
	if len(webhooks) == 0 {
		e.logger.Debug().
			Int64("empresa_id", empresaID).
			Int64("telefono_id", telefonoID).
			Str("event_type", string(eventType)).
			Int("webhooks_match", 0).
			Str("resultado", "no_subscribers").
			Msg("emision de webhook omitida")
		return nil
	}

	dataJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	body := domain.WebhookDeliveryEnvelope{
		EventType: eventType,
		Data:      dataJSON,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return err
	}

	var firstErr error
	for _, webhook := range webhooks {
		item := &domain.WebhookQueueItem{
			WebhookID:        webhook.ID,
			Payload:          bodyJSON,
			ProximoIntentoAt: e.nowFunc(),
		}
		if err := e.webhookStore.EnqueueEvent(item); err != nil {
			e.logger.Error().Err(err).Int64("empresa_id", empresaID).Int64("webhook_id", webhook.ID).Str("event_type", string(eventType)).Msg("error encolando evento para webhook")
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if firstErr != nil {
		return firstErr
	}

	e.logger.Info().
		Int64("empresa_id", empresaID).
		Int64("telefono_id", telefonoID).
		Str("event_type", string(eventType)).
		Int("webhooks_match", len(webhooks)).
		Str("resultado", "queued").
		Msg("evento webhook encolado")
	return nil
}

func webhookMessageSummary(evt *waEvents.Message) (string, string) {
	if evt == nil || evt.Message == nil {
		return "unsupported", "unsupported"
	}
	msg := evt.Message

	if text := strings.TrimSpace(msg.GetConversation()); text != "" {
		return text, "text"
	}
	if text := strings.TrimSpace(msg.GetExtendedTextMessage().GetText()); text != "" {
		return text, "text"
	}
	if caption := strings.TrimSpace(msg.GetImageMessage().GetCaption()); caption != "" {
		return caption, "image"
	}
	if msg.GetImageMessage() != nil {
		return "image", "image"
	}
	if caption := strings.TrimSpace(msg.GetVideoMessage().GetCaption()); caption != "" {
		return caption, "video"
	}
	if msg.GetVideoMessage() != nil {
		return "video", "video"
	}
	if name := strings.TrimSpace(msg.GetDocumentMessage().GetFileName()); name != "" {
		return name, "document"
	}
	if msg.GetDocumentMessage() != nil {
		return "document", "document"
	}
	if msg.GetAudioMessage() != nil {
		return "audio", "audio"
	}
	if msg.GetStickerMessage() != nil {
		return "sticker", "sticker"
	}
	if msg.GetContactMessage() != nil || msg.GetContactsArrayMessage() != nil {
		return "contact", "contact"
	}
	if msg.GetLocationMessage() != nil || msg.GetLiveLocationMessage() != nil {
		return "location", "location"
	}
	return "unsupported", "unsupported"
}

func webhookReceiptStatus(receiptType waTypes.ReceiptType) string {
	switch receiptType {
	case waTypes.ReceiptTypeDelivered:
		return "delivered"
	case waTypes.ReceiptTypeSender:
		return "sent"
	case waTypes.ReceiptTypeRetry:
		return "retry"
	case waTypes.ReceiptTypeRead, waTypes.ReceiptTypeReadSelf:
		return "read"
	case waTypes.ReceiptTypePlayed, waTypes.ReceiptTypePlayedSelf:
		return "played"
	default:
		status := strings.TrimSpace(string(receiptType))
		if status == "" {
			return "delivered"
		}
		return status
	}
}
