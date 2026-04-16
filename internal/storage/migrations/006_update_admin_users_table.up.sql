-- Migration 006: Update admin_users and add empresa relationship
ALTER TABLE admin_users ADD COLUMN empresa_id INT NULL;
CREATE INDEX idx_empresa_id ON admin_users(empresa_id);
ALTER TABLE admin_users ADD FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE SET NULL;
ALTER TABLE admin_users MODIFY COLUMN role ENUM('super_admin','admin','operador','viewer') NOT NULL DEFAULT 'operador';