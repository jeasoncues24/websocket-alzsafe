# Story S-6.2: Modelo Empresa + Teléfono

## Epic
Epic 6: Sistema de Autenticación JWT por Empresa

## Prioridad
P0

## Estado
pending

## Overview

Crear modelos `Empresa` y `Telefono` en `internal/domain/`.

## Acceptance Criteria

- [ ] Struct `Empresa` con campos: ID, RUC, Nombre, TokenVersion, Permissions
- [ ] Struct `Telefono` con campos: ID, EmpresaID, PhoneNumber, Status, QRSession, SessionData
- [ ] Métodos: IncrementTokenVersion(), GetTelefonos()
- [ ] Tests unitarios pasando

## Código

```go
// internal/domain/empresa.go
type Empresa struct {
    ID            int64     `json:"id"`
    RUC           string   `json:"ruc"`
    Nombre        string   `json:"nombre"`
    TokenVersion  int      `json:"token_version"`
    APIKey       string   `json:"-"` // No mostrar en JSON
    Permissions  []string `json:"permissions"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

func (e *Empresa) IncrementTokenVersion() error {
    // Atomic increment + cache invalidation
}

func (e *Empresa) GetTelefonos() ([]*Telefono, error) {
    // Fetch telefonos from DB
}

func (e *Empresa) HasPermission(perm string) bool {
    for _, p := range e.Permissions {
        if p == perm {
            return true
        }
    }
    return false
}

// internal/domain/telefono.go
type Telefono struct {
    ID           int64     `json:"id"`
    EmpresaID    int64     `json:"empresa_id"`
    PhoneNumber  string   `json:"phone_number"`
    Status       string   `json:"status"` // active, qr_pending, disconnected
    QRSession    string   `json:"qr_string,omitempty"`
    SessionData  []byte   `json:"-"`
    LastConnected time.Time `json:"last_connected"`
}

func (t *Telefono) IsConnected() bool {
    return t.Status == "active"
}
```

## Dependencies
- S-6.1 (schema DB)

## Estimated Effort
2 days