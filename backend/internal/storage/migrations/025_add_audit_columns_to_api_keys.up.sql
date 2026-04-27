-- 025: Agregar columnas de auditoría a api_keys
ALTER TABLE api_keys ADD COLUMN updated_by BIGINT NULL AFTER revoked_at;

ALTER TABLE api_keys ADD INDEX idx_api_keys_updated_by (updated_by);