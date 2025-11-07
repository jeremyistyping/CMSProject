-- Migration: Add Sales Data Integrity Constraints
-- Purpose: Add critical database constraints to prevent data inconsistencies
-- Date: 2025-09-19
-- Priority: Critical (Priority 1)

-- =======================
-- SALES TABLE CONSTRAINTS
-- =======================

-- 1. Amount constraints - ensure all amounts are non-negative
ALTER TABLE sales 
ADD CONSTRAINT chk_sales_amounts_positive 
CHECK (
    total_amount >= 0 AND 
    paid_amount >= 0 AND 
    outstanding_amount >= 0 AND
    subtotal >= 0 AND
    discount_amount >= 0
);

-- 2. Payment consistency constraint - paid + outstanding = total
-- Note: Allow small floating point tolerance (0.01)
ALTER TABLE sales 
ADD CONSTRAINT chk_sales_payment_balance 
CHECK (
    ABS((paid_amount + outstanding_amount) - total_amount) <= 0.01
);

-- 3. Date logic constraints
ALTER TABLE sales 
ADD CONSTRAINT chk_sales_date_logic 
CHECK (
    due_date >= date AND
    (valid_until IS NULL OR valid_until >= date)
);

-- 4. Status constraints - only allow valid status values
ALTER TABLE sales 
ADD CONSTRAINT chk_sales_status_valid 
CHECK (
    status IN ('DRAFT', 'PENDING', 'CONFIRMED', 'INVOICED', 'OVERDUE', 'PAID', 'COMPLETED', 'CANCELLED')
);

-- 5. Type constraints - only allow valid sale types
ALTER TABLE sales 
ADD CONSTRAINT chk_sales_type_valid 
CHECK (
    type IN ('QUOTATION', 'ORDER', 'INVOICE')
);

-- 6. Tax constraints - tax rates should be reasonable
ALTER TABLE sales 
ADD CONSTRAINT chk_sales_tax_rates 
CHECK (
    ppn_percent >= 0 AND ppn_percent <= 100 AND
    pph_percent >= 0 AND pph_percent <= 100 AND
    discount_percent >= 0 AND discount_percent <= 100 AND
    ppn_rate >= 0 AND ppn_rate <= 100 AND
    pph21_rate >= 0 AND pph21_rate <= 100 AND
    pph23_rate >= 0 AND pph23_rate <= 100
);

-- 7. Currency and exchange rate constraints
ALTER TABLE sales 
ADD CONSTRAINT chk_sales_currency_valid 
CHECK (
    currency IN ('IDR', 'USD', 'EUR', 'SGD', 'JPY') AND
    exchange_rate > 0
);

-- 8. Unique constraints for critical fields
-- Ensure invoice numbers are unique (when not null)
CREATE UNIQUE INDEX idx_sales_invoice_number_unique 
ON sales (invoice_number) 
WHERE invoice_number IS NOT NULL AND invoice_number != '' AND deleted_at IS NULL;

-- Ensure sale codes are unique (when not deleted)
CREATE UNIQUE INDEX idx_sales_code_unique 
ON sales (code) 
WHERE deleted_at IS NULL;

-- =============================
-- SALE ITEMS TABLE CONSTRAINTS
-- =============================

-- 1. Quantity and price constraints
ALTER TABLE sale_items 
ADD CONSTRAINT chk_sale_items_positive_values 
CHECK (
    quantity > 0 AND
    unit_price >= 0 AND
    line_total >= 0 AND
    discount_amount >= 0 AND
    final_amount >= 0
);

-- 2. Discount constraints
ALTER TABLE sale_items 
ADD CONSTRAINT chk_sale_items_discount 
CHECK (
    discount_percent >= 0 AND discount_percent <= 100 AND
    discount_amount <= (quantity * unit_price)
);

-- 3. Line total consistency
ALTER TABLE sale_items 
ADD CONSTRAINT chk_sale_items_line_total_consistency 
CHECK (
    ABS(line_total - (quantity * unit_price - discount_amount)) <= 0.01
);

-- ==============================
-- SALE PAYMENTS TABLE CONSTRAINTS
-- ==============================

-- 1. Payment amount constraints
ALTER TABLE sale_payments 
ADD CONSTRAINT chk_sale_payments_amount_positive 
CHECK (amount > 0);

-- 2. Payment date constraints
ALTER TABLE sale_payments 
ADD CONSTRAINT chk_sale_payments_date_reasonable 
CHECK (
    payment_date <= CURRENT_DATE + INTERVAL '1 day' AND
    payment_date >= '2020-01-01'::date
);

