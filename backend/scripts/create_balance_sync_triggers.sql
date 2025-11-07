-- ===========================================
-- COA Balance Synchronization Triggers
-- ===========================================
-- Purpose: Automatically sync accounts.balance when cash_banks.balance changes
-- This prevents future balance inconsistency issues

-- Function to sync account balance when cash_bank balance changes
CREATE OR REPLACE FUNCTION sync_account_balance_on_cash_bank_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Only update if balance actually changed
    IF OLD.balance IS DISTINCT FROM NEW.balance THEN
        -- Update the corresponding account balance
        UPDATE accounts 
        SET 
            balance = NEW.balance,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.account_id;
        
        -- Log the sync operation for audit
        INSERT INTO balance_sync_log (
            cash_bank_id,
            account_id,
            old_balance,
            new_balance,
            sync_timestamp,
            trigger_source
        ) VALUES (
            NEW.id,
            NEW.account_id,
            OLD.balance,
            NEW.balance,
            CURRENT_TIMESTAMP,
            'cash_bank_trigger'
        );
        
        -- Raise notice for debugging (optional)
        RAISE NOTICE 'Balance synced: Cash Bank ID %, Account ID %, Balance: % -> %', 
            NEW.id, NEW.account_id, OLD.balance, NEW.balance;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on cash_banks table
DROP TRIGGER IF EXISTS trigger_sync_account_balance_on_cash_bank_update ON cash_banks;

CREATE TRIGGER trigger_sync_account_balance_on_cash_bank_update
    AFTER UPDATE OF balance ON cash_banks
    FOR EACH ROW
    EXECUTE FUNCTION sync_account_balance_on_cash_bank_update();

-- ===========================================
-- Reverse Sync: Account -> Cash Bank
-- ===========================================
-- In case accounts table is updated directly, sync back to cash_banks

CREATE OR REPLACE FUNCTION sync_cash_bank_balance_on_account_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Only update if balance actually changed
    IF OLD.balance IS DISTINCT FROM NEW.balance THEN
        -- Update the corresponding cash_bank balance
        UPDATE cash_banks 
        SET 
            balance = NEW.balance,
            updated_at = CURRENT_TIMESTAMP
        WHERE account_id = NEW.id;
        
        -- Log the reverse sync operation
        INSERT INTO balance_sync_log (
            cash_bank_id,
            account_id,
            old_balance,
            new_balance,
            sync_timestamp,
            trigger_source
        ) VALUES (
            (SELECT id FROM cash_banks WHERE account_id = NEW.id),
            NEW.id,
            OLD.balance,
            NEW.balance,
            CURRENT_TIMESTAMP,
            'account_trigger'
        );
        
        RAISE NOTICE 'Reverse balance synced: Account ID %, Balance: % -> %', 
            NEW.id, OLD.balance, NEW.balance;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create reverse trigger on accounts table
DROP TRIGGER IF EXISTS trigger_sync_cash_bank_balance_on_account_update ON accounts;

CREATE TRIGGER trigger_sync_cash_bank_balance_on_account_update
    AFTER UPDATE OF balance ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION sync_cash_bank_balance_on_account_update();

-- ===========================================
-- Balance Sync Audit Table
-- ===========================================
-- Create table to log all balance sync operations

