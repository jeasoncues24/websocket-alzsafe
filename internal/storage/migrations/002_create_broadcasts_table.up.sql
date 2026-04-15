-- Migration: 002_create_broadcasts_table
-- Created: 2026-04-15

CREATE TABLE IF NOT EXISTS broadcasts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    reference_id VARCHAR(255) NOT NULL UNIQUE,
    ruc_empresa VARCHAR(50) NOT NULL,
    total INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_broadcasts_ruc_status (ruc_empresa, status)
);