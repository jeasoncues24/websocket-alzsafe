package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wsapi/internal/domain"
)

// JobQueueRepository define operaciones de persistencia para la cola genérica de jobs.
type JobQueueRepository interface {
	// CreateJobWithItems persiste un job y todos sus items en una transacción atómica.
	CreateJobWithItems(ctx context.Context, job *domain.Job, items []domain.JobItem) error

	// GetPendingItems retorna los items pendientes de un job ordenados por sequence_order.
	GetPendingItems(ctx context.Context, jobID int64) ([]domain.JobItem, error)

	// GetByEntityID obtiene un job por su entity_id (reference_id del broadcast).
	GetByEntityID(ctx context.Context, entityID string) (*domain.Job, error)

	// UpdateStatus actualiza el estado de un job.
	UpdateStatus(ctx context.Context, jobID int64, status domain.JobStatus, completedAt *time.Time) error

	// Heartbeat actualiza last_heartbeat para evitar que el job sea marcado como stuck.
	Heartbeat(ctx context.Context, jobID int64) error

	// UpdateItemStatus actualiza el estado de un item individual tras el intento de envío.
	UpdateItemStatus(ctx context.Context, itemID int64, status domain.JobItemStatus, errText string) error

	// GetAllItems retorna todos los items de un job ordenados por sequence_order.
	GetAllItems(ctx context.Context, jobID int64) ([]domain.JobItem, error)

	// ListByEmpresa retorna todos los jobs de una empresa, más recientes primero.
	ListByEmpresa(ctx context.Context, empresaID int64) ([]domain.Job, error)

	// RecoverStuckJobs resetea a 'pending' los jobs que llevan más de threshold en 'running'
	// sin actualizar su heartbeat. Retorna el número de jobs recuperados.
	RecoverStuckJobs(ctx context.Context, threshold time.Duration) (int, error)

	// GetItemStatsByJobs retorna estadísticas agrupadas de items para varios jobs en una sola query.
	GetItemStatsByJobs(ctx context.Context, jobIDs []int64) (map[int64]struct{ Success, Failed, Total int }, error)

	// GetPendingJobs retorna todos los jobs en estado 'pending'.
	GetPendingJobs(ctx context.Context) ([]domain.Job, error)
}

type mysqlJobQueueRepository struct {
	db *sql.DB
}

// NewJobQueueRepository crea un repositorio MySQL para la cola de jobs.
func NewJobQueueRepository(db *sql.DB) JobQueueRepository {
	return &mysqlJobQueueRepository{db: db}
}

func (r *mysqlJobQueueRepository) CreateJobWithItems(ctx context.Context, job *domain.Job, items []domain.JobItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("job_queue: begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO job_queue (type, entity_id, status, priority, empresa_id, max_attempts, metadata, created_at)
		VALUES (?, ?, 'pending', ?, ?, ?, ?, ?)`,
		string(job.Type), job.EntityID, job.Priority, job.EmpresaID, job.MaxAttempts, nullableString(job.Metadata), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("job_queue: insert job: %w", err)
	}
	jobID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("job_queue: last insert id: %w", err)
	}
	job.ID = jobID

	for i := range items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO job_items (job_id, sequence_order, payload, status, created_at)
			VALUES (?, ?, ?, 'pending', ?)`,
			jobID, items[i].SequenceOrder, items[i].Payload, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("job_queue: insert item[%d]: %w", i, err)
		}
	}

	return tx.Commit()
}

func (r *mysqlJobQueueRepository) GetByEntityID(ctx context.Context, entityID string) (*domain.Job, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, entity_id, status, priority, empresa_id,
		       attempt_count, max_attempts, last_heartbeat, next_retry_at,
		       created_at, metadata, started_at, completed_at
		FROM job_queue WHERE entity_id = ? LIMIT 1`, entityID)

	job := &domain.Job{}
	var metadata sql.NullString
	err := row.Scan(
		&job.ID, &job.Type, &job.EntityID, &job.Status,
		&job.Priority, &job.EmpresaID, &job.AttemptCount, &job.MaxAttempts,
		&job.LastHeartbeat, &job.NextRetryAt,
		&job.CreatedAt, &metadata, &job.StartedAt, &job.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("job_queue: get by entity_id: %w", err)
	}
	if metadata.Valid {
		job.Metadata = metadata.String
	}
	return job, nil
}

func (r *mysqlJobQueueRepository) UpdateStatus(ctx context.Context, jobID int64, status domain.JobStatus, completedAt *time.Time) error {
	if status == domain.JobStatusRunning {
		now := time.Now()
		_, err := r.db.ExecContext(ctx, `
			UPDATE job_queue
			SET status = ?, started_at = ?, last_heartbeat = ?
			WHERE id = ?`,
			string(status), now, now, jobID,
		)
		return err
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE job_queue
		SET status = ?, completed_at = ?
		WHERE id = ?`,
		string(status), completedAt, jobID,
	)
	return err
}

func (r *mysqlJobQueueRepository) Heartbeat(ctx context.Context, jobID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE job_queue SET last_heartbeat = ? WHERE id = ?`, time.Now(), jobID)
	return err
}

func (r *mysqlJobQueueRepository) UpdateItemStatus(ctx context.Context, itemID int64, status domain.JobItemStatus, errText string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE job_items
		SET status = ?, error_text = ?, processed_at = ?, attempt_count = attempt_count + 1
		WHERE id = ?`,
		string(status), nullableString(errText), time.Now(), itemID,
	)
	return err
}

