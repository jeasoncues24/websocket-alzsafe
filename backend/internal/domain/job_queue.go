package domain

import (
	"math"
	"time"
)

// JobType identifica el tipo de trabajo en la cola.
type JobType string

const (
	JobTypeBroadcast        JobType = "broadcast"
	JobTypeScheduledMessage JobType = "scheduled_message"
)

// JobStatus ciclo de vida de un job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// JobItemStatus estado de un item individual dentro de un job.
type JobItemStatus string

const (
	JobItemPending JobItemStatus = "pending"
	JobItemSent    JobItemStatus = "sent"
	JobItemFailed  JobItemStatus = "failed"
	JobItemSkipped JobItemStatus = "skipped"
)

// Job representa un trabajo encolado genérico persistido en job_queue.
type Job struct {
	ID            int64      `json:"id"`
	Type          JobType    `json:"type"`
	EntityID      string     `json:"entity_id"`
	Status        JobStatus  `json:"status"`
	Priority      int        `json:"priority"`
	EmpresaID     int64      `json:"empresa_id"`
	AttemptCount  int        `json:"attempt_count"`
	MaxAttempts   int        `json:"max_attempts"`
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	Metadata      string     `json:"metadata,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// JobItem representa un destinatario o unidad de trabajo dentro de un job.
type JobItem struct {
	ID            int64         `json:"id"`
	JobID         int64         `json:"job_id"`
	SequenceOrder int           `json:"sequence_order"`
	Payload       string        `json:"payload"`
	Status        JobItemStatus `json:"status"`
	AttemptCount  int           `json:"attempt_count"`
	ErrorText     string        `json:"error_text,omitempty"`
	ProcessedAt   *time.Time    `json:"processed_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// BroadcastTimingConfig parámetros de delay para calcular tiempo estimado.
type BroadcastTimingConfig struct {
	BatchSizeMin, BatchSizeMax     int
	IntraBatchDelayMin             time.Duration
	IntraBatchDelayMax             time.Duration
	InterBatchDelayMin             time.Duration
	InterBatchDelayMax             time.Duration
	MacroPauseEvery                int
	MacroPauseMin, MacroPauseMax   time.Duration
}

// EstimateBroadcastSeconds calcula el tiempo estimado en segundos para enviar n mensajes.
// Usa los valores medios de cada rango de delay.
func EstimateBroadcastSeconds(n int, cfg BroadcastTimingConfig) int {
	if n <= 0 {
		return 0
	}

	avgIntra := avgDuration(cfg.IntraBatchDelayMin, cfg.IntraBatchDelayMax)
	avgInter := avgDuration(cfg.InterBatchDelayMin, cfg.InterBatchDelayMax)
	avgMacro := avgDuration(cfg.MacroPauseMin, cfg.MacroPauseMax)

	batchSize := float64(cfg.BatchSizeMin+cfg.BatchSizeMax) / 2.0
	batches := math.Ceil(float64(n) / batchSize)

	// delays entre mensajes dentro de batches (n-1 delays intra, 1 menos por el último de cada batch)
	msgsInBatches := float64(n-int(batches)) * float64(avgIntra/time.Millisecond)
	// delays entre batches
	interPauses := (batches - 1) * float64(avgInter/time.Millisecond)
	// macro-pausas cada MacroPauseEvery mensajes
	var macroPauses float64
	if cfg.MacroPauseEvery > 0 {
		macroPauses = math.Floor(float64(n)/float64(cfg.MacroPauseEvery)) * float64(avgMacro/time.Millisecond)
	}

	totalMs := msgsInBatches + interPauses + macroPauses
	return int(math.Ceil(totalMs / 1000.0))
}

func avgDuration(a, b time.Duration) time.Duration {
	return (a + b) / 2
}
