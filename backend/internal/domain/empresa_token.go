package domain

import "context"

// EmpresaJWTClaims representa los claims del JWT de empresa (long-lived, 5 años)
// Campos mínimos: sub (empresa_id), ver (token_version), iss, iat, exp
// Scope y PhoneID son opcionales: zero values = token regular de empresa
type EmpresaJWTClaims struct {
	EmpresaID     int64    `json:"sub"`
	TokenVersion  int      `json:"ver"`
	EmpresaRUC    string   `json:"ruc"`
	EmpresaNombre string   `json:"nombre"`
	Permissions   []string `json:"permissions,omitempty"`
	Scope         string   `json:"scope,omitempty"`    // "qr_link" para tokens provisionales
	PhoneID       int64    `json:"phone_id,omitempty"` // teléfono restringido (solo QR link)
}

type empresaJWTClaimsKey struct{}

// WithEmpresaJWTClaims almacena EmpresaJWTClaims en el contexto.
func WithEmpresaJWTClaims(ctx context.Context, claims *EmpresaJWTClaims) context.Context {
	return context.WithValue(ctx, empresaJWTClaimsKey{}, claims)
}

// GetEmpresaJWTClaims recupera EmpresaJWTClaims del contexto.
func GetEmpresaJWTClaims(ctx context.Context) (*EmpresaJWTClaims, bool) {
	claims, ok := ctx.Value(empresaJWTClaimsKey{}).(*EmpresaJWTClaims)
	return claims, ok
}

// EmpresaJWTResponse es la respuesta del endpoint que genera el JWT de empresa.
type EmpresaJWTResponse struct {
	OK      bool   `json:"ok"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}
