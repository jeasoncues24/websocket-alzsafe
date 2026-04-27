-- 005: Empresas table (multi-tenant)
CREATE TABLE IF NOT EXISTS empresas (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ruc VARCHAR(20) UNIQUE NOT NULL,
    nombre VARCHAR(255) NOT NULL,
    nombre_comercial VARCHAR(255),
    telefono VARCHAR(30),
    direccion VARCHAR(500),
    token_version INT NOT NULL DEFAULT 1,
    permissions JSON,
    activo BOOLEAN NOT NULL DEFAULT TRUE,
    created_by BIGINT NULL,
    updated_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_empresas_ruc (ruc),
    INDEX idx_empresas_nombre (nombre),
    INDEX idx_empresas_activo (activo),
    INDEX idx_empresas_created_by (created_by),
    INDEX idx_empresas_updated_by (updated_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample empresa
INSERT INTO empresas (ruc, nombre, nombre_comercial, telefono, activo, created_by) VALUES
('20100000001', 'Empresa Demo S.A.C.', 'Demo Company', '+51 999 000 001', TRUE, NULL);