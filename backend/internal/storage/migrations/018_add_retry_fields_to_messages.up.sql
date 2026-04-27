-- 018: Add retry tracking fields to messages table
-- Supports retry functionality for failed messages

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0 AFTER error_reason,
    ADD COLUMN IF NOT EXISTS last_attempt_at TIMESTAMP NULL AFTER retry_count;
