-- =====================================================
-- Migration: 037_add_invoice_types_system_fixed.sql  
-- =====================================================
-- Author: System
-- Date: 2025-10-02
-- Description: Add invoice types and counter system for custom invoice numbering (PostgreSQL)
--              This migration adds support for custom invoice numbering formats
--              like "0001/STA-C/X-2025" where each invoice type has its own counter

-- =====================================================
-- 1. Create invoice_types table
-- =====================================================
CREATE TABLE IF NOT EXISTS invoice_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    CONSTRAINT fk_invoice_types_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

-- Create indexes for invoice_types
CREATE INDEX IF NOT EXISTS idx_invoice_types_code ON invoice_types(code);
CREATE INDEX IF NOT EXISTS idx_invoice_types_is_active ON invoice_types(is_active);
CREATE INDEX IF NOT EXISTS idx_invoice_types_created_by ON invoice_types(created_by);
CREATE INDEX IF NOT EXISTS idx_invoice_types_deleted_at ON invoice_types(deleted_at);

-- Add table comment
COMMENT ON TABLE invoice_types IS 'Invoice types for custom numbering schemes';
COMMENT ON COLUMN invoice_types.name IS 'Display name for invoice type (e.g., Corporate Sales)';
COMMENT ON COLUMN invoice_types.code IS 'Short code for numbering (e.g., STA-C)';
COMMENT ON COLUMN invoice_types.description IS 'Optional description of the invoice type';
COMMENT ON COLUMN invoice_types.is_active IS 'Whether this type is active';
COMMENT ON COLUMN invoice_types.created_by IS 'User who created this type';

-- =====================================================
-- 2. Create invoice_counters table
-- =====================================================
CREATE TABLE IF NOT EXISTS invoice_counters (
    id BIGSERIAL PRIMARY KEY,
    invoice_type_id BIGINT NOT NULL,
    year INTEGER NOT NULL,
    counter INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(invoice_type_id, year),
    CONSTRAINT fk_invoice_counters_invoice_type_id FOREIGN KEY (invoice_type_id) REFERENCES invoice_types(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Create indexes for invoice_counters
CREATE INDEX IF NOT EXISTS idx_invoice_counters_invoice_type_id ON invoice_counters(invoice_type_id);
CREATE INDEX IF NOT EXISTS idx_invoice_counters_year ON invoice_counters(year);

-- Add table and column comments
COMMENT ON TABLE invoice_counters IS 'Counter tracking for invoice numbering per type per year';
COMMENT ON COLUMN invoice_counters.invoice_type_id IS 'Foreign key to invoice_types';
COMMENT ON COLUMN invoice_counters.year IS 'Year for counter (e.g., 2025)';
COMMENT ON COLUMN invoice_counters.counter IS 'Current counter value for this type/year';

-- =====================================================
-- 3. Add invoice_type_id column to sales table (if not exists)
-- =====================================================
DO $$
BEGIN
    -- Check if column exists, add if it doesn't
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'sales' 
        AND column_name = 'invoice_type_id'
    ) THEN
        ALTER TABLE sales ADD COLUMN invoice_type_id BIGINT NULL;
        RAISE NOTICE 'Added invoice_type_id column to sales table';
    ELSE
        RAISE NOTICE 'invoice_type_id column already exists in sales table';
    END IF;
END $$;

-- Add index for invoice_type_id if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_sales_invoice_type_id ON sales(invoice_type_id);

-- Add foreign key constraint if it doesn't exist
DO $$
BEGIN
    -- Check if foreign key constraint exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.key_column_usage 
        WHERE table_name = 'sales' 
        AND constraint_name = 'fk_sales_invoice_type_id'
    ) THEN
        ALTER TABLE sales ADD CONSTRAINT fk_sales_invoice_type_id 
        FOREIGN KEY (invoice_type_id) REFERENCES invoice_types(id) ON DELETE SET NULL ON UPDATE CASCADE;
        RAISE NOTICE 'Added foreign key constraint fk_sales_invoice_type_id';
    ELSE
        RAISE NOTICE 'Foreign key constraint fk_sales_invoice_type_id already exists';
    END IF;
END $$;

-- Add comment to the new column
COMMENT ON COLUMN sales.invoice_type_id IS 'Optional invoice type for custom numbering';

-- =====================================================
-- 4. Insert default invoice types (seed data)
-- =====================================================
DO $$
DECLARE
    admin_user_id BIGINT;
