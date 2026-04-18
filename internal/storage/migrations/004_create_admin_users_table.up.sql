-- 004: Admin users table (without FK - added later)
CREATE TABLE IF NOT EXISTS admin_users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    empresa_id BIGINT,
    rol VARCHAR(20) NOT NULL DEFAULT 'operador',
    role_id BIGINT,
    is_root BOOLEAN NOT NULL DEFAULT FALSE,
    activo BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,
    INDEX idx_admin_users_username (username),
    INDEX idx_admin_users_empresa (empresa_id),
    INDEX idx_admin_users_rol (rol),
    INDEX idx_admin_users_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO admin_users (username, password_hash, email, rol, role_id, is_root, activo) VALUES
('admin', '$2a$10$rZ5qMNwXrK.iNKhVbXbJme5vJ.3KQxVqYxW8Jq0XwH9Y3aZ8Q3b6C', 'admin@wsapi.com', 'super_admin', 1, TRUE, TRUE);