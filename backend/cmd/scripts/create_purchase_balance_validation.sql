-- ========================================
-- PURCHASE BALANCE VALIDATION FUNCTION
-- ========================================
-- Similar to the sales balance validation but for purchases
-- This function validates that Accounts Payable balances match purchase outstanding amounts

DELIMITER $$

DROP FUNCTION IF EXISTS validate_purchase_balances$$

CREATE FUNCTION validate_purchase_balances() RETURNS JSON
READS SQL DATA
DETERMINISTIC
BEGIN
    DECLARE validation_result JSON DEFAULT JSON_OBJECT();
    DECLARE total_outstanding DECIMAL(15,2) DEFAULT 0;
    DECLARE current_ap_balance DECIMAL(15,2) DEFAULT 0;
    DECLARE expected_ap_balance DECIMAL(15,2) DEFAULT 0;
    DECLARE balance_discrepancy DECIMAL(15,2) DEFAULT 0;
    DECLARE accounts_payable_account_id INT DEFAULT NULL;
    DECLARE issue_count INT DEFAULT 0;
    DECLARE validation_status VARCHAR(20) DEFAULT 'PASSED';
    DECLARE bank_balance_discrepancy DECIMAL(15,2) DEFAULT 0;
    DECLARE total_payments_recorded DECIMAL(15,2) DEFAULT 0;
    DECLARE total_purchase_paid_amount DECIMAL(15,2) DEFAULT 0;
    
    -- Get total outstanding amount from purchases (credit purchases only)
    SELECT COALESCE(SUM(outstanding_amount), 0) 
    INTO total_outstanding
    FROM purchases 
    WHERE payment_method = 'CREDIT' 
      AND deleted_at IS NULL;
      
    -- Get total paid amount from purchases for verification
    SELECT COALESCE(SUM(paid_amount), 0) 
    INTO total_purchase_paid_amount
    FROM purchases 
    WHERE deleted_at IS NULL;
    
    -- Get total payments recorded in purchase_payments table
    SELECT COALESCE(SUM(amount), 0) 
    INTO total_payments_recorded
    FROM purchase_payments 
    WHERE deleted_at IS NULL;
    
    -- Find Accounts Payable account (Hutang Usaha)
    SELECT id INTO accounts_payable_account_id
    FROM accounts 
    WHERE (code LIKE '%2101%' OR name LIKE '%Hutang Usaha%' OR name LIKE '%Accounts Payable%')
      AND deleted_at IS NULL
    ORDER BY 
        CASE 
            WHEN code = '2101' THEN 1 
            WHEN code LIKE '2101%' THEN 2 
            WHEN name LIKE '%Hutang Usaha%' THEN 3
            ELSE 4 
        END
    LIMIT 1;
    
    -- Get current Accounts Payable balance
    IF accounts_payable_account_id IS NOT NULL THEN
        SELECT COALESCE(balance, 0) 
        INTO current_ap_balance
        FROM accounts 
        WHERE id = accounts_payable_account_id;
    END IF;
    
    -- Expected AP balance should be NEGATIVE (liability)
    -- Outstanding purchases create liability, so expected balance = -total_outstanding
    SET expected_ap_balance = -total_outstanding;
    SET balance_discrepancy = current_ap_balance - expected_ap_balance;
    
    -- Check for issues
    IF ABS(balance_discrepancy) > 1.00 THEN
        SET issue_count = issue_count + 1;
        SET validation_status = 'FAILED';
    END IF;
    
    -- Check payment consistency
    SET bank_balance_discrepancy = total_purchase_paid_amount - total_payments_recorded;
    IF ABS(bank_balance_discrepancy) > 1.00 THEN
        SET issue_count = issue_count + 1;
        SET validation_status = 'FAILED';
    END IF;
    
    -- Build validation result
    SET validation_result = JSON_OBJECT(
        'validation_timestamp', NOW(),
        'status', validation_status,
        'issue_count', issue_count,
        'accounts_payable', JSON_OBJECT(
            'account_id', accounts_payable_account_id,
            'current_balance', current_ap_balance,
            'expected_balance', expected_ap_balance,
            'discrepancy', balance_discrepancy,
            'is_correct', ABS(balance_discrepancy) <= 1.00
        ),
        'purchase_summary', JSON_OBJECT(
            'total_outstanding', total_outstanding,
            'total_paid_amount', total_purchase_paid_amount,
            'total_payments_recorded', total_payments_recorded,
            'payment_discrepancy', bank_balance_discrepancy,
            'payments_consistent', ABS(bank_balance_discrepancy) <= 1.00
        ),
        'recommendations', 
        CASE 
            WHEN validation_status = 'PASSED' THEN JSON_ARRAY('No action needed - balances are correct')
            ELSE JSON_ARRAY(
                CASE WHEN ABS(balance_discrepancy) > 1.00 
                     THEN 'Fix Accounts Payable balance discrepancy' 
                     ELSE NULL END,
                CASE WHEN ABS(bank_balance_discrepancy) > 1.00 
                     THEN 'Reconcile purchase payment records' 
                     ELSE NULL END
            )
        END
    );
    
    RETURN validation_result;
