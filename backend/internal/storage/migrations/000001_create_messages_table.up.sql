-- 001: Messages table
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    empresa_id BIGINT NOT NULL DEFAULT 0,
    telefono_id BIGINT NOT NULL DEFAULT 0,
    destino VARCHAR(50) NOT NULL,
    contenido TEXT NOT NULL,
    adjuntos_json LONGTEXT NULL,
    estado VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_reason TEXT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP NULL,
    reference_id VARCHAR(100) UNIQUE,
    tiempo_envio TIMESTAMP NULL,
    timestamp_created TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    timestamp_sent TIMESTAMP NULL,
    timestamp_confirmed TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_messages_empresa (empresa_id),
    INDEX idx_messages_telefono (telefono_id),
    INDEX idx_messages_estado (estado),
    INDEX idx_messages_reference (reference_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
