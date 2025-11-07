-- Migration 009: Add payment method fields to purchases table
-- This enables proper tracking of cash vs credit purchases and bank accounts

-- Add payment method fields to purchases table
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS payment_method VARCHAR(20) DEFAULT 'CREDIT',
ADD COLUMN IF NOT EXISTS bank_account_id INTEGER REFERENCES cash_banks(id),
ADD COLUMN IF NOT EXISTS payment_reference VARCHAR(100);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_purchases_bank_account_id ON purchases(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_purchases_payment_method ON purchases(payment_method);

-- Add comments for documentation
COMMENT ON COLUMN purchases.payment_method IS 'Payment method: CREDIT, CASH, or TRANSFER';
COMMENT ON COLUMN purchases.bank_account_id IS 'Bank account used for cash/transfer purchases (foreign key to cash_banks)';
COMMENT ON COLUMN purchases.payment_reference IS 'Reference for payment (check number, transfer reference, etc.)';

-- Update existing purchases to have default payment method
UPDATE purchases 
SET payment_method = 'CREDIT' 
WHERE payment_method IS NULL;

-- Make payment_method not null after setting defaults
ALTER TABLE purchases 
ALTER COLUMN payment_method SET NOT NULL;

COMMIT;
