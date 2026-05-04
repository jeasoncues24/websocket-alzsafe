-- 016: Seeds iniciales consolidados
INSERT INTO roles (name, description, is_root, permissions, created_by) VALUES
('super_admin', 'Super Administrador', TRUE, '["all"]', NULL),
('admin', 'Administrador', FALSE, '["users","companies","messages","broadcasts","sessions"]', NULL),
('operador', 'Operador', FALSE, '["messages","broadcasts"]', NULL),
('viewer', 'Visor', FALSE, '["messages:read","broadcasts:read"]', NULL);

INSERT INTO empresas (ruc, nombre, nombre_comercial, telefono_contacto, activo, created_by) VALUES
('20100000001', 'Empresa Demo S.A.C.', 'Demo Company', '+51 999 000 001', TRUE, NULL);

INSERT INTO admin_users (username, password_hash, email, role_id, activo, created_by) VALUES
('admin_usqay', '$2a$12$nchOPi3dzhpy6TCd5WwlHuArAjSvAY7N/0XFzapBIaZKpDT3tRgcG', 'admin@wsapi.com', 1, TRUE, NULL);

INSERT INTO modules (name, slug, description) VALUES
('dashboard', 'dashboard', 'Panel de control'),
('companies', 'companies', 'Gestión de empresas'),
('users', 'users', 'Gestión de usuarios'),
('roles', 'roles', 'Gestión de roles'),
('modules', 'modules', 'Gestión de módulos'),
('sessions', 'sessions', 'Sesiones WhatsApp'),
('messages', 'messages', 'Mensajes'),
('broadcasts', 'broadcasts', 'Difusiones');

INSERT INTO user_modules (user_id, module_id)
SELECT 1, id FROM modules;