func (r *mysqlJobQueueRepository) GetAllItems(ctx context.Context, jobID int64) ([]domain.JobItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, job_id, sequence_order, payload, status, attempt_count, error_text, processed_at, created_at
		FROM job_items WHERE job_id = ? ORDER BY sequence_order ASC`, jobID)
	if err != nil {
		return nil, fmt.Errorf("job_queue: get items: %w", err)
	}
	defer rows.Close()

	var items []domain.JobItem
	for rows.Next() {
		var item domain.JobItem
		var errText sql.NullString
		if err := rows.Scan(
			&item.ID, &item.JobID, &item.SequenceOrder, &item.Payload,
			&item.Status, &item.AttemptCount, &errText, &item.ProcessedAt, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("job_queue: scan item: %w", err)
		}
		if errText.Valid {
			item.ErrorText = errText.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *mysqlJobQueueRepository) GetPendingItems(ctx context.Context, jobID int64) ([]domain.JobItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, job_id, sequence_order, payload, status, attempt_count, error_text, processed_at, created_at
		FROM job_items WHERE job_id = ? AND status = 'pending' ORDER BY sequence_order ASC`, jobID)
	if err != nil {
		return nil, fmt.Errorf("job_queue: get pending items: %w", err)
	}
	defer rows.Close()

	var items []domain.JobItem
	for rows.Next() {
		var item domain.JobItem
		var errText sql.NullString
		if err := rows.Scan(
			&item.ID, &item.JobID, &item.SequenceOrder, &item.Payload,
			&item.Status, &item.AttemptCount, &errText, &item.ProcessedAt, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("job_queue: scan pending item: %w", err)
		}
		if errText.Valid {
			item.ErrorText = errText.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *mysqlJobQueueRepository) ListByEmpresa(ctx context.Context, empresaID int64) ([]domain.Job, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, entity_id, status, priority, empresa_id,
		       attempt_count, max_attempts, last_heartbeat, next_retry_at,
		       created_at, metadata, started_at, completed_at
		FROM job_queue WHERE empresa_id = ? ORDER BY created_at DESC`, empresaID)
	if err != nil {
		return nil, fmt.Errorf("job_queue: list by empresa: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var job domain.Job
		var metadata sql.NullString
		if err := rows.Scan(
			&job.ID, &job.Type, &job.EntityID, &job.Status,
			&job.Priority, &job.EmpresaID, &job.AttemptCount, &job.MaxAttempts,
			&job.LastHeartbeat, &job.NextRetryAt,
			&job.CreatedAt, &metadata, &job.StartedAt, &job.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("job_queue: scan job: %w", err)
		}
		if metadata.Valid {
			job.Metadata = metadata.String
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *mysqlJobQueueRepository) RecoverStuckJobs(ctx context.Context, threshold time.Duration) (int, error) {
	cutoff := time.Now().Add(-threshold)
	res, err := r.db.ExecContext(ctx, `
		UPDATE job_queue
		SET status = 'pending', attempt_count = attempt_count + 1, last_heartbeat = NULL
		WHERE status = 'running'
		  AND (last_heartbeat IS NULL OR last_heartbeat < ?)`,
		cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("job_queue: recover stuck: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// BroadcastJobPayload es el JSON almacenado en job_items.payload para jobs de tipo broadcast.
type BroadcastJobPayload struct {
	Destino string `json:"destino"`
	Mensaje string `json:"mensaje"`
}

// EncodeBroadcastPayload serializa un BroadcastItem a JSON para almacenar en job_items.
func EncodeBroadcastPayload(destino, mensaje string) (string, error) {
	b, err := json.Marshal(BroadcastJobPayload{Destino: destino, Mensaje: mensaje})
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DecodeBroadcastPayload deserializa el payload JSON de un job_item a destino+mensaje.
func DecodeBroadcastPayload(payload string) (destino, mensaje string, err error) {
	var p BroadcastJobPayload
	if err = json.Unmarshal([]byte(payload), &p); err != nil {
		return
	}
	return p.Destino, p.Mensaje, nil
}

func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func (r *mysqlJobQueueRepository) GetItemStatsByJobs(ctx context.Context, jobIDs []int64) (map[int64]struct{ Success, Failed, Total int }, error) {
	stats := make(map[int64]struct{ Success, Failed, Total int })
	if len(jobIDs) == 0 {
		return stats, nil
	}

	placeholders := make([]string, len(jobIDs))
	args := make([]interface{}, len(jobIDs))
	for i, id := range jobIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT job_id,
		       SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) AS success_count,
		       SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failed_count,
		       COUNT(*) AS total_count
		FROM job_items
		WHERE job_id IN (%s)
		GROUP BY job_id`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var jobID int64
		var success, failed, total int
		if err := rows.Scan(&jobID, &success, &failed, &total); err != nil {
			return nil, err
		}
		stats[jobID] = struct{ Success, Failed, Total int }{
			Success: success,
			Failed:  failed,
			Total:   total,
		}
	}
	return stats, rows.Err()
}

func (r *mysqlJobQueueRepository) GetPendingJobs(ctx context.Context) ([]domain.Job, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, entity_id, status, priority, empresa_id,
		       attempt_count, max_attempts, last_heartbeat, next_retry_at,
		       created_at, metadata, started_at, completed_at
		FROM job_queue WHERE status = 'pending' ORDER BY priority DESC, created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("job_queue: get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var job domain.Job
		var metadata sql.NullString
		if err := rows.Scan(
			&job.ID, &job.Type, &job.EntityID, &job.Status,
			&job.Priority, &job.EmpresaID, &job.AttemptCount, &job.MaxAttempts,
			&job.LastHeartbeat, &job.NextRetryAt,
			&job.CreatedAt, &metadata, &job.StartedAt, &job.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("job_queue: scan pending job: %w", err)
		}
		if metadata.Valid {
			job.Metadata = metadata.String
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}
