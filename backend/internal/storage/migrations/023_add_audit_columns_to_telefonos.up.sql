-- 023: Agregar columnas de auditoría a telefonos
ALTER TABLE telefonos ADD COLUMN created_by BIGINT NULL AFTER last_connected;
ALTER TABLE telefonos ADD COLUMN updated_by BIGINT NULL AFTER created_by;

ALTER TABLE telefonos ADD INDEX idx_telefonos_created_by (created_by);
ALTER TABLE telefonos ADD INDEX idx_telefonos_updated_by (updated_by);