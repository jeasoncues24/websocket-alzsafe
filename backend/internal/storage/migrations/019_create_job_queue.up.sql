-- 019: Cola genérica de jobs reutilizable (broadcast, mensajes programados, etc.)
CREATE TABLE IF NOT EXISTS job_queue (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    type            VARCHAR(50)  NOT NULL,
    entity_id       VARCHAR(100) NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    priority        TINYINT      NOT NULL DEFAULT 5,
    empresa_id      BIGINT       NOT NULL,
    attempt_count   INT          NOT NULL DEFAULT 0,
    max_attempts    INT          NOT NULL DEFAULT 3,
    last_heartbeat  TIMESTAMP    NULL,
    next_retry_at   TIMESTAMP    NULL,
    metadata        JSON         NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at      TIMESTAMP    NULL,
    completed_at    TIMESTAMP    NULL,

    INDEX idx_empresa_status      (empresa_id, status),
    INDEX idx_type_status_priority (type, status, priority, created_at),
    INDEX idx_entity              (entity_id),
    INDEX idx_heartbeat_running   (status, last_heartbeat)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Items individuales por job (destinatarios de broadcast, mensajes programados, etc.)
CREATE TABLE IF NOT EXISTS job_items (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    job_id          BIGINT       NOT NULL,
    sequence_order  INT UNSIGNED NOT NULL,
    payload         JSON         NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    attempt_count   INT          NOT NULL DEFAULT 0,
    error_text      TEXT         NULL,
    processed_at    TIMESTAMP    NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_job_status_seq      (job_id, status, sequence_order),
    CONSTRAINT fk_job_items_job   FOREIGN KEY (job_id) REFERENCES job_queue(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
