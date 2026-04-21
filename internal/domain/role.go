package domain

import "time"

// Role representa un rol de usuario en el sistema
type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsRoot      bool      `json:"is_root"`
	Permissions []string  `json:"permissions,omitempty"`
	UsageCount  int       `json:"usage_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
