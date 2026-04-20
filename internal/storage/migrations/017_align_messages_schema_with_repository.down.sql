-- 017 rollback: remove columns introduced to align repository contract

ALTER TABLE messages
    DROP COLUMN IF EXISTS timestamp_confirmed,
    DROP COLUMN IF EXISTS timestamp_sent,
    DROP COLUMN IF EXISTS timestamp_created,
    DROP COLUMN IF EXISTS error_reason,
    DROP COLUMN IF EXISTS adjuntos_json;
