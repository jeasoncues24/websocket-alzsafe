-- 017: Webhooks outbound (configuración por api_key/teléfono)
-- Ownership grain: cada api_key (que ya está atada a un teléfono específico) puede
-- registrar su propio webhook. Esto refleja que el contrato B2B es por número de
-- WhatsApp, no por empresa. Una empresa con N teléfonos puede tener N webhooks
-- distintos, uno por integrador. `empresa_id` se mantiene para read-only desde
-- el panel admin (soporte) y para queries agregadas.
CREATE TABLE IF NOT EXISTS webhooks_outbound (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    empresa_id      BIGINT NOT NULL,
    telefono_id     BIGINT NOT NULL,
    api_key_id      BIGINT NOT NULL,
    url             TEXT NOT NULL,
    secret          VARCHAR(255) NOT NULL,
    eventos         JSON NOT NULL,
    activo          BOOLEAN NOT NULL DEFAULT TRUE,
    failure_count   INT DEFAULT 0,
    last_error      TEXT,
    last_success_at TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_webhooks_empresa (empresa_id, activo),
    INDEX idx_webhooks_empresa_created (empresa_id, created_at),
    INDEX idx_webhooks_telefono (telefono_id, activo),
    INDEX idx_webhooks_api_key (api_key_id, activo),
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE,
    FOREIGN KEY (telefono_id) REFERENCES telefonos(id) ON DELETE CASCADE,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
