-- 022: Agregar columnas de auditoría a empresas
ALTER TABLE empresas ADD COLUMN created_by BIGINT NULL AFTER activo;
ALTER TABLE empresas ADD COLUMN updated_by BIGINT NULL AFTER created_by;

ALTER TABLE empresas ADD INDEX idx_empresas_created_by (created_by);
ALTER TABLE empresas ADD INDEX idx_empresas_updated_by (updated_by);