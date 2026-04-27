package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"wsapi/internal/domain"
)

func (s *ApiKeyStore) RecordUsageEvent(event *domain.ApiKeyUsageEvent) error {
	_, err := s.db.Exec(
		`INSERT INTO api_key_usage_events (
			api_key_id, empresa_id, telefono_id, method, endpoint, status_code, latency_ms,
			request_units, response_units, request_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ApiKeyID, event.EmpresaID, event.TelefonoID, event.Method, event.Endpoint,
		event.StatusCode, event.LatencyMS, event.RequestUnits, event.ResponseUnits, event.RequestID,
	)
	if err != nil {
		return fmt.Errorf("error al registrar uso de api key: %w", err)
	}
	return s.TouchLastUsed(event.ApiKeyID)
}

func (s *ApiKeyStore) UpsertDailyUsage(usage *domain.ApiKeyUsageDaily) error {
	var currentRequestCount, currentSuccessCount, currentErrorCount, currentLatencyAvg, currentMessagesSent, currentBroadcastsSent int
	var currentBytesIn, currentBytesOut int64
	err := s.db.QueryRow(
		`SELECT request_count, success_count, error_count, latency_avg_ms, messages_sent, broadcasts_sent, bytes_in, bytes_out
		 FROM api_key_usage_daily WHERE day = ? AND api_key_id = ?`,
		usage.Day, usage.ApiKeyID,
	).Scan(&currentRequestCount, &currentSuccessCount, &currentErrorCount, &currentLatencyAvg, &currentMessagesSent, &currentBroadcastsSent, &currentBytesIn, &currentBytesOut)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error al leer uso diario actual: %w", err)
	}

	newRequestCount := currentRequestCount + usage.RequestCount
	newSuccessCount := currentSuccessCount + usage.SuccessCount
	newErrorCount := currentErrorCount + usage.ErrorCount
	newMessagesSent := currentMessagesSent + usage.MessagesSent
	newBroadcastsSent := currentBroadcastsSent + usage.BroadcastsSent
	newBytesIn := currentBytesIn + usage.BytesIn
	newBytesOut := currentBytesOut + usage.BytesOut
	newLatencyAvg := usage.LatencyAvgMS
	if newRequestCount > 0 {
		weighted := (currentLatencyAvg * currentRequestCount) + (usage.LatencyAvgMS * usage.RequestCount)
		newLatencyAvg = weighted / newRequestCount
	}

	if err == sql.ErrNoRows {
		_, err = s.db.Exec(
			`INSERT INTO api_key_usage_daily (
				day, api_key_id, empresa_id, telefono_id, request_count, success_count, error_count,
				latency_avg_ms, messages_sent, broadcasts_sent, bytes_in, bytes_out
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			usage.Day, usage.ApiKeyID, usage.EmpresaID, usage.TelefonoID,
			usage.RequestCount, usage.SuccessCount, usage.ErrorCount, usage.LatencyAvgMS,
			usage.MessagesSent, usage.BroadcastsSent, usage.BytesIn, usage.BytesOut,
		)
		if err != nil {
			return fmt.Errorf("error al crear uso diario: %w", err)
		}
		return nil
	}

	_, err = s.db.Exec(
		`UPDATE api_key_usage_daily SET
			request_count = ?, success_count = ?, error_count = ?, latency_avg_ms = ?,
			messages_sent = ?, broadcasts_sent = ?, bytes_in = ?, bytes_out = ?, updated_at = NOW()
		 WHERE day = ? AND api_key_id = ?`,
		newRequestCount, newSuccessCount, newErrorCount, newLatencyAvg,
		newMessagesSent, newBroadcastsSent, newBytesIn, newBytesOut,
		usage.Day, usage.ApiKeyID,
	)
	if err != nil {
		return fmt.Errorf("error al registrar uso diario de api key: %w", err)
	}
	return nil
}

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

func (s *ApiKeyStore) GetUsageDailyByKey(apiKeyID int64) ([]domain.ApiKeyUsageDaily, error) {
	rows, err := s.db.Query(
		`SELECT day, api_key_id, empresa_id, telefono_id, request_count, success_count, error_count,
			latency_avg_ms, messages_sent, broadcasts_sent, bytes_in, bytes_out
		 FROM api_key_usage_daily WHERE api_key_id = ? ORDER BY day DESC`,
		apiKeyID,
	)
	if err != nil {
		return nil, fmt.Errorf("error al obtener uso diario: %w", err)
	}
	defer rows.Close()

	var items []domain.ApiKeyUsageDaily
	for rows.Next() {
		var item domain.ApiKeyUsageDaily
		var day time.Time
		if err := rows.Scan(
			&day, &item.ApiKeyID, &item.EmpresaID, &item.TelefonoID, &item.RequestCount,
			&item.SuccessCount, &item.ErrorCount, &item.LatencyAvgMS, &item.MessagesSent,
			&item.BroadcastsSent, &item.BytesIn, &item.BytesOut,
		); err != nil {
			return nil, fmt.Errorf("error al escanear uso diario: %w", err)
		}
		item.Day = day.Format("2006-01-02")
		items = append(items, item)
	}
	return items, nil
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
