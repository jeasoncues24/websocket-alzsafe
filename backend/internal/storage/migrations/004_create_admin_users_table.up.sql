-- 004: Admin users table (refactorizado - sin rol, is_root, empresa_id)
-- IsRoot se obtiene via JOIN con roles.is_root
CREATE TABLE IF NOT EXISTS admin_users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    role_id BIGINT,
    activo BOOLEAN NOT NULL DEFAULT TRUE,
    created_by BIGINT NULL,
    updated_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,
    INDEX idx_admin_users_username (username),
    INDEX idx_admin_users_role (role_id),
    INDEX idx_admin_users_created_by (created_by),
    INDEX idx_admin_users_updated_by (updated_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- El primer usuario se inserta sin created_by (sistema)
-- Asumiendo que rol super_admin tiene id=1
INSERT INTO admin_users (username, password_hash, email, role_id, activo, created_by) VALUES
('admin_usqay', '$2a$12$nchOPi3dzhpy6TCd5WwlHuArAjSvAY7N/0XFzapBIaZKpDT3tRgcG', 'admin@wsapi.com', 1, TRUE, NULL);