END$$

-- ========================================
-- PURCHASE BALANCE SYNC FUNCTION
-- ========================================
-- Function to automatically fix purchase balance discrepancies

DROP FUNCTION IF EXISTS sync_purchase_balances$$

CREATE FUNCTION sync_purchase_balances() RETURNS JSON
READS SQL DATA
MODIFIES SQL DATA
DETERMINISTIC
BEGIN
    DECLARE sync_result JSON DEFAULT JSON_OBJECT();
    DECLARE total_outstanding DECIMAL(15,2) DEFAULT 0;
    DECLARE expected_ap_balance DECIMAL(15,2) DEFAULT 0;
    DECLARE accounts_payable_account_id INT DEFAULT NULL;
    DECLARE old_ap_balance DECIMAL(15,2) DEFAULT 0;
    DECLARE balance_updated BOOLEAN DEFAULT FALSE;
    DECLARE updates_made INT DEFAULT 0;
    
    -- Get total outstanding from credit purchases
    SELECT COALESCE(SUM(outstanding_amount), 0) 
    INTO total_outstanding
    FROM purchases 
    WHERE payment_method = 'CREDIT' 
      AND deleted_at IS NULL;
    
    -- Find Accounts Payable account
    SELECT id, balance 
    INTO accounts_payable_account_id, old_ap_balance
    FROM accounts 
    WHERE (code LIKE '%2101%' OR name LIKE '%Hutang Usaha%' OR name LIKE '%Accounts Payable%')
      AND deleted_at IS NULL
    ORDER BY 
        CASE 
            WHEN code = '2101' THEN 1 
            WHEN code LIKE '2101%' THEN 2 
            WHEN name LIKE '%Hutang Usaha%' THEN 3
            ELSE 4 
        END
    LIMIT 1;
    
    -- Calculate expected balance (negative for liability)
    SET expected_ap_balance = -total_outstanding;
    
    -- Update Accounts Payable balance if discrepancy exists
    IF accounts_payable_account_id IS NOT NULL AND ABS(old_ap_balance - expected_ap_balance) > 1.00 THEN
        UPDATE accounts 
        SET balance = expected_ap_balance,
            updated_at = NOW()
        WHERE id = accounts_payable_account_id;
        
        SET balance_updated = TRUE;
        SET updates_made = updates_made + 1;
    END IF;
    
    -- Fix individual purchase outstanding amounts that might be inconsistent
    UPDATE purchases 
    SET outstanding_amount = total_amount - paid_amount,
        updated_at = NOW()
    WHERE ABS(outstanding_amount - (total_amount - paid_amount)) > 0.01
      AND deleted_at IS NULL;
      
    SET updates_made = updates_made + ROW_COUNT();
    
    -- Build sync result
    SET sync_result = JSON_OBJECT(
        'sync_timestamp', NOW(),
        'updates_made', updates_made,
        'accounts_payable_updated', balance_updated,
        'accounts_payable', JSON_OBJECT(
            'account_id', accounts_payable_account_id,
            'old_balance', old_ap_balance,
            'new_balance', expected_ap_balance,
            'total_outstanding', total_outstanding
        ),
        'status', CASE WHEN updates_made > 0 THEN 'UPDATED' ELSE 'NO_CHANGES_NEEDED' END
    );
    
    RETURN sync_result;
