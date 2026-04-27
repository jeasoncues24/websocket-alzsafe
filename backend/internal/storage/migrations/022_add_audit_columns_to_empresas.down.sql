ALTER TABLE empresas DROP INDEX idx_empresas_created_by;
ALTER TABLE empresas DROP INDEX idx_empresas_updated_by;
ALTER TABLE empresas DROP COLUMN updated_by;
ALTER TABLE empresas DROP COLUMN created_by;