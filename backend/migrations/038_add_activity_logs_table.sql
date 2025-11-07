-- Migration: Add activity_logs table for comprehensive user activity tracking
-- Purpose: Track all user activities including login, CRUD operations, and API calls
-- Author: System
-- Date: 2025-10-19

-- Create activity_logs table
CREATE TABLE IF NOT EXISTS activity_logs (
    id SERIAL PRIMARY KEY,
    
    -- User Information
    user_id INTEGER NOT NULL,
    username VARCHAR(50) NOT NULL,
    role VARCHAR(20) NOT NULL,
    
    -- Request Information
    method VARCHAR(10) NOT NULL, -- GET, POST, PUT, DELETE, etc.
    path VARCHAR(500) NOT NULL,
    action VARCHAR(100), -- login, create_product, update_sale, etc.
    resource VARCHAR(50), -- users, products, sales, etc.
    
    -- Request Details
    request_body TEXT,
    query_params TEXT,
    
    -- Response Information
    status_code INTEGER NOT NULL,
    response_body TEXT,
    
    -- Network Information
    ip_address VARCHAR(45), -- IPv4 or IPv6
    user_agent TEXT,
    
    -- Performance Metrics
    duration BIGINT,
    
    -- Additional Context
    description TEXT,
    metadata JSONB, -- Additional JSON data
    
    -- Error Tracking
    is_error BOOLEAN DEFAULT FALSE,
    error_message TEXT,
    
    -- Audit Trail
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Foreign Key Constraint (optional - comment out if users table doesn't exist)
    CONSTRAINT fk_activity_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_username ON activity_logs(username);
CREATE INDEX IF NOT EXISTS idx_activity_logs_role ON activity_logs(role);
CREATE INDEX IF NOT EXISTS idx_activity_logs_path ON activity_logs(path);
CREATE INDEX IF NOT EXISTS idx_activity_logs_resource ON activity_logs(resource);
CREATE INDEX IF NOT EXISTS idx_activity_logs_status_code ON activity_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_activity_logs_is_error ON activity_logs(is_error);
CREATE INDEX IF NOT EXISTS idx_activity_logs_ip_address ON activity_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_deleted_at ON activity_logs(deleted_at);

-- Create composite index for common query patterns
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_date ON activity_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_role_date ON activity_logs(role, created_at DESC);

-- Add comment to table
COMMENT ON TABLE activity_logs IS 'Comprehensive tracking of all user activities in the system';

-- Add comments to important columns
COMMENT ON COLUMN activity_logs.user_id IS 'ID of the user who performed the action';
COMMENT ON COLUMN activity_logs.action IS 'Type of action performed (e.g., login, create_product, update_sale)';
COMMENT ON COLUMN activity_logs.resource IS 'Resource being accessed (e.g., users, products, sales)';
COMMENT ON COLUMN activity_logs.duration IS 'Time taken to process the request in milliseconds';
COMMENT ON COLUMN activity_logs.metadata IS 'Additional contextual data stored as JSON';

-- Optional: Create a view for easy summary queries
CREATE OR REPLACE VIEW activity_logs_summary AS
SELECT 
    DATE(created_at) as activity_date,
    user_id,
    username,
    role,
    COUNT(*) as total_actions,
    SUM(CASE WHEN is_error = false THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN is_error = true THEN 1 ELSE 0 END) as error_count,
    AVG(duration) as avg_duration_ms
FROM activity_logs
WHERE deleted_at IS NULL
GROUP BY DATE(created_at), user_id, username, role
ORDER BY activity_date DESC, total_actions DESC;

-- Optional: Create a function for automatic cleanup of old logs
CREATE OR REPLACE FUNCTION cleanup_old_activity_logs(days_to_keep INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM activity_logs
    WHERE created_at < CURRENT_TIMESTAMP - (days_to_keep || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Optional: Create a scheduled job to cleanup old logs (PostgreSQL extension required: pg_cron)
-- Uncomment if you have pg_cron extension installed
-- SELECT cron.schedule(
--     'cleanup-activity-logs',
--     '0 2 * * 0', -- Every Sunday at 2 AM
--     'SELECT cleanup_old_activity_logs(90);'
-- );

COMMIT;
