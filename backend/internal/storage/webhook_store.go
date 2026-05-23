package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wsapi/internal/domain"
)

var ErrWebhookQueueNotPending = errors.New("webhook queue item is not pending")

type WebhookStore struct {
	db *sql.DB
}

func NewWebhookStore(db *sql.DB) *WebhookStore {
	return &WebhookStore{db: db}
}

func (s *WebhookStore) Create(w *domain.Webhook) error {
	eventosJSON, err := json.Marshal(w.Eventos)
	if err != nil {
		return fmt.Errorf("error al serializar eventos: %w", err)
	}

	query := `INSERT INTO webhooks_outbound (
		empresa_id, telefono_id, api_key_id, url, secret, eventos, activo
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query,
		w.EmpresaID,
		w.TelefonoID,
		w.ApiKeyID,
		w.URL,
		w.Secret,
		string(eventosJSON),
		w.Activo,
	)
	if err != nil {
		return fmt.Errorf("error al crear webhook: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error al obtener ID de webhook: %w", err)
	}

	w.ID = id
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()
	return nil
}

func (s *WebhookStore) scanList(rows *sql.Rows) ([]domain.Webhook, error) {
	var webhooks []domain.Webhook
	for rows.Next() {
		var w domain.Webhook
		var eventosJSON sql.NullString
		var lastError sql.NullString
		var lastSuccessAt sql.NullTime

		err := rows.Scan(
			&w.ID, &w.EmpresaID, &w.TelefonoID, &w.ApiKeyID, &w.URL, &eventosJSON,
			&w.Activo, &w.FailureCount, &lastError, &lastSuccessAt,
			&w.CreatedAt, &w.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear webhook: %w", err)
		}

		if eventosJSON.Valid {
			_ = json.Unmarshal([]byte(eventosJSON.String), &w.Eventos)
		}
		if lastError.Valid {
			w.LastError = &lastError.String
		}
		if lastSuccessAt.Valid {
			w.LastSuccessAt = &lastSuccessAt.Time
		}

		webhooks = append(webhooks, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar webhooks: %w", err)
	}
	return webhooks, nil
}

func (s *WebhookStore) ListByApiKey(apiKeyID int64) ([]domain.Webhook, error) {
	query := `SELECT id, empresa_id, telefono_id, api_key_id, url, eventos, activo, failure_count, last_error, last_success_at, created_at, updated_at
			  FROM webhooks_outbound WHERE api_key_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("error al listar webhooks: %w", err)
	}
	defer rows.Close()
	return s.scanList(rows)
}

func (s *WebhookStore) ListByTelefono(telefonoID int64) ([]domain.Webhook, error) {
	query := `SELECT id, empresa_id, telefono_id, api_key_id, url, eventos, activo, failure_count, last_error, last_success_at, created_at, updated_at
			  FROM webhooks_outbound WHERE telefono_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, telefonoID)
	if err != nil {
		return nil, fmt.Errorf("error al listar webhooks: %w", err)
	}
	defer rows.Close()
	return s.scanList(rows)
}

func (s *WebhookStore) ListByEmpresa(empresaID int64) ([]domain.Webhook, error) {
	query := `SELECT id, empresa_id, telefono_id, api_key_id, url, eventos, activo, failure_count, last_error, last_success_at, created_at, updated_at
			  FROM webhooks_outbound WHERE empresa_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, empresaID)
	if err != nil {
		return nil, fmt.Errorf("error al listar webhooks: %w", err)
	}
	defer rows.Close()
	return s.scanList(rows)
}

func (s *WebhookStore) ListActiveByTelefonoAndEvent(telefonoID int64, eventType domain.WebhookEvent) ([]domain.Webhook, error) {
	webhooks, err := s.ListByTelefono(telefonoID)
	if err != nil {
		return nil, err
	}

	filtered := make([]domain.Webhook, 0, len(webhooks))
	for _, webhook := range webhooks {
		if !webhook.Activo {
			continue
		}
		for _, subscribedEvent := range webhook.Eventos {
			if subscribedEvent == eventType {
				filtered = append(filtered, webhook)
				break
			}
		}
	}
	return filtered, nil
}

