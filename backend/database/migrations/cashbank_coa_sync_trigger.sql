-- =======================================================================================
-- CashBank-COA Synchronization Database Triggers
-- =======================================================================================
-- This SQL creates database triggers to automatically sync CashBank balance changes 
-- with their linked COA accounts as a safety net for data consistency.

-- Drop existing triggers and functions if they exist
DROP TRIGGER IF EXISTS trg_sync_cashbank_coa ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trg_sync_cashbank_balance_update ON cash_banks;
DROP FUNCTION IF EXISTS sync_cashbank_balance_to_coa();
DROP FUNCTION IF EXISTS update_cashbank_balance_from_transactions();

-- =======================================================================================
-- Function: sync_cashbank_balance_to_coa
-- Purpose: Automatically sync CashBank balance changes to linked COA account
-- Triggered by: INSERT, UPDATE, DELETE on cash_bank_transactions table
-- =======================================================================================
CREATE OR REPLACE FUNCTION sync_cashbank_balance_to_coa()
RETURNS TRIGGER AS $$
DECLARE
    coa_account_id INTEGER;
    transaction_sum DECIMAL(15,2);
    cash_bank_name VARCHAR(100);
    account_code VARCHAR(20);
    old_cash_bank_id INTEGER;
    new_cash_bank_id INTEGER;
BEGIN
    -- Determine which cash_bank_id to process based on operation
    IF TG_OP = 'DELETE' THEN
        old_cash_bank_id := OLD.cash_bank_id;
        new_cash_bank_id := OLD.cash_bank_id;
    ELSIF TG_OP = 'UPDATE' THEN
        old_cash_bank_id := OLD.cash_bank_id;
        new_cash_bank_id := NEW.cash_bank_id;
    ELSE -- INSERT
        old_cash_bank_id := NEW.cash_bank_id;
        new_cash_bank_id := NEW.cash_bank_id;
    END IF;
    
    -- Process old cash bank (for UPDATE and DELETE operations)
    IF old_cash_bank_id IS NOT NULL AND (TG_OP = 'DELETE' OR (TG_OP = 'UPDATE' AND old_cash_bank_id != new_cash_bank_id)) THEN
        -- Get the linked COA account ID and cash bank name
        SELECT cb.account_id, cb.name INTO coa_account_id, cash_bank_name
        FROM cash_banks cb 
        WHERE cb.id = old_cash_bank_id AND cb.deleted_at IS NULL;
        
        -- Skip if no linked COA account
        IF coa_account_id IS NOT NULL AND coa_account_id > 0 THEN
            -- Calculate total transaction sum for old cash bank
            SELECT COALESCE(SUM(amount), 0) INTO transaction_sum
            FROM cash_bank_transactions
            WHERE cash_bank_id = old_cash_bank_id AND deleted_at IS NULL;
            
            -- Update old CashBank balance
            UPDATE cash_banks 
            SET balance = transaction_sum, updated_at = NOW()
            WHERE id = old_cash_bank_id;
            
            -- Update linked COA account balance
            UPDATE accounts 
            SET balance = transaction_sum, updated_at = NOW()
            WHERE id = coa_account_id;
            
            -- Get account code for logging
            SELECT code INTO account_code FROM accounts WHERE id = coa_account_id;
            
            -- Log the sync action for old cash bank
            INSERT INTO audit_logs (
                table_name, 
                action, 
                record_id, 
                old_values, 
                new_values,
                created_at,
                notes
            ) VALUES (
                'cashbank_coa_sync',
                'AUTO_SYNC_OLD',
                coa_account_id,
                '{}',
                json_build_object(
                    'balance', transaction_sum, 
                    'cash_bank_id', old_cash_bank_id,
                    'cash_bank_name', cash_bank_name,
                    'account_code', account_code,
                    'operation', TG_OP
                ),
                NOW(),
                'Automatic sync from cash_bank_transactions trigger (old cash bank)'
            );
        END IF;
    END IF;
    
    -- Process new cash bank (for INSERT and UPDATE operations)
    IF new_cash_bank_id IS NOT NULL AND (TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND old_cash_bank_id != new_cash_bank_id)) THEN
        -- Get the linked COA account ID and cash bank name
        SELECT cb.account_id, cb.name INTO coa_account_id, cash_bank_name
        FROM cash_banks cb 
        WHERE cb.id = new_cash_bank_id AND cb.deleted_at IS NULL;
        
        -- Skip if no linked COA account
        IF coa_account_id IS NOT NULL AND coa_account_id > 0 THEN
            -- Calculate total transaction sum for new cash bank
            SELECT COALESCE(SUM(amount), 0) INTO transaction_sum
            FROM cash_bank_transactions
            WHERE cash_bank_id = new_cash_bank_id AND deleted_at IS NULL;
            
            -- Update new CashBank balance
            UPDATE cash_banks 
            SET balance = transaction_sum, updated_at = NOW()
            WHERE id = new_cash_bank_id;
            
            -- Update linked COA account balance
            UPDATE accounts 
            SET balance = transaction_sum, updated_at = NOW()
            WHERE id = coa_account_id;
            
            -- Get account code for logging
            SELECT code INTO account_code FROM accounts WHERE id = coa_account_id;
            
            -- Log the sync action for new cash bank
            INSERT INTO audit_logs (
                table_name, 
                action, 
                record_id, 
                old_values, 
                new_values,
                created_at,
                notes
            ) VALUES (
                'cashbank_coa_sync',
                'AUTO_SYNC_NEW',
                coa_account_id,
                '{}',
                json_build_object(
                    'balance', transaction_sum, 
                    'cash_bank_id', new_cash_bank_id,
                    'cash_bank_name', cash_bank_name,
                    'account_code', account_code,
                    'operation', TG_OP
                ),
                NOW(),
                'Automatic sync from cash_bank_transactions trigger (new cash bank)'
            );
        END IF;
    END IF;
    
    -- Handle same cash bank UPDATE (amount changed, same cash bank)
    IF TG_OP = 'UPDATE' AND old_cash_bank_id = new_cash_bank_id THEN
        -- Get the linked COA account ID and cash bank name
        SELECT cb.account_id, cb.name INTO coa_account_id, cash_bank_name
        FROM cash_banks cb 
        WHERE cb.id = new_cash_bank_id AND cb.deleted_at IS NULL;
        
        -- Skip if no linked COA account
        IF coa_account_id IS NOT NULL AND coa_account_id > 0 THEN
            -- Calculate total transaction sum
            SELECT COALESCE(SUM(amount), 0) INTO transaction_sum
            FROM cash_bank_transactions
            WHERE cash_bank_id = new_cash_bank_id AND deleted_at IS NULL;
            
            -- Update CashBank balance
            UPDATE cash_banks 
            SET balance = transaction_sum, updated_at = NOW()
            WHERE id = new_cash_bank_id;
            
            -- Update linked COA account balance
            UPDATE accounts 
            SET balance = transaction_sum, updated_at = NOW()
            WHERE id = coa_account_id;
            
            -- Get account code for logging
            SELECT code INTO account_code FROM accounts WHERE id = coa_account_id;
            
            -- Log the sync action
            INSERT INTO audit_logs (
                table_name, 
                action, 
                record_id, 
                old_values, 
                new_values,
                created_at,
                notes
            ) VALUES (
                'cashbank_coa_sync',
                'AUTO_SYNC_UPDATE',
                coa_account_id,
                json_build_object('old_amount', OLD.amount),
                json_build_object(
                    'balance', transaction_sum, 
                    'cash_bank_id', new_cash_bank_id,
                    'cash_bank_name', cash_bank_name,
                    'account_code', account_code,
                    'new_amount', NEW.amount,
                    'operation', TG_OP
                ),
                NOW(),
                'Automatic sync from cash_bank_transactions trigger (update same cash bank)'
            );
        END IF;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- =======================================================================================
