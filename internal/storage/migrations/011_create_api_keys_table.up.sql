-- 011: API keys por teléfono
CREATE TABLE IF NOT EXISTS api_keys (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    empresa_id BIGINT NOT NULL,
    telefono_id BIGINT NOT NULL,
    nombre VARCHAR(120) NOT NULL,
    key_prefix VARCHAR(12) NOT NULL,
    secret_hash CHAR(64) NOT NULL,
    scopes JSON,
    activo BOOLEAN NOT NULL DEFAULT TRUE,
    created_by_user_id BIGINT NULL,
    last_used_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    revoked_at TIMESTAMP NULL,
    rotated_from_id BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_api_keys_secret_hash (secret_hash),
    UNIQUE KEY uq_api_keys_key_prefix (key_prefix),
    INDEX idx_api_keys_empresa (empresa_id),
    INDEX idx_api_keys_telefono (telefono_id),
    INDEX idx_api_keys_activo (activo),
    INDEX idx_api_keys_last_used (last_used_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
