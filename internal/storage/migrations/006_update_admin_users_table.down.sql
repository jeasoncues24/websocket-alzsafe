-- Migration 006: Revert admin_users changes
ALTER TABLE admin_users DROP FOREIGN KEY empresa_ibfk_1;
ALTER TABLE admin_users DROP COLUMN empresa_id;
ALTER TABLE admin_users MODIFY COLUMN role ENUM('admin','operator','viewer') NOT NULL DEFAULT 'operator';
DROP INDEX idx_empresa_id ON admin_users;