package domain

import "time"

// Module representa un módulo del sistema
type Module struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"created_at"`
}
