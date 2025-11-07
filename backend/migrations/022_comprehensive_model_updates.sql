-- Migration: 022_comprehensive_model_updates.sql
-- Purpose: Handle all model changes from recent codebase improvements
-- Date: 2024-09-19

BEGIN;

-- ==============================================
-- 1. CREATE ASSET_CATEGORIES TABLE
-- ==============================================
CREATE TABLE IF NOT EXISTS asset_categories (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id BIGINT REFERENCES asset_categories(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for asset_categories
CREATE INDEX IF NOT EXISTS idx_asset_categories_parent_id ON asset_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_asset_categories_deleted_at ON asset_categories(deleted_at);
CREATE INDEX IF NOT EXISTS idx_asset_categories_code ON asset_categories(code);
CREATE INDEX IF NOT EXISTS idx_asset_categories_is_active ON asset_categories(is_active);

-- ==============================================
-- 2. UPDATE ASSETS TABLE
-- ==============================================
-- Add category_id to assets table
ALTER TABLE assets 
ADD COLUMN IF NOT EXISTS category_id BIGINT REFERENCES asset_categories(id) ON DELETE SET NULL;

-- Create index for assets category_id
CREATE INDEX IF NOT EXISTS idx_assets_category_id ON assets(category_id);

-- ==============================================
-- 3. UPDATE PURCHASES TABLE - TAX FIELDS TO NULLABLE
-- ==============================================
-- Note: PostgreSQL doesn't require explicit changes for nullable fields
-- The application will handle NULL vs 0 distinction
-- But we should ensure the columns allow NULL values

-- Update existing NOT NULL constraints if they exist (safely)
DO $$
BEGIN
    -- Check if columns exist before trying to modify them
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'purchases' AND column_name = 'ppn_rate') THEN
        ALTER TABLE purchases ALTER COLUMN ppn_rate DROP NOT NULL;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'purchases' AND column_name = 'pph21_rate') THEN
        ALTER TABLE purchases ALTER COLUMN pph21_rate DROP NOT NULL;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'purchases' AND column_name = 'pph23_rate') THEN
        ALTER TABLE purchases ALTER COLUMN pph23_rate DROP NOT NULL;
    END IF;
END $$;

-- ==============================================
-- 4. UPDATE SALE_ITEMS TABLE
-- ==============================================
-- Add new fields to sale_items
ALTER TABLE sale_items 
ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(15,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS tax_account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL;

-- Create index for tax_account_id
CREATE INDEX IF NOT EXISTS idx_sale_items_tax_account_id ON sale_items(tax_account_id);

-- ==============================================
-- 5. UPDATE SALE_PAYMENTS TABLE STRUCTURE
-- ==============================================
-- Add new fields to sale_payments table
ALTER TABLE sale_payments 
ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50),
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'COMPLETED',
ADD COLUMN IF NOT EXISTS account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL;

-- Remove old fields that might conflict (if they exist)
-- ALTER TABLE sale_payments DROP COLUMN IF EXISTS payment_id;
-- ALTER TABLE sale_payments DROP COLUMN IF EXISTS payment_number;
-- ALTER TABLE sale_payments DROP COLUMN IF EXISTS method;

-- Update existing method column to payment_method if it exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'sale_payments' AND column_name = 'method') THEN
        -- Copy data from method to payment_method
        UPDATE sale_payments SET payment_method = method WHERE payment_method IS NULL;
        -- Drop the old column
        ALTER TABLE sale_payments DROP COLUMN method;
    END IF;
END $$;

-- Create indexes for sale_payments
CREATE INDEX IF NOT EXISTS idx_sale_payments_status ON sale_payments(status);
CREATE INDEX IF NOT EXISTS idx_sale_payments_payment_method ON sale_payments(payment_method);
CREATE INDEX IF NOT EXISTS idx_sale_payments_account_id ON sale_payments(account_id);

