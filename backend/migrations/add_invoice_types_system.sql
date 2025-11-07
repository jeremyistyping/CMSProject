-- =====================================================
-- Invoice Types System Migration
-- =====================================================
-- Author: System
-- Date: 2025-10-02
-- Description: Add invoice types and counter system for custom invoice numbering

-- Create invoice_types table
CREATE TABLE IF NOT EXISTS invoice_types (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT 'Display name for invoice type (e.g., Corporate Sales)',
    code VARCHAR(20) NOT NULL UNIQUE COMMENT 'Short code for numbering (e.g., STA-C)',
    description TEXT COMMENT 'Optional description of the invoice type',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT 'Whether this type is active',
    created_by BIGINT UNSIGNED NOT NULL COMMENT 'User who created this type',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    PRIMARY KEY (id),
    KEY idx_invoice_types_code (code),
    KEY idx_invoice_types_is_active (is_active),
    KEY idx_invoice_types_created_by (created_by),
    KEY idx_invoice_types_deleted_at (deleted_at),
    
    CONSTRAINT fk_invoice_types_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Invoice types for custom numbering schemes';

-- Create invoice_counters table
CREATE TABLE IF NOT EXISTS invoice_counters (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    invoice_type_id BIGINT UNSIGNED NOT NULL COMMENT 'Foreign key to invoice_types',
    year INT NOT NULL COMMENT 'Year for counter (e.g., 2025)',
    counter INT NOT NULL DEFAULT 0 COMMENT 'Current counter value for this type/year',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_invoice_counters_type_year (invoice_type_id, year),
    KEY idx_invoice_counters_invoice_type_id (invoice_type_id),
    KEY idx_invoice_counters_year (year),
    
    CONSTRAINT fk_invoice_counters_invoice_type_id FOREIGN KEY (invoice_type_id) REFERENCES invoice_types(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Counter tracking for invoice numbering per type per year';

-- Add invoice_type_id column to sales table
ALTER TABLE sales 
ADD COLUMN invoice_type_id BIGINT UNSIGNED NULL COMMENT 'Optional invoice type for custom numbering' AFTER sales_person_id,
ADD KEY idx_sales_invoice_type_id (invoice_type_id),
ADD CONSTRAINT fk_sales_invoice_type_id FOREIGN KEY (invoice_type_id) REFERENCES invoice_types(id) ON DELETE SET NULL ON UPDATE CASCADE;

-- Insert default invoice types (optional seed data)
INSERT INTO invoice_types (name, code, description, created_by, created_at, updated_at) VALUES
('Corporate Sales', 'STA-C', 'Invoice type for corporate/B2B sales', 1, NOW(), NOW()),
('Retail Sales', 'STA-B', 'Invoice type for retail/B2C sales', 1, NOW(), NOW()),
('Service Sales', 'STA-S', 'Invoice type for service-based sales', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description);

-- Create initial counters for current year (optional)
INSERT INTO invoice_counters (invoice_type_id, year, counter, created_at, updated_at)
SELECT id, YEAR(NOW()), 0, NOW(), NOW()
FROM invoice_types
WHERE is_active = TRUE
ON DUPLICATE KEY UPDATE counter = counter; -- Don't reset existing counters

-- Add indexes for better performance
CREATE INDEX idx_sales_invoice_number ON sales(invoice_number);
CREATE INDEX idx_sales_date_status ON sales(date, status);
CREATE INDEX idx_sales_status_invoice_type ON sales(status, invoice_type_id);

-- Update existing sales to use default invoice type (optional)
-- WARNING: Only run this if you want to assign a default type to existing invoices
-- UPDATE sales 
-- SET invoice_type_id = (SELECT id FROM invoice_types WHERE code = 'STA-C' LIMIT 1)
-- WHERE invoice_type_id IS NULL AND invoice_number IS NOT NULL;

-- =====================================================
-- Migration Complete
-- =====================================================

-- Verification queries (uncomment to test):
-- SELECT 'Invoice Types Created:' AS status, COUNT(*) AS count FROM invoice_types;
-- SELECT 'Invoice Counters Created:' AS status, COUNT(*) AS count FROM invoice_counters;
-- SELECT 'Sales with Invoice Types:' AS status, COUNT(*) AS count FROM sales WHERE invoice_type_id IS NOT NULL;
-- 
-- Sample usage:
-- SELECT it.name, it.code, ic.year, ic.counter 
-- FROM invoice_types it 
-- LEFT JOIN invoice_counters ic ON it.id = ic.invoice_type_id 
-- WHERE it.is_active = TRUE 
-- ORDER BY it.name, ic.year DESC;