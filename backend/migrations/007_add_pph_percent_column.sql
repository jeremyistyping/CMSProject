-- Migration to add missing pph_percent column to sales table
-- This fixes the error: column "pph_percent" does not exist

-- Add pph_percent column to sales table if it doesn't exist
ALTER TABLE sales 
ADD COLUMN IF NOT EXISTS pph_percent DECIMAL(5,2) DEFAULT 0;

-- Update any existing sales that might have pph but no pph_percent
UPDATE sales 
SET pph_percent = 0 
WHERE pph_percent IS NULL;

-- Add index for better performance
CREATE INDEX IF NOT EXISTS idx_sales_pph_percent ON sales(pph_percent);
