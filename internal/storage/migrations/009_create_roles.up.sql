-- Migration 009: Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255),
    is_root BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_is_root (is_root)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed initial roles
INSERT INTO roles (name, description, is_root) VALUES 
('root', 'Superadministrador con acceso total', TRUE),
('admin', 'Administrador con acceso completo', FALSE),
('operador', 'Operador con acceso limitado', FALSE),
('viewer', 'Solo visualización', FALSE);