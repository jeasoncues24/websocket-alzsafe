-- Migration: 003_create_broadcast_results_table
-- Created: 2026-04-15

CREATE TABLE IF NOT EXISTS broadcast_results (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    broadcast_id BIGINT NOT NULL,
    item_index INT NOT NULL,
    destino VARCHAR(20) NOT NULL,
    state VARCHAR(20) NOT NULL,
    error TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_broadcast_results_broadcast_id (broadcast_id),
    FOREIGN KEY (broadcast_id) REFERENCES broadcasts(id) ON DELETE CASCADE
);