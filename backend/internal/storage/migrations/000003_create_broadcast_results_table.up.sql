-- 003: Broadcast results table
CREATE TABLE IF NOT EXISTS broadcast_results (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    broadcast_id BIGINT NOT NULL,
    destino VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    sent_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_results_broadcast (broadcast_id),
    INDEX idx_results_destino (destino)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;