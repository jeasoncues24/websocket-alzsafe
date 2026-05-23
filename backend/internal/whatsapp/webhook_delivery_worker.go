package whatsapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

const (
	defaultWebhookPollInterval        = 5 * time.Second
	defaultWebhookRequestTimeout      = 10 * time.Second
	defaultWebhookBatchSize           = 20
	defaultWebhookMaxAttempts         = 6
	defaultWebhookDeactivateThreshold = 20
	maxWebhookDeliveryErrorLength     = 512
)

var defaultWebhookRetrySchedule = []time.Duration{
	time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	6 * time.Hour,
}

type WebhookDeliveryWorkerConfig struct {
	PollInterval        time.Duration
	RequestTimeout      time.Duration
	BatchSize           int
	MaxAttempts         int
	DeactivateThreshold int
	RetrySchedule       []time.Duration
	NowFunc             func() time.Time
	Logger              *zerolog.Logger
}

type WebhookDeliveryWorker struct {
	store               *storage.WebhookStore
	client              *http.Client
	logger              zerolog.Logger
	pollInterval        time.Duration
	batchSize           int
	maxAttempts         int
	deactivateThreshold int
	retrySchedule       []time.Duration
	nowFunc             func() time.Time
}

func NewWebhookDeliveryWorker(store *storage.WebhookStore, client *http.Client, cfg WebhookDeliveryWorkerConfig) *WebhookDeliveryWorker {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaultWebhookPollInterval
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = defaultWebhookRequestTimeout
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultWebhookBatchSize
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaultWebhookMaxAttempts
	}
	if cfg.DeactivateThreshold <= 0 {
		cfg.DeactivateThreshold = defaultWebhookDeactivateThreshold
	}
	if len(cfg.RetrySchedule) == 0 {
		cfg.RetrySchedule = append([]time.Duration(nil), defaultWebhookRetrySchedule...)
	}
	if cfg.NowFunc == nil {
		cfg.NowFunc = time.Now
	}

	logger := config.GetLogger().With().Str("component", "webhook_delivery_worker").Logger()
	if cfg.Logger != nil {
		logger = cfg.Logger.With().Str("component", "webhook_delivery_worker").Logger()
	}

	if client == nil {
		client = &http.Client{Timeout: cfg.RequestTimeout}
	} else if client.Timeout == 0 {
		clone := *client
		clone.Timeout = cfg.RequestTimeout
		client = &clone
	}

	return &WebhookDeliveryWorker{
		store:               store,
		client:              client,
		logger:              logger,
		pollInterval:        cfg.PollInterval,
		batchSize:           cfg.BatchSize,
		maxAttempts:         cfg.MaxAttempts,
		deactivateThreshold: cfg.DeactivateThreshold,
		retrySchedule:       append([]time.Duration(nil), cfg.RetrySchedule...),
		nowFunc:             cfg.NowFunc,
	}
}

func (w *WebhookDeliveryWorker) Run(ctx context.Context) {
	if w == nil || w.store == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if err := w.processDueItems(ctx); err != nil && !errors.Is(err, context.Canceled) {
		w.logger.Error().Err(err).Msg("error procesando webhooks pendientes")
	}

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info().Msg("worker de webhooks detenido por context cancelado")
			return
		case <-ticker.C:
			if err := w.processDueItems(ctx); err != nil && !errors.Is(err, context.Canceled) {
				w.logger.Error().Err(err).Msg("error procesando webhooks pendientes")
			}
		}
	}
}

func (w *WebhookDeliveryWorker) processDueItems(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	items, err := w.store.PollPending(w.batchSize)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := w.processQueueItem(ctx, item); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			w.logger.Error().Err(err).Int64("queue_id", item.ID).Int64("webhook_id", item.WebhookID).Msg("error procesando item de webhook")
		}
	}

	return nil
}

