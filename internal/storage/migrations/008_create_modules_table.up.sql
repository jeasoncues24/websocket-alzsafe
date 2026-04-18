-- 008: Modules table
CREATE TABLE IF NOT EXISTS modules (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_modules_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default modules
INSERT INTO modules (name, description) VALUES
('dashboard', 'Panel de control'),
('companies', 'Gestión de empresas'),
('users', 'Gestión de usuarios'),
('roles', 'Gestión de roles'),
('modules', 'Gestión de módulos'),
('sessions', 'Sesiones WhatsApp'),
('messages', 'Mensajes'),
('broadcasts', 'Difusiones');