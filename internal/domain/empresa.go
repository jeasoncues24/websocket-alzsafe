package domain

import "time"

// Empresa representa una empresa en el sistema multi-tenant
type Empresa struct {
	ID              int64     `json:"id"`
	RUC             string    `json:"ruc"`
	Nombre          string    `json:"nombre"`
	NombreComercial string    `json:"nombre_comercial,omitempty"`
	Telefono        string    `json:"telefono,omitempty"`
	Direccion       string    `json:"direccion,omitempty"`
	Activo          bool      `json:"activo"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// EmpresaRequest representa el request HTTP para crear/actualizar empresa
type EmpresaRequest struct {
	RUC             string `json:"ruc"`
	Nombre          string `json:"nombre"`
	NombreComercial string `json:"nombre_comercial,omitempty"`
	Telefono        string `json:"telefono,omitempty"`
	Direccion       string `json:"direccion,omitempty"`
}

// EmpresaResponse representa la respuesta HTTP para operaciones de empresa
type EmpresaResponse struct {
	OK      bool     `json:"ok"`
	Empresa *Empresa `json:"empresa,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// EmpresasListResponse representa la respuesta con lista de empresas
type EmpresasListResponse struct {
	OK       bool      `json:"ok"`
	Empresas []Empresa `json:"empresas"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
	Error    string    `json:"error,omitempty"`
}

// NewEmpresa crea una nueva instancia de Empresa
func NewEmpresa(ruc, nombre string) *Empresa {
	return &Empresa{
		RUC:    ruc,
		Nombre: nombre,
		Activo: true,
	}
}
