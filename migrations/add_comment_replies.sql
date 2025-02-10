-- Add parent_id column to comments table
ALTER TABLE comments ADD COLUMN parent_id INTEGER DEFAULT NULL REFERENCES comments(id) ON DELETE CASCADE;
