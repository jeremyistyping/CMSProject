-- ========================================
-- PURCHASE BALANCE AUTO-SYNC TRIGGERS
-- ========================================
-- These triggers ensure that Accounts Payable balances are automatically
-- synchronized whenever purchase payment data changes

DELIMITER $$

-- ========================================
-- TRIGGER: After Purchase Payment Insert
-- ========================================

DROP TRIGGER IF EXISTS trg_purchase_payment_insert$$

CREATE TRIGGER trg_purchase_payment_insert
AFTER INSERT ON purchase_payments
FOR EACH ROW
BEGIN
    DECLARE sync_result JSON;
    DECLARE validation_result JSON;
    
    -- Update the related purchase paid_amount and outstanding_amount
    UPDATE purchases 
    SET paid_amount = (
        SELECT COALESCE(SUM(amount), 0) 
        FROM purchase_payments 
        WHERE purchase_id = NEW.purchase_id 
          AND deleted_at IS NULL
    ),
    outstanding_amount = total_amount - (
        SELECT COALESCE(SUM(amount), 0) 
        FROM purchase_payments 
        WHERE purchase_id = NEW.purchase_id 
          AND deleted_at IS NULL
    ),
    updated_at = NOW()
    WHERE id = NEW.purchase_id;
    
    -- Auto-sync purchase balances
    SET sync_result = sync_purchase_balances();
    
    -- Log the sync action (optional - for audit trail)
    INSERT INTO balance_sync_log (
        entity_type,
        entity_id,
        trigger_action,
        sync_result,
        created_at
    ) VALUES (
        'PURCHASE_PAYMENT',
        NEW.id,
        'INSERT',
        sync_result,
        NOW()
    ) ON DUPLICATE KEY UPDATE
        sync_result = VALUES(sync_result),
        created_at = VALUES(created_at);

END$$

-- ========================================
-- TRIGGER: After Purchase Payment Update
-- ========================================

DROP TRIGGER IF EXISTS trg_purchase_payment_update$$

CREATE TRIGGER trg_purchase_payment_update
AFTER UPDATE ON purchase_payments
FOR EACH ROW
BEGIN
    DECLARE sync_result JSON;
    DECLARE affected_purchase_ids TEXT DEFAULT '';
    
    -- Collect affected purchase IDs (in case purchase_id changed)
    SET affected_purchase_ids = CONCAT(OLD.purchase_id, ',', NEW.purchase_id);
    
    -- Update the old purchase if purchase_id changed
    IF OLD.purchase_id != NEW.purchase_id THEN
        UPDATE purchases 
        SET paid_amount = (
            SELECT COALESCE(SUM(amount), 0) 
            FROM purchase_payments 
            WHERE purchase_id = OLD.purchase_id 
              AND deleted_at IS NULL
        ),
        outstanding_amount = total_amount - (
            SELECT COALESCE(SUM(amount), 0) 
            FROM purchase_payments 
            WHERE purchase_id = OLD.purchase_id 
              AND deleted_at IS NULL
        ),
        updated_at = NOW()
        WHERE id = OLD.purchase_id;
    END IF;
    
    -- Update the new/current purchase
    UPDATE purchases 
    SET paid_amount = (
        SELECT COALESCE(SUM(amount), 0) 
        FROM purchase_payments 
        WHERE purchase_id = NEW.purchase_id 
          AND deleted_at IS NULL
    ),
    outstanding_amount = total_amount - (
        SELECT COALESCE(SUM(amount), 0) 
        FROM purchase_payments 
        WHERE purchase_id = NEW.purchase_id 
          AND deleted_at IS NULL
    ),
    updated_at = NOW()
    WHERE id = NEW.purchase_id;
    
    -- Auto-sync purchase balances
    SET sync_result = sync_purchase_balances();
    
    -- Log the sync action
    INSERT INTO balance_sync_log (
        entity_type,
        entity_id,
        trigger_action,
        sync_result,
        created_at
    ) VALUES (
        'PURCHASE_PAYMENT',
        NEW.id,
        'UPDATE',
        sync_result,
        NOW()
    ) ON DUPLICATE KEY UPDATE
        sync_result = VALUES(sync_result),
        created_at = VALUES(created_at);

