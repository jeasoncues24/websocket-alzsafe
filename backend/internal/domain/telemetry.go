package domain

import "time"

type TelemetryEvent struct {
	ApiKeyID     int64     `json:"api_key_id"`
	EmpresaID    int64     `json:"empresa_id"`
	TelefonoID   int64     `json:"telefono_id"`
	ContractName string    `json:"contract_name"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	LatencyMS    int       `json:"latency_ms"`
	ErrorCode    string    `json:"error_code,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type TelemetryTimeSeriesPoint struct {
	Bucket       time.Time `json:"bucket"`
	RequestCount int       `json:"request_count"`
	SuccessCount int       `json:"success_count"`
	ErrorCount   int       `json:"error_count"`
	LatencyAvgMS float64   `json:"latency_avg_ms"`
	ErrorRate    float64   `json:"error_rate"`
}

type TelemetryUsageStats struct {
	ErrorRate                float64  `json:"error_rate"`
	LatencyP50MS             float64  `json:"latency_p50_ms"`
	LatencyP95MS             float64  `json:"latency_p95_ms"`
	LatencyP99MS             float64  `json:"latency_p99_ms"`
	TrendDirection           string   `json:"trend_direction"`
	PeakHour                 int      `json:"peak_hour"`
	PeakDay                  string   `json:"peak_day"`
	MessagesPerRequestRatio  float64  `json:"messages_per_request_ratio"`
	UptimeRatio              float64  `json:"uptime_ratio"`
	TotalRequests            int      `json:"total_requests"`
	TotalErrors              int      `json:"total_errors"`
	PeriodDays               int      `json:"period_days"`
}

type TelemetryAuditStats struct {
	RotationsPerMonth        float64                      `json:"rotations_per_month"`
	TimeSinceLastRotationDays *int                        `json:"time_since_last_rotation_days,omitempty"`
	ActorDistribution        []ActorActivity              `json:"actor_distribution"`
	RevocationRate           float64                      `json:"revocation_rate"`
	TotalKeys                int                          `json:"total_keys"`
	TotalRevoked             int                          `json:"total_revoked"`
}

type ActorActivity struct {
	UserID   int64  `json:"user_id"`
	Actions  int    `json:"actions"`
}

type TelemetryFilter struct {
	From        time.Time
	To          time.Time
	Granularity string
	Contract    string
}

type TelemetryMetricsResponse struct {
	OK     bool                   `json:"ok"`
	Stats  *TelemetryUsageStats   `json:"stats,omitempty"`
	Series []TelemetryTimeSeriesPoint `json:"series,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

type TelemetryAuditResponse struct {
	OK     bool                 `json:"ok"`
	Stats  *TelemetryAuditStats `json:"stats,omitempty"`
	Error  string               `json:"error,omitempty"`
}