-- ==============================================
-- 6. ENSURE SALE_RETURNS AND SALE_RETURN_ITEMS EXIST
-- ==============================================
CREATE TABLE IF NOT EXISTS sale_returns (
    id BIGSERIAL PRIMARY KEY,
    sale_id BIGINT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    approver_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    return_number VARCHAR(50),
    type VARCHAR(20),
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    reason TEXT,
    credit_note_number VARCHAR(50),
    total_amount DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'PENDING',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for sale_returns
CREATE INDEX IF NOT EXISTS idx_sale_returns_sale_id ON sale_returns(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_returns_user_id ON sale_returns(user_id);
CREATE INDEX IF NOT EXISTS idx_sale_returns_approver_id ON sale_returns(approver_id);
CREATE INDEX IF NOT EXISTS idx_sale_returns_date ON sale_returns(date);
CREATE INDEX IF NOT EXISTS idx_sale_returns_status ON sale_returns(status);
CREATE INDEX IF NOT EXISTS idx_sale_returns_deleted_at ON sale_returns(deleted_at);

CREATE TABLE IF NOT EXISTS sale_return_items (
    id BIGSERIAL PRIMARY KEY,
    sale_return_id BIGINT NOT NULL REFERENCES sale_returns(id) ON DELETE CASCADE,
    sale_item_id BIGINT NOT NULL REFERENCES sale_items(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    reason VARCHAR(255),
    unit_price DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for sale_return_items
CREATE INDEX IF NOT EXISTS idx_sale_return_items_sale_return_id ON sale_return_items(sale_return_id);
CREATE INDEX IF NOT EXISTS idx_sale_return_items_sale_item_id ON sale_return_items(sale_item_id);
CREATE INDEX IF NOT EXISTS idx_sale_return_items_deleted_at ON sale_return_items(deleted_at);

-- ==============================================
-- 7. DATA MIGRATION AND CLEANUP
-- ==============================================

-- Migrate existing category data to asset_categories if needed
-- This assumes there might be existing category strings that need to be normalized
INSERT INTO asset_categories (code, name, description, is_active)
SELECT DISTINCT 
    UPPER(REPLACE(category, ' ', '_')) as code,
    category as name,
    'Migrated from existing asset categories' as description,
    true as is_active
FROM assets 
WHERE category IS NOT NULL 
  AND category != ''
  AND NOT EXISTS (
      SELECT 1 FROM asset_categories ac 
      WHERE ac.name = assets.category
  )
ON CONFLICT (code) DO NOTHING;

-- Update assets to link to asset_categories
UPDATE assets 
SET category_id = ac.id
FROM asset_categories ac
WHERE assets.category = ac.name 
  AND assets.category_id IS NULL;

-- ==============================================
-- 8. ADD PERFORMANCE INDEXES
-- ==============================================

-- Add missing indexes for better performance
CREATE INDEX IF NOT EXISTS idx_sales_customer_id_date ON sales(customer_id, date);
CREATE INDEX IF NOT EXISTS idx_sale_items_product_id_sale_id ON sale_items(product_id, sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_payments_sale_id_date ON sale_payments(sale_id, payment_date);

-- ==============================================
-- 9. UPDATE CONSTRAINTS AND VALIDATIONS
-- ==============================================

-- Add check constraints to ensure data integrity
ALTER TABLE sale_returns 
ADD CONSTRAINT IF NOT EXISTS chk_sale_returns_status 
CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'COMPLETED'));

ALTER TABLE sale_payments 
ADD CONSTRAINT IF NOT EXISTS chk_sale_payments_status 
CHECK (status IN ('PENDING', 'COMPLETED', 'CANCELLED'));

ALTER TABLE sale_return_items 
ADD CONSTRAINT IF NOT EXISTS chk_sale_return_items_quantity_positive 
CHECK (quantity > 0);

ALTER TABLE asset_categories 
ADD CONSTRAINT IF NOT EXISTS chk_asset_categories_code_format 
CHECK (code ~ '^[A-Z0-9_]+$');

-- ==============================================
-- 10. CREATE TRIGGERS FOR UPDATED_AT
-- ==============================================

-- Create updated_at trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers to new tables
DROP TRIGGER IF EXISTS update_asset_categories_updated_at ON asset_categories;
CREATE TRIGGER update_asset_categories_updated_at 
    BEFORE UPDATE ON asset_categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_sale_returns_updated_at ON sale_returns;
CREATE TRIGGER update_sale_returns_updated_at 
    BEFORE UPDATE ON sale_returns 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_sale_return_items_updated_at ON sale_return_items;
CREATE TRIGGER update_sale_return_items_updated_at 
    BEFORE UPDATE ON sale_return_items 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==============================================
-- 11. INSERT SAMPLE DATA FOR ASSET CATEGORIES
-- ==============================================

-- Insert some common asset categories if the table is empty
INSERT INTO asset_categories (code, name, description, parent_id, is_active) 
VALUES 
    ('BUILDINGS', 'Buildings & Structures', 'Physical buildings and structures', NULL, true),
    ('VEHICLES', 'Vehicles', 'Transportation vehicles', NULL, true),
    ('EQUIPMENT', 'Equipment', 'Machinery and equipment', NULL, true),
    ('FURNITURE', 'Furniture & Fixtures', 'Office furniture and fixtures', NULL, true),
    ('IT_ASSETS', 'IT Assets', 'Computers, servers, and IT equipment', NULL, true),
    ('OFFICE_EQUIP', 'Office Equipment', 'Office equipment and tools', 'EQUIPMENT', true)
ON CONFLICT (code) DO NOTHING;

-- Update parent_id for subcategories
UPDATE asset_categories 
SET parent_id = (SELECT id FROM asset_categories WHERE code = 'EQUIPMENT' LIMIT 1)
WHERE code = 'OFFICE_EQUIP' AND parent_id IS NULL;

COMMIT;

-- ==============================================
-- ROLLBACK SCRIPT (commented out - for reference)
-- ==============================================

/*
-- To rollback this migration:

BEGIN;

-- Remove triggers
DROP TRIGGER IF EXISTS update_asset_categories_updated_at ON asset_categories;
DROP TRIGGER IF EXISTS update_sale_returns_updated_at ON sale_returns;
DROP TRIGGER IF EXISTS update_sale_return_items_updated_at ON sale_return_items;

-- Drop constraints
ALTER TABLE sale_returns DROP CONSTRAINT IF EXISTS chk_sale_returns_status;
ALTER TABLE sale_payments DROP CONSTRAINT IF EXISTS chk_sale_payments_status;
ALTER TABLE sale_return_items DROP CONSTRAINT IF EXISTS chk_sale_return_items_quantity_positive;
ALTER TABLE asset_categories DROP CONSTRAINT IF EXISTS chk_asset_categories_code_format;

-- Remove columns
ALTER TABLE assets DROP COLUMN IF EXISTS category_id;
ALTER TABLE sale_items DROP COLUMN IF EXISTS discount_amount;
ALTER TABLE sale_items DROP COLUMN IF EXISTS tax_account_id;
ALTER TABLE sale_payments DROP COLUMN IF EXISTS payment_method;
ALTER TABLE sale_payments DROP COLUMN IF EXISTS status;
ALTER TABLE sale_payments DROP COLUMN IF EXISTS account_id;

-- Drop tables (be careful with data loss)
DROP TABLE IF EXISTS sale_return_items;
DROP TABLE IF EXISTS sale_returns;
DROP TABLE IF EXISTS asset_categories;

COMMIT;
*/