-- 007: Roles table
CREATE TABLE IF NOT EXISTS roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(255),
    permissions JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default roles
INSERT INTO roles (name, description, permissions) VALUES
('super_admin', 'Super Administrador', '["all"]'),
('admin', 'Administrador', '["users","companies","messages","broadcasts","sessions"]'),
('operador', 'Operador', '["messages","broadcasts"]'),
('viewer', 'Visor', '["messages:read","broadcasts:read"]');