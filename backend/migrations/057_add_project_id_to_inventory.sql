-- Migration: Add project_id to inventory table
-- Date: 2025-11-24
-- Purpose: Track material usage per project

BEGIN;

ALTER TABLE inventory 
ADD COLUMN IF NOT EXISTS project_id BIGINT REFERENCES projects(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_inventory_project_id ON inventory(project_id);

COMMIT;
