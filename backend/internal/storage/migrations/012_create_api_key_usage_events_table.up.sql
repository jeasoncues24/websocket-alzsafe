-- 012: API key usage events
CREATE TABLE IF NOT EXISTS api_key_usage_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    api_key_id BIGINT NOT NULL,
    empresa_id BIGINT NOT NULL,
    telefono_id BIGINT NOT NULL,
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    status_code INT NOT NULL,
    latency_ms INT NOT NULL,
    request_units INT NOT NULL DEFAULT 1,
    response_units INT NOT NULL DEFAULT 0,
    request_id VARCHAR(64) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_api_key_usage_key (api_key_id),
    INDEX idx_api_key_usage_empresa (empresa_id),
    INDEX idx_api_key_usage_telefono (telefono_id),
    INDEX idx_api_key_usage_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
