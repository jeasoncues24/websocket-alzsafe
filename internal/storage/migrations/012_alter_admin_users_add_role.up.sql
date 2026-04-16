-- Migration 012: Add role_id and is_root to admin_users
ALTER TABLE admin_users ADD COLUMN role_id INT NULL;
ALTER TABLE admin_users ADD COLUMN is_root BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE admin_users ADD CONSTRAINT fk_admin_users_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL;
CREATE INDEX idx_role_id ON admin_users(role_id);
CREATE INDEX idx_is_root ON admin_users(is_root);