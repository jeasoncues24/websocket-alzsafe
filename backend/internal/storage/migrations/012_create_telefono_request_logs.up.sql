-- 012: Trazas individuales de request HTTP al service API
-- Patrón principal: WHERE api_key_id = ? AND created_at BETWEEN ? AND ?
-- Índice compuesto (api_key_id, created_at) cubre lectura y agrupación por contrato.
CREATE TABLE telefono_request_logs (
    id           BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
    api_key_id   BIGINT UNSIGNED   NOT NULL,
    empresa_id   BIGINT UNSIGNED   NOT NULL DEFAULT 0,
    telefono_id  BIGINT UNSIGNED   NOT NULL DEFAULT 0,
    contract_name VARCHAR(100)     NOT NULL DEFAULT '',
    endpoint     VARCHAR(255)      NOT NULL DEFAULT '',
    method       VARCHAR(10)       NOT NULL DEFAULT '',
    status_code  SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    latency_ms   INT UNSIGNED      NOT NULL DEFAULT 0,
    error_code   VARCHAR(50)       NULL,
    error_message TEXT             NULL,
    created_at   DATETIME(3)       NOT NULL,
    PRIMARY KEY (id),
    -- Índice principal: cubre la consulta de rango por api_key + fecha
    INDEX idx_trl_key_time (api_key_id, created_at),
    -- Índice para agrupar por contrato dentro de una key
    INDEX idx_trl_key_contract (api_key_id, contract_name, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
