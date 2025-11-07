-- =======================================================================================
-- REVERSE SYNC TRIGGER: COA Account Balance Changes → CashBank Balance Update
-- =======================================================================================
-- This trigger ensures when COA account balance is manually changed,
-- the linked CashBank balance is also updated automatically

-- Function to sync CashBank balance when COA account balance changes
CREATE OR REPLACE FUNCTION sync_coa_to_cashbank()
RETURNS TRIGGER AS $$
DECLARE
    linked_cashbanks_cursor CURSOR FOR
        SELECT cb.id, cb.name, cb.balance as old_balance
        FROM cash_banks cb 
        WHERE cb.account_id = NEW.id 
        AND cb.deleted_at IS NULL 
        AND cb.is_active = true;
    cashbank_record RECORD;
    balance_difference DECIMAL(15,2);
BEGIN
    -- Only process if balance actually changed
    IF OLD.balance IS DISTINCT FROM NEW.balance THEN
        
        -- Log the COA balance change
        INSERT INTO audit_logs (
            table_name, 
            action, 
            record_id, 
            old_values, 
            new_values,
            created_at,
            notes
        ) VALUES (
            'coa_to_cashbank_sync',
            'COA_BALANCE_CHANGED',
            NEW.id,
            json_build_object('old_balance', OLD.balance),
            json_build_object('new_balance', NEW.balance, 'account_code', NEW.code, 'account_name', NEW.name),
            NOW(),
            'COA account balance changed manually - syncing to linked CashBanks'
        );
        
        balance_difference := NEW.balance - COALESCE(OLD.balance, 0);
        
        -- Update all linked CashBanks
        FOR cashbank_record IN linked_cashbanks_cursor LOOP
            
            -- Update CashBank balance to match COA balance
            UPDATE cash_banks 
            SET balance = NEW.balance, updated_at = NOW()
            WHERE id = cashbank_record.id;
            
            -- Log the sync action for each CashBank
            INSERT INTO audit_logs (
                table_name, 
                action, 
                record_id, 
                old_values, 
                new_values,
                created_at,
                notes
            ) VALUES (
                'coa_to_cashbank_sync',
                'CASHBANK_BALANCE_SYNCED',
                cashbank_record.id,
                json_build_object('old_balance', cashbank_record.old_balance),
                json_build_object(
                    'new_balance', NEW.balance,
                    'coa_account_id', NEW.id,
                    'coa_account_code', NEW.code,
                    'cashbank_name', cashbank_record.name
                ),
                NOW(),
                'CashBank balance updated from COA account change'
            );
            
        END LOOP;
        
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on accounts table for COA to CashBank sync
DROP TRIGGER IF EXISTS trg_sync_coa_to_cashbank ON accounts;
CREATE TRIGGER trg_sync_coa_to_cashbank
    AFTER UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION sync_coa_to_cashbank();

-- =======================================================================================
-- Add Comments for Documentation
-- =======================================================================================
COMMENT ON FUNCTION sync_coa_to_cashbank() IS 'Automatically sync CashBank balances when linked COA account balance changes';
COMMENT ON TRIGGER trg_sync_coa_to_cashbank ON accounts IS 'Trigger to maintain COA-to-CashBank balance synchronization';

-- =======================================================================================
-- Test the Reverse Sync
-- =======================================================================================
/*
-- To test the reverse sync:

-- 1. Check current state
SELECT 
    cb.id as cashbank_id,
    cb.name as cashbank_name, 
    cb.balance as cashbank_balance,
    a.id as account_id,
    a.code as account_code,
    a.balance as coa_balance
FROM cash_banks cb 
JOIN accounts a ON cb.account_id = a.id 
WHERE cb.deleted_at IS NULL AND cb.is_active = true;

-- 2. Update COA balance manually
UPDATE accounts SET balance = 999999 WHERE code = '1102';

-- 3. Check if CashBank balance updated automatically
SELECT 
    cb.id as cashbank_id,
    cb.name as cashbank_name, 
    cb.balance as cashbank_balance,
    a.id as account_id,
    a.code as account_code,
    a.balance as coa_balance
FROM cash_banks cb 
JOIN accounts a ON cb.account_id = a.id 
WHERE cb.deleted_at IS NULL AND cb.is_active = true;

-- 4. Check audit logs
SELECT * FROM audit_logs 
WHERE table_name = 'coa_to_cashbank_sync' 
ORDER BY created_at DESC 
LIMIT 10;

*/