END$$

-- ========================================
-- PURCHASE BALANCE MONITORING FUNCTION
-- ========================================
-- Function to provide detailed balance monitoring

DROP FUNCTION IF EXISTS get_purchase_balance_status$$

CREATE FUNCTION get_purchase_balance_status() RETURNS JSON
READS SQL DATA
DETERMINISTIC
BEGIN
    DECLARE status_result JSON DEFAULT JSON_OBJECT();
    DECLARE total_purchases INT DEFAULT 0;
    DECLARE credit_purchases INT DEFAULT 0;
    DECLARE total_outstanding DECIMAL(15,2) DEFAULT 0;
    DECLARE total_paid DECIMAL(15,2) DEFAULT 0;
    DECLARE ap_balance DECIMAL(15,2) DEFAULT 0;
    DECLARE total_bank_balance DECIMAL(15,2) DEFAULT 0;
    
    -- Get purchase counts and amounts
    SELECT 
        COUNT(*),
        COALESCE(SUM(CASE WHEN payment_method = 'CREDIT' THEN 1 ELSE 0 END), 0),
        COALESCE(SUM(outstanding_amount), 0),
        COALESCE(SUM(paid_amount), 0)
    INTO total_purchases, credit_purchases, total_outstanding, total_paid
    FROM purchases 
    WHERE deleted_at IS NULL;
    
    -- Get Accounts Payable balance
    SELECT COALESCE(balance, 0) 
    INTO ap_balance
    FROM accounts 
    WHERE (code LIKE '%2101%' OR name LIKE '%Hutang Usaha%' OR name LIKE '%Accounts Payable%')
      AND deleted_at IS NULL
    ORDER BY 
        CASE 
            WHEN code = '2101' THEN 1 
            WHEN code LIKE '2101%' THEN 2 
            WHEN name LIKE '%Hutang Usaha%' THEN 3
            ELSE 4 
        END
    LIMIT 1;
    
    -- Get total bank balance (cash accounts)
    SELECT COALESCE(SUM(balance), 0) 
    INTO total_bank_balance
    FROM accounts 
    WHERE code LIKE '110%' -- Cash and bank accounts
      AND deleted_at IS NULL;
    
    -- Build status result
    SET status_result = JSON_OBJECT(
        'timestamp', NOW(),
        'purchase_summary', JSON_OBJECT(
            'total_purchases', total_purchases,
            'credit_purchases', credit_purchases,
            'total_outstanding', total_outstanding,
            'total_paid', total_paid
        ),
        'account_balances', JSON_OBJECT(
            'accounts_payable_balance', ap_balance,
            'expected_ap_balance', -total_outstanding,
            'ap_balance_correct', ABS(ap_balance - (-total_outstanding)) <= 1.00,
            'total_bank_balance', total_bank_balance
        ),
        'health_status', 
        CASE 
            WHEN ABS(ap_balance - (-total_outstanding)) <= 1.00 THEN 'HEALTHY'
            ELSE 'NEEDS_ATTENTION'
        END
    );
    
    RETURN status_result;
END$$

DELIMITER ;

-- ========================================
-- CREATE MONITORING VIEW
-- ========================================

CREATE OR REPLACE VIEW purchase_balance_monitoring AS
SELECT 
    'Purchase Balance Status' as check_name,
    JSON_UNQUOTE(JSON_EXTRACT(get_purchase_balance_status(), '$.health_status')) as status,
    JSON_UNQUOTE(JSON_EXTRACT(get_purchase_balance_status(), '$.purchase_summary.total_outstanding')) as total_outstanding,
    JSON_UNQUOTE(JSON_EXTRACT(get_purchase_balance_status(), '$.account_balances.accounts_payable_balance')) as ap_balance,
    JSON_UNQUOTE(JSON_EXTRACT(get_purchase_balance_status(), '$.account_balances.expected_ap_balance')) as expected_ap_balance,
    NOW() as last_checked;

-- Test the functions
SELECT 'Testing purchase balance validation...' as message;
SELECT validate_purchase_balances() as validation_result;

SELECT 'Testing purchase balance status...' as message;
SELECT get_purchase_balance_status() as status_result;

SELECT '✅ Purchase balance validation functions created successfully!' as result;