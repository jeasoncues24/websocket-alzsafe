-- 017: Align messages schema with storage repository contract
-- The repository expects adjuntos_json, error_reason and timestamp_* columns.

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS adjuntos_json LONGTEXT NULL AFTER contenido,
    ADD COLUMN IF NOT EXISTS error_reason TEXT NULL AFTER estado,
    ADD COLUMN IF NOT EXISTS timestamp_created TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP AFTER reference_id,
    ADD COLUMN IF NOT EXISTS timestamp_sent TIMESTAMP NULL AFTER timestamp_created,
    ADD COLUMN IF NOT EXISTS timestamp_confirmed TIMESTAMP NULL AFTER timestamp_sent;

-- Backfill timestamp_created for historical rows created with legacy schema.
UPDATE messages
SET timestamp_created = COALESCE(timestamp_created, tiempo_envio, created_at, NOW())
WHERE timestamp_created IS NULL;