BEGIN
    -- Get admin user ID (fallback to 1 if no admin found)
    SELECT id INTO admin_user_id FROM users WHERE role = 'admin' OR role = 'ADMIN' LIMIT 1;
    IF admin_user_id IS NULL THEN
        admin_user_id := 1;
        RAISE NOTICE 'No admin user found, using user ID 1';
    END IF;

    -- Insert invoice types with UPSERT behavior
    INSERT INTO invoice_types (name, code, description, created_by, created_at, updated_at) VALUES
    ('Corporate Sales', 'STA-C', 'Invoice type for corporate/B2B sales transactions', admin_user_id, NOW(), NOW()),
    ('Retail Sales', 'STA-B', 'Invoice type for retail/B2C sales transactions', admin_user_id, NOW(), NOW()),
    ('Service Sales', 'STA-S', 'Invoice type for service-based sales transactions', admin_user_id, NOW(), NOW()),
    ('Export Sales', 'EXP', 'Invoice type for export/international sales', admin_user_id, NOW(), NOW())
    ON CONFLICT (code) DO UPDATE SET 
        name = EXCLUDED.name, 
        description = EXCLUDED.description,
        updated_at = NOW();
    
    RAISE NOTICE 'Invoice types seeded successfully';
END $$;

-- =====================================================
-- 5. Create initial counters for current year
-- =====================================================
INSERT INTO invoice_counters (invoice_type_id, year, counter, created_at, updated_at)
SELECT id, EXTRACT(YEAR FROM NOW()), 0, NOW(), NOW()
FROM invoice_types
WHERE is_active = TRUE
ON CONFLICT (invoice_type_id, year) DO NOTHING; -- Don't reset existing counters

-- =====================================================
-- 6. Add performance indexes
-- =====================================================
-- Index for invoice_number lookup
CREATE INDEX IF NOT EXISTS idx_sales_invoice_number ON sales(invoice_number);

-- Index for date and status lookup
CREATE INDEX IF NOT EXISTS idx_sales_date_status ON sales(date, status);

-- Composite index for status and invoice_type_id
CREATE INDEX IF NOT EXISTS idx_sales_status_invoice_type ON sales(status, invoice_type_id);

-- =====================================================
-- 7. Create updated_at trigger for invoice_types
-- =====================================================
CREATE OR REPLACE FUNCTION update_invoice_types_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Drop trigger if exists, then create
DROP TRIGGER IF EXISTS trigger_invoice_types_updated_at ON invoice_types;
CREATE TRIGGER trigger_invoice_types_updated_at
    BEFORE UPDATE ON invoice_types
    FOR EACH ROW
    EXECUTE FUNCTION update_invoice_types_updated_at();

-- =====================================================
-- 8. Create updated_at trigger for invoice_counters
-- =====================================================
CREATE OR REPLACE FUNCTION update_invoice_counters_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Drop trigger if exists, then create
DROP TRIGGER IF EXISTS trigger_invoice_counters_updated_at ON invoice_counters;
CREATE TRIGGER trigger_invoice_counters_updated_at
    BEFORE UPDATE ON invoice_counters
    FOR EACH ROW
    EXECUTE FUNCTION update_invoice_counters_updated_at();

-- =====================================================
-- 9. Create helper function for invoice number generation
-- =====================================================
CREATE OR REPLACE FUNCTION get_next_invoice_number(invoice_type_id_param BIGINT)
RETURNS TEXT AS $$
DECLARE
    current_year INTEGER;
    next_counter INTEGER;
    invoice_code TEXT;
    roman_month TEXT;
    result_number TEXT;
