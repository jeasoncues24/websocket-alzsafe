-- 002: Broadcasts table
CREATE TABLE IF NOT EXISTS broadcasts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    empresa_id BIGINT NOT NULL DEFAULT 0,
    telefono_id BIGINT NOT NULL DEFAULT 0,
    reference_id VARCHAR(100) UNIQUE NOT NULL,
    total INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_broadcasts_empresa (empresa_id),
    INDEX idx_broadcasts_telefono (telefono_id),
    INDEX idx_broadcasts_status (status),
    INDEX idx_broadcasts_reference (reference_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;