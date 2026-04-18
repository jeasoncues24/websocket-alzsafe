package domain

import (
	"encoding/json"
	"time"
)

// ApiKey representa una API key de consumo asociada a un teléfono WhatsApp.
type ApiKey struct {
	ID              int64      `json:"id"`
	EmpresaID       int64      `json:"empresa_id"`
	TelefonoID      int64      `json:"telefono_id"`
	Nombre          string     `json:"nombre,omitempty"`
	KeyPrefix       string     `json:"key_prefix"`
	SecretHash      string     `json:"-"`
	Scopes          []string   `json:"scopes,omitempty"`
	Activo          bool       `json:"activo"`
	CreatedByUserID *int64     `json:"created_by_user_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	RotatedFromID   *int64     `json:"rotated_from_id,omitempty"`
}

// ApiKeyCreateResponse devuelve la key creada y el secreto visible una sola vez.
type ApiKeyCreateResponse struct {
	OK      bool    `json:"ok"`
	ApiKey  *ApiKey `json:"api_key,omitempty"`
	Secret  string  `json:"secret,omitempty"`
	Message string  `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
}

// ApiKeyResponse devuelve una key sin exponer el secreto.
type ApiKeyResponse struct {
	OK     bool    `json:"ok"`
	ApiKey *ApiKey `json:"api_key,omitempty"`
	Error  string  `json:"error,omitempty"`
}

// ApiKeyListResponse agrupa múltiples keys.
type ApiKeyListResponse struct {
	OK      bool     `json:"ok"`
	ApiKeys []ApiKey `json:"api_keys"`
	Error   string   `json:"error,omitempty"`
}

// ApiKeyUsageDaily representa un rollup diario de uso.
type ApiKeyUsageDaily struct {
	Day            string `json:"day"`
	ApiKeyID       int64  `json:"api_key_id"`
	EmpresaID      int64  `json:"empresa_id"`
	TelefonoID     int64  `json:"telefono_id"`
	RequestCount   int    `json:"request_count"`
	SuccessCount   int    `json:"success_count"`
	ErrorCount     int    `json:"error_count"`
	LatencyAvgMS   int    `json:"latency_avg_ms"`
	MessagesSent   int    `json:"messages_sent"`
	BroadcastsSent int    `json:"broadcasts_sent"`
	BytesIn        int64  `json:"bytes_in"`
	BytesOut       int64  `json:"bytes_out"`
}

// ApiKeyUsageEvent representa una traza de uso por request.
type ApiKeyUsageEvent struct {
	ID            int64     `json:"id"`
	ApiKeyID      int64     `json:"api_key_id"`
	EmpresaID     int64     `json:"empresa_id"`
	TelefonoID    int64     `json:"telefono_id"`
	Method        string    `json:"method"`
	Endpoint      string    `json:"endpoint"`
	StatusCode    int       `json:"status_code"`
	LatencyMS     int       `json:"latency_ms"`
	RequestUnits  int       `json:"request_units"`
	ResponseUnits int       `json:"response_units"`
	RequestID     string    `json:"request_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// ApiKeyAuditEvent representa una acción auditable sobre una key.
type ApiKeyAuditEvent struct {
	ID          int64           `json:"id"`
	ApiKeyID    int64           `json:"api_key_id"`
	EmpresaID   int64           `json:"empresa_id"`
	TelefonoID  int64           `json:"telefono_id"`
	Action      string          `json:"action"`
	ActorUserID *int64          `json:"actor_user_id,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// NewApiKey crea una nueva instancia de ApiKey.
func NewApiKey(empresaID, telefonoID int64, secretHash string) *ApiKey {
	return &ApiKey{
		EmpresaID:  empresaID,
		TelefonoID: telefonoID,
		SecretHash: secretHash,
		Activo:     true,
	}
}

// IsExpired verifica si la API key ha expirado.
func (a *ApiKey) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// IsValid verifica si la API key está activa y no ha expirado.
func (a *ApiKey) IsValid() bool {
	return a.Activo && a.RevokedAt == nil && !a.IsExpired()
}