BEGIN
    current_year := EXTRACT(YEAR FROM NOW());
    
    -- Get invoice type code
    SELECT code INTO invoice_code FROM invoice_types WHERE id = invoice_type_id_param;
    IF invoice_code IS NULL THEN
        RAISE EXCEPTION 'Invoice type not found with ID: %', invoice_type_id_param;
    END IF;
    
    -- Get and increment counter atomically
    INSERT INTO invoice_counters (invoice_type_id, year, counter)
    VALUES (invoice_type_id_param, current_year, 1)
    ON CONFLICT (invoice_type_id, year) 
    DO UPDATE SET counter = invoice_counters.counter + 1;
    
    -- Get the updated counter
    SELECT counter INTO next_counter 
    FROM invoice_counters 
    WHERE invoice_type_id = invoice_type_id_param AND year = current_year;
    
    -- Convert month to Roman numerals
    roman_month := CASE EXTRACT(MONTH FROM NOW())
        WHEN 1 THEN 'I' WHEN 2 THEN 'II' WHEN 3 THEN 'III' WHEN 4 THEN 'IV'
        WHEN 5 THEN 'V' WHEN 6 THEN 'VI' WHEN 7 THEN 'VII' WHEN 8 THEN 'VIII'
        WHEN 9 THEN 'IX' WHEN 10 THEN 'X' WHEN 11 THEN 'XI' WHEN 12 THEN 'XII'
    END;
    
    -- Format: 0001/STA-C/X-2025
    result_number := LPAD(next_counter::TEXT, 4, '0') || '/' || 
                     invoice_code || '/' || 
                     roman_month || '-' || current_year;
    
    RETURN result_number;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 10. Create helper function for previewing next invoice number
-- =====================================================
CREATE OR REPLACE FUNCTION preview_next_invoice_number(invoice_type_id_param BIGINT)
RETURNS TEXT AS $$
DECLARE
    current_year INTEGER;
    next_counter INTEGER;
    invoice_code TEXT;
    roman_month TEXT;
    result_number TEXT;
BEGIN
    current_year := EXTRACT(YEAR FROM NOW());
    
    -- Get invoice type code
    SELECT code INTO invoice_code FROM invoice_types WHERE id = invoice_type_id_param;
    IF invoice_code IS NULL THEN
        RAISE EXCEPTION 'Invoice type not found with ID: %', invoice_type_id_param;
    END IF;
    
    -- Get current counter (without incrementing)
    SELECT COALESCE(counter, 0) + 1 INTO next_counter
    FROM invoice_counters 
    WHERE invoice_type_id = invoice_type_id_param AND year = current_year;
    
    -- If no counter exists, next would be 1
    IF next_counter IS NULL THEN
        next_counter := 1;
    END IF;
    
    -- Convert month to Roman numerals
    roman_month := CASE EXTRACT(MONTH FROM NOW())
        WHEN 1 THEN 'I' WHEN 2 THEN 'II' WHEN 3 THEN 'III' WHEN 4 THEN 'IV'
        WHEN 5 THEN 'V' WHEN 6 THEN 'VI' WHEN 7 THEN 'VII' WHEN 8 THEN 'VIII'
        WHEN 9 THEN 'IX' WHEN 10 THEN 'X' WHEN 11 THEN 'XI' WHEN 12 THEN 'XII'
    END;
    
    -- Format: 0001/STA-C/X-2025
    result_number := LPAD(next_counter::TEXT, 4, '0') || '/' || 
                     invoice_code || '/' || 
                     roman_month || '-' || current_year;
    
    RETURN result_number;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 11. Verification and sample data display
-- =====================================================
SELECT 'MIGRATION 037 COMPLETED SUCCESSFULLY' as status;

-- Show created invoice types
SELECT 'Invoice Types Created:' AS info, COUNT(*) AS count FROM invoice_types;

-- Show initialized counters
SELECT 'Invoice Counters Initialized:' AS info, COUNT(*) AS count FROM invoice_counters;

-- Show sample of created data
SELECT 
    it.id,
    it.name,
    it.code,
    it.is_active,
    COALESCE(ic.year::TEXT, 'N/A') as counter_year,
    COALESCE(ic.counter, 0) as current_counter,
    preview_next_invoice_number(it.id) as next_invoice_number
FROM invoice_types it
LEFT JOIN invoice_counters ic ON it.id = ic.invoice_type_id
ORDER BY it.id, ic.year DESC;

-- =====================================================
-- Sample Usage Examples:
-- =====================================================
-- 
-- 1. Create a new sale with invoice type:
-- INSERT INTO sales (customer_id, invoice_type_id, date, ...) VALUES (1, 1, NOW(), ...);
--
-- 2. Generate next invoice number:
-- SELECT get_next_invoice_number(1); -- Returns: "0001/STA-C/X-2025"
--
-- 3. Preview next invoice number (without incrementing):
-- SELECT preview_next_invoice_number(1); -- Returns: "0001/STA-C/X-2025"
--
-- 4. Test all invoice types:
-- SELECT id, name, code, preview_next_invoice_number(id) as next_number
-- FROM invoice_types WHERE is_active = TRUE;
--
-- =====================================================