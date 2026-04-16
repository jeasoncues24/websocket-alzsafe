-- Migration 013: Drop trigger for root protection
DROP TRIGGER IF EXISTS prevent_root_role_update;