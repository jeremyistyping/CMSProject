-- Migration 012: Purchase Payment Integration (PostgreSQL Compatible)
-- Add payment tracking and create purchase_payments table for cross-reference

BEGIN;

-- Ensure purchase payment fields exist (redundant but safe)
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(15,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS outstanding_amount DECIMAL(15,2) DEFAULT 0;

-- Initialize outstanding_amount for existing CREDIT purchases
UPDATE purchases 
SET outstanding_amount = total_amount 
WHERE payment_method = 'CREDIT' AND outstanding_amount = 0;

-- Initialize outstanding_amount to 0 for CASH/TRANSFER purchases (already paid)
UPDATE purchases 
SET outstanding_amount = 0,
    paid_amount = total_amount 
WHERE payment_method IN ('CASH', 'TRANSFER') AND outstanding_amount != 0;

-- Create purchase_payments table for cross-reference tracking (PostgreSQL syntax)
CREATE TABLE IF NOT EXISTS purchase_payments (
    id SERIAL PRIMARY KEY,
    purchase_id INT NOT NULL,
    payment_number VARCHAR(50),
    date TIMESTAMP,
    amount DECIMAL(15,2) DEFAULT 0,
    method VARCHAR(20),
    reference VARCHAR(100),
    notes TEXT,
    cash_bank_id INT,
    user_id INT NOT NULL,
    payment_id INT, -- Cross-reference to payments table
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (purchase_id) REFERENCES purchases(id) ON DELETE CASCADE,
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL,
    FOREIGN KEY (cash_bank_id) REFERENCES cash_banks(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_purchase_payments_purchase_id ON purchase_payments(purchase_id);
CREATE INDEX IF NOT EXISTS idx_purchase_payments_payment_id ON purchase_payments(payment_id);
CREATE INDEX IF NOT EXISTS idx_purchase_payments_date ON purchase_payments(date);
CREATE INDEX IF NOT EXISTS idx_purchase_payments_deleted_at ON purchase_payments(deleted_at);

-- Enhance payment_allocations to support purchase bills
ALTER TABLE payment_allocations 
ADD COLUMN IF NOT EXISTS bill_id INT;

-- Create index for bill_id
CREATE INDEX IF NOT EXISTS idx_payment_allocations_bill_id ON payment_allocations(bill_id);

-- Add foreign key constraint for bill_id (PostgreSQL syntax)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_payment_allocations_bill'
    ) THEN
        ALTER TABLE payment_allocations 
        ADD CONSTRAINT fk_payment_allocations_bill 
            FOREIGN KEY (bill_id) REFERENCES purchases(id) ON DELETE SET NULL;
    END IF;
END $$;

COMMIT;