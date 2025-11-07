-- Migration: Add sales prefix and next number to settings table
-- This will enable configurable sales transaction prefixes instead of hardcoded "SO-"

-- Add sales_prefix column with default value 'SOA'
ALTER TABLE settings ADD COLUMN IF NOT EXISTS sales_prefix VARCHAR(10) DEFAULT 'SOA';

-- Add sales_next_number column with default value 1
ALTER TABLE settings ADD COLUMN IF NOT EXISTS sales_next_number INTEGER DEFAULT 1;

-- Update any existing settings record to have the new fields
UPDATE settings 
SET sales_prefix = COALESCE(sales_prefix, 'SOA'),
    sales_next_number = COALESCE(sales_next_number, 1)
WHERE sales_prefix IS NULL OR sales_next_number IS NULL;

-- Add NOT NULL constraints after setting default values
ALTER TABLE settings ALTER COLUMN sales_prefix SET NOT NULL;
ALTER TABLE settings ALTER COLUMN sales_next_number SET NOT NULL;