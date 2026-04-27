-- 024: Agregar columnas de auditoría a roles
ALTER TABLE roles ADD COLUMN created_by BIGINT NULL AFTER is_root;
ALTER TABLE roles ADD COLUMN updated_by BIGINT NULL AFTER created_by;

ALTER TABLE roles ADD INDEX idx_roles_created_by (created_by);
ALTER TABLE roles ADD INDEX idx_roles_updated_by (updated_by);