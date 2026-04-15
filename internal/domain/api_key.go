package domain

import "time"

// ApiKey representa una API key de empresa
type ApiKey struct {
	ID        int64      `json:"id"`
	EmpresaID int64      `json:"empresa_id"`
	KeyHash   string     `json:"-"`
	Nombre    string     `json:"nombre,omitempty"`
	Activo    bool       `json:"activo"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// NewApiKey crea una nueva instancia de ApiKey
func NewApiKey(empresaID int64, keyHash string) *ApiKey {
	return &ApiKey{
		EmpresaID: empresaID,
		KeyHash:   keyHash,
		Activo:    true,
	}
}

// IsExpired verifica si la API key ha expirado
func (a *ApiKey) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// IsValid verifica si la API key está activa y no ha expirado
func (a *ApiKey) IsValid() bool {
	return a.Activo && !a.IsExpired()
}
