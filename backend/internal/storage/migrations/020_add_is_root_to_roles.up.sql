-- 020: Agregar is_root a roles
ALTER TABLE roles ADD COLUMN is_root BOOLEAN NOT NULL DEFAULT FALSE AFTER description;

UPDATE roles SET is_root = TRUE WHERE name = 'super_admin';

ALTER TABLE roles ADD INDEX idx_roles_is_root (is_root);