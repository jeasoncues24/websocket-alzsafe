-- Migration: 001_create_messages_table
-- Created: 2026-04-15

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    reference_id VARCHAR(255) NOT NULL UNIQUE,
    ruc_empresa VARCHAR(50) NOT NULL,
    destino VARCHAR(20) NOT NULL,
    mensaje TEXT NOT NULL,
    estado VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_messages_ruc_created (ruc_empresa, created_at DESC),
    INDEX idx_messages_estado (estado)
);