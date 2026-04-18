package domain

import "time"

// TelefonoStatus representa el estado de conexión de un teléfono WhatsApp
type TelefonoStatus string

const (
	TelefonoStatusActive       TelefonoStatus = "active"
	TelefonoStatusQRPending    TelefonoStatus = "qr_pending"
	TelefonoStatusDisconnected TelefonoStatus = "disconnected"
)

func (s TelefonoStatus) String() string {
	return string(s)
}

// Telefono representa un número WhatsApp asociado a una empresa
type Telefono struct {
	ID             int64          `json:"id"`
	EmpresaID      int64          `json:"empresa_id"`
	CodigoPais     string         `json:"codigo_pais"`
	Numero         string         `json:"numero"`
	NumeroCompleto string         `json:"numero_completo"`
	Status         TelefonoStatus `json:"status"`
	SessionData    []byte         `json:"-"`
	QRString       string         `json:"qr_string,omitempty"`
	LastConnected  *time.Time     `json:"last_connected,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// TelefonoResponse representa la respuesta HTTP para operaciones de teléfono
type TelefonoResponse struct {
	OK       bool      `json:"ok"`
	Telefono *Telefono `json:"telefono,omitempty"`
	Error    string    `json:"error,omitempty"`
}

// TelefonosListResponse representa la respuesta con lista de teléfonos
type TelefonosListResponse struct {
	OK        bool       `json:"ok"`
	Telefonos []Telefono `json:"telefonos"`
	Total     int        `json:"total"`
	Error     string     `json:"error,omitempty"`
}