CREATE TABLE IF NOT EXISTS balance_sync_log (
    id SERIAL PRIMARY KEY,
    cash_bank_id INTEGER,
    account_id INTEGER,
    old_balance DECIMAL(15,2),
    new_balance DECIMAL(15,2),
    sync_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    trigger_source VARCHAR(50), -- 'cash_bank_trigger', 'account_trigger', 'manual_sync'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_balance_sync_log_cash_bank_id ON balance_sync_log(cash_bank_id);
CREATE INDEX IF NOT EXISTS idx_balance_sync_log_account_id ON balance_sync_log(account_id);
CREATE INDEX IF NOT EXISTS idx_balance_sync_log_timestamp ON balance_sync_log(sync_timestamp);

-- ===========================================
-- Balance Validation Function
-- ===========================================
-- Function to check for any balance inconsistencies

CREATE OR REPLACE FUNCTION validate_balance_consistency()
RETURNS TABLE(
    cash_bank_id INTEGER,
    account_id INTEGER,
    cash_bank_name TEXT,
    account_name TEXT,
    cash_bank_balance DECIMAL(15,2),
    account_balance DECIMAL(15,2),
    difference DECIMAL(15,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cb.id as cash_bank_id,
        cb.account_id,
        cb.account_name as cash_bank_name,
        a.name as account_name,
        cb.balance as cash_bank_balance,
        a.balance as account_balance,
        (cb.balance - a.balance) as difference
    FROM cash_banks cb
    JOIN accounts a ON cb.account_id = a.id
    WHERE ABS(cb.balance - a.balance) > 0.01; -- Allow small floating point differences
END;
$$ LANGUAGE plpgsql;

-- ===========================================
-- Manual Sync Function
-- ===========================================
-- Function to manually sync all balances (emergency use)

CREATE OR REPLACE FUNCTION manual_sync_all_balances()
RETURNS TABLE(
    synced_count INTEGER,
    error_count INTEGER,
    sync_details TEXT
) AS $$
DECLARE
    sync_count INTEGER := 0;
    error_count INTEGER := 0;
    rec RECORD;
BEGIN
    -- Sync cash_banks -> accounts
    FOR rec IN 
        SELECT cb.id, cb.account_id, cb.balance, a.balance as account_balance
        FROM cash_banks cb
        JOIN accounts a ON cb.account_id = a.id
        WHERE ABS(cb.balance - a.balance) > 0.01
    LOOP
        BEGIN
            UPDATE accounts 
            SET balance = rec.balance, updated_at = CURRENT_TIMESTAMP
            WHERE id = rec.account_id;
            
            -- Log manual sync
            INSERT INTO balance_sync_log (
                cash_bank_id, account_id, old_balance, new_balance, 
                sync_timestamp, trigger_source
            ) VALUES (
                rec.id, rec.account_id, rec.account_balance, rec.balance,
                CURRENT_TIMESTAMP, 'manual_sync'
            );
            
            sync_count := sync_count + 1;
        EXCEPTION WHEN OTHERS THEN
            error_count := error_count + 1;
            RAISE NOTICE 'Error syncing Cash Bank ID %: %', rec.id, SQLERRM;
        END;
    END LOOP;
    
    RETURN QUERY SELECT sync_count, error_count, 
        format('Synced %s accounts, %s errors', sync_count, error_count)::TEXT;
END;
$$ LANGUAGE plpgsql;

-- ===========================================
-- Usage Examples and Testing
-- ===========================================

-- Check current balance inconsistencies
-- SELECT * FROM validate_balance_consistency();

-- Manual sync all balances (if needed)
-- SELECT * FROM manual_sync_all_balances();

-- View recent sync operations
-- SELECT * FROM balance_sync_log ORDER BY sync_timestamp DESC LIMIT 10;

-- Test trigger (update a cash_bank balance and see if account balance follows)
-- UPDATE cash_banks SET balance = balance + 100 WHERE id = 1;

-- ===========================================
-- Cleanup Commands (if needed to remove)
-- ===========================================

-- DROP TRIGGER IF EXISTS trigger_sync_account_balance_on_cash_bank_update ON cash_banks;
-- DROP TRIGGER IF EXISTS trigger_sync_cash_bank_balance_on_account_update ON accounts;
-- DROP FUNCTION IF EXISTS sync_account_balance_on_cash_bank_update();
-- DROP FUNCTION IF EXISTS sync_cash_bank_balance_on_account_update();
-- DROP FUNCTION IF EXISTS validate_balance_consistency();
-- DROP FUNCTION IF EXISTS manual_sync_all_balances();
-- DROP TABLE IF EXISTS balance_sync_log;

COMMIT;

-- Success message
SELECT 'Balance sync triggers and functions created successfully!' as status;