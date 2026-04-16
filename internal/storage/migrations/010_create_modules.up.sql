-- Migration 010: Create modules table
CREATE TABLE IF NOT EXISTS modules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255),
    slug VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed initial modules
INSERT INTO modules (name, description, slug) VALUES 
('Dashboard', 'Panel de métricas', 'dashboard'),
('Empresas', 'Gestión de empresas', 'companies'),
('Mensajes', 'Historial de mensajes', 'messages'),
('Sesiones', 'Gestión de sesiones WhatsApp', 'sessions'),
('Difusiones', 'Envío masivo de mensajes', 'broadcasts'),
('Configuración', 'Configuración del sistema', 'settings');