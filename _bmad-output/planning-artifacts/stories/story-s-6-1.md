# Story S-6.1: Schema DB - Empresas y Teléfonos

## Epic
Epic 6: Sistema de Autenticación JWT por Empresa

## Prioridad
P0

## Estado
pending

## Overview

Crear tablas `empresas` y `telefonos` para soportar múltiples números por empresa.

## Acceptance Criteria

- [ ] Tabla `empresas` con `token_version` y `permissions`
- [ ] Tabla `telefonos` con FK a `empresa_id`
- [ ] Índices para lookup rápido

## SQL

```sql
CREATE TABLE empresas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ruc VARCHAR(11) UNIQUE NOT NULL,
    nombre VARCHAR(255) NOT NULL,
    token_version INT NOT NULL DEFAULT 1,
    api_key TEXT,
    permissions JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE telefonos (
    id INT AUTO_INCREMENT PRIMARY KEY,
    empresa_id INT NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    status ENUM('active', 'qr_pending', 'disconnected') DEFAULT 'disconnected',
    qr_string TEXT,
    session_data BLOB,
    last_connected TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (empresa_id) REFERENCES empresas(id)
);

CREATE INDEX idx_telefonos_empresa ON telefonos(empresa_id);
CREATE INDEX idx_telefonos_status ON telefonos(empresa_id, status);
```

## Dependencies
- Ninguna

## Estimated Effort
1 day