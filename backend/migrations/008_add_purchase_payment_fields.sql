-- Migration to add missing payment tracking fields to purchases table
-- These fields are needed for payment and matching functionality

-- Add payment tracking fields if they don't exist
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(15,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS outstanding_amount DECIMAL(15,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS matching_status VARCHAR(20) DEFAULT 'PENDING';

-- Initialize outstanding_amount to equal total_amount for existing records where outstanding_amount is 0
UPDATE purchases 
SET outstanding_amount = total_amount 
WHERE outstanding_amount = 0 AND total_amount > 0;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_purchases_paid_amount ON purchases(paid_amount);
CREATE INDEX IF NOT EXISTS idx_purchases_outstanding_amount ON purchases(outstanding_amount);
CREATE INDEX IF NOT EXISTS idx_purchases_matching_status ON purchases(matching_status);

-- Add comments for documentation
COMMENT ON COLUMN purchases.paid_amount IS 'Amount that has been paid for this purchase';
COMMENT ON COLUMN purchases.outstanding_amount IS 'Remaining amount to be paid';
COMMENT ON COLUMN purchases.matching_status IS 'Status of three-way matching: PENDING, PARTIAL, MATCHED, MISMATCH';
