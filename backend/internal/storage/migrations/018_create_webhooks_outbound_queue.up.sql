-- 018: Cola de eventos para webhooks outbound
CREATE TABLE IF NOT EXISTS webhooks_outbound_queue (
    id                  BIGINT AUTO_INCREMENT PRIMARY KEY,
    webhook_id          BIGINT NOT NULL,
    payload             JSON NOT NULL,
    intentos            INT DEFAULT 0,
    proximo_intento_at  TIMESTAMP NOT NULL,
    estado              ENUM('pending','sending','done','failed') DEFAULT 'pending',
    last_error          TEXT,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_queue_due (estado, proximo_intento_at),
    FOREIGN KEY (webhook_id) REFERENCES webhooks_outbound(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
