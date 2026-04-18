-- 014: API key audit trail
CREATE TABLE IF NOT EXISTS api_key_audit_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    api_key_id BIGINT NOT NULL,
    empresa_id BIGINT NOT NULL,
    telefono_id BIGINT NOT NULL,
    action VARCHAR(40) NOT NULL,
    actor_user_id BIGINT NULL,
    metadata JSON NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_api_key_audit_key (api_key_id),
    INDEX idx_api_key_audit_empresa (empresa_id),
    INDEX idx_api_key_audit_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
