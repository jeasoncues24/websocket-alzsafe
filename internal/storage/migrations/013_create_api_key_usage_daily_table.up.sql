-- 013: API key usage daily rollup
CREATE TABLE IF NOT EXISTS api_key_usage_daily (
    day DATE NOT NULL,
    api_key_id BIGINT NOT NULL,
    empresa_id BIGINT NOT NULL,
    telefono_id BIGINT NOT NULL,
    request_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    error_count INT NOT NULL DEFAULT 0,
    latency_avg_ms INT NOT NULL DEFAULT 0,
    messages_sent INT NOT NULL DEFAULT 0,
    broadcasts_sent INT NOT NULL DEFAULT 0,
    bytes_in BIGINT NOT NULL DEFAULT 0,
    bytes_out BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (day, api_key_id),
    INDEX idx_api_key_usage_daily_empresa (empresa_id),
    INDEX idx_api_key_usage_daily_telefono (telefono_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