func (s *WebhookStore) GetByID(id int64) (*domain.Webhook, error) {
	query := `SELECT id, empresa_id, telefono_id, api_key_id, url, secret, eventos, activo, failure_count, last_error, last_success_at, created_at, updated_at
			  FROM webhooks_outbound WHERE id = ?`

	var w domain.Webhook
	var eventosJSON sql.NullString
	var lastError sql.NullString
	var lastSuccessAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&w.ID, &w.EmpresaID, &w.TelefonoID, &w.ApiKeyID, &w.URL, &w.Secret, &eventosJSON,
		&w.Activo, &w.FailureCount, &lastError, &lastSuccessAt,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener webhook: %w", err)
	}

	if eventosJSON.Valid {
		_ = json.Unmarshal([]byte(eventosJSON.String), &w.Eventos)
	}
	if lastError.Valid {
		w.LastError = &lastError.String
	}
	if lastSuccessAt.Valid {
		w.LastSuccessAt = &lastSuccessAt.Time
	}

	return &w, nil
}

func (s *WebhookStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM webhooks_outbound WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error al eliminar webhook: %w", err)
	}
	return nil
}

func (s *WebhookStore) IncrementFailureCount(id int64) error {
	_, err := s.db.Exec(`UPDATE webhooks_outbound SET failure_count = failure_count + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error al incrementar failure_count: %w", err)
	}
	return nil
}

func (s *WebhookStore) Deactivate(id int64) error {
	_, err := s.db.Exec(`UPDATE webhooks_outbound SET activo = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error al desactivar webhook: %w", err)
	}
	return nil
}

func (s *WebhookStore) EnqueueEvent(item *domain.WebhookQueueItem) error {
	if item.Estado == "" {
		item.Estado = domain.WebhookQueuePending
	}

	query := `INSERT INTO webhooks_outbound_queue (
		webhook_id, payload, estado, proximo_intento_at
	) VALUES (?, ?, ?, ?)`

	result, err := s.db.Exec(query,
		item.WebhookID,
		item.Payload,
		item.Estado,
		item.ProximoIntentoAt,
	)
	if err != nil {
		return fmt.Errorf("error al encolar evento: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error al obtener ID de cola: %w", err)
	}

	item.ID = id
	item.CreatedAt = time.Now()
	return nil
}

func (s *WebhookStore) PollPending(limit int) ([]domain.WebhookQueueItem, error) {
	query := `SELECT id, webhook_id, payload, intentos, proximo_intento_at, estado, last_error, created_at
			  FROM webhooks_outbound_queue
	WHERE estado = 'pending' AND proximo_intento_at <= CURRENT_TIMESTAMP
			  ORDER BY proximo_intento_at ASC
			  LIMIT ?`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error al obtener eventos pendientes: %w", err)
	}
	defer rows.Close()

	var items []domain.WebhookQueueItem
	for rows.Next() {
		var item domain.WebhookQueueItem
		var payload string
		var lastError sql.NullString

		err := rows.Scan(
			&item.ID, &item.WebhookID, &payload, &item.Intentos,
			&item.ProximoIntentoAt, &item.Estado, &lastError, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear item de cola: %w", err)
		}
		item.Payload = json.RawMessage(payload)
		if lastError.Valid {
			item.LastError = &lastError.String
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al iterar items de cola: %w", err)
	}

	return items, nil
}