-- 3. Payment method constraints
ALTER TABLE sale_payments 
ADD CONSTRAINT chk_sale_payments_method_valid 
CHECK (
    payment_method IN ('CASH', 'BANK_TRANSFER', 'CREDIT_CARD', 'CHECK', 'OTHER')
);

-- 4. Payment status constraints
ALTER TABLE sale_payments 
ADD CONSTRAINT chk_sale_payments_status_valid 
CHECK (
    status IN ('COMPLETED', 'PENDING', 'CANCELLED')
);

-- ===============================
-- PERFORMANCE & INTEGRITY INDICES
-- ===============================

-- Critical indices for sales queries
CREATE INDEX IF NOT EXISTS idx_sales_status_date 
ON sales(status, date) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sales_customer_status 
ON sales(customer_id, status) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sales_outstanding_amount 
ON sales(outstanding_amount) 
WHERE outstanding_amount > 0 AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sales_due_date_overdue 
ON sales(due_date, status) 
WHERE status IN ('INVOICED', 'OVERDUE') AND deleted_at IS NULL;

-- Sale items indices
CREATE INDEX IF NOT EXISTS idx_sale_items_sale_product 
ON sale_items(sale_id, product_id) 
WHERE deleted_at IS NULL;

-- Sale payments indices
CREATE INDEX IF NOT EXISTS idx_sale_payments_sale_date 
ON sale_payments(sale_id, payment_date) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sale_payments_amount 
ON sale_payments(amount) 
WHERE amount > 0 AND deleted_at IS NULL;

-- ======================
-- FOREIGN KEY CONSTRAINTS
-- ======================

-- Ensure referential integrity (if not already exists)

-- Sales foreign keys
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_sales_customer_id'
    ) THEN
        ALTER TABLE sales 
        ADD CONSTRAINT fk_sales_customer_id 
        FOREIGN KEY (customer_id) REFERENCES contacts(id) ON DELETE RESTRICT;
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_sales_user_id'
    ) THEN
        ALTER TABLE sales 
        ADD CONSTRAINT fk_sales_user_id 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_sales_sales_person_id'
    ) THEN
        ALTER TABLE sales 
        ADD CONSTRAINT fk_sales_sales_person_id 
        FOREIGN KEY (sales_person_id) REFERENCES contacts(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Sale items foreign keys
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_sale_items_sale_id'
    ) THEN
        ALTER TABLE sale_items 
        ADD CONSTRAINT fk_sale_items_sale_id 
        FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE CASCADE;
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_sale_items_product_id'
    ) THEN
        ALTER TABLE sale_items 
        ADD CONSTRAINT fk_sale_items_product_id 
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- Sale payments foreign keys
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_sale_payments_sale_id'
    ) THEN
        ALTER TABLE sale_payments 
        ADD CONSTRAINT fk_sale_payments_sale_id 
        FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE CASCADE;
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_sale_payments_user_id'
    ) THEN
        ALTER TABLE sale_payments 
        ADD CONSTRAINT fk_sale_payments_user_id 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- ===================
-- DATA VALIDATION FUNCTIONS
-- ===================

