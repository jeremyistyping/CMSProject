-- Migration 011: Purchase Payment Integration
-- Add payment tracking and create purchase_payments table for cross-reference


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

-- Create purchase_payments table for cross-reference tracking
CREATE TABLE IF NOT EXISTS purchase_payments (
    id BIGSERIAL PRIMARY KEY,
    purchase_id BIGINT NOT NULL,
    payment_number VARCHAR(50),
    date TIMESTAMP WITH TIME ZONE,
    amount DECIMAL(15,2) DEFAULT 0,
    method VARCHAR(20),
    reference VARCHAR(100),
    notes TEXT,
    cash_bank_id BIGINT,
    user_id BIGINT NOT NULL,
    payment_id BIGINT, -- Cross-reference to payments table
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (purchase_id) REFERENCES purchases(id) ON DELETE CASCADE,
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL,
    FOREIGN KEY (cash_bank_id) REFERENCES cash_banks(id) ON DELETE SET NULL
);

-- Create indexes for purchase_payments
CREATE INDEX IF NOT EXISTS idx_purchase_payments_purchase_id ON purchase_payments(purchase_id);
CREATE INDEX IF NOT EXISTS idx_purchase_payments_payment_id ON purchase_payments(payment_id);
CREATE INDEX IF NOT EXISTS idx_purchase_payments_date ON purchase_payments(date);

-- Enhance payment_allocations to support purchase bills
ALTER TABLE payment_allocations 
ADD COLUMN IF NOT EXISTS bill_id BIGINT;

-- Create index for bill_id
CREATE INDEX IF NOT EXISTS idx_payment_allocations_bill_id ON payment_allocations(bill_id);

-- Add foreign key constraint for bill_id
ALTER TABLE payment_allocations 
ADD CONSTRAINT fk_payment_allocations_bill 
    FOREIGN KEY (bill_id) REFERENCES purchases(id) ON DELETE SET NULL;

-- Comments for documentation
COMMENT ON COLUMN purchases.paid_amount IS 'Total amount paid for this purchase';
COMMENT ON COLUMN purchases.outstanding_amount IS 'Remaining amount to be paid (total_amount - paid_amount)';
COMMENT ON TABLE purchase_payments IS 'Cross-reference table linking purchases to payment management records';
COMMENT ON COLUMN payment_allocations.bill_id IS 'Reference to purchase (bill) for payable payments';
