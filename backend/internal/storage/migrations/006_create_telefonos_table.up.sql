-- 006: Telefonos table
CREATE TABLE IF NOT EXISTS telefonos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    empresa_id BIGINT NOT NULL,
    codigo_pais VARCHAR(5) NOT NULL DEFAULT '+51',
    numero VARCHAR(20) NOT NULL,
    numero_completo VARCHAR(30) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'disconnected',
    session_data LONGBLOB,
    qr_string TEXT,
    last_connected TIMESTAMP NULL,
    created_by BIGINT NULL,
    updated_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_telefonos_empresa (empresa_id),
    INDEX idx_telefonos_numero (numero),
    INDEX idx_telefonos_completo (numero_completo),
    INDEX idx_telefonos_status (status),
    INDEX idx_telefonos_created_by (created_by),
    INDEX idx_telefonos_updated_by (updated_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;