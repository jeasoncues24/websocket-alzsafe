ALTER TABLE roles DROP INDEX idx_roles_updated_by;
ALTER TABLE roles DROP INDEX idx_roles_created_by;
ALTER TABLE roles DROP COLUMN updated_by;
ALTER TABLE roles DROP COLUMN created_by;