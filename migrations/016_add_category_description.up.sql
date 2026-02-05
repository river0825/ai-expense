-- Add description column to categories table
ALTER TABLE categories ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';
