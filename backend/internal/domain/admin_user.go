package domain

import (
	"context"
	"time"
)

type contextKey string

const adminJWTClaimsKey contextKey = "admin_jwt_claims"

// WithAdminJWTClaims stores AdminJWTClaims in context
func WithAdminJWTClaims(ctx context.Context, claims *AdminJWTClaims) context.Context {
	return context.WithValue(ctx, adminJWTClaimsKey, claims)
}

// GetAdminJWTClaims retrieves AdminJWTClaims from context
func GetAdminJWTClaims(ctx context.Context) (*AdminJWTClaims, bool) {
	claims, ok := ctx.Value(adminJWTClaimsKey).(*AdminJWTClaims)
	return claims, ok
}

// AdminUser representa un usuario administrador del sistema
// IsRoot se deriva del rol via JOIN con tabla roles
type AdminUser struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Email        string     `json:"email,omitempty"`
	RoleID       *int64     `json:"role_id,omitempty"`
	RoleName     UserRole   `json:"rol,omitempty"`
	IsRoot       bool       `json:"is_root"`
	Activo       bool       `json:"activo"`
	CreatedBy    *int64     `json:"created_by,omitempty"`
	UpdatedBy    *int64     `json:"updated_by,omitempty"`
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

// UserRole representa el rol del usuario (solo para JWT legacy)
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleOperador   UserRole = "operador"
	RoleViewer     UserRole = "viewer"
)

// AdminJWTClaims representa los claims del JWT
type AdminJWTClaims struct {
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
func NewAdminUser(username, passwordHash string, roleID *int64) *AdminUser {
	return &AdminUser{
		Username:     username,
		PasswordHash: passwordHash,
		RoleID:       roleID,
		Activo:       true,
	}
}