func (w *WebhookDeliveryWorker) processQueueItem(ctx context.Context, item domain.WebhookQueueItem) error {
	if err := w.store.MarkSending(item.ID); err != nil {
		if errors.Is(err, storage.ErrWebhookQueueNotPending) {
			w.logger.Debug().Int64("queue_id", item.ID).Msg("item de webhook ya fue reclamado por otro worker")
			return nil
		}
		return err
	}

	webhook, err := w.store.GetByID(item.WebhookID)
	if err != nil {
		return w.store.MarkQueueFailed(item.ID, truncateWebhookDeliveryError(err.Error()))
	}
	if webhook == nil {
		return w.store.MarkQueueFailed(item.ID, "webhook no encontrado")
	}

	envelope, err := decodeWebhookDeliveryEnvelope(item.Payload)
	if err != nil {
		w.logger.Warn().Err(err).Int64("queue_id", item.ID).Str("payload_raw", string(item.Payload)).Msg("payload invalido en cola de webhook")
		w.logDeliveryResult(item, webhook.ID, "", 0, item.Intentos+1, 0, "failed", err)
		_, failErr := w.store.MarkDeliveryFailed(item.ID, webhook.ID, truncateWebhookDeliveryError(err.Error()), w.deactivateThreshold)
		return failErr
	}

	attempt := item.Intentos + 1
	startedAt := w.nowFunc()
	statusCode, deliveryErr := w.deliver(ctx, webhook, item.ID, envelope)
	latency := w.nowFunc().Sub(startedAt)

	if errors.Is(deliveryErr, context.Canceled) {
		return deliveryErr
	}

	if deliveryErr == nil && statusCode >= 200 && statusCode < 300 {
		if err := w.store.MarkDeliverySucceeded(item.ID, webhook.ID); err != nil {
			return err
		}
		w.logDeliveryResult(item, webhook.ID, envelope.EventType, statusCode, attempt, latency, "done", nil)
		return nil
	}

	errMsg := buildWebhookDeliveryError(statusCode, deliveryErr)
	if w.shouldRetry(statusCode, deliveryErr, attempt) {
		nextRetryAt := w.nowFunc().Add(w.retryDelayForAttempt(attempt))
		if err := w.store.MarkDeliveryRetryPending(item.ID, truncateWebhookDeliveryError(errMsg), nextRetryAt); err != nil {
			return err
		}
		w.logDeliveryResult(item, webhook.ID, envelope.EventType, statusCode, attempt, latency, "retry_pending", deliveryErrOrStatusError(statusCode, deliveryErr))
		return nil
	}

	deactivated, err := w.store.MarkDeliveryFailed(item.ID, webhook.ID, truncateWebhookDeliveryError(errMsg), w.deactivateThreshold)
	if err != nil {
		return err
	}
	result := "failed"
	if deactivated {
		result = "failed_webhook_deactivated"
	}
	w.logDeliveryResult(item, webhook.ID, envelope.EventType, statusCode, attempt, latency, result, deliveryErrOrStatusError(statusCode, deliveryErr))
	return nil
}

func (w *WebhookDeliveryWorker) deliver(ctx context.Context, webhook *domain.Webhook, queueID int64, envelope domain.WebhookDeliveryEnvelope) (int, error) {
	body := envelope.Data
	if len(body) == 0 {
		body = []byte(`{}`)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Wsapi-Signature", buildWebhookSignature(body, webhook.Secret))
	req.Header.Set("X-Wsapi-Event", string(envelope.EventType))
	req.Header.Set("X-Wsapi-Delivery", strconv.FormatInt(queueID, 10))

	resp, err := w.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.CopyN(io.Discard, resp.Body, 1<<20)
	return resp.StatusCode, nil
}

func decodeWebhookDeliveryEnvelope(payload json.RawMessage) (domain.WebhookDeliveryEnvelope, error) {
	var envelope domain.WebhookDeliveryEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return envelope, fmt.Errorf("payload invalido: %w", err)
	}
	if strings.TrimSpace(string(envelope.EventType)) == "" {
		return envelope, errors.New("payload invalido: falta event_type")
	}
	if len(bytes.TrimSpace(envelope.Data)) == 0 {
		return envelope, errors.New("payload invalido: falta data")
	}
	return envelope, nil
}

func buildWebhookSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (w *WebhookDeliveryWorker) retryDelayForAttempt(attempt int) time.Duration {
	if len(w.retrySchedule) == 0 {
		return time.Minute
	}
	index := attempt - 1
	if index < 0 {
		index = 0
	}
	if index >= len(w.retrySchedule) {
		index = len(w.retrySchedule) - 1
	}
	return w.retrySchedule[index]
}

func (w *WebhookDeliveryWorker) shouldRetry(statusCode int, err error, attempt int) bool {
	if attempt >= w.maxAttempts {
		return false
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false
		}
		return true
	}
	return statusCode >= 500
}

func buildWebhookDeliveryError(statusCode int, err error) string {
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("webhook respondio con status %d", statusCode)
}

func deliveryErrOrStatusError(statusCode int, err error) error {
	if err != nil {
		return err
	}
	if statusCode == 0 {
		return nil
	}
	return fmt.Errorf("status %d", statusCode)
}

func truncateWebhookDeliveryError(msg string) string {
	msg = strings.TrimSpace(msg)
	if len(msg) <= maxWebhookDeliveryErrorLength {
		return msg
	}
	return msg[:maxWebhookDeliveryErrorLength]
}

func (w *WebhookDeliveryWorker) logDeliveryResult(item domain.WebhookQueueItem, webhookID int64, eventType domain.WebhookEvent, statusCode int, attempt int, latency time.Duration, result string, err error) {
	evt := w.logger.Info()
	if err != nil {
		evt = w.logger.Warn().Err(err)
	}
	evt.
		Int64("queue_id", item.ID).
		Int64("webhook_id", webhookID).
		Str("event_type", string(eventType)).
		Int("status_code", statusCode).
		Int("attempt", attempt).
		Int64("latency_ms", latency.Milliseconds()).
		Str("resultado", result).
		Msg("entrega de webhook procesada")
}
