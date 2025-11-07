-- Migration to fix outstanding amounts for existing approved credit purchases
-- This fixes purchases where outstanding_amount is 0 or NULL but they are CREDIT purchases

-- Update outstanding amounts for APPROVED CREDIT purchases that haven't been paid
UPDATE purchases 
SET 
    outstanding_amount = total_amount,
    paid_amount = 0
WHERE 
    payment_method = 'CREDIT' 
    AND status = 'APPROVED' 
    AND (outstanding_amount IS NULL OR outstanding_amount = 0)
    AND total_amount > 0;

-- Update outstanding amounts for PAID CREDIT purchases by calculating from existing payments
-- This handles cases where payments were made but outstanding amount wasn't updated
UPDATE purchases p
SET 
    outstanding_amount = GREATEST(0, p.total_amount - COALESCE((
        SELECT SUM(pa.allocated_amount)
        FROM payment_allocations pa
        INNER JOIN payments pay ON pa.payment_id = pay.id
        WHERE pa.bill_id = p.id AND pay.status = 'COMPLETED'
    ), 0)),
    paid_amount = COALESCE((
        SELECT SUM(pa.allocated_amount)
        FROM payment_allocations pa
        INNER JOIN payments pay ON pa.payment_id = pay.id
        WHERE pa.bill_id = p.id AND pay.status = 'COMPLETED'
    ), 0)
WHERE 
    p.payment_method = 'CREDIT' 
    AND p.status IN ('APPROVED', 'PAID')
    AND p.total_amount > 0;

-- Update status to PAID for purchases that are fully paid
UPDATE purchases 
SET status = 'PAID'
WHERE 
    payment_method = 'CREDIT' 
    AND status = 'APPROVED' 
    AND outstanding_amount = 0
    AND paid_amount > 0;

-- Add indexes for performance if they don't exist
CREATE INDEX IF NOT EXISTS idx_purchases_payment_method ON purchases(payment_method);
CREATE INDEX IF NOT EXISTS idx_purchases_status ON purchases(status);
CREATE INDEX IF NOT EXISTS idx_payment_allocations_bill_id ON payment_allocations(bill_id);