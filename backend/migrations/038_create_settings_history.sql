-- Migration: Create settings_history table for audit tracking
-- Description: Creates settings_history table to track changes to system settings
-- Date: 2025-11-04

BEGIN;

-- Create settings_history table
CREATE TABLE IF NOT EXISTS settings_history (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    -- Reference to settings
    settings_id BIGINT NOT NULL,
    
    -- Change tracking
    field VARCHAR(255) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    action VARCHAR(50) DEFAULT 'UPDATE',
    
    -- User tracking
    changed_by BIGINT NOT NULL,
    
    -- Additional context
    ip_address VARCHAR(255),
    user_agent TEXT,
    reason TEXT,
    
    -- Foreign keys
    CONSTRAINT fk_settings_history_settings 
        FOREIGN KEY (settings_id) REFERENCES settings(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_settings_history_settings_id ON settings_history(settings_id);
CREATE INDEX IF NOT EXISTS idx_settings_history_changed_by ON settings_history(changed_by);
CREATE INDEX IF NOT EXISTS idx_settings_history_field ON settings_history(field);
CREATE INDEX IF NOT EXISTS idx_settings_history_created_at ON settings_history(created_at);
CREATE INDEX IF NOT EXISTS idx_settings_history_deleted_at ON settings_history(deleted_at);

-- Add comments for documentation
COMMENT ON TABLE settings_history IS 'Audit log for tracking changes to system settings';
COMMENT ON COLUMN settings_history.settings_id IS 'Reference to the settings record';
COMMENT ON COLUMN settings_history.field IS 'Name of the field that was changed';
COMMENT ON COLUMN settings_history.old_value IS 'Previous value (JSON string)';
COMMENT ON COLUMN settings_history.new_value IS 'New value (JSON string)';
COMMENT ON COLUMN settings_history.action IS 'Type of action: UPDATE, RESET, CREATE';
COMMENT ON COLUMN settings_history.changed_by IS 'User who made the change';
COMMENT ON COLUMN settings_history.ip_address IS 'IP address of the user';
COMMENT ON COLUMN settings_history.user_agent IS 'Browser user agent string';
COMMENT ON COLUMN settings_history.reason IS 'Optional reason for the change';

COMMIT;
