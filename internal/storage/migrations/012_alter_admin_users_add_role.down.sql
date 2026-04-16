-- Migration 012: Remove role_id and is_root from admin_users
ALTER TABLE admin_users DROP FOREIGN KEY fk_admin_users_role;
ALTER TABLE admin_users DROP COLUMN role_id;
ALTER TABLE admin_users DROP COLUMN is_root;
DROP INDEX idx_role_id ON admin_users;
DROP INDEX idx_is_root ON admin_users;