-- Function to validate sale data integrity
CREATE OR REPLACE FUNCTION validate_sale_data_integrity()
RETURNS TABLE(
    check_name VARCHAR(100),
    issue_count BIGINT,
    sample_ids TEXT,
    severity VARCHAR(20)
) AS $$
BEGIN
    -- Check 1: Sales with inconsistent payment amounts
    RETURN QUERY
    SELECT 
        'Inconsistent Payment Amounts'::VARCHAR(100),
        COUNT(*),
        STRING_AGG(id::TEXT, ', ' ORDER BY id LIMIT 10),
        'CRITICAL'::VARCHAR(20)
    FROM sales 
    WHERE ABS((paid_amount + outstanding_amount) - total_amount) > 0.01
    AND deleted_at IS NULL;

    -- Check 2: Sales with negative amounts
    RETURN QUERY
    SELECT 
        'Negative Amounts'::VARCHAR(100),
        COUNT(*),
        STRING_AGG(id::TEXT, ', ' ORDER BY id LIMIT 10),
        'CRITICAL'::VARCHAR(20)
    FROM sales 
    WHERE (total_amount < 0 OR paid_amount < 0 OR outstanding_amount < 0)
    AND deleted_at IS NULL;

    -- Check 3: Sales with invalid date logic
    RETURN QUERY
    SELECT 
        'Invalid Date Logic'::VARCHAR(100),
        COUNT(*),
        STRING_AGG(id::TEXT, ', ' ORDER BY id LIMIT 10),
        'HIGH'::VARCHAR(20)
    FROM sales 
    WHERE due_date < date
    AND deleted_at IS NULL;

    -- Check 4: Orphaned sale items
    RETURN QUERY
    SELECT 
        'Orphaned Sale Items'::VARCHAR(100),
        COUNT(*),
        STRING_AGG(id::TEXT, ', ' ORDER BY id LIMIT 10),
        'MEDIUM'::VARCHAR(20)
    FROM sale_items si
    WHERE NOT EXISTS (SELECT 1 FROM sales s WHERE s.id = si.sale_id)
    AND si.deleted_at IS NULL;

    -- Check 5: Payments exceeding sale totals
    RETURN QUERY
    SELECT 
        'Overpayments'::VARCHAR(100),
        COUNT(*),
        STRING_AGG(s.id::TEXT, ', ' ORDER BY s.id LIMIT 10),
        'HIGH'::VARCHAR(20)
    FROM sales s
    WHERE s.paid_amount > s.total_amount + 0.01
    AND s.deleted_at IS NULL;

END;
$$ LANGUAGE plpgsql;

-- ===================
-- CONSTRAINT VALIDATION
-- ===================

-- Test all constraints work by trying to insert invalid data
-- This should fail with constraint violations

-- Create a test function to validate constraints
CREATE OR REPLACE FUNCTION test_sales_constraints()
RETURNS TEXT AS $$
DECLARE
    test_result TEXT := 'All constraints working correctly';
    error_count INTEGER := 0;
BEGIN
    -- Test 1: Try negative total amount (should fail)
    BEGIN
        INSERT INTO sales (code, customer_id, user_id, type, date, due_date, total_amount, paid_amount, outstanding_amount)
        VALUES ('TEST-NEG', 1, 1, 'INVOICE', CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', -100, 0, -100);
        DELETE FROM sales WHERE code = 'TEST-NEG';
        error_count := error_count + 1;
        test_result := test_result || '; ERROR: Negative amount constraint not working';
    EXCEPTION 
        WHEN check_violation THEN
            -- Expected behavior - constraint is working
            NULL;
    END;

    -- Test 2: Try inconsistent payment amounts (should fail)
    BEGIN
        INSERT INTO sales (code, customer_id, user_id, type, date, due_date, total_amount, paid_amount, outstanding_amount)
        VALUES ('TEST-INCONSISTENT', 1, 1, 'INVOICE', CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', 100, 50, 60);
        DELETE FROM sales WHERE code = 'TEST-INCONSISTENT';
        error_count := error_count + 1;
        test_result := test_result || '; ERROR: Payment balance constraint not working';
    EXCEPTION 
        WHEN check_violation THEN
            -- Expected behavior - constraint is working
            NULL;
    END;

    -- Test 3: Try invalid status (should fail)
    BEGIN
        INSERT INTO sales (code, customer_id, user_id, type, status, date, due_date, total_amount, paid_amount, outstanding_amount)
        VALUES ('TEST-STATUS', 1, 1, 'INVOICE', 'INVALID_STATUS', CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', 100, 0, 100);
        DELETE FROM sales WHERE code = 'TEST-STATUS';
        error_count := error_count + 1;
        test_result := test_result || '; ERROR: Status constraint not working';
    EXCEPTION 
        WHEN check_violation THEN
            -- Expected behavior - constraint is working
            NULL;
    END;

    IF error_count = 0 THEN
        RETURN 'SUCCESS: All sales data integrity constraints are working correctly';
    ELSE
        RETURN test_result;
    END IF;

END;
$$ LANGUAGE plpgsql;

-- =====================
-- MIGRATION COMPLETION
-- =====================

-- Log migration completion
INSERT INTO migration_logs (migration_name, executed_at, description, status)
VALUES (
    '020_add_sales_data_integrity_constraints',
    NOW(),
    'Added comprehensive data integrity constraints for sales, sale_items, and sale_payments tables',
    'COMPLETED'
) ON CONFLICT (migration_name) DO UPDATE SET 
    executed_at = NOW(),
    status = 'COMPLETED';

-- Final validation
SELECT 'Migration completed successfully. Run SELECT * FROM test_sales_constraints(); to validate.' as result;