-- Function: update_cashbank_balance_from_transactions
-- Purpose: Sync COA balance when CashBank account_id is changed
-- Triggered by: UPDATE on cash_banks table when account_id changes
-- =======================================================================================
CREATE OR REPLACE FUNCTION update_cashbank_balance_from_transactions()
RETURNS TRIGGER AS $$
DECLARE
    transaction_sum DECIMAL(15,2);
    old_account_code VARCHAR(20);
    new_account_code VARCHAR(20);
BEGIN
    -- Only process if account_id actually changed
    IF OLD.account_id IS DISTINCT FROM NEW.account_id THEN
        
        -- Handle old account (reset balance to 0 if it was linked)
        IF OLD.account_id IS NOT NULL AND OLD.account_id > 0 THEN
            UPDATE accounts 
            SET balance = 0, updated_at = NOW()
            WHERE id = OLD.account_id;
            
            SELECT code INTO old_account_code FROM accounts WHERE id = OLD.account_id;
            
            -- Log old account reset
            INSERT INTO audit_logs (
                table_name, 
                action, 
                record_id, 
                old_values, 
                new_values,
                created_at,
                notes
            ) VALUES (
                'cashbank_coa_sync',
                'UNLINK_ACCOUNT',
                OLD.account_id,
                json_build_object('account_id', OLD.account_id, 'account_code', old_account_code),
                json_build_object('balance', 0),
                NOW(),
                'Reset old linked account balance when CashBank account_id changed'
            );
        END IF;
        
        -- Handle new account (sync balance if newly linked)
        IF NEW.account_id IS NOT NULL AND NEW.account_id > 0 THEN
            -- Calculate current transaction sum for this cash bank
            SELECT COALESCE(SUM(amount), 0) INTO transaction_sum
            FROM cash_bank_transactions
            WHERE cash_bank_id = NEW.id AND deleted_at IS NULL;
            
            -- Update CashBank balance based on transactions
            NEW.balance := transaction_sum;
            
            -- Update new linked COA account balance
            UPDATE accounts 
            SET balance = transaction_sum, updated_at = NOW()
            WHERE id = NEW.account_id;
            
            SELECT code INTO new_account_code FROM accounts WHERE id = NEW.account_id;
            
            -- Log new account sync
            INSERT INTO audit_logs (
                table_name, 
                action, 
                record_id, 
                old_values, 
                new_values,
                created_at,
                notes
            ) VALUES (
                'cashbank_coa_sync',
                'LINK_ACCOUNT',
                NEW.account_id,
                json_build_object('account_id', OLD.account_id),
                json_build_object(
                    'account_id', NEW.account_id, 
                    'account_code', new_account_code,
                    'balance', transaction_sum,
                    'cash_bank_name', NEW.name
                ),
                NOW(),
                'Sync new linked account balance when CashBank account_id changed'
            );
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =======================================================================================
-- Create Triggers
-- =======================================================================================

-- Trigger on cash_bank_transactions for automatic balance synchronization
CREATE TRIGGER trg_sync_cashbank_coa
    AFTER INSERT OR UPDATE OR DELETE ON cash_bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION sync_cashbank_balance_to_coa();

-- Trigger on cash_banks for handling account_id changes
CREATE TRIGGER trg_sync_cashbank_balance_update
    BEFORE UPDATE ON cash_banks
    FOR EACH ROW
    EXECUTE FUNCTION update_cashbank_balance_from_transactions();

-- =======================================================================================
-- Create Audit Log Table if it doesn't exist
-- =======================================================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    record_id INTEGER,
    old_values JSONB,
    new_values JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    notes TEXT
);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_table_action ON audit_logs (table_name, action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at);

-- =======================================================================================
-- Add Comments for Documentation
-- =======================================================================================
COMMENT ON FUNCTION sync_cashbank_balance_to_coa() IS 'Automatically sync CashBank balance changes to linked COA accounts when transactions are modified';
COMMENT ON FUNCTION update_cashbank_balance_from_transactions() IS 'Sync COA balance when CashBank account_id is changed, handles linking/unlinking';
COMMENT ON TRIGGER trg_sync_cashbank_coa ON cash_bank_transactions IS 'Trigger to maintain CashBank-COA balance synchronization';
COMMENT ON TRIGGER trg_sync_cashbank_balance_update ON cash_banks IS 'Trigger to handle CashBank-COA linking changes';

-- =======================================================================================
-- Test Data Integrity After Trigger Installation
-- =======================================================================================
-- This section will help verify that triggers work correctly

-- Function to validate all CashBank-COA sync integrity
CREATE OR REPLACE FUNCTION validate_cashbank_coa_integrity()
RETURNS TABLE (
    cash_bank_id INTEGER,
    cash_bank_name VARCHAR(100),
    coa_account_id INTEGER,
    coa_account_code VARCHAR(20),
    cash_bank_balance DECIMAL(15,2),
    coa_balance DECIMAL(15,2),
    transaction_sum DECIMAL(15,2),
    is_synced BOOLEAN,
    discrepancy DECIMAL(15,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cb.id::INTEGER as cash_bank_id,
        cb.name as cash_bank_name,
        a.id::INTEGER as coa_account_id,
        a.code as coa_account_code,
        cb.balance as cash_bank_balance,
        a.balance as coa_balance,
        COALESCE(tx_sum.transaction_sum, 0) as transaction_sum,
        (cb.balance = a.balance AND cb.balance = COALESCE(tx_sum.transaction_sum, 0)) as is_synced,
        (cb.balance - a.balance) as discrepancy
    FROM cash_banks cb
    JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
    LEFT JOIN (
        SELECT 
            cash_bank_id,
            SUM(amount) as transaction_sum
        FROM cash_bank_transactions 
        WHERE deleted_at IS NULL 
        GROUP BY cash_bank_id
    ) tx_sum ON cb.id = tx_sum.cash_bank_id
    WHERE cb.deleted_at IS NULL 
      AND cb.account_id IS NOT NULL 
      AND cb.account_id > 0
    ORDER BY cb.name;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_cashbank_coa_integrity() IS 'Validates CashBank-COA balance synchronization integrity across all linked accounts';

-- =======================================================================================
-- Usage Examples and Testing
-- =======================================================================================
/*
-- To test the triggers, you can:

-- 1. Check current integrity:
SELECT * FROM validate_cashbank_coa_integrity() WHERE NOT is_synced;

-- 2. Insert a test transaction and verify auto-sync:
INSERT INTO cash_bank_transactions (cash_bank_id, amount, transaction_date, notes) 
VALUES (1, 100000, NOW(), 'Test transaction for trigger');

-- 3. Check audit logs:
SELECT * FROM audit_logs WHERE table_name = 'cashbank_coa_sync' ORDER BY created_at DESC LIMIT 10;

-- 4. Update transaction amount and verify sync:
UPDATE cash_bank_transactions SET amount = 150000 WHERE id = (SELECT MAX(id) FROM cash_bank_transactions);

-- 5. Delete transaction and verify sync:
DELETE FROM cash_bank_transactions WHERE id = (SELECT MAX(id) FROM cash_bank_transactions);

-- 6. Change CashBank account_id linking:
UPDATE cash_banks SET account_id = 2 WHERE id = 1;

*/
