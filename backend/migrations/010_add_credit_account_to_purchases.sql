-- Add credit_account_id field to purchases table for tracking liability accounts on credit purchases
-- Migration: 010_add_credit_account_to_purchases

BEGIN;

-- Add credit_account_id column to purchases table
ALTER TABLE purchases ADD COLUMN credit_account_id INTEGER;

-- Add index for better performance on credit_account_id lookups
CREATE INDEX idx_purchases_credit_account_id ON purchases(credit_account_id);

-- Add foreign key constraint to ensure referential integrity with accounts table
ALTER TABLE purchases ADD CONSTRAINT fk_purchases_credit_account 
    FOREIGN KEY (credit_account_id) REFERENCES accounts(id);

-- Add comment to explain the purpose
COMMENT ON COLUMN purchases.credit_account_id IS 'Reference to liability account for credit purchases';

COMMIT;
