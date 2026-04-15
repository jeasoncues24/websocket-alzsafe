package domain

import "time"

// BroadcastItem represents a single recipient and message in a broadcast request.
type BroadcastItem struct {
	Destino string `json:"destino"`
	Mensaje string `json:"mensaje"`
}

// BroadcastRequest represents the HTTP POST /broadcast request payload.
type BroadcastRequest struct {
	RUCEmpresa    string          `json:"ruc_empresa"`
	ListaDifusion []BroadcastItem `json:"lista_difusion"`
}

// BroadcastResponse represents the HTTP response for a broadcast request.
type BroadcastResponse struct {
	OK          bool   `json:"ok"`
	ReferenceID string `json:"reference_id,omitempty"`
	Total       int    `json:"total,omitempty"`
	Error       string `json:"error,omitempty"`
	Details     string `json:"details,omitempty"`
}

// BroadcastStatus represents the status of a broadcast job.
type BroadcastStatus string

const (
	BroadcastStatusPending   BroadcastStatus = "pending"
	BroadcastStatusCompleted BroadcastStatus = "completed"
	BroadcastStatusFailed    BroadcastStatus = "failed"
)

// BroadcastResult represents the result of processing a single recipient.
type BroadcastResult struct {
	Index     int       `json:"index"`
	Destino   string    `json:"destino"`
	State     string    `json:"state"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// BroadcastJob represents a broadcast job with its results.
type BroadcastJob struct {
	ReferenceID string            `json:"reference_id"`
	RUCEmpresa  string            `json:"ruc_empresa"`
	Total       int               `json:"total"`
	Status      BroadcastStatus   `json:"status"`
	Results     []BroadcastResult `json:"results,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// BroadcastDetailResponse represents the response for GET /broadcast/{reference_id}
type BroadcastDetailResponse struct {
	OK          bool              `json:"ok"`
	ReferenceID string            `json:"reference_id"`
	RUCEmpresa  string            `json:"ruc_empresa"`
	Total       int               `json:"total"`
	Status      string            `json:"status"`
	Results     []BroadcastResult `json:"results,omitempty"`
	Error       string            `json:"error,omitempty"`
}