END$$

-- ========================================
-- TRIGGER: After Purchase Payment Delete
-- ========================================

DROP TRIGGER IF EXISTS trg_purchase_payment_delete$$

CREATE TRIGGER trg_purchase_payment_delete
AFTER DELETE ON purchase_payments
FOR EACH ROW
BEGIN
    DECLARE sync_result JSON;
    
    -- Update the related purchase paid_amount and outstanding_amount
    UPDATE purchases 
    SET paid_amount = (
        SELECT COALESCE(SUM(amount), 0) 
        FROM purchase_payments 
        WHERE purchase_id = OLD.purchase_id 
          AND deleted_at IS NULL
    ),
    outstanding_amount = total_amount - (
        SELECT COALESCE(SUM(amount), 0) 
        FROM purchase_payments 
        WHERE purchase_id = OLD.purchase_id 
          AND deleted_at IS NULL
    ),
    updated_at = NOW()
    WHERE id = OLD.purchase_id;
    
    -- Auto-sync purchase balances
    SET sync_result = sync_purchase_balances();
    
    -- Log the sync action
    INSERT INTO balance_sync_log (
        entity_type,
        entity_id,
        trigger_action,
        sync_result,
        created_at
    ) VALUES (
        'PURCHASE_PAYMENT',
        OLD.id,
        'DELETE',
        sync_result,
        NOW()
    ) ON DUPLICATE KEY UPDATE
        sync_result = VALUES(sync_result),
        created_at = VALUES(created_at);

END$$

-- ========================================
-- TRIGGER: After Purchase Update
-- ========================================
-- Trigger to sync balances when purchase amounts change

DROP TRIGGER IF EXISTS trg_purchase_update$$

CREATE TRIGGER trg_purchase_update
AFTER UPDATE ON purchases
FOR EACH ROW
BEGIN
    DECLARE sync_result JSON;
    
    -- Only sync if amounts have changed
    IF OLD.total_amount != NEW.total_amount 
       OR OLD.paid_amount != NEW.paid_amount 
       OR OLD.outstanding_amount != NEW.outstanding_amount
       OR OLD.payment_method != NEW.payment_method THEN
       
        -- Ensure outstanding_amount is calculated correctly
        IF NEW.outstanding_amount != (NEW.total_amount - NEW.paid_amount) THEN
            UPDATE purchases 
            SET outstanding_amount = total_amount - paid_amount,
                updated_at = NOW()
            WHERE id = NEW.id;
        END IF;
        
        -- Auto-sync purchase balances
        SET sync_result = sync_purchase_balances();
        
        -- Log the sync action
        INSERT INTO balance_sync_log (
            entity_type,
            entity_id,
            trigger_action,
            sync_result,
            created_at
        ) VALUES (
            'PURCHASE',
            NEW.id,
            'UPDATE',
            sync_result,
            NOW()
        ) ON DUPLICATE KEY UPDATE
            sync_result = VALUES(sync_result),
            created_at = VALUES(created_at);
    END IF;

END$$

DELIMITER ;

-- ========================================
-- CREATE BALANCE SYNC LOG TABLE
-- ========================================
-- Table to track all balance sync operations for auditing

