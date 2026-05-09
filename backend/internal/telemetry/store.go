package telemetry

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"wsapi/internal/domain"
)

type MySQLStore struct {
	db     *sql.DB
	cfg    Config
	buffer []*domain.TelemetryEvent
	mu     sync.Mutex
	done   chan struct{}
}

func NewMySQLStore(db *sql.DB, cfg Config) *MySQLStore {
	s := &MySQLStore{
		db:     db,
		cfg:    cfg,
		buffer: make([]*domain.TelemetryEvent, 0, cfg.BufferSize),
		done:   make(chan struct{}),
	}
	if cfg.Enabled {
		go s.flushLoop()
	}
	return s
}

func (s *MySQLStore) Record(event *domain.TelemetryEvent) error {
	if !s.cfg.Enabled {
		return nil
	}
	s.mu.Lock()
	s.buffer = append(s.buffer, event)
	shouldFlush := len(s.buffer) >= s.cfg.BatchSize
	s.mu.Unlock()

	if shouldFlush {
		return s.Flush()
	}
	return nil
}

func (s *MySQLStore) Flush() error {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return nil
	}
	batch := s.buffer
	s.buffer = make([]*domain.TelemetryEvent, 0, s.cfg.BufferSize)
	s.mu.Unlock()

	return s.insertBatch(batch)
}

func (s *MySQLStore) insertBatch(events []*domain.TelemetryEvent) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("telemetry: begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.insertRequestLogs(tx, events); err != nil {
		return err
	}

	buckets := aggregateToBuckets(events)
	if err := s.upsertMetricsBuckets(tx, buckets); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *MySQLStore) insertRequestLogs(tx *sql.Tx, events []*domain.TelemetryEvent) error {
	stmt, err := tx.Prepare(`INSERT INTO telefono_request_logs
		(api_key_id, empresa_id, telefono_id, contract_name, endpoint, method,
		 status_code, latency_ms, error_code, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("telemetry: prepare request logs: %w", err)
	}
	defer stmt.Close()

	for _, e := range events {
		var apiKeyID, empresaID, telefonoID *int64
		if e.ApiKeyID > 0 {
			apiKeyID = &e.ApiKeyID
		}
		if e.EmpresaID > 0 {
			empresaID = &e.EmpresaID
		}
		if e.TelefonoID > 0 {
			telefonoID = &e.TelefonoID
		}
		var errorCode, errorMsg *string
		if e.ErrorCode != "" {
			errorCode = &e.ErrorCode
		}
		if e.ErrorMessage != "" {
			errorMsg = &e.ErrorMessage
		}

		if _, err = stmt.Exec(apiKeyID, empresaID, telefonoID, e.ContractName, e.Endpoint,
			e.Method, e.StatusCode, e.LatencyMS, errorCode, errorMsg, e.CreatedAt); err != nil {
			return fmt.Errorf("telemetry: insert request log: %w", err)
		}
	}
	return nil
}

// metricsBucket es un agregado por (api_key_id, contract_name, bucket_min).
type metricsBucket struct {
	ApiKeyID     int64
	ContractName string
	BucketMin    time.Time
	RequestCount int
	SuccessCount int
	ErrorCount   int
	LatencyP50   float64
	LatencyP95   float64
	LatencyP99   float64
}

type bucketKey struct {
	apiKeyID     int64
	contractName string
	bucketMin    time.Time
}

// aggregateToBuckets agrupa events por minuto y calcula conteos y percentiles de latencia.
func aggregateToBuckets(events []*domain.TelemetryEvent) []metricsBucket {
	groups := make(map[bucketKey][]int)

	for _, e := range events {
		if e.ApiKeyID == 0 {
			continue
		}
		key := bucketKey{
			apiKeyID:     e.ApiKeyID,
			contractName: e.ContractName,
			bucketMin:    truncateToMinute(e.CreatedAt),
		}
		groups[key] = append(groups[key], e.LatencyMS)
	}

	buckets := make([]metricsBucket, 0, len(groups))
	for key, latencies := range groups {
		sort.Ints(latencies)
		b := metricsBucket{
			ApiKeyID:     key.apiKeyID,
			ContractName: key.contractName,
			BucketMin:    key.bucketMin,
			RequestCount: len(latencies),
			LatencyP50:   percentileInt(latencies, 0.50),
			LatencyP95:   percentileInt(latencies, 0.95),
			LatencyP99:   percentileInt(latencies, 0.99),
		}
		// Contar success/error por status_code en el batch original
		for _, e := range events {
			if e.ApiKeyID == key.apiKeyID && e.ContractName == key.contractName &&
				truncateToMinute(e.CreatedAt) == key.bucketMin {
				if e.StatusCode < 400 {
					b.SuccessCount++
				} else {
					b.ErrorCount++
				}
			}
		}
		buckets = append(buckets, b)
	}
	return buckets
}

// upsertMetricsBuckets hace INSERT ... ON DUPLICATE KEY UPDATE con media ponderada de latencias.
func (s *MySQLStore) upsertMetricsBuckets(tx *sql.Tx, buckets []metricsBucket) error {
	if len(buckets) == 0 {
		return nil
	}

	// La media ponderada en SQL usa la columna actual antes de la actualización.
	// VALUES() referencia el valor propuesto en el INSERT.
	stmt, err := tx.Prepare(`
		INSERT INTO telefono_metrics_min
			(api_key_id, contract_name, bucket_min,
			 request_count, success_count, error_count,
			 latency_p50_ms, latency_p95_ms, latency_p99_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			request_count  = request_count  + VALUES(request_count),
			success_count  = success_count  + VALUES(success_count),
			error_count    = error_count    + VALUES(error_count),
			latency_p50_ms = ROUND(
				(latency_p50_ms * request_count + VALUES(latency_p50_ms) * VALUES(request_count))
				/ (request_count + VALUES(request_count)), 2),
			latency_p95_ms = ROUND(
				(latency_p95_ms * request_count + VALUES(latency_p95_ms) * VALUES(request_count))
				/ (request_count + VALUES(request_count)), 2),
			latency_p99_ms = ROUND(
				(latency_p99_ms * request_count + VALUES(latency_p99_ms) * VALUES(request_count))
				/ (request_count + VALUES(request_count)), 2)
	`)
	if err != nil {
		return fmt.Errorf("telemetry: prepare metrics upsert: %w", err)
	}
	defer stmt.Close()

	for _, b := range buckets {
		if _, err = stmt.Exec(
			b.ApiKeyID, b.ContractName, b.BucketMin,
			b.RequestCount, b.SuccessCount, b.ErrorCount,
			math.Round(b.LatencyP50*100)/100,
			math.Round(b.LatencyP95*100)/100,
			math.Round(b.LatencyP99*100)/100,
		); err != nil {
			return fmt.Errorf("telemetry: upsert metrics bucket: %w", err)
		}
	}
	return nil
}

func (s *MySQLStore) flushLoop() {
	ticker := time.NewTicker(time.Duration(s.cfg.FlushSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.Flush(); err != nil {
				log.Printf("[telemetry] flush error: %v", err)
			}
		case <-s.done:
			s.Flush()
			return
		}
	}
}

func (s *MySQLStore) Close() error {
	close(s.done)
	return nil
}

// truncateToMinute elimina los segundos y sub-segundos de t.
func truncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

// percentileInt devuelve el percentil p (0.0–1.0) de una slice de enteros YA ORDENADA.
func percentileInt(sorted []int, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return float64(sorted[idx])
}