func (s *WebhookStore) MarkSending(id int64) error {
	result, err := s.db.Exec(`UPDATE webhooks_outbound_queue SET estado = 'sending' WHERE id = ? AND estado = 'pending'`, id)
	if err != nil {
		return fmt.Errorf("error al marcar como sending: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrWebhookQueueNotPending
	}
	return nil
}

func (s *WebhookStore) MarkDone(id int64) error {
	_, err := s.db.Exec(`UPDATE webhooks_outbound_queue SET estado = 'done', intentos = intentos + 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error al marcar como done: %w", err)
	}
	return nil
}

func (s *WebhookStore) MarkFailed(id int64, errMsg string, nextRetryAt time.Time) error {
	_, err := s.db.Exec(
		`UPDATE webhooks_outbound_queue SET estado = 'failed', intentos = intentos + 1, last_error = ?, proximo_intento_at = ? WHERE id = ?`,
		errMsg, nextRetryAt, id,
	)
	if err != nil {
		return fmt.Errorf("error al marcar como failed: %w", err)
	}
	return nil
}

func (s *WebhookStore) MarkQueueFailed(id int64, errMsg string) error {
	_, err := s.db.Exec(
		`UPDATE webhooks_outbound_queue SET estado = 'failed', intentos = intentos + 1, last_error = ? WHERE id = ?`,
		errMsg, id,
	)
	if err != nil {
		return fmt.Errorf("error al marcar cola como failed: %w", err)
	}
	return nil
}

func (s *WebhookStore) MarkDeliveryRetryPending(id int64, errMsg string, nextRetryAt time.Time) error {
	_, err := s.db.Exec(
		`UPDATE webhooks_outbound_queue SET estado = 'pending', intentos = intentos + 1, last_error = ?, proximo_intento_at = ? WHERE id = ?`,
		errMsg, nextRetryAt, id,
	)
	if err != nil {
		return fmt.Errorf("error al reprogramar entrega webhook: %w", err)
	}
	return nil
}

func (s *WebhookStore) MarkDeliverySucceeded(queueID, webhookID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transaccion de entrega exitosa: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE webhooks_outbound_queue SET estado = 'done', intentos = intentos + 1, last_error = NULL WHERE id = ?`, queueID); err != nil {
		return fmt.Errorf("error marcando cola como done: %w", err)
	}
	if _, err := tx.Exec(`UPDATE webhooks_outbound SET failure_count = 0, last_error = NULL, last_success_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, webhookID); err != nil {
		return fmt.Errorf("error actualizando webhook tras entrega exitosa: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error confirmando entrega exitosa: %w", err)
	}
	return nil
}

func (s *WebhookStore) MarkDeliveryFailed(queueID, webhookID int64, errMsg string, deactivateThreshold int) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, fmt.Errorf("error iniciando transaccion de fallo de entrega: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE webhooks_outbound_queue SET estado = 'failed', intentos = intentos + 1, last_error = ? WHERE id = ?`, errMsg, queueID); err != nil {
		return false, fmt.Errorf("error marcando cola como failed: %w", err)
	}

	// 1. Incrementar failure_count y desactivar si corresponde en un solo UPDATE universal atómico
	// Compatible con MySQL y SQLite (no requiere FOR UPDATE)
	_, err = tx.Exec(
		`UPDATE webhooks_outbound 
		 SET failure_count = failure_count + 1, 
		     last_error = ?, 
		     activo = CASE WHEN ? > 0 AND failure_count + 1 >= ? THEN FALSE ELSE activo END,
		     updated_at = CURRENT_TIMESTAMP 
		 WHERE id = ?`,
		errMsg, deactivateThreshold, deactivateThreshold, webhookID,
	)
	if err != nil {
		return false, fmt.Errorf("error actualizando fallo y estado de webhook: %w", err)
	}

	// 2. Leer el estado activo tras la actualización para saber si se desactivó
	var activo bool
	err = tx.QueryRow(`SELECT activo FROM webhooks_outbound WHERE id = ?`, webhookID).Scan(&activo)
	if err != nil {
		return false, fmt.Errorf("error leyendo estado activo de webhook post-actualizacion: %w", err)
	}

	deactivated := !activo

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("error confirmando fallo de entrega: %w", err)
	}
	return deactivated, nil
}
