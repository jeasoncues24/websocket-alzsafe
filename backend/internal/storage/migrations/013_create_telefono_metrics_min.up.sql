-- 013: Agregados por minuto del service API (dashboard de métricas)
-- Patrón upsert: INSERT ... ON DUPLICATE KEY UPDATE por (api_key_id, contract_name, bucket_min)
-- Patrón lectura: WHERE api_key_id = ? AND bucket_min BETWEEN ? AND ? GROUP BY DATE_FORMAT(...)
-- El UNIQUE KEY sirve como índice de cobertura para ambos patrones.
CREATE TABLE telefono_metrics_min (
    id            BIGINT UNSIGNED    NOT NULL AUTO_INCREMENT,
    api_key_id    BIGINT UNSIGNED    NOT NULL,
    contract_name VARCHAR(100)       NOT NULL DEFAULT '',
    bucket_min    DATETIME           NOT NULL,
    request_count INT UNSIGNED       NOT NULL DEFAULT 0,
    success_count INT UNSIGNED       NOT NULL DEFAULT 0,
    error_count   INT UNSIGNED       NOT NULL DEFAULT 0,
    latency_p50_ms DECIMAL(10,2)     NOT NULL DEFAULT 0.00,
    latency_p95_ms DECIMAL(10,2)     NOT NULL DEFAULT 0.00,
    latency_p99_ms DECIMAL(10,2)     NOT NULL DEFAULT 0.00,
    messages_sent INT UNSIGNED       NOT NULL DEFAULT 0,
    created_at    DATETIME(3)        NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    -- UNIQUE KEY habilita ON DUPLICATE KEY UPDATE y cubre lecturas por api_key + rango de tiempo
    UNIQUE KEY uq_tmm_bucket (api_key_id, contract_name, bucket_min)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
