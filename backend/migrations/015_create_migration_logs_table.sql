-- Create migration_logs table for tracking migrations
CREATE TABLE IF NOT EXISTS migration_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    migration_name VARCHAR(255) NOT NULL UNIQUE,
    status ENUM('SUCCESS', 'FAILED', 'SKIPPED') NOT NULL DEFAULT 'SUCCESS',
    message TEXT,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    execution_time_ms INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_migration_logs_name (migration_name),
    INDEX idx_migration_logs_status (status),
    INDEX idx_migration_logs_executed_at (executed_at)
);