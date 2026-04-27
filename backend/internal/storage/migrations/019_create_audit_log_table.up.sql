-- 019: Auditoría genérica unificada
CREATE TABLE IF NOT EXISTS audit_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    entity_type VARCHAR(40) NOT NULL COMMENT 'empresa, telefono, rol, api_key, admin_user, user_module',
    entity_id BIGINT NOT NULL COMMENT 'ID del registro afectado',
    action VARCHAR(20) NOT NULL COMMENT 'created, updated, deleted, disabled, enabled, revoked, rotated, assigned, removed',
    actor_user_id BIGINT NULL COMMENT 'Quién realizó la acción (NULL = sistema)',
    changes JSON NULL COMMENT 'Campos antes/después',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_audit_entity (entity_type, entity_id),
    INDEX idx_audit_actor (actor_user_id),
    INDEX idx_audit_action (action),
    INDEX idx_audit_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;