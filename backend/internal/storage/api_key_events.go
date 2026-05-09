package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"wsapi/internal/domain"
)

func (s *ApiKeyStore) RecordAuditEvent(event *domain.ApiKeyAuditEvent) error {
	_, err := s.db.Exec(
		`INSERT INTO api_key_audit_events (
			api_key_id, empresa_id, telefono_id, action, actor_user_id, metadata
		) VALUES (?, ?, ?, ?, ?, ?)`,
		event.ApiKeyID, event.EmpresaID, event.TelefonoID, event.Action, event.ActorUserID, event.Metadata,
	)
	if err != nil {
		return fmt.Errorf("error al registrar auditoria de api key: %w", err)
	}
	return nil
}

func (s *ApiKeyStore) GetAuditEventsByKey(apiKeyID int64) ([]domain.ApiKeyAuditEvent, error) {
	rows, err := s.db.Query(
		`SELECT id, api_key_id, empresa_id, telefono_id, action, actor_user_id, metadata, created_at
		 FROM api_key_audit_events WHERE api_key_id = ? ORDER BY created_at DESC`,
		apiKeyID,
	)
	if err != nil {
		return nil, fmt.Errorf("error al obtener auditoria: %w", err)
	}
	defer rows.Close()

	var items []domain.ApiKeyAuditEvent
	for rows.Next() {
		var item domain.ApiKeyAuditEvent
		var actorID sql.NullInt64
		var metadata sql.NullString
		if err := rows.Scan(&item.ID, &item.ApiKeyID, &item.EmpresaID, &item.TelefonoID, &item.Action, &actorID, &metadata, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("error al escanear auditoria: %w", err)
		}
		if actorID.Valid {
			v := actorID.Int64
			item.ActorUserID = &v
		}
		if metadata.Valid {
			item.Metadata = json.RawMessage(metadata.String)
		}
		items = append(items, item)
	}
	return items, nil
}