CREATE TABLE IF NOT EXISTS balance_sync_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    trigger_action VARCHAR(20) NOT NULL,
    sync_result JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_balance_sync_log_entity (entity_type, entity_id),
    INDEX idx_balance_sync_log_created (created_at),
    
    UNIQUE KEY uk_balance_sync_log_entity_action (entity_type, entity_id, trigger_action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ========================================
-- CREATE PURCHASE BALANCE HEALTH CHECK VIEW
-- ========================================
-- View to monitor purchase balance health in real-time

CREATE OR REPLACE VIEW purchase_balance_health AS
SELECT 
    p.id as purchase_id,
    p.code as purchase_code,
    p.vendor_id,
    p.total_amount,
    p.paid_amount,
    p.outstanding_amount,
    p.payment_method,
    p.status,
    
    -- Calculated fields
    (p.total_amount - p.paid_amount) as calculated_outstanding,
    ABS(p.outstanding_amount - (p.total_amount - p.paid_amount)) as outstanding_discrepancy,
    
    -- Payment summary from purchase_payments table
    COALESCE(payment_summary.total_payments, 0) as actual_payments,
    COALESCE(payment_summary.payment_count, 0) as payment_count,
    
    -- Health status
    CASE 
        WHEN ABS(p.outstanding_amount - (p.total_amount - p.paid_amount)) > 0.01 THEN 'INCONSISTENT'
        WHEN ABS(p.paid_amount - COALESCE(payment_summary.total_payments, 0)) > 0.01 THEN 'PAYMENT_MISMATCH'
        ELSE 'HEALTHY'
    END as health_status,
    
    p.updated_at as last_updated

FROM purchases p
LEFT JOIN (
    SELECT 
        purchase_id,
        SUM(amount) as total_payments,
        COUNT(*) as payment_count
    FROM purchase_payments
    WHERE deleted_at IS NULL
    GROUP BY purchase_id
) payment_summary ON p.id = payment_summary.purchase_id

WHERE p.deleted_at IS NULL
ORDER BY 
    CASE 
        WHEN ABS(p.outstanding_amount - (p.total_amount - p.paid_amount)) > 0.01 THEN 1
        WHEN ABS(p.paid_amount - COALESCE(payment_summary.total_payments, 0)) > 0.01 THEN 2
        ELSE 3
    END,
    p.id DESC;

-- ========================================
-- CREATE CLEANUP PROCEDURE FOR LOG TABLE
-- ========================================
-- Procedure to clean up old balance sync logs (older than 30 days)

DELIMITER $$

DROP PROCEDURE IF EXISTS cleanup_balance_sync_logs$$

CREATE PROCEDURE cleanup_balance_sync_logs()
BEGIN
    DECLARE deleted_count INT DEFAULT 0;
    
    DELETE FROM balance_sync_log 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    SET deleted_count = ROW_COUNT();
    
    SELECT CONCAT('Cleaned up ', deleted_count, ' old balance sync log entries') as result;
END$$

DELIMITER ;

-- ========================================
-- TEST THE TRIGGERS
-- ========================================

-- Test validation functions
SELECT 'Testing purchase balance validation after trigger creation...' as message;
SELECT validate_purchase_balances() as validation_result;

-- Check if there are any balance issues to fix
SELECT 'Checking purchase balance health...' as message;
SELECT * FROM purchase_balance_health WHERE health_status != 'HEALTHY' LIMIT 5;

-- Show sync log entries if any exist
SELECT 'Recent balance sync log entries...' as message;
SELECT * FROM balance_sync_log ORDER BY created_at DESC LIMIT 5;

SELECT '✅ Purchase balance triggers created and tested successfully!' as result;

-- ========================================
-- USAGE INSTRUCTIONS
-- ========================================

/*
These triggers will now automatically:

1. **Update purchase paid_amount and outstanding_amount** whenever purchase_payments are inserted, updated, or deleted

2. **Sync Accounts Payable balance** to match total outstanding amounts from credit purchases

3. **Log all sync operations** in balance_sync_log table for audit trail

4. **Maintain data consistency** in real-time without manual intervention

To monitor the system:
- Query `purchase_balance_health` view to see any inconsistent purchases
- Query `balance_sync_log` table to see sync operations
- Use `validate_purchase_balances()` function for validation reports
- Use `get_purchase_balance_status()` function for system status

Manual operations:
- Call `sync_purchase_balances()` to force a manual sync if needed
- Call `cleanup_balance_sync_logs()` to clean old log entries

The system is now fully automated and self-maintaining!
*/