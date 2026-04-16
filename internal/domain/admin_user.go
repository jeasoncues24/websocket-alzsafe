package domain

import (
	"context"
	"time"
)

// UserRole representa el rol del usuario admin
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleOperador   UserRole = "operador"
	RoleViewer     UserRole = "viewer"
)

type contextKey string

const tokenClaimsKey contextKey = "token_claims"

// WithTokenClaims stores TokenClaims in context
func WithTokenClaims(ctx context.Context, claims *TokenClaims) context.Context {
	return context.WithValue(ctx, tokenClaimsKey, claims)
}

// GetTokenClaims retrieves TokenClaims from context
func GetTokenClaims(ctx context.Context) (*TokenClaims, bool) {
	claims, ok := ctx.Value(tokenClaimsKey).(*TokenClaims)
	return claims, ok
}

// AdminUser representa un usuario administrador del sistema
type AdminUser struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Email        string     `json:"email,omitempty"`
	EmpresaID    *int64     `json:"empresa_id,omitempty"`
	Rol          UserRole   `json:"rol"`
	RoleID       *int64     `json:"role_id,omitempty"`
	IsRoot       bool       `json:"is_root"`
	Activo       bool       `json:"activo"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLogin    *time.Time `json:"last_login_at,omitempty"`
}

// LoginRequest representa el request de login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse representa la respuesta de login
type LoginResponse struct {
	OK      bool   `json:"ok"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// TokenClaims representa los claims del JWT
type TokenClaims struct {
	JTI           string   `json:"jti,omitempty"`
	UserID        int64    `json:"user_id"`
	Username      string   `json:"username"`
	Rol           UserRole `json:"rol"`
	IsRoot        bool     `json:"is_root"`
	EmpresaID     *int64   `json:"empresa_id,omitempty"`
	EmpresaRUC    *string  `json:"empresa_ruc,omitempty"`
	EmpresaNombre *string  `json:"empresa_nombre,omitempty"`
}

// NewAdminUser crea una nueva instancia de AdminUser
func NewAdminUser(username, passwordHash string, rol UserRole) *AdminUser {
	return &AdminUser{
		Username:     username,
		PasswordHash: passwordHash,
		Rol:          rol,
		Activo:       true,
	}
}
