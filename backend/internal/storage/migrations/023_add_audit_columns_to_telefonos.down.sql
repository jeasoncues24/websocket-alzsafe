ALTER TABLE telefonos DROP INDEX idx_telefonos_updated_by;
ALTER TABLE telefonos DROP INDEX idx_telefonos_created_by;
ALTER TABLE telefonos DROP COLUMN updated_by;
ALTER TABLE telefonos DROP COLUMN created_by;