package storage

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"wsapi/internal/domain"
)

type TelemetryStore struct {
	db *sql.DB
}

func NewTelemetryStore(db *sql.DB) *TelemetryStore {
	return &TelemetryStore{db: db}
}

func (s *TelemetryStore) GetUsageStats(apiKeyID int64, desde, hasta time.Time) (*domain.TelemetryUsageStats, error) {
	if hasta.IsZero() {
		hasta = time.Now()
	}
	if desde.IsZero() {
		desde = hasta.AddDate(0, 0, -30)
	}

	var totalReq, totalErr int
	var avgLatency sql.NullFloat64
	var sumRequests, sumErrors int
	var peakReq int
	var peakHour int
	var peakDayStr sql.NullString
	var periodDays int

	err := s.db.QueryRow(`
		SELECT
			COALESCE(SUM(request_count), 0),
			COALESCE(SUM(error_count), 0),
			AVG(latency_p50_ms),
			COUNT(DISTINCT DATE(bucket_min))
		FROM telefono_metrics_min
		WHERE api_key_id = ? AND bucket_min >= ? AND bucket_min <= ?`,
		apiKeyID, desde, hasta,
	).Scan(&totalReq, &totalErr, &avgLatency, &periodDays)
	if err != nil {
		return nil, fmt.Errorf("error al obtener stats usage: %w", err)
	}

	_ = s.db.QueryRow(`
		SELECT hour_min, max_req FROM (
			SELECT HOUR(bucket_min) as hour_min, SUM(request_count) as max_req
			FROM telefono_metrics_min
			WHERE api_key_id = ? AND bucket_min >= ? AND bucket_min <= ?
			GROUP BY HOUR(bucket_min)
			ORDER BY max_req DESC LIMIT 1
		) t`, apiKeyID, desde, hasta,
	).Scan(&peakHour, &peakReq)

	_ = s.db.QueryRow(`
		SELECT DATE_FORMAT(bucket_min, '%Y-%m-%d') FROM (
			SELECT DATE(bucket_min) as d, SUM(request_count) as cnt
			FROM telefono_metrics_min
			WHERE api_key_id = ? AND bucket_min >= ? AND bucket_min <= ?
			GROUP BY DATE(bucket_min)
			ORDER BY cnt DESC LIMIT 1
		) t`, apiKeyID, desde, hasta,
	).Scan(&peakDayStr)

	if err := s.db.QueryRow(`
		SELECT COALESCE(SUM(request_count), 0), COALESCE(SUM(error_count), 0)
		FROM telefono_metrics_min
		WHERE api_key_id = ? AND bucket_min >= ? AND bucket_min <= ?
	`, apiKeyID, desde, hasta).Scan(&sumRequests, &sumErrors); err != nil {
		sumRequests, sumErrors = totalReq, totalErr
	}

	_ = s.db.QueryRow(`
		SELECT COALESCE(SUM(request_count), 0), COALESCE(SUM(error_count), 0)
		FROM telefono_metrics_min
		WHERE api_key_id = ? AND bucket_min >= ? AND bucket_min < ?
	`, apiKeyID, desde.AddDate(0, 0, -7), desde).Scan(&sumRequests, &sumErrors)

	_ = s.db.QueryRow(`
		SELECT COALESCE(SUM(request_count), 0), COALESCE(SUM(error_count), 0)
		FROM telefono_metrics_min
		WHERE api_key_id = ? AND bucket_min >= ? AND bucket_min < ?
	`, apiKeyID, desde.AddDate(0, 0, -14), desde.AddDate(0, 0, -7)).Scan(&sumRequests, &sumErrors)

	var errorRate float64
	if totalReq > 0 {
		errorRate = math.Round(float64(totalErr)/float64(totalReq)*10000) / 100
	}

	var p50, p95, p99 sql.NullFloat64
	s.db.QueryRow(`
		SELECT
			AVG(latency_p50_ms), AVG(latency_p95_ms), AVG(latency_p99_ms)
		FROM telefono_metrics_min
		WHERE api_key_id = ? AND bucket_min >= ? AND bucket_min <= ?
	`, apiKeyID, desde, hasta).Scan(&p50, &p95, &p99)

	trend := "stable"
	if periodDays >= 7 {
		var prevTotal, curTotal int
		mid := hasta.AddDate(0, 0, -periodDays/2)
		s.db.QueryRow(`SELECT COALESCE(SUM(request_count),0) FROM telefono_metrics_min WHERE api_key_id=? AND bucket_min>=? AND bucket_min<?`,
			apiKeyID, mid, hasta).Scan(&curTotal)
		s.db.QueryRow(`SELECT COALESCE(SUM(request_count),0) FROM telefono_metrics_min WHERE api_key_id=? AND bucket_min>=? AND bucket_min<?`,
			apiKeyID, desde, mid).Scan(&prevTotal)
		if prevTotal > 0 && curTotal > 0 {
			ratio := float64(curTotal) / float64(prevTotal)
			if ratio > 1.2 {
				trend = "up"
			} else if ratio < 0.8 {
				trend = "down"
			}
		}
	}

	var uptimeRatio float64
	if periodDays > 0 {
		uptimeRatio = math.Round(float64(periodDays-totalErr)/float64(periodDays)*100) / 100
	}

	stats := &domain.TelemetryUsageStats{
		ErrorRate:               errorRate,
		LatencyP50MS:            p50.Float64,
		LatencyP95MS:            p95.Float64,
		LatencyP99MS:            p99.Float64,
		TrendDirection:          trend,
		PeakHour:                peakHour,
		PeakDay:                 peakDayStr.String,
		MessagesPerRequestRatio: 0,
		UptimeRatio:             uptimeRatio,
		TotalRequests:           totalReq,
		TotalErrors:             totalErr,
		PeriodDays:              periodDays,
	}
	return stats, nil
}

func (s *TelemetryStore) GetUsageTimeSeries(apiKeyID int64, desde, hasta time.Time, granularidad string) ([]domain.TelemetryTimeSeriesPoint, error) {
	var dateFmt string
	switch granularidad {
	case "weekly":
		dateFmt = "%Y-%u"
	case "monthly":
		dateFmt = "%Y-%m"
	default:
		dateFmt = "%Y-%m-%d"
	}

	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT
			DATE_FORMAT(bucket_min, '%s') as bucket,
			SUM(request_count),
			SUM(success_count),
			SUM(error_count),
			AVG(latency_p50_ms),
			CASE WHEN SUM(request_count) > 0
				THEN ROUND(SUM(error_count) * 100.0 / SUM(request_count), 2)
				ELSE 0
			END
		FROM telefono_metrics_min
		WHERE api_key_id = ? AND bucket_min >= ? AND bucket_min <= ?
		GROUP BY bucket
		ORDER BY bucket`, dateFmt),
		apiKeyID, desde, hasta)
	if err != nil {
		return nil, fmt.Errorf("error al obtener time series: %w", err)
	}
	defer rows.Close()

	var points []domain.TelemetryTimeSeriesPoint
	for rows.Next() {
		var p domain.TelemetryTimeSeriesPoint
		var bucketStr string
		if err := rows.Scan(&bucketStr, &p.RequestCount, &p.SuccessCount, &p.ErrorCount,
			&p.LatencyAvgMS, &p.ErrorRate); err != nil {
			return nil, fmt.Errorf("error al escanear time series: %w", err)
		}
		if parsed, err := time.Parse("2006-01-02", bucketStr); err == nil {
			p.Bucket = parsed
		}
		points = append(points, p)
	}
	return points, nil
}

func (s *TelemetryStore) GetAuditStats(apiKeyID int64) (*domain.TelemetryAuditStats, error) {
	var totalKeys, totalRevoked int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM api_keys WHERE activo=1`).Scan(&totalKeys)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM api_keys WHERE activo=0`).Scan(&totalRevoked)

	var rotationsPerMonth float64
	_ = s.db.QueryRow(`
		SELECT COALESCE(COUNT(*) / NULLIF(TIMESTAMPDIFF(MONTH, MIN(created_at), NOW()), 0), 0)
		FROM api_key_audit_events WHERE api_key_id = ? AND action = 'rotated'
	`, apiKeyID).Scan(&rotationsPerMonth)

	var lastRotated sql.NullTime
	_ = s.db.QueryRow(`
		SELECT MAX(created_at) FROM api_key_audit_events
		WHERE api_key_id = ? AND action = 'rotated'`, apiKeyID).Scan(&lastRotated)
	var daysSince *int
	if lastRotated.Valid {
		d := int(time.Since(lastRotated.Time).Hours() / 24)
		daysSince = &d
	}

	var revRate float64
	if totalKeys > 0 {
		revRate = math.Round(float64(totalRevoked)/float64(totalKeys)*100) / 100
	}

	rows, err := s.db.Query(`
		SELECT COALESCE(actor_user_id, 0) as uid, COUNT(*) as cnt
		FROM api_key_audit_events WHERE api_key_id = ?
		GROUP BY actor_user_id ORDER BY cnt DESC LIMIT 5`, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener actores: %w", err)
	}
	defer rows.Close()

	var actors []domain.ActorActivity
	for rows.Next() {
		var a domain.ActorActivity
		if err := rows.Scan(&a.UserID, &a.Actions); err != nil {
			return nil, fmt.Errorf("error al escanear actor: %w", err)
		}
		actors = append(actors, a)
	}

	stats := &domain.TelemetryAuditStats{
		RotationsPerMonth:        rotationsPerMonth,
		TimeSinceLastRotationDays: daysSince,
		ActorDistribution:        actors,
		RevocationRate:           revRate,
		TotalKeys:                totalKeys,
		TotalRevoked:             totalRevoked,
	}
	return stats, nil
}
