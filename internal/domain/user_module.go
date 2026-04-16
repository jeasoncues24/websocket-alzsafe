package domain

import "time"

// UserModule representa la relación usuario-módulo
type UserModule struct {
	UserID    int64     `json:"user_id"`
	ModuleID  int64     `json:"module_id"`
	CreatedAt time.Time `json:"created_at"`
}

// UserWithRole representa un usuario con su rol
type UserWithRole struct {
	AdminUser
	RoleID   *int64 `json:"role_id,omitempty"`
	RoleName string `json:"role_name,omitempty"`
	IsRoot   bool   `json:"is_root"`
}
