-- 015: Remove missing columns from modules and roles tables

ALTER TABLE modules DROP COLUMN slug;
ALTER TABLE roles DROP COLUMN is_root;
