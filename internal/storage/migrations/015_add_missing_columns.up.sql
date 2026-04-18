-- 015: Add missing columns to modules and roles tables
-- Fix: code expects 'slug' in modules and 'is_root' in roles

-- Add slug column to modules table when missing
ALTER TABLE modules ADD COLUMN IF NOT EXISTS slug VARCHAR(50) UNIQUE AFTER name;

-- Update existing modules with slug values
UPDATE modules SET slug = LOWER(name) WHERE slug IS NULL;

-- Add is_root column to roles table when missing
ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_root BOOLEAN DEFAULT FALSE AFTER description;

-- Mark super_admin as root role
UPDATE roles SET is_root = TRUE WHERE name = 'super_admin';
