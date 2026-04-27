-- 018: Remove retry tracking fields from messages table

ALTER TABLE messages
    DROP COLUMN IF EXISTS retry_count,
    DROP COLUMN IF EXISTS last_attempt